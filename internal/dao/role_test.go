package dao

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRolePermissions(t *testing.T) {
	t.Run("returns permissions on success", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		rows := sqlmock.NewRows([]string{"name"}).AddRow("cluster:view").AddRow("cluster:edit")
		mock.ExpectQuery(".*").WillReturnRows(rows)

		res, err := GetRolePermissions(context.Background(), db, uuid.New())
		require.NoError(t, err)
		require.Len(t, res, 2)
		assert.Equal(t, "cluster:view", res[0].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("db down"))

		res, err := GetRolePermissions(context.Background(), db, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("returns empty slice, not nil, when no rows", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}))

		res, err := GetRolePermissions(context.Background(), db, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestGetRolePermissionsByScope(t *testing.T) {
	t.Run("upper-cases scope and returns rows", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		rows := sqlmock.NewRows([]string{"name", "scope"}).AddRow("cluster:view", "PROJECT")
		mock.ExpectQuery(".*").WillReturnRows(rows)

		res, err := GetRolePermissionsByScope(context.Background(), db, "project")
		require.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, "PROJECT", res[0].Scope)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetRolePermissionsByScope(context.Background(), db, "project")
		assert.Error(t, err)
	})
}

func TestGetRolePermissionsByNames(t *testing.T) {
	t.Run("returns matching permissions", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		rows := sqlmock.NewRows([]string{"name", "description", "scope"}).
			AddRow("cluster:view", "view clusters", "PROJECT")
		mock.ExpectQuery(".*").WillReturnRows(rows)

		res, err := GetRolePermissionsByNames(context.Background(), db, "cluster:view")
		require.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, "cluster:view", res[0].Name)
	})

	t.Run("works with zero names given", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name", "description", "scope"}))

		res, err := GetRolePermissionsByNames(context.Background(), db)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetRolePermissionsByNames(context.Background(), db, "x")
		assert.Error(t, err)
	})
}
