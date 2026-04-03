package notification

import (
	"testing"
	"time"

	"dnsherene/internal/report"
)

func TestRequestValidateRequiresGeneratedAt(t *testing.T) {
	req := Request{
		Report: report.Info{},
	}

	err := req.Validate()
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
}

func TestRequestAccountCount(t *testing.T) {
	req := Request{
		Report: report.Info{
			GeneratedAt: time.Now().UTC(),
			Accounts: []report.AccountInfo{
				{Index: 1, Total: 2},
				{Index: 2, Total: 2},
			},
		},
	}

	if got := req.AccountCount(); got != 2 {
		t.Fatalf("AccountCount = %d, want 2", got)
	}
}

func TestRequestHasFailures(t *testing.T) {
	req := Request{
		Report: report.Info{
			GeneratedAt: time.Now().UTC(),
			Accounts: []report.AccountInfo{
				{Index: 1, Total: 2, Renewed: 1},
				{Index: 2, Total: 2, Failed: 1, Error: "renew failed"},
			},
		},
	}

	if !req.HasFailures() {
		t.Fatalf("HasFailures = false, want true")
	}
}
