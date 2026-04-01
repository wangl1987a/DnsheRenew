package app

import (
	"reflect"
	"testing"

	"dnsherene/pkg/dnshe"
)

func TestSelectTargetsByIDs(t *testing.T) {
	all := []dnshe.Subdomain{
		{ID: 1, Subdomain: "api", Rootdomain: "example.com", FullDomain: "api.example.com"},
		{ID: 2, Subdomain: "blog", Rootdomain: "example.com", FullDomain: "blog.example.com"},
	}

	got, err := selectTargets(all, "2, 1, 2", "", "")
	if err != nil {
		t.Fatalf("selectTargets returned error: %v", err)
	}

	ids := make([]int, 0, len(got))
	for _, item := range got {
		ids = append(ids, item.ID)
	}

	want := []int{1, 2}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("unexpected ids: got=%v want=%v", ids, want)
	}
}

func TestSelectTargetsByFilter(t *testing.T) {
	all := []dnshe.Subdomain{
		{ID: 1, Subdomain: "api", Rootdomain: "example.com"},
		{ID: 2, Subdomain: "api", Rootdomain: "example.net"},
		{ID: 3, Subdomain: "blog", Rootdomain: "example.com"},
	}

	got, err := selectTargets(all, "", "example.com", "api")
	if err != nil {
		t.Fatalf("selectTargets returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 target, got %d", len(got))
	}
	if got[0].ID != 1 {
		t.Fatalf("expected ID=1, got ID=%d", got[0].ID)
	}
}

func TestParseIDsInvalid(t *testing.T) {
	_, err := parseIDs("a,2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
