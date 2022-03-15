package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetKubectlSetting(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewkubectlClusterSettingsService(db)

	ouuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "kc"."name", "kc"."organization_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("kcs-" + ouuid))

	_, err := ps.Get(context.Background(), ouuid, cuuid)
	if err != nil {
		t.Fatal("could not get KubectlSetting:", err)
	}
}
