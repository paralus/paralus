package dao

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// newMockBunDB returns a *bun.DB backed by go-sqlmock so dao functions can be
// unit tested without a real Postgres instance.
func newMockBunDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqldb.Close() })
	return bun.NewDB(sqldb, pgdialect.New()), mock
}
