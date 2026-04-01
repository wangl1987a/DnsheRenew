package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestMaskAPIKey(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "short", in: "abc", want: "***"},
		{name: "medium", in: "abcdef", want: "ab***ef"},
		{name: "long", in: "abcdefghijkl", want: "abcd***ijkl"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := MaskAPIKey(tc.in); got != tc.want {
				t.Fatalf("MaskAPIKey(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSanitizePublicError(t *testing.T) {
	err := errors.New("run renew failed for api key cfsd_abcdefghijkl and domain blog.example.com via https://hooks.example.com/token")

	got := SanitizePublicError(err)

	if strings.Contains(got, "cfsd_abcdefghijkl") {
		t.Fatalf("raw api key leaked: %q", got)
	}
	if strings.Contains(got, "blog.example.com") {
		t.Fatalf("raw domain leaked: %q", got)
	}
	if strings.Contains(got, "https://hooks.example.com/token") {
		t.Fatalf("raw url leaked: %q", got)
	}
	if !strings.Contains(got, "cfsd***ijkl") {
		t.Fatalf("masked api key missing: %q", got)
	}
	if !strings.Contains(got, "<redacted_domain>") {
		t.Fatalf("redacted domain missing: %q", got)
	}
	if !strings.Contains(got, "<redacted_url>") {
		t.Fatalf("redacted url missing: %q", got)
	}
}

func TestWritePublicErrorReport(t *testing.T) {
	var buf bytes.Buffer

	err := errors.Join(
		errors.New("config failed"),
		errors.New("run renew for api[1] failed: 2 renew request(s) failed: http 429 Rate limit exceeded"),
	)

	WritePublicErrorReport(&buf, err)

	got := buf.String()
	if !strings.Contains(got, "error_count=2") {
		t.Fatalf("error_count missing: %q", got)
	}
	if !strings.Contains(got, "error[1]=config failed") {
		t.Fatalf("first error missing: %q", got)
	}
	if !strings.Contains(got, "error[2]=run renew for api[1] failed: 2 renew request(s) failed: http 429 Rate limit exceeded") {
		t.Fatalf("second error missing: %q", got)
	}
}
