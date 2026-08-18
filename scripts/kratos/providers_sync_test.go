package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"gopkg.in/yaml.v2"
)

func newMockBunDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqldb.Close() })
	return bun.NewDB(sqldb, pgdialect.New()), mock
}

func TestSync(t *testing.T) {
	t.Run("writes the fetched providers as a Kratos-shaped YAML config", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		rows := sqlmock.NewRows([]string{"name", "provider_name", "client_id", "client_secret", "scopes", "issuer_url"}).
			AddRow("google", "google", "client-1", "secret-1", "{email,profile}", "https://accounts.google.com")
		mock.ExpectQuery(`(?i)SELECT`).WillReturnRows(rows)

		out := filepath.Join(t.TempDir(), "providers.yaml")
		err := sync(context.Background(), db, out)
		require.NoError(t, err)

		data, err := os.ReadFile(out)
		require.NoError(t, err)

		var c Config
		require.NoError(t, yaml.Unmarshal(data, &c))
		require.Len(t, c.Selfservice.Methods.Oidc.Config.Providers, 1)
		p := c.Selfservice.Methods.Oidc.Config.Providers[0]
		assert.Equal(t, "google", p.Id)
		assert.Equal(t, "client-1", p.ClientId)
		assert.Equal(t, "https://accounts.google.com", p.IssuerURL)
	})

	t.Run("writes an empty provider list when the DB has none", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		rows := sqlmock.NewRows([]string{"name", "provider_name", "client_id", "client_secret", "scopes", "issuer_url"})
		mock.ExpectQuery(`(?i)SELECT`).WillReturnRows(rows)

		out := filepath.Join(t.TempDir(), "providers.yaml")
		err := sync(context.Background(), db, out)
		require.NoError(t, err)

		data, err := os.ReadFile(out)
		require.NoError(t, err)
		var c Config
		require.NoError(t, yaml.Unmarshal(data, &c))
		assert.Empty(t, c.Selfservice.Methods.Oidc.Config.Providers)
	})

	t.Run("propagates a DB query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(`(?i)SELECT`).WillReturnError(errors.New("connection reset"))

		out := filepath.Join(t.TempDir(), "providers.yaml")
		err := sync(context.Background(), db, out)
		assert.Error(t, err)
		_, statErr := os.Stat(out)
		assert.True(t, os.IsNotExist(statErr), "should not write the output file when the DB query fails")
	})

	t.Run("propagates a file-write error for an unwritable path", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		rows := sqlmock.NewRows([]string{"name", "provider_name", "client_id", "client_secret", "scopes", "issuer_url"})
		mock.ExpectQuery(`(?i)SELECT`).WillReturnRows(rows)

		err := sync(context.Background(), db, filepath.Join(t.TempDir(), "no-such-dir", "providers.yaml"))
		assert.Error(t, err)
	})
}
