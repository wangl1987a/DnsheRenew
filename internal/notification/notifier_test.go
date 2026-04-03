package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dnsherene/internal/config"
	"dnsherene/internal/report"
)

func TestConsoleWritesWhenEnabled(t *testing.T) {
	var out bytes.Buffer
	info := report.Info{
		RenewedTotal: 1,
		Accounts: []report.AccountInfo{
			{
				Index:        1,
				Total:        1,
				APIKeyMasked: "cfsd***ijkl",
				Renewed:      1,
				Domains: []report.DomainInfo{
					{Domain: "api.example.com", ExpiresAt: "2026-08-01 00:00:00"},
				},
			},
		},
	}

	if err := (Console{enabled: true, outWriter: &out, errWriter: &out}).Notify(context.Background(), info); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}
	if !strings.Contains(out.String(), "renewed_total=1") {
		t.Fatalf("debug output missing summary: %q", out.String())
	}
	if !strings.Contains(out.String(), "\"domain\":\"api.example.com\"") {
		t.Fatalf("debug output missing domains: %q", out.String())
	}
	if !strings.Contains(out.String(), "\"expires_at\":\"2026-08-01 00:00:00\"") {
		t.Fatalf("debug output missing expires_at: %q", out.String())
	}
}

func TestWebhookUsesConfig(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := (Webhook{
		config: config.WebhookConfig{
			URL:   server.URL,
			Token: "token-123",
		},
		httpClient: server.Client(),
	}).Notify(context.Background(), report.Info{RenewedTotal: 1})
	if err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}
	if gotAuth != "Bearer token-123" {
		t.Fatalf("Authorization = %q, want Bearer token-123", gotAuth)
	}
}

func TestTelegramUsesConfigAndFormatsHTML(t *testing.T) {
	type request struct {
		ChatID          any    `json:"chat_id"`
		MessageThreadID *int   `json:"message_thread_id,omitempty"`
		Text            string `json:"text"`
		ParseMode       string `json:"parse_mode"`
	}

	var requests []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottoken-123/sendMessage" {
			t.Fatalf("path = %q, want /bottoken-123/sendMessage", r.URL.Path)
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		requests = append(requests, req)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer server.Close()

	threadID := 7
	info := report.Info{
		RenewedTotal: 1,
		Accounts: []report.AccountInfo{
			{
				Index:        1,
				Total:        1,
				APIKeyMasked: "cfsd***ijkl",
				Matched:      1,
				Renewed:      1,
				Domains: []report.DomainInfo{
					{Domain: "api.example.com", ExpiresAt: "2026-08-01 00:00:00"},
				},
				RenewedList: []report.RenewedDomain{
					{Domain: "api.example.com", NewExpiresAt: "2027-08-01 00:00:00", RemainingDays: 365},
				},
			},
		},
	}

	err := (Telegram{
		config: config.TelegramConfig{
			BotToken:        "token-123",
			ChatID:          -100123456,
			MessageThreadID: &threadID,
		},
		httpClient: server.Client(),
		apiBaseURL: server.URL,
	}).Notify(context.Background(), info)
	if err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	if requests[0].ParseMode != "HTML" {
		t.Fatalf("parse_mode = %q, want HTML", requests[0].ParseMode)
	}
	if threadID := requests[0].MessageThreadID; threadID == nil || *threadID != 7 {
		t.Fatalf("message_thread_id = %v, want 7", requests[0].MessageThreadID)
	}
	if !strings.Contains(requests[0].Text, "<b>DNSHE Renew Summary</b>") {
		t.Fatalf("summary formatting missing: %q", requests[0].Text)
	}
	if !strings.Contains(requests[0].Text, "<b>API [1/1]</b>") {
		t.Fatalf("account formatting missing: %q", requests[0].Text)
	}
	if !strings.Contains(requests[0].Text, "<blockquote>") {
		t.Fatalf("blockquote formatting missing: %q", requests[0].Text)
	}
	if !strings.Contains(requests[0].Text, "<code>api.example.com</code>") {
		t.Fatalf("domain formatting missing: %q", requests[0].Text)
	}
}

func TestTelegramPartialConfigReturnsError(t *testing.T) {
	err := (Telegram{
		config: config.TelegramConfig{
			BotToken: "token-123",
		},
	}).Notify(context.Background(), report.Info{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestNewManagerRejectsUnimplementedChannels(t *testing.T) {
	_, err := NewManager(config.NotificationConfig{
		Mail: &config.MailConfig{
			SenderAddress: "sender@example.com",
		},
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
