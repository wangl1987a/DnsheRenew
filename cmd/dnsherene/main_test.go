package main

import "testing"

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
			if got := maskAPIKey(tc.in); got != tc.want {
				t.Fatalf("maskAPIKey(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
