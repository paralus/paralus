package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
)

func performOrganizationBasicChecks(t *testing.T, organization *systemv3.Organization, puuid string) {
	if organization.GetMetadata().GetName() != "organization-"+puuid {
		t.Error("invalid name returned")
	}
}

func TestCreateOrganization(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewOrganizationService(db, getLogger())

	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))

	mock.ExpectQuery(`INSERT INTO "authsrv_organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))

	organization := &systemv3.Organization{
		Metadata: &v3.Metadata{Id: ouuid, Name: "organization-" + ouuid, Partner: "partname"},
		Spec:     &systemv3.OrganizationSpec{},
	}
	organization, err := ps.Create(context.Background(), organization)
	if err != nil {
		t.Fatal("could not create organization:", err)
	}
	performOrganizationBasicChecks(t, organization, ouuid)
}

func TestCreateOrganizationDuplicate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewOrganizationService(db, getLogger())

	ouuid := uuid.New().String()

	organization := &systemv3.Organization{
		Metadata: &v3.Metadata{Id: ouuid, Name: "organization-" + ouuid},
		Spec:     &systemv3.OrganizationSpec{},
	}

	// Try to recreate
	mock.ExpectQuery(`INSERT INTO "authsrv_organization"`).
		WithArgs().WillReturnError(fmt.Errorf("unique constraint violation"))
	_, err := gs.Create(context.Background(), organization)
	if err == nil {
		t.Fatal("should not be able to recreate project with same name")
	}
}

func TestOrganizationDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewOrganizationService(db, getLogger())

	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "organization"."id", "organization"."name", .* FROM "authsrv_organization" AS "organization" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ouuid, "organization-"+ouuid))

	mock.ExpectQuery(`UPDATE "authsrv_organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ouuid, "organization-"+ouuid))

	organization := &systemv3.Organization{
		Metadata: &v3.Metadata{Id: ouuid, Name: "organization-" + ouuid},
	}
	_, err := ps.Delete(context.Background(), organization)
	if err != nil {
		t.Fatal("could not delete organization:", err)
	}
}

func TestOrganizationDeleteNonExist(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewOrganizationService(db, getLogger())

	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "organization"."id", "organization"."name", .* FROM "authsrv_organization" AS "organization" WHERE`).
		WithArgs().WillReturnError(fmt.Errorf("no data available"))

	organization := &systemv3.Organization{
		Metadata: &v3.Metadata{Id: ouuid, Name: "organization-" + ouuid},
	}
	_, err := ps.Delete(context.Background(), organization)
	if err == nil {
		t.Fatal("deleted non existent organization")
	}
}

func TestOrganizationGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewOrganizationService(db, getLogger())

	partuuid := uuid.New().String()
	ouuid := uuid.New().String()
	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "organization"."id", "organization"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(partuuid, "partner-"+partuuid))

	organization := &systemv3.Organization{
		Metadata: &v3.Metadata{Id: puuid, Name: "organization-" + puuid},
	}
	_, err := ps.GetByName(context.Background(), organization.GetMetadata().Name)
	if err != nil {
		t.Fatal("could not get organization:", err)
	}
}

func TestOrganizationGetById(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewOrganizationService(db, getLogger())

	partuuid := uuid.New().String()
	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "organization"."id", "organization"."name", .* FROM "authsrv_organization" AS "organization" WHERE .*id = '` + puuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "partner_id"}).AddRow(puuid, "organization-"+puuid, partuuid))

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(partuuid, "partner-"+partuuid))

	organization := &systemv3.Organization{
		Metadata: &v3.Metadata{Id: puuid, Name: "organization-" + puuid},
	}
	organization, err := ps.GetByID(context.Background(), organization.Metadata.Id)
	if err != nil {
		t.Fatal("could not get organization:", err)
	}
	performOrganizationBasicChecks(t, organization, puuid)
}

func TestOrganizationUpdate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewOrganizationService(db, getLogger())

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "organization"."id", "organization"."name", .* FROM "authsrv_organization" AS "organization" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "organization-"+puuid))

	mock.ExpectExec(`UPDATE "authsrv_organization"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	organization := &systemv3.Organization{
		Metadata: &v3.Metadata{Id: puuid, Name: "organization-" + puuid},
		Spec:     &systemv3.OrganizationSpec{},
	}
	_, err := ps.Update(context.Background(), organization)
	if err != nil {
		t.Fatal("could not update organization:", err)
	}
}
