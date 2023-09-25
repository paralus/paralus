package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
)

func performBasicChecks(t *testing.T, metro *infrav3.Location, puuid string) {
	if metro.Metadata.Name != "metro-"+puuid {
		t.Error("invalid name returned")
	}
}

func TestCreateMetro(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewMetroService(db)

	puuid := uuid.New().String()
	muuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))

	mock.ExpectQuery(`INSERT INTO "cluster_metro"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(muuid))

	metro := &infrav3.Location{
		Metadata: &commonv3.Metadata{Id: muuid, Name: "metro-" + muuid},
		Spec:     &infrav3.Metro{},
	}
	metro, err := ps.Create(context.Background(), metro)
	if err != nil {
		t.Fatal("could not create metro:", err)
	}
	performBasicChecks(t, metro, muuid)
}

func TestCreateMetroDuplicate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewMetroService(db)

	muuid := uuid.New().String()

	metro := &infrav3.Location{
		Metadata: &commonv3.Metadata{Id: muuid, Name: "metro-" + muuid},
		Spec:     &infrav3.Metro{},
	}

	// Try to recreate
	mock.ExpectQuery(`INSERT INTO "cluster_metro"`).
		WithArgs().WillReturnError(fmt.Errorf("unique constraint violation"))
	_, err := gs.Create(context.Background(), metro)
	if err == nil {
		t.Fatal("should not be able to recreate metro with same name")
	}
}

func TestMetroDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewMetroService(db)

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "metro"."id", "metro"."name", .* FROM "cluster_metro" AS "metro" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "metro-"+puuid))

	mock.ExpectExec(`UPDATE "cluster_metro" AS "metro" SET trash = TRUE WHERE`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	metro := &infrav3.Location{
		Metadata: &commonv3.Metadata{Id: puuid, Name: "metro-" + puuid},
		Spec:     &infrav3.Metro{},
	}
	_, err := ps.Delete(context.Background(), metro)
	if err != nil {
		t.Fatal("could not delete metro:", err)
	}
}

func TestMetroDeleteNonExist(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewMetroService(db)

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "metro"."id", "metro"."name", .* FROM "cluster_metro" AS "metro" WHERE`).
		WithArgs().WillReturnError(fmt.Errorf("no data available"))

	metro := &infrav3.Location{
		Metadata: &commonv3.Metadata{Id: puuid, Name: "metro-" + puuid},
		Spec:     &infrav3.Metro{},
	}
	_, err := ps.Delete(context.Background(), metro)
	if err == nil {
		t.Fatal("deleted non existent metro")
	}
}

func TestMetroGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewMetroService(db)

	muuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "metro"."id", "metro"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(muuid))

	metro := &infrav3.Location{
		Metadata: &commonv3.Metadata{Id: muuid, Name: "metro-" + muuid},
		Spec:     &infrav3.Metro{},
	}
	_, err := ps.GetByName(context.Background(), metro.Metadata.Name)
	if err != nil {
		t.Fatal("could not get metro:", err)
	}
	performBasicChecks(t, metro, muuid)
}

func TestMetroUpdate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewMetroService(db)

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "metro"."id", "metro"."name", .* FROM "cluster_metro" AS "metro" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "metro-"+puuid))

	mock.ExpectExec(`UPDATE "cluster_metro"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	metro := &infrav3.Location{
		Metadata: &commonv3.Metadata{Id: puuid, Name: "metro-" + puuid},
		Spec:     &infrav3.Metro{},
	}
	_, err := ps.Update(context.Background(), metro)
	if err != nil {
		t.Fatal("could not update metro:", err)
	}
}
