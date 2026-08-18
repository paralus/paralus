package dao

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// genericEntity returns a fresh model with Partner/Organization/Project
// filter fields, for exercising the generic filter-building helpers in
// common.go without hard-coding one specific domain model everywhere.
func genericEntity() *models.ProjectAccountNamespaceRole {
	return &models.ProjectAccountNamespaceRole{}
}

func TestCreate(t *testing.T) {
	t.Run("succeeds (INSERT ... RETURNING executes as Query under pgdialect)", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		res, err := Create(context.Background(), db, genericEntity())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := Create(context.Background(), db, genericEntity())
		assert.Error(t, err)
	})
}

func TestGetX(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("x"))

		res, err := GetX(context.Background(), db, "name", "x", genericEntity())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetX(context.Background(), db, "name", "x", genericEntity())
		assert.Error(t, err)
	})
}

func TestGetM(t *testing.T) {
	t.Run("succeeds with multiple checks", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("x"))

		res, err := GetM(context.Background(), db, map[string]interface{}{"name": "x", "trash": false}, genericEntity())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetM(context.Background(), db, map[string]interface{}{"name": "x"}, genericEntity())
		assert.Error(t, err)
	})
}

func TestGetByID(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		res, err := GetByID(context.Background(), db, uuid.New(), genericEntity())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetByID(context.Background(), db, uuid.New(), genericEntity())
		assert.Error(t, err)
	})
}

func TestGetByName(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("x"))

		res, err := GetByName(context.Background(), db, "x", genericEntity())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetByName(context.Background(), db, "x", genericEntity())
		assert.Error(t, err)
	})
}

func TestGetByNamePartnerOrg(t *testing.T) {
	t.Run("applies partner and org filters when valid", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("x"))

		pid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		oid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		res, err := GetByNamePartnerOrg(context.Background(), db, "x", pid, oid, genericEntity())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("skips filters when invalid", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("x"))

		res, err := GetByNamePartnerOrg(context.Background(), db, "x", uuid.NullUUID{}, uuid.NullUUID{}, genericEntity())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetByNamePartnerOrg(context.Background(), db, "x", uuid.NullUUID{}, uuid.NullUUID{}, genericEntity())
		assert.Error(t, err)
	})
}

func TestGetIdByName(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	res, err := GetIdByName(context.Background(), db, "x", genericEntity())
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestGetAttributesByName(t *testing.T) {
	t.Run("selects the requested columns", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id", "namespace"}).AddRow(uuid.New(), "ns-1"))

		res, err := GetAttributesByName(context.Background(), db, "x", genericEntity(), "id", "namespace")
		require.NoError(t, err)
		got := res.(*models.ProjectAccountNamespaceRole)
		assert.Equal(t, "ns-1", got.Namespace)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetAttributesByName(context.Background(), db, "x", genericEntity(), "id")
		assert.Error(t, err)
	})
}

func TestGetIdByNamePartnerOrg(t *testing.T) {
	t.Run("applies filters when valid", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		pid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		oid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		res, err := GetIdByNamePartnerOrg(context.Background(), db, "x", pid, oid, genericEntity())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetIdByNamePartnerOrg(context.Background(), db, "x", uuid.NullUUID{}, uuid.NullUUID{}, genericEntity())
		assert.Error(t, err)
	})
}

func TestGetNameById(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("x"))

	res, err := GetNameById(context.Background(), db, uuid.New(), genericEntity())
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestGetNamesByIds(t *testing.T) {
	t.Run("short-circuits on empty id list without querying", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		res, err := GetNamesByIds(context.Background(), db, []uuid.UUID{}, genericEntity())
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("returns names for given ids", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("a").AddRow("b"))

		res, err := GetNamesByIds(context.Background(), db, []uuid.UUID{uuid.New(), uuid.New()}, genericEntity())
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetNamesByIds(context.Background(), db, []uuid.UUID{uuid.New()}, genericEntity())
		assert.Error(t, err)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		res, err := Update(context.Background(), db, uuid.New(), genericEntity())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		_, err := Update(context.Background(), db, uuid.New(), genericEntity())
		assert.Error(t, err)
	})
}

