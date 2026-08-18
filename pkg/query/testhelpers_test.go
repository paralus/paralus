package query

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// testEntity is a minimal bun model with the partner/organization/project/name
// columns that pkg/query's filter builders operate on, used across this
// package's tests instead of pulling in a real domain model.
type testEntity struct {
	bun.BaseModel `bun:"table:test_entities,alias:te"`

	ID             string `bun:"id,pk"`
	Name           string `bun:"name"`
	PartnerId      string `bun:"partner_id"`
	OrganizationId string `bun:"organization_id"`
	ProjectId      string `bun:"project_id"`
}

// newTestDB returns a *bun.DB backed by go-sqlmock, used only to render SQL
// via AppendQuery in these tests -- no query is ever actually executed
// against it.
func newTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqldb.Close() })
	return bun.NewDB(sqldb, pgdialect.New())
}

// renderSelect renders the SQL a *bun.SelectQuery would execute, for
// assertions on which WHERE clauses a query-builder function added.
func renderSelect(t *testing.T, q *bun.SelectQuery) string {
	t.Helper()
	b, err := q.AppendQuery(q.DB().Formatter(), nil)
	if err != nil {
		t.Fatalf("AppendQuery: %v", err)
	}
	return string(b)
}

func renderUpdate(t *testing.T, q *bun.UpdateQuery) string {
	t.Helper()
	b, err := q.AppendQuery(q.DB().Formatter(), nil)
	if err != nil {
		t.Fatalf("AppendQuery: %v", err)
	}
	return string(b)
}
