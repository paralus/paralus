package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetKubeconfigSetting(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewKubeconfigSettingService(db)

	uuuid := uuid.New().String()
	ouuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "ks"."id", "ks"."organization_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuuid))

	_, err := ps.Get(context.Background(), ouuid, cuuid, true)
	if err != nil {
		t.Fatal("could not get Kubeconfig Setting:", err)
	}
}
