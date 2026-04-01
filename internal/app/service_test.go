package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dnsherene/pkg/dnshe"
)

func TestRunNoSubdomainsIsNotAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("action"); got == "renew" {
			t.Fatalf("unexpected renew request when subdomain list is empty")
		}

		if got := r.URL.Query().Get("endpoint"); got != "subdomains" {
			t.Fatalf("endpoint = %q, want subdomains", got)
		}
		if got := r.URL.Query().Get("action"); got != "list" {
			t.Fatalf("action = %q, want list", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":    true,
			"count":      0,
			"subdomains": []map[string]any{},
		})
	}))
	defer server.Close()

	client, err := dnshe.NewClient(dnshe.Config{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		APISecret:  "test-secret",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	service, err := NewService(client)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	result, err := service.Run(context.Background(), false)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Matched != 0 || result.Renewed != 0 || result.Failed != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}
