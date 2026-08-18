package dao

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCluster(t *testing.T) {
	t.Run("creates a token, sets override selector, inserts, and labels the cluster", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		// CreateToken -> dao.Create -> INSERT ... RETURNING (Query).
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uuid.New(), "tok-1"))
		// cluster insert -> INSERT ... RETURNING (Query), Cluster.ID has a default.
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		// cluster-id label update -> plain UPDATE, no RETURNING.
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		cluster := &models.Cluster{ID: uuid.New(), Name: "cluster-1"}
		err := CreateCluster(context.Background(), db, cluster)
		require.NoError(t, err)
		assert.Equal(t, "paralus.dev/overrideCluster=cluster-1", cluster.OverrideSelector)
		assert.Equal(t, "tok-1", cluster.Token)
	})

	t.Run("keeps an explicit override selector", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uuid.New(), "tok-1"))
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		cluster := &models.Cluster{ID: uuid.New(), Name: "cluster-1", OverrideSelector: "custom=selector"}
		err := CreateCluster(context.Background(), db, cluster)
		require.NoError(t, err)
		assert.Equal(t, "custom=selector", cluster.OverrideSelector)
	})

	t.Run("propagates token creation error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := CreateCluster(context.Background(), db, &models.Cluster{Name: "cluster-1"})
		assert.Error(t, err)
	})

	t.Run("propagates cluster insert error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uuid.New(), "tok-1"))
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := CreateCluster(context.Background(), db, &models.Cluster{Name: "cluster-1"})
		assert.Error(t, err)
	})

	t.Run("propagates label-update error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uuid.New(), "tok-1"))
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := CreateCluster(context.Background(), db, &models.Cluster{Name: "cluster-1"})
		assert.Error(t, err)
	})
}

func TestUpdateCluster(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := UpdateCluster(context.Background(), db, &models.Cluster{ID: uuid.New()})
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := UpdateCluster(context.Background(), db, &models.Cluster{ID: uuid.New()})
		assert.Error(t, err)
	})
}

func TestUpdateClusterAnnotations(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := UpdateClusterAnnotations(context.Background(), db, &models.Cluster{ID: uuid.New()})
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := UpdateClusterAnnotations(context.Background(), db, &models.Cluster{ID: uuid.New()})
		assert.Error(t, err)
	})
}

func TestGetCluster(t *testing.T) {
	t.Run("looks up by ID when set", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("cluster-1"))

		res, err := GetCluster(context.Background(), db, &models.Cluster{ID: uuid.New()})
		require.NoError(t, err)
		assert.Equal(t, "cluster-1", res.Name)
	})

	t.Run("falls back to name lookup when ID is nil", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("cluster-1"))

		res, err := GetCluster(context.Background(), db, &models.Cluster{Name: "cluster-1"})
		require.NoError(t, err)
		assert.Equal(t, "cluster-1", res.Name)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetCluster(context.Background(), db, &models.Cluster{Name: "cluster-1"})
		assert.Error(t, err)
	})
}

func TestDeleteCluster(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := DeleteCluster(context.Background(), db, &models.Cluster{ID: uuid.New()})
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := DeleteCluster(context.Background(), db, &models.Cluster{ID: uuid.New()})
		assert.Error(t, err)
	})
}

func TestListClusters(t *testing.T) {
	// Each subtest builds its own commonv3.QueryOptions literal rather than
	// copying a shared value, since QueryOptions embeds a protobuf
	// MessageState (containing a sync.Mutex) that must not be copied.
	newValidOpts := func() commonv3.QueryOptions {
		return commonv3.QueryOptions{
			Partner:      uuid.New().String(),
			Organization: uuid.New().String(),
			Project:      uuid.New().String(),
		}
	}

	t.Run("uses ListFiltered when a query or order is given", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("cluster-1"))

		opts := newValidOpts()
		opts.Q = "search-term"
		res, err := ListClusters(context.Background(), db, opts)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("uses ListByProject when no query or order is given", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("cluster-1"))

		res, err := ListClusters(context.Background(), db, newValidOpts())
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("panics on an invalid partner/org/project UUID", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		assert.Panics(t, func() {
			_, _ = ListClusters(context.Background(), db, commonv3.QueryOptions{})
		})
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := ListClusters(context.Background(), db, newValidOpts())
		assert.Error(t, err)
	})
}

func TestGetClusterForToken(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("cluster-1"))

		res, err := GetClusterForToken(context.Background(), db, "tok-1")
		require.NoError(t, err)
		assert.Equal(t, "cluster-1", res.Name)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetClusterForToken(context.Background(), db, "missing")
		assert.Error(t, err)
	})
}

func TestNotify(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))

		err := Notify(db, "mychannel", "myvalue")
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := Notify(db, "mychannel", "myvalue")
		assert.Error(t, err)
	})
}
