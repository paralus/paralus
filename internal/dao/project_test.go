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

func TestGetProjectOrganization(t *testing.T) {
	t.Run("returns project/org info", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		rows := sqlmock.NewRows([]string{"project", "organization", "project_id", "organization_id", "partner_id"}).
			AddRow("proj-1", "org-1", "p-id", "o-id", "pt-id")
		mock.ExpectQuery(".*").WillReturnRows(rows)

		res, err := GetProjectOrganization(context.Background(), db, "proj-1")
		require.NoError(t, err)
		assert.Equal(t, "proj-1", res.Project)
		assert.Equal(t, "org-1", res.Organization)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetProjectOrganization(context.Background(), db, "missing")
		assert.Error(t, err)
	})
}

func TestGetFileteredProjects(t *testing.T) {
	t.Run("returns all projects when a permission grants global project access", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"account_id", "project_id"}).AddRow(uuid.New(), uuid.Nil),
		)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"name"}).AddRow("proj-a").AddRow("proj-b"),
		)

		res, err := GetFileteredProjects(context.Background(), db, uuid.New(), uuid.New(), uuid.New())
		require.NoError(t, err)
		require.Len(t, res, 2)
	})

	t.Run("filters by specific project ids when no global permission is found", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		projID := uuid.New()
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"account_id", "project_id"}).AddRow(uuid.New(), projID),
		)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"name"}).AddRow("proj-a"),
		)

		res, err := GetFileteredProjects(context.Background(), db, uuid.New(), uuid.New(), uuid.New())
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("short-circuits with empty slice when no permissions grant any project", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"account_id", "project_id"}))

		res, err := GetFileteredProjects(context.Background(), db, uuid.New(), uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.Empty(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propagates error from permission query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetFileteredProjects(context.Background(), db, uuid.New(), uuid.New(), uuid.New())
		assert.Error(t, err)
	})

	t.Run("propagates error from project query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"account_id", "project_id"}).AddRow(uuid.New(), uuid.Nil),
		)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetFileteredProjects(context.Background(), db, uuid.New(), uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}

func TestGetProjectGroupRoles(t *testing.T) {
	t.Run("aggregates project-group and project-group-namespace roles", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"role", "project", "group"}).AddRow("editor", "proj-1", "team-a"),
		)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"role", "project", "group", "namespace"}).AddRow("viewer", "proj-1", "team-a", "ns-1"),
		)

		res, err := GetProjectGroupRoles(context.Background(), db, uuid.New())
		require.NoError(t, err)
		require.Len(t, res, 2)
	})

	t.Run("propagates error from first query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetProjectGroupRoles(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})

	t.Run("propagates error from second query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role", "project", "group"}))
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetProjectGroupRoles(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}

func TestGetProjectUserRoles(t *testing.T) {
	t.Run("aggregates project-user and project-user-namespace roles", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"role", "user"}).AddRow("editor", "a@b.com"),
		)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"role", "user", "namespace"}).AddRow("viewer", "a@b.com", "ns-1"),
		)

		res, err := GetProjectUserRoles(context.Background(), db, uuid.New())
		require.NoError(t, err)
		require.Len(t, res, 2)
	})

	t.Run("propagates error from first query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetProjectUserRoles(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})

	t.Run("propagates error from second query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role", "user"}))
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetProjectUserRoles(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}
