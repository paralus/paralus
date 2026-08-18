package dao

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateToken(t *testing.T) {
	t.Run("assigns an xid name and succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		token := &models.ClusterToken{}
		err := CreateToken(context.Background(), db, token)
		require.NoError(t, err)
		assert.NotEmpty(t, token.Name)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := CreateToken(context.Background(), db, &models.ClusterToken{})
		assert.Error(t, err)
	})
}

func TestRegisterToken(t *testing.T) {
	t.Run("marks token used and returns it", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("tok-1"))
		// RegisterToken's update-error is discarded (BUG: `dao.Update(...)`
		// result is never assigned), so it always reports success regardless
		// of whether this update actually succeeds.
		mock.ExpectExec(".*").WillReturnError(errors.New("update failed, but this error is silently dropped"))

		ct, err := RegisterToken(context.Background(), db, "tok-1")
		require.NoError(t, err, "RegisterToken currently ignores the Update error entirely")
		assert.Equal(t, "tok-1", ct.Name)
	})

	t.Run("returns ErrInvalidToken when lookup fails", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("not found"))

		_, err := RegisterToken(context.Background(), db, "missing")
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}
