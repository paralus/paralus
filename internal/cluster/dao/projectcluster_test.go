package dao

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProjectCluster(t *testing.T) {
	// ProjectCluster has no pk/default-valued columns, so unlike most models
	// in this codebase its INSERT has no RETURNING clause and runs as a
	// plain Exec, not a Query.
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := CreateProjectCluster(context.Background(), db, &models.ProjectCluster{ProjectID: uuid.New(), ClusterID: uuid.New()})
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := CreateProjectCluster(context.Background(), db, &models.ProjectCluster{})
		assert.Error(t, err)
	})
}

func TestGetProjectsForCluster(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"project_id"}).AddRow(uuid.New()))

		res, err := GetProjectsForCluster(context.Background(), db, uuid.New())
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetProjectsForCluster(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}

func TestDeleteProjectsForCluster(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := DeleteProjectsForCluster(context.Background(), db, uuid.New())
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := DeleteProjectsForCluster(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}

func TestValidateClusterAccess(t *testing.T) {
	t.Run("true when a matching cluster is found", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		ok, err := ValidateClusterAccess(context.Background(), db, commonv3.QueryOptions{Project: "proj-1"})
		require.NoError(t, err)
		assert.True(t, ok)
	})

	// ValidateClusterAccess scans into a single *models.Cluster (not a slice)
	// and never special-cases sql.ErrNoRows, so a no-match result surfaces
	// as an error rather than a clean (false, nil).
	t.Run("no match surfaces sql.ErrNoRows rather than a clean false", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}))

		ok, err := ValidateClusterAccess(context.Background(), db, commonv3.QueryOptions{Project: "proj-1"})
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.False(t, ok)
	})

	t.Run("invalid selector fails before querying", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		_, err := ValidateClusterAccess(context.Background(), db, commonv3.QueryOptions{Selector: "==="})
		assert.Error(t, err)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := ValidateClusterAccess(context.Background(), db, commonv3.QueryOptions{})
		assert.Error(t, err)
	})
}
