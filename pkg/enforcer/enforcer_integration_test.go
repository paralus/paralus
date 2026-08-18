//go:build integration

package enforcer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Init() wires a real gorm-adapter (which auto-migrates a casbin_rule table)
// and a casbin model/enforcer around a live database. That schema-creation
// and cross-library wiring is the kind of seam a hand-rolled sqlmock can't
// exercise faithfully, so it's covered here against a real Postgres instead.
func TestCasbinEnforcer_Init_Integration(t *testing.T) {
	ctx := context.Background()

	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("paralus_test"),
		postgres.WithUsername("paralus"),
		postgres.WithPassword("paralus"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = pg.Terminate(ctx)
	})

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	e := NewCasbinEnforcer(db)
	enforcer, err := e.Init()
	require.NoError(t, err)
	require.NotNil(t, enforcer)

	// g2 groups "admin" under the "admin:ops_star" role via KeyMatchCu's
	// "*" shortcut, and the policy grants ns=* proj=* org=* obj=cluster
	// act=view to that role.
	_, err = enforcer.AddNamedGroupingPolicy("g2", "admin", "admin:ops_star")
	require.NoError(t, err)
	_, err = enforcer.AddPolicy("admin:ops_star", "*", "*", "*", "cluster")
	require.NoError(t, err)

	ok, err := enforcer.Enforce("admin", "team-a", "proj-a", "org-a", "cluster", "view")
	require.NoError(t, err)
	require.True(t, ok, "wildcard g2 role + wildcard policy should allow any ns/proj/org")

	ok, err = enforcer.Enforce("someone-else", "team-a", "proj-a", "org-a", "cluster", "view")
	require.NoError(t, err)
	require.False(t, ok, "a subject with no grouping policy must not be allowed")
}
