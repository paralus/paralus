package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeUsername(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already lowercase alnum is unchanged", "bob123", "bob123"},
		{"uppercase is lowercased", "Bob", "bob"},
		{"at-sign and dot are replaced with -<ascii>", "bob@example.com", "bob-64example-46com"},
		{"empty string", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SanitizeUsername(tt.input))
		})
	}

	t.Run("truncates to 253 runes", func(t *testing.T) {
		long := strings.Repeat("a", 300)
		got := SanitizeUsername(long)
		require.Len(t, []rune(got), 253)
	})
}

func TestTemplateScope(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		assert.Equal(t, "template/abc", ToTemplateScope("abc"))
		scope, err := GetTemplateScope("template/abc")
		require.NoError(t, err)
		assert.Equal(t, "abc", scope)
	})

	t.Run("errors without the prefix", func(t *testing.T) {
		_, err := GetTemplateScope("abc")
		assert.Error(t, err)
	})
}

func TestClusterScope(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		assert.Equal(t, "cluster/abc", ToClusterScope("abc"))
		scope, err := GetClusterScope("cluster/abc")
		require.NoError(t, err)
		assert.Equal(t, "abc", scope)
	})

	t.Run("errors without the prefix", func(t *testing.T) {
		_, err := GetClusterScope("abc")
		assert.Error(t, err)
	})
}

func TestUserScope(t *testing.T) {
	t.Run("round trip for a regular user", func(t *testing.T) {
		assert.Equal(t, "user/abc", ToUserScope("abc"))
		scope, err := GetUserScope("user/abc")
		require.NoError(t, err)
		assert.Equal(t, "abc", scope)
	})

	t.Run("recognizes the ssouser prefix", func(t *testing.T) {
		scope, err := GetUserScope("ssouser/abc")
		require.NoError(t, err)
		assert.Equal(t, "abc", scope)
	})

	t.Run("errors without either prefix", func(t *testing.T) {
		_, err := GetUserScope("abc")
		assert.Error(t, err)
	})
}

func TestOrganizationScope(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		assert.Equal(t, "organization/abc", ToOrganizationScope("abc"))
		scope, err := GetOrganizationScope("organization/abc")
		require.NoError(t, err)
		assert.Equal(t, "abc", scope)
	})

	t.Run("errors without the prefix", func(t *testing.T) {
		_, err := GetOrganizationScope("abc")
		assert.Error(t, err)
	})
}
