package notifier

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dnsherene/internal/report"
)

func TestConsoleReadsDNSHEDebug(t *testing.T) {
	t.Setenv("DNSHE_DEBUG", "true")

	var out bytes.Buffer
	info := report.Info{
		RenewedTotal: 1,
		Accounts: []report.AccountInfo{
			{Index: 1, Total: 1, APIKeyMasked: "cfsd***ijkl", Renewed: 1},
		},
	}

	if err := (Console{outWriter: &out, errWriter: &out}).Notify(context.Background(), info); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}
	if !strings.Contains(out.String(), "renewed_total=1") {
		t.Fatalf("debug output missing summary: %q", out.String())
	}
}

func TestWebhookReadsDNSHEPrefixedEnv(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("DNSHE_NOTIFY_WEBHOOK_URL", server.URL)
	t.Setenv("DNSHE_NOTIFY_WEBHOOK_TOKEN", "token-123")

	err := (Webhook{httpClient: server.Client()}).Notify(context.Background(), report.Info{RenewedTotal: 1})
	if err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}
	if gotAuth != "Bearer token-123" {
		t.Fatalf("Authorization = %q, want Bearer token-123", gotAuth)
	}
}
