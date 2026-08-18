package dao

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUsers(t *testing.T) {
	t.Run("returns identities in the group", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		rows := sqlmock.NewRows([]string{
			"id", "schema_id", "traits", "created_at", "updated_at",
			"state", "state_changed_at", "nid", "metadata_public",
		}).AddRow(
			uuid.New(), "default", []byte(`{"email":"a@b.com"}`), time.Now(), time.Now(),
			"active", time.Now(), uuid.New(), []byte(`{}`),
		)
		mock.ExpectQuery(".*").WillReturnRows(rows)

		res, err := GetUsers(context.Background(), db, uuid.New())
		require.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, "default", res[0].SchemaId)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetUsers(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}

func TestGetGroupRoles(t *testing.T) {
	t.Run("aggregates group, project-group, and project-group-namespace roles", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"role", "group"}).AddRow("admin", "team-a"),
		)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"role", "project", "group"}).AddRow("editor", "proj-a", "team-a"),
		)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"role", "project", "namespace", "group"}).AddRow("viewer", "proj-a", "ns-a", "team-a"),
		)

		res, err := GetGroupRoles(context.Background(), db, uuid.New())
		require.NoError(t, err)
		require.Len(t, res, 3)
		assert.Equal(t, "admin", res[0].Role)
		assert.Equal(t, "editor", res[1].Role)
		assert.Equal(t, "viewer", res[2].Role)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propagates error from group role query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetGroupRoles(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})

	t.Run("propagates error from project group role query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role", "group"}))
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetGroupRoles(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})

	t.Run("propagates error from project group namespace role query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role", "group"}))
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role", "project", "group"}))
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetGroupRoles(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}
