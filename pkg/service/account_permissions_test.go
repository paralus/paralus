package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetAccountPermissions(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewAccountPermissionService(db)

	aid := uuid.New().String()
	oid := uuid.New().String()
	pid := uuid.New().String()

	mock.ExpectQuery(`SELECT "sap"."account_id", "sap"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(aid))

	_, err := ps.GetAccountPermissions(context.Background(), aid, oid, pid)
	if err != nil {
		t.Fatal("could not get GetAccountPermissions:", err)
	}
}

func TestIsPartnerSuperAdmin(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewAccountPermissionService(db)

	aid := uuid.New().String()
	pid := uuid.New().String()

	mock.ExpectQuery(`SELECT "sap"."account_id", "sap"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(aid))

	_, _, err := ps.IsPartnerSuperAdmin(context.Background(), aid, pid)
	if err != nil {
		t.Fatal("could not get IsPartnerSuperAdmin:", err)
	}
}

func TestGetAccountProjectsByPermission(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewAccountPermissionService(db)

	aid := uuid.New().String()
	oid := uuid.New().String()
	pid := uuid.New().String()

	mock.ExpectQuery(`SELECT "sap"."account_id", "sap"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(aid))

	_, err := ps.GetAccountProjectsByPermission(context.Background(), aid, oid, pid, "read")
	if err != nil {
		t.Fatal("could not get GetAccountProjectsByPermission:", err)
	}
}

func TestGetAccountPermissionsByProjectIDPermissions(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewAccountPermissionService(db)

	projects := []string{"myproject"}
	permissions := []string{"read"}
	aid := uuid.New().String()
	oid := uuid.New().String()
	pid := uuid.New().String()

	mock.ExpectQuery(`SELECT "sap"."account_id", "sap"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(aid))

	_, err := ps.GetAccountPermissionsByProjectIDPermissions(context.Background(), aid, oid, pid, projects, permissions)
	if err != nil {
		t.Fatal("could not get GetAccountPermissionsByProjectIDPermissions:", err)
	}
}

func TestIsOrgAdmin(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewAccountPermissionService(db)

	aid := uuid.New().String()
	pid := uuid.New().String()

	mock.ExpectQuery(`SELECT "sap"."account_id", "sap"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role_name"}).AddRow("ADMIN"))

	_, err := ps.IsOrgAdmin(context.Background(), aid, pid)
	if err != nil {
		t.Fatal("could not get IsOrgAdmin:", err)
	}
}

func TestIsAccountActive(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewAccountPermissionService(db)

	aid := uuid.New().String()
	oid := uuid.New().String()

	mock.ExpectQuery(`SELECT "group"."id", "group"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(aid))

	mock.ExpectQuery(`SELECT "groupaccount"."id", "groupaccount"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(aid))

	_, err := ps.IsAccountActive(context.Background(), aid, oid)
	if err != nil {
		t.Fatal("could not get IsAccountActive:", err)
	}
}

func TestGetAccount(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewAccountPermissionService(db)

	aid := uuid.New().String()

	mock.ExpectQuery(`SELECT "identities"."id", "identities"."traits"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(aid))

	_, err := ps.GetAccount(context.Background(), aid)
	if err != nil {
		t.Fatal("could not get account:", err)
	}
}

func TestGetAccountGroups(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewAccountPermissionService(db)

	aid := uuid.New().String()

	mock.ExpectQuery(`SELECT "groupaccount"."id", "groupaccount"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(aid))

	_, err := ps.GetAccountGroups(context.Background(), aid)
	if err != nil {
		t.Fatal("could not get account groups:", err)
	}
}
