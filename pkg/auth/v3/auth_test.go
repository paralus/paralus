package authv3

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvWithDefault(t *testing.T) {
	t.Run("returns env value when set", func(t *testing.T) {
		t.Setenv("PARALUS_TEST_ENV", "actual")
		assert.Equal(t, "actual", getEnvWithDefault("PARALUS_TEST_ENV", "fallback"))
	})

	t.Run("returns default when unset", func(t *testing.T) {
		os.Unsetenv("PARALUS_TEST_ENV_UNSET")
		assert.Equal(t, "fallback", getEnvWithDefault("PARALUS_TEST_ENV_UNSET", "fallback"))
	})
}

func TestGetDSN(t *testing.T) {
	t.Run("uses DSN directly when set", func(t *testing.T) {
		t.Setenv("DSN", "postgres://explicit-dsn")
		assert.Equal(t, "postgres://explicit-dsn", getDSN())
	})

	t.Run("builds DSN from parts when DSN unset", func(t *testing.T) {
		os.Unsetenv("DSN")
		t.Setenv("DB_USER", "u")
		t.Setenv("DB_PASSWORD", "p")
		t.Setenv("DB_ADDR", "db:5432")
		t.Setenv("DB_NAME", "n")

		want := "postgres://u:p@db:5432/n?sslmode=disable"
		assert.Equal(t, want, getDSN())
	})
}

func TestNewAuthContext(t *testing.T) {
	ks := &fakeApiKeyService{}
	as := &fakeAuthzService{}

	ac := NewAuthContext(nil, nil, ks, as)

	assert.Nil(t, ac.db)
	assert.Nil(t, ac.kc)
	assert.Same(t, ks, ac.ks)
	assert.Same(t, as, ac.as)
}