func TestUpdateX(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		res, err := UpdateX(context.Background(), db, "name", "x", genericEntity())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		_, err := UpdateX(context.Background(), db, "name", "x", genericEntity())
		assert.Error(t, err)
	})
}

func TestDelete(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := Delete(context.Background(), db, uuid.New(), genericEntity())
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := Delete(context.Background(), db, uuid.New(), genericEntity())
		assert.Error(t, err)
	})
}

func TestDeleteX(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := DeleteX(context.Background(), db, "name", "x", genericEntity())
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := DeleteX(context.Background(), db, "name", "x", genericEntity())
		assert.Error(t, err)
	})
}

func TestDeleteR(t *testing.T) {
	// Returning("*") forces the driver call through Query, not Exec.
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		err := DeleteR(context.Background(), db, uuid.New(), genericEntity())
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := DeleteR(context.Background(), db, uuid.New(), genericEntity())
		assert.Error(t, err)
	})
}

func TestDeleteXR(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		err := DeleteXR(context.Background(), db, "name", "x", genericEntity())
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := DeleteXR(context.Background(), db, "name", "x", genericEntity())
		assert.Error(t, err)
	})
}

func TestHardDeleteAll(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 5))

		err := HardDeleteAll(context.Background(), db, genericEntity())
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := HardDeleteAll(context.Background(), db, genericEntity())
		assert.Error(t, err)
	})
}

func TestList(t *testing.T) {
	t.Run("applies partner and org filters when valid", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("x"))

		var out []models.ProjectAccountNamespaceRole
		pid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		oid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		_, err := List(context.Background(), db, pid, oid, &out)
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		var out []models.ProjectAccountNamespaceRole
		_, err := List(context.Background(), db, uuid.NullUUID{}, uuid.NullUUID{}, &out)
		assert.Error(t, err)
	})
}

func TestListFiltered(t *testing.T) {
	t.Run("applies query, filters, ordering and pagination", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("x"))

		var out []models.ProjectAccountNamespaceRole
		pid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		oid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		projID := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		_, err := ListFiltered(context.Background(), db, pid, oid, projID, &out, "search", "name", "asc", 10, 0)
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		var out []models.ProjectAccountNamespaceRole
		_, err := ListFiltered(context.Background(), db, uuid.NullUUID{}, uuid.NullUUID{}, uuid.NullUUID{}, &out, "", "", "", 0, 0)
		assert.Error(t, err)
	})
}

func TestListByProject(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("x"))

		var out []models.ProjectAccountNamespaceRole
		pid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		err := ListByProject(context.Background(), db, pid, uuid.NullUUID{}, uuid.NullUUID{}, &out)
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		var out []models.ProjectAccountNamespaceRole
		err := ListByProject(context.Background(), db, uuid.NullUUID{}, uuid.NullUUID{}, uuid.NullUUID{}, &out)
		assert.Error(t, err)
	})
}

func TestListAll(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("x"))

		var out []models.ProjectAccountNamespaceRole
		_, err := ListAll(context.Background(), db, &out)
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		var out []models.ProjectAccountNamespaceRole
		_, err := ListAll(context.Background(), db, &out)
		assert.Error(t, err)
	})
}

func TestGetUserByEmail(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		res, err := GetUserByEmail(context.Background(), db, "a@b.com", &models.KratosIdentities{})
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetUserByEmail(context.Background(), db, "a@b.com", &models.KratosIdentities{})
		assert.Error(t, err)
	})
}

func TestGetUserFullByEmail(t *testing.T) {
	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetUserFullByEmail(context.Background(), db, "a@b.com", &models.KratosIdentities{})
		assert.Error(t, err)
	})
}

func TestGetUserIdByEmail(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		res, err := GetUserIdByEmail(context.Background(), db, "a@b.com", &models.KratosIdentities{})
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetUserIdByEmail(context.Background(), db, "a@b.com", &models.KratosIdentities{})
		assert.Error(t, err)
	})
}

func TestGetUserLastAuthTime(t *testing.T) {
	t.Run("returns the max authenticated_at time", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		want := time.Now().Truncate(time.Second)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(want))

		got, err := GetUserLastAuthTime(context.Background(), db, uuid.New())
		require.NoError(t, err)
		assert.True(t, want.Equal(got))
	})

	t.Run("propagates a generic query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetUserLastAuthTime(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}
