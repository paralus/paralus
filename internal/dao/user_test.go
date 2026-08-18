package dao

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserType(t *testing.T) {
	t.Run("returns credential type name", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"identity_credential__identity_credential_type__name"}).AddRow("password"),
		)

		got, err := GetUserType(context.Background(), db, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, "password", got)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetUserType(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}

func TestGetGroups(t *testing.T) {
	t.Run("returns groups", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("team-a"))

		res, err := GetGroups(context.Background(), db, uuid.New())
		require.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, "team-a", res[0].Name)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetGroups(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}

func TestGetUserRoles(t *testing.T) {
	t.Run("aggregates account, project-account, and project-account-namespace roles", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("admin"))
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("editor", "proj-a"))
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("viewer", "proj-a", "ns-a"))

		res, err := GetUserRoles(context.Background(), db, uuid.New())
		require.NoError(t, err)
		require.Len(t, res, 3)
	})

	t.Run("propagates error from first query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetUserRoles(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})

	t.Run("propagates error from second query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role"}))
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetUserRoles(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})

	t.Run("propagates error from third query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role"}))
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role", "project"}))
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetUserRoles(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}

func TestGetQueryFilteredUsers(t *testing.T) {
	t.Run("returns matching account ids with no optional filters", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(uuid.New()))

		res, err := GetQueryFilteredUsers(context.Background(), db, uuid.New(), uuid.New(), uuid.Nil, uuid.Nil, nil)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("applies role, project, and group filters when given", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(uuid.New()))

		res, err := GetQueryFilteredUsers(context.Background(), db, uuid.New(), uuid.New(), uuid.New(), uuid.New(), []uuid.UUID{uuid.New()})
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetQueryFilteredUsers(context.Background(), db, uuid.New(), uuid.New(), uuid.Nil, uuid.Nil, nil)
		assert.Error(t, err)
	})
}

func TestGetUserNamesByIds(t *testing.T) {
	t.Run("short-circuits on empty id list", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		res, err := GetUserNamesByIds(context.Background(), db, []uuid.UUID{}, &models.KratosIdentities{})
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("returns names for given ids", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("a@b.com"))

		res, err := GetUserNamesByIds(context.Background(), db, []uuid.UUID{uuid.New()}, &models.KratosIdentities{})
		require.NoError(t, err)
		assert.Equal(t, []string{"a@b.com"}, res)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetUserNamesByIds(context.Background(), db, []uuid.UUID{uuid.New()}, &models.KratosIdentities{})
		assert.Error(t, err)
	})
}

func TestIsSSOAccount(t *testing.T) {
	t.Run("true when a matching oidc credential row is found", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		got, err := IsSSOAccount(context.Background(), db, uuid.New())
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("false, no error, on sql.ErrNoRows", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(sql.ErrNoRows)

		got, err := IsSSOAccount(context.Background(), db, uuid.New())
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("propagates other errors", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := IsSSOAccount(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}

func TestListFilteredUsers(t *testing.T) {
	t.Run("plain listing with no type/query filters", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		res, err := ListFilteredUsers(context.Background(), db, nil, "", "", "", "", 0, 0)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("filters by precomputed user ids and query string", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		res, err := ListFilteredUsers(context.Background(), db, []uuid.UUID{uuid.New()}, "bob", "", "email", "asc", 10, 0)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("password type excludes oidc users via an Except subquery", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		res, err := ListFilteredUsers(context.Background(), db, nil, "", KratosPasswordType, "", "", 0, 0)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := ListFilteredUsers(context.Background(), db, nil, "", "", "", "", 0, 0)
		assert.Error(t, err)
	})
}

func TestListFilteredUsersWithGroup(t *testing.T) {
	t.Run("returns users belonging to the group", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"account__id", "account__schema_id"}).AddRow(uuid.New(), "default"),
		)

		res, err := ListFilteredUsersWithGroup(context.Background(), db, nil, uuid.New(), "", "", "", "", 0, 0)
		require.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, "default", res[0].SchemaId)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := ListFilteredUsersWithGroup(context.Background(), db, nil, uuid.New(), "", "", "", "", 0, 0)
		assert.Error(t, err)
	})
}
