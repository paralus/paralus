package dao

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// newMockBunDB returns a *bun.DB backed by go-sqlmock so dao functions can be
// unit tested without a real Postgres instance. dao takes bun.IDB, which
// *bun.DB satisfies.
func newMockBunDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqldb.Close() })
	return bun.NewDB(sqldb, pgdialect.New()), mock
}

// expectScanAndCount wires up sqlmock for bun's SelectQuery.ScanAndCount when
// pagination is active: bun runs the row-fetch and a separate "count(*)"
// query concurrently in two goroutines, so arrival order at the driver is
// non-deterministic. MatchExpectationsInOrder(false) plus regexes that only
// match one query each (count(*) has no LIMIT; the paginated data query
// always has one, since Paginate() is what triggers this concurrent path)
// let sqlmock route each concurrent call correctly regardless of order.
// Go's regexp (RE2) has no negative lookahead, so the two patterns must be
// positive and mutually exclusive rather than "matches X" / "doesn't match X".
func expectScanAndCount(mock sqlmock.Sqlmock, dataRows *sqlmock.Rows, count int64) {
	mock.MatchExpectationsInOrder(false)
	mock.ExpectQuery(`(?i)count\(\*\)`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
	mock.ExpectQuery(`(?i)LIMIT`).WillReturnRows(dataRows)
}
