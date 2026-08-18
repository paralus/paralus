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

func TestGetProjectNamespaces(t *testing.T) {
	t.Run("merges account and group namespaces", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"namespace"}).AddRow("ns-a"),
		)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"namespace"}).AddRow("ns-b"),
		)

		res, err := GetProjectNamespaces(context.Background(), db, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, []string{"ns-a", "ns-b"}, res)
	})

	t.Run("propagates error from first query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		res, err := GetProjectNamespaces(context.Background(), db, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("propagates error from second query", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"namespace"}))
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		res, err := GetProjectNamespaces(context.Background(), db, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestGetAccountProjectNamespaces(t *testing.T) {
	t.Run("returns namespaces", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"namespace"}).AddRow("ns-a"),
		)

		res, err := GetAccountProjectNamespaces(context.Background(), db, uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.Equal(t, []string{"ns-a"}, res)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetAccountProjectNamespaces(context.Background(), db, uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}

func TestGetGroupProjectNamespaces(t *testing.T) {
	t.Run("returns namespaces", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"namespace"}).AddRow("ns-g"),
		)

		res, err := GetGroupProjectNamespaces(context.Background(), db, uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.Equal(t, []string{"ns-g"}, res)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetGroupProjectNamespaces(context.Background(), db, uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}
