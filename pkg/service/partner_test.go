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

func TestPartnerUpsert(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewPartnerService(db, getLogger())

	// Create a partner for testing
	partner := &systemv3.Partner{
		Metadata: &v3.Metadata{
			Name:        "partner-test",
			Description: "Test Partner Description",
		},
		Spec: &systemv3.PartnerSpec{
			Host:              "test.host.com",
			Domain:            "test.domain.com",
			TosLink:           "https://tos.link",
			LogoLink:          "https://logo.link",
			NotificationEmail: "notify@test.com",
			HelpdeskEmail:     "help@test.com",
			ProductName:       "Test Product",
			SupportTeamName:   "Test Support",
			OpsHost:           "ops.test.com",
			FavIconLink:       "https://favicon.link",
			IsTOTPEnabled:     true,
		},
	}

	// This regex should match the actual INSERT query with ON CONFLICT clause
	mock.ExpectQuery(`INSERT INTO "authsrv_partner"`).
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

	// Test insert
	result, err := ps.Upsert(context.Background(), partner)
	if err != nil {
		t.Fatal("could not upsert partner:", err)
	}

	// Verify result
	if result.GetMetadata().GetName() != partner.GetMetadata().GetName() {
		t.Error("invalid name returned")
	}
	if result.GetSpec().GetHost() != partner.GetSpec().GetHost() {
		t.Error("invalid host returned")
	}

	// Test update of existing partner
	updatedPartner := &systemv3.Partner{
		Metadata: &v3.Metadata{
			Name:        "partner-test", // Same name to trigger update
			Description: "Updated Description",
		},
		Spec: &systemv3.PartnerSpec{
			Host:              "updated.host.com",
			Domain:            "updated.domain.com",
			TosLink:           "https://updated-tos.link",
			LogoLink:          "https://updated-logo.link",
			NotificationEmail: "updated-notify@test.com",
			HelpdeskEmail:     "updated-help@test.com",
			ProductName:       "Updated Product",
			SupportTeamName:   "Updated Support",
			OpsHost:           "updated-ops.test.com",
			FavIconLink:       "https://updated-favicon.link",
			IsTOTPEnabled:     false,
		},
	}

	// For the update test, use a similar expectation
	mock.ExpectQuery(`INSERT INTO "authsrv_partner"`).
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

	// Test update via upsert
	result, err = ps.Upsert(context.Background(), updatedPartner)
	if err != nil {
		t.Fatal("could not upsert (update) partner:", err)
	}

	// Verify update result
	if result.GetSpec().GetHost() != updatedPartner.GetSpec().GetHost() {
		t.Error("invalid host returned after update")
	}
	if result.GetSpec().GetDomain() != updatedPartner.GetSpec().GetDomain() {
		t.Error("invalid domain returned after update")
	}

	// Test upsert failure
	mock.ExpectQuery(`INSERT INTO "authsrv_partner"`).
		WithArgs().
		WillReturnError(fmt.Errorf("database error"))

	_, err = ps.Upsert(context.Background(), partner)
	if err == nil {
		t.Error("expected error on upsert failure, got nil")
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
