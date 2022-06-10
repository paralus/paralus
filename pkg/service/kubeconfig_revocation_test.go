package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetKubeconfigRevocation(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewKubeconfigRevocationService(db, getLogger())

	ouuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "kr"."id", "kr"."organization_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))

	_, err := ps.Get(context.Background(), ouuid, cuuid, false)
	if err != nil {
		t.Fatal("could not get kubeconfig revocation:", err)
	}
}
