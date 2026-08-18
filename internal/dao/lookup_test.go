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

func TestGetPartnerId(t *testing.T) {
	t.Run("returns id on success", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		want := uuid.New()
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(want))

		got, err := GetPartnerId(context.Background(), db, "acme")
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetPartnerId(context.Background(), db, "missing")
		assert.Error(t, err)
	})
}

func TestGetOrganizationId(t *testing.T) {
	db, mock := newMockBunDB(t)
	want := uuid.New()
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(want))

	got, err := GetOrganizationId(context.Background(), db, "org")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetProjectId(t *testing.T) {
	db, mock := newMockBunDB(t)
	want := uuid.New()
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(want))

	got, err := GetProjectId(context.Background(), db, "proj")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetPartnerName(t *testing.T) {
	t.Run("returns name on success", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("acme"))

		got, err := GetPartnerName(context.Background(), db, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, "acme", got)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetPartnerName(context.Background(), db, uuid.New())
		assert.Error(t, err)
	})
}

func TestGetOrganizationName(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("org-1"))

	got, err := GetOrganizationName(context.Background(), db, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "org-1", got)
}

func TestGetProjectName(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("proj-1"))

	got, err := GetProjectName(context.Background(), db, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "proj-1", got)
}
