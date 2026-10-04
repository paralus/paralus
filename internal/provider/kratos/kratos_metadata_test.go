package providers

import "testing"

func TestMetadataStringLowercaseKeys(t *testing.T) {
	m := map[string]interface{}{
		"organization": "org-from-sql",
		"partner":      "partner-from-sql",
		"ForceReset":   true,
	}
	if got := metadataString(m, "Organization", "organization"); got != "org-from-sql" {
		t.Fatalf("organization: got %q", got)
	}
	if got := metadataString(m, "Partner", "partner"); got != "partner-from-sql" {
		t.Fatalf("partner: got %q", got)
	}
	// A later Update() rebuilds metadata from this read. Missing the lowercase
	// keys used to persist empty Organization/Partner and drop the real ids.
	if metadataString(m, "Organization") != "" {
		t.Fatal("bare Organization key must not match the lowercase upgrade SQL key")
	}
}
