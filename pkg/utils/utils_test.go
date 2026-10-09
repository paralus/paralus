package utils

import (
	"strings"
	"testing"
)

func TestGetRandomPassword(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		p := GetRandomPassword(8)
		if len(p) != 8 {
			t.Fatalf("expected length 8, got %d", len(p))
		}
		if !strings.ContainsAny(p, "0123456789") {
			t.Fatalf("password %q has no digit", p)
		}
		if !strings.ContainsAny(p, "~=+%^*/()[]{}/!@#$?|") {
			t.Fatalf("password %q has no special character", p)
		}
		seen[p] = true
	}
	if len(seen) < 195 {
		t.Fatalf("too many duplicate passwords: %d unique of 200", len(seen))
	}
}
