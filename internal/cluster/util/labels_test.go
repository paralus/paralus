package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidKubernetesLabelNameValueRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple alnum", "abc123", true},
		{"with dashes underscores dots", "a-b_c.d", true},
		{"single char", "a", true},
		{"empty string does not match", "", false},
		{"starts with dash", "-abc", false},
		{"ends with dash", "abc-", false},
		{"contains slash", "a/b", false},
		{"contains space", "a b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidKubernetesLabelNameValueRegex(tt.input))
		})
	}
}

func TestValidateCustomLabelKey(t *testing.T) {
	t.Run("empty key errors", func(t *testing.T) {
		assert.Error(t, ValidateCustomLabelKey("", "v"))
	})

	t.Run("valid key with no prefix", func(t *testing.T) {
		assert.NoError(t, ValidateCustomLabelKey("mykey", "v"))
	})

	t.Run("valid key with custom prefix", func(t *testing.T) {
		assert.NoError(t, ValidateCustomLabelKey("example.com/mykey", "v"))
	})

	t.Run("rejects reserved paralus.dev prefix", func(t *testing.T) {
		err := ValidateCustomLabelKey("paralus.dev/mykey", "v")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reserved domain")
	})

	t.Run("rejects too-long prefix", func(t *testing.T) {
		longPrefix := strings.Repeat("a", 254)
		err := ValidateCustomLabelKey(longPrefix+"/key", "v")
		assert.Error(t, err)
	})

	t.Run("rejects invalid prefix characters", func(t *testing.T) {
		err := ValidateCustomLabelKey("-bad-/key", "v")
		assert.Error(t, err)
	})

	t.Run("rejects empty suffix", func(t *testing.T) {
		err := ValidateCustomLabelKey("example.com/", "v")
		assert.Error(t, err)
	})

	t.Run("rejects too-long suffix", func(t *testing.T) {
		longKey := strings.Repeat("a", 64)
		err := ValidateCustomLabelKey(longKey, "v")
		assert.Error(t, err)
	})

	t.Run("rejects invalid suffix characters", func(t *testing.T) {
		err := ValidateCustomLabelKey("bad key", "v")
		assert.Error(t, err)
	})
}

func TestValidateCustomLabels(t *testing.T) {
	t.Run("all valid", func(t *testing.T) {
		err := ValidateCustomLabels(map[string]string{"key1": "val1", "example.com/key2": "val2"})
		assert.NoError(t, err)
	})

	t.Run("invalid key short-circuits", func(t *testing.T) {
		err := ValidateCustomLabels(map[string]string{"": "val1"})
		assert.Error(t, err)
	})

	t.Run("too-long value short-circuits", func(t *testing.T) {
		longValue := strings.Repeat("a", 64)
		err := ValidateCustomLabels(map[string]string{"key1": longValue})
		assert.Error(t, err)
	})

	t.Run("invalid value characters short-circuit", func(t *testing.T) {
		err := ValidateCustomLabels(map[string]string{"key1": "bad value"})
		assert.Error(t, err)
	})
}

func TestSanitizeLabelValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"trims whitespace", "  Hello World  ", "hello-world"},
		{"collapses internal whitespace", "Hello   World", "hello-world"},
		{"replaces slash with dash", "path/to/thing", "path-to-thing"},
		{"lowercases", "ABC", "abc"},
		{"already sanitized", "abc-def", "abc-def"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SanitizeLabelValues(tt.input))
		})
	}
}

func TestValidateFullKey(t *testing.T) {
	t.Run("empty key errors", func(t *testing.T) {
		assert.Error(t, ValidateFullKey("", "v"))
	})

	t.Run("valid key with prefix", func(t *testing.T) {
		assert.NoError(t, ValidateFullKey("example.com/mykey", "v"))
	})

	t.Run("rejects reserved paralus.dev prefix", func(t *testing.T) {
		err := ValidateFullKey("paralus.dev/mykey", "v")
		assert.Error(t, err)
	})

	t.Run("rejects invalid key suffix", func(t *testing.T) {
		err := ValidateFullKey("example.com/bad key", "v")
		assert.Error(t, err)
	})
}

func TestValidatePrefix(t *testing.T) {
	t.Run("empty prefix is valid", func(t *testing.T) {
		assert.NoError(t, ValidatePrefix(""))
	})

	t.Run("too long prefix errors", func(t *testing.T) {
		assert.Error(t, ValidatePrefix(strings.Repeat("a", 254)))
	})

	t.Run("reserved domain errors", func(t *testing.T) {
		assert.Error(t, ValidatePrefix("paralus.dev"))
	})

	t.Run("invalid characters error", func(t *testing.T) {
		assert.Error(t, ValidatePrefix("-bad-"))
	})

	t.Run("valid prefix passes", func(t *testing.T) {
		assert.NoError(t, ValidatePrefix("example.com"))
	})
}

func TestValidateKey(t *testing.T) {
	t.Run("empty key errors", func(t *testing.T) {
		assert.Error(t, ValidateKey("", "v"))
	})

	t.Run("too long key errors", func(t *testing.T) {
		assert.Error(t, ValidateKey(strings.Repeat("a", 64), "v"))
	})

	t.Run("invalid characters error", func(t *testing.T) {
		assert.Error(t, ValidateKey("bad key", "v"))
	})

	t.Run("valid key passes", func(t *testing.T) {
		assert.NoError(t, ValidateKey("mykey", "v"))
	})
}

func TestValidateValue(t *testing.T) {
	t.Run("empty value is valid", func(t *testing.T) {
		assert.NoError(t, ValidateValue("k", ""))
	})

	t.Run("too long value errors", func(t *testing.T) {
		assert.Error(t, ValidateValue("k", strings.Repeat("a", 64)))
	})

	t.Run("invalid characters error", func(t *testing.T) {
		assert.Error(t, ValidateValue("k", "bad value"))
	})

	t.Run("valid value passes", func(t *testing.T) {
		assert.NoError(t, ValidateValue("k", "myvalue"))
	})
}
