package dao

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BUG: CreateOperatorBootstrap's existence check calls
// dao.GetX(ctx, db, "edge_id", ..., &bstrap) where bstrap is already declared
// as *models.ClusterOperatorBootstrap, so it passes a **models.ClusterOperatorBootstrap.
// bun rejects this locally ("bun: Model(unsupported **models.ClusterOperatorBootstrap)")
// without ever reaching the database, so the lookup always "fails" and the
// delete-existing-row branch is unreachable in production -- every call
// takes the "no existing bootstrap data" path and goes straight to INSERT,
// regardless of whether a row already exists. These tests document the
// actual, single-insert-only control flow.
func TestCreateOperatorBootstrap(t *testing.T) {
	t.Run("always skips the delete branch and inserts directly", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectBegin()
		tx, err := db.Begin()
		require.NoError(t, err)

		// INSERT ... RETURNING executes as a Query under pgdialect.
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"cluster_id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		bootstrap := &models.ClusterOperatorBootstrap{ClusterId: uuid.New()}
		err = CreateOperatorBootstrap(context.Background(), tx, bootstrap)
		require.NoError(t, err)
		require.NoError(t, tx.Commit())
	})

	t.Run("propagates the insert error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectBegin()
		tx, err := db.Begin()
		require.NoError(t, err)

		mock.ExpectQuery(".*").WillReturnError(errors.New("insert failed"))
		mock.ExpectRollback()

		bootstrap := &models.ClusterOperatorBootstrap{ClusterId: uuid.New()}
		err = CreateOperatorBootstrap(context.Background(), tx, bootstrap)
		assert.Error(t, err)
		require.NoError(t, tx.Rollback())
	})
}

func TestGetOperatorBootstrap(t *testing.T) {
	// BUG: GetOperatorBootstrap passes `bootstrap` (a models.ClusterOperatorBootstrap
	// value) instead of &bootstrap into dao.GetX, which calls db.NewSelect().Model(entity).
	// bun requires a pointer to scan into, so this fails immediately regardless
	// of what the mock returns.
	t.Run("always fails: passes a non-pointer Model to bun", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		_, err := GetOperatorBootstrap(context.Background(), db, "cluster-1")
		assert.Error(t, err)
	})
}
