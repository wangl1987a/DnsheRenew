package dnshe

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient(Config{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		APISecret:  "test-secret",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return client
}

func TestListSubdomainsBuildsRequest(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if got := r.URL.Query().Get("m"); got != "domain_hub" {
			t.Fatalf("m = %q, want domain_hub", got)
		}
		if got := r.URL.Query().Get("endpoint"); got != "subdomains" {
			t.Fatalf("endpoint = %q, want subdomains", got)
		}
		if got := r.URL.Query().Get("action"); got != "list" {
			t.Fatalf("action = %q, want list", got)
		}
		if got := r.Header.Get("X-API-Key"); got != "test-key" {
			t.Fatalf("X-API-Key = %q, want test-key", got)
		}
		if got := r.Header.Get("X-API-Secret"); got != "test-secret" {
			t.Fatalf("X-API-Secret = %q, want test-secret", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"count":   1,
			"subdomains": []map[string]any{
				{
					"id":          1,
					"subdomain":   "api",
					"rootdomain":  "example.com",
					"full_domain": "api.example.com",
					"status":      "active",
				},
			},
		})
	})

	got, err := client.ListSubdomains(context.Background())
	if err != nil {
		t.Fatalf("ListSubdomains returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].FullDomain != "api.example.com" {
		t.Fatalf("FullDomain = %q, want api.example.com", got[0].FullDomain)
	}
}

func TestRenewSubdomainReturnsExtendedFields(t *testing.T) {
	type renewRequest struct {
		SubdomainID int `json:"subdomain_id"`
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.URL.Query().Get("endpoint"); got != "subdomains" {
			t.Fatalf("endpoint = %q, want subdomains", got)
		}
		if got := r.URL.Query().Get("action"); got != "renew" {
			t.Fatalf("action = %q, want renew", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}

		var payload renewRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body failed: %v", err)
		}
		if payload.SubdomainID != 3 {
			t.Fatalf("subdomain_id = %d, want 3", payload.SubdomainID)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":             true,
			"message":             "Subdomain renewed successfully",
			"subdomain_id":        3,
			"subdomain":           "myapp",
			"previous_expires_at": "2025-05-01 00:00:00",
			"new_expires_at":      "2026-05-01 00:00:00",
			"renewed_at":          "2025-04-10 12:34:56",
			"never_expires":       1,
			"status":              "active",
			"remaining_days":      366,
		})
	})

	got, err := client.RenewSubdomain(context.Background(), 3)
	if err != nil {
		t.Fatalf("RenewSubdomain returned error: %v", err)
	}
	if got.Message != "Subdomain renewed successfully" {
		t.Fatalf("Message = %q, want Subdomain renewed successfully", got.Message)
	}
	if got.RenewedAt != "2025-04-10 12:34:56" {
		t.Fatalf("RenewedAt = %q, want 2025-04-10 12:34:56", got.RenewedAt)
	}
	if !got.NeverExpires {
		t.Fatalf("NeverExpires = false, want true")
	}
	if got.Status != "active" {
		t.Fatalf("Status = %q, want active", got.Status)
	}
}

func TestListSubdomainsReturnsStructuredHTTPError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"Rate limit exceeded","limit":60,"remaining":0,"reset_at":"2025-10-19 15:31:00"}`))
	})

	_, err := client.ListSubdomains(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusTooManyRequests)
	}
	if apiErr.ErrorText != "Rate limit exceeded" {
		t.Fatalf("ErrorText = %q, want Rate limit exceeded", apiErr.ErrorText)
	}
	if apiErr.Limit == nil || *apiErr.Limit != 60 {
		t.Fatalf("Limit = %v, want 60", apiErr.Limit)
	}
	if apiErr.Remaining == nil || *apiErr.Remaining != 0 {
		t.Fatalf("Remaining = %v, want 0", apiErr.Remaining)
	}
	if apiErr.ResetAt != "2025-10-19 15:31:00" {
		t.Fatalf("ResetAt = %q, want 2025-10-19 15:31:00", apiErr.ResetAt)
	}
}

func TestListSubdomainsReturnsStructuredBusinessError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false,"error":"Invalid API key"}`))
	})

	_, err := client.ListSubdomains(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Operation != "list subdomains" {
		t.Fatalf("Operation = %q, want list subdomains", apiErr.Operation)
	}
	if apiErr.ErrorText != "Invalid API key" {
		t.Fatalf("ErrorText = %q, want Invalid API key", apiErr.ErrorText)
	}
}

func TestCreateDNSRecordNormalizesType(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body failed: %v", err)
		}
		if payload["type"] != "TXT" {
			t.Fatalf("type = %v, want TXT", payload["type"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"DNS record created successfully","record_id":5}`))
	})

	result, err := client.CreateDNSRecord(context.Background(), CreateDNSRecordRequest{
		SubdomainID: 1,
		Type:        "txt",
		Content:     "hello",
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord returned error: %v", err)
	}
	if result.RecordID != 5 {
		t.Fatalf("RecordID = %d, want 5", result.RecordID)
	}
}

func TestCreateDNSRecordRequiresMXPriority(t *testing.T) {
	client := &Client{}

	_, err := client.CreateDNSRecord(context.Background(), CreateDNSRecordRequest{
		SubdomainID: 1,
		Type:        "mx",
		Content:     "mail.example.com",
	})
	if err == nil {
		t.Fatalf("expected MX validation error, got nil")
	}
}

func TestCreateDNSRecordAllowsMXPriority(t *testing.T) {
	priority := 10

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body failed: %v", err)
		}
		if payload["type"] != "MX" {
			t.Fatalf("type = %v, want MX", payload["type"])
		}
		if payload["priority"] != float64(10) {
			t.Fatalf("priority = %v, want 10", payload["priority"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"DNS record created successfully","record_id":6}`))
	})

	result, err := client.CreateDNSRecord(context.Background(), CreateDNSRecordRequest{
		SubdomainID: 1,
		Type:        "mx",
		Content:     "mail.example.com",
		Priority:    &priority,
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord with MX priority returned error: %v", err)
	}
	if result.RecordID != 6 {
		t.Fatalf("RecordID = %d, want 6", result.RecordID)
	}
}
