package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	bun "github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

func getDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal("unable to create sqlmock:", err)
	}
	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.FromEnv("BUNDEBUG"),
	))
	return db, mock
}

func performPartnerBasicChecks(t *testing.T, partner *systemv3.Partner, puuid string) {
	if partner.GetMetadata().GetName() != "partner-"+puuid {
		t.Error("invalid name returned")
	}
}

func TestCreatePartner(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewPartnerService(db, getLogger())

	puuid := uuid.New().String()

	mock.ExpectQuery(`INSERT INTO "authsrv_partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))

	partner := &systemv3.Partner{
		Metadata: &v3.Metadata{Id: puuid, Name: "partner-" + puuid},
		Spec:     &systemv3.PartnerSpec{},
	}
	partner, err := ps.Create(context.Background(), partner)
	if err != nil {
		t.Fatal("could not create partner:", err)
	}
	performPartnerBasicChecks(t, partner, puuid)
}

func TestCreatePartnerDuplicate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewPartnerService(db, getLogger())

	puuid := uuid.New().String()

	partner := &systemv3.Partner{
		Metadata: &v3.Metadata{Id: puuid, Name: "partner-" + puuid},
		Spec:     &systemv3.PartnerSpec{},
	}

	// Try to recreate
	mock.ExpectQuery(`INSERT INTO "authsrv_partner"`).
		WithArgs().WillReturnError(fmt.Errorf("unique constraint violation"))
	_, err := gs.Create(context.Background(), partner)
	if err == nil {
		t.Fatal("should not be able to recreate partner with same name")
	}
}

func TestPartnerDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewPartnerService(db, getLogger())

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name", .* FROM "authsrv_partner" AS "partner" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "partner-"+puuid))

	mock.ExpectExec(`UPDATE "authsrv_partner"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	partner := &systemv3.Partner{
		Metadata: &v3.Metadata{Id: puuid, Name: "partner-" + puuid},
	}
	_, err := ps.Delete(context.Background(), partner)
	if err != nil {
		t.Fatal("could not delete partner:", err)
	}
}

func TestPartnerDeleteNonExist(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewPartnerService(db, getLogger())

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name", .* FROM "authsrv_partner" AS "partner" WHERE`).
		WithArgs().WillReturnError(fmt.Errorf("no data available"))

	partner := &systemv3.Partner{
		Metadata: &v3.Metadata{Id: puuid, Name: "partner-" + puuid},
	}
	_, err := gs.Delete(context.Background(), partner)
	if err == nil {
		t.Fatal("deleted non existent group")
	}
}

func TestPartnerGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewPartnerService(db, getLogger())

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))

	partner := &systemv3.Partner{
		Metadata: &v3.Metadata{Id: puuid, Name: "partner-" + puuid},
	}
	_, err := ps.GetByName(context.Background(), partner.GetMetadata().Name)
	if err != nil {
		t.Fatal("could not get partner:", err)
	}
}

func TestPartnerGetById(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewPartnerService(db, getLogger())

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name", .* FROM "authsrv_partner" AS "partner" WHERE .*id = '` + puuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "partner-"+puuid))

	partner := &systemv3.Partner{
		Metadata: &v3.Metadata{Id: puuid, Name: "partner-" + puuid},
	}
	partner, err := ps.GetByID(context.Background(), partner.Metadata.Id)
	if err != nil {
		t.Fatal("could not get partner:", err)
	}
	performPartnerBasicChecks(t, partner, puuid)
}

func TestPartnerUpdate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewPartnerService(db, getLogger())

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name", .* FROM "authsrv_partner" AS "partner" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "partner-"+puuid))

	mock.ExpectExec(`UPDATE "authsrv_partner"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	partner := &systemv3.Partner{
		Metadata: &v3.Metadata{Id: puuid, Name: "partner-" + puuid},
	}
	_, err := ps.Update(context.Background(), partner)
	if err != nil {
		t.Fatal("could not update partner:", err)
	}
}
