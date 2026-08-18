package dao

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetKubeconfigRevocation(t *testing.T) {
	t.Run("returns revocation row", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		id := uuid.New()
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))

		res, err := GetKubeconfigRevocation(context.Background(), db, uuid.New(), uuid.New(), false)
		require.NoError(t, err)
		assert.Equal(t, id, res.ID)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetKubeconfigRevocation(context.Background(), db, uuid.New(), uuid.New(), false)
		assert.Error(t, err)
	})
}

func TestCreateKubeconfigRevocation(t *testing.T) {
	// bun issues INSERT ... RETURNING under pgdialect, which the driver
	// executes as a Query, not an Exec.
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		err := CreateKubeconfigRevocation(context.Background(), db, &models.KubeconfigRevocation{})
		require.NoError(t, err)
	})

	t.Run("propagates exec error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := CreateKubeconfigRevocation(context.Background(), db, &models.KubeconfigRevocation{})
		assert.Error(t, err)
	})
}

func TestUpdateKubeconfigRevocation(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := UpdateKubeconfigRevocation(context.Background(), db, &models.KubeconfigRevocation{
			OrganizationId: uuid.New(), AccountId: uuid.New(),
		})
		require.NoError(t, err)
	})

	t.Run("propagates exec error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := UpdateKubeconfigRevocation(context.Background(), db, &models.KubeconfigRevocation{})
		assert.Error(t, err)
	})
}

func TestGetKubeconfigSetting(t *testing.T) {
	t.Run("filters by account when accountID is set", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"scope"}).AddRow("USER"))

		res, err := GetKubeconfigSetting(context.Background(), db, uuid.New(), uuid.New(), true)
		require.NoError(t, err)
		assert.Equal(t, "USER", res.Scope)
	})

	t.Run("omits account filter when accountID is uuid.Nil", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"scope"}).AddRow("ORGANIZATION"))

		res, err := GetKubeconfigSetting(context.Background(), db, uuid.New(), uuid.Nil, false)
		require.NoError(t, err)
		assert.Equal(t, "ORGANIZATION", res.Scope)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetKubeconfigSetting(context.Background(), db, uuid.New(), uuid.New(), false)
		assert.Error(t, err)
	})
}

func TestCreateKubeconfigSetting(t *testing.T) {
	// bun issues INSERT ... RETURNING under pgdialect, which the driver
	// executes as a Query, not an Exec.
	t.Run("sets organization scope when AccountId is nil", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		ks := &models.KubeconfigSetting{}
		err := CreateKubeconfigSetting(context.Background(), db, ks)
		require.NoError(t, err)
		assert.Equal(t, sentry.KubeconfigSettingOrganizationScope, ks.Scope)
	})

	t.Run("sets user scope when AccountId is set", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		ks := &models.KubeconfigSetting{AccountId: uuid.New()}
		err := CreateKubeconfigSetting(context.Background(), db, ks)
		require.NoError(t, err)
		assert.Equal(t, sentry.KubeconfigSettingUserScope, ks.Scope)
	})

	t.Run("propagates exec error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := CreateKubeconfigSetting(context.Background(), db, &models.KubeconfigSetting{})
		assert.Error(t, err)
	})
}

func TestUpdateKubeconfigSetting(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := UpdateKubeconfigSetting(context.Background(), db, &models.KubeconfigSetting{})
		require.NoError(t, err)
	})

	t.Run("propagates exec error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := UpdateKubeconfigSetting(context.Background(), db, &models.KubeconfigSetting{})
		assert.Error(t, err)
	})
}

func TestGetkubectlClusterSettings(t *testing.T) {
	t.Run("returns setting", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("cluster-1"))

		res, err := GetkubectlClusterSettings(context.Background(), db, uuid.New(), "cluster-1")
		require.NoError(t, err)
		assert.Equal(t, "cluster-1", res.Name)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetkubectlClusterSettings(context.Background(), db, uuid.New(), "missing")
		assert.Error(t, err)
	})
}

func TestCreatekubectlClusterSettings(t *testing.T) {
	db, mock := newMockBunDB(t)
	// bun issues INSERT ... RETURNING under pgdialect, which the driver
	// executes as a Query, not an Exec.
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("c1"))

	err := CreatekubectlClusterSettings(context.Background(), db, &models.KubectlClusterSetting{Name: "c1"})
	require.NoError(t, err)
}

func TestUpdatekubectlClusterSettings(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := UpdatekubectlClusterSettings(context.Background(), db, &models.KubectlClusterSetting{Name: "c1"})
		require.NoError(t, err)
	})

	t.Run("propagates exec error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := UpdatekubectlClusterSettings(context.Background(), db, &models.KubectlClusterSetting{Name: "c1"})
		assert.Error(t, err)
	})
}
