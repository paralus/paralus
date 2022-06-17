package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetProjectNamespaces(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	puuid := uuid.New()
	mock.ExpectQuery(`SELECT "projectaccountnamespacerole"."id", "projectaccountnamespacerole"."name", "projectaccountnamespacerole"."description", "projectaccountnamespacerole"."created_at", "projectaccountnamespacerole"."modified_at", "projectaccountnamespacerole"."trash", "projectaccountnamespacerole"."organization_id", "projectaccountnamespacerole"."partner_id", "projectaccountnamespacerole"."role_id", "projectaccountnamespacerole"."account_id", "projectaccountnamespacerole"."project_id", "projectaccountnamespacerole"."namespace", "projectaccountnamespacerole"."active" FROM "authsrv_projectaccountnamespacerole" AS "projectaccountnamespacerole" WHERE \(project_id = '` + puuid.String() + `'\) AND \(trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"namespace"}).AddRow("namespace1"))
	mock.ExpectQuery(`SELECT "projectgroupnamespacerole"."id", "projectgroupnamespacerole"."name", "projectgroupnamespacerole"."description", "projectgroupnamespacerole"."created_at", "projectgroupnamespacerole"."modified_at", "projectgroupnamespacerole"."trash", "projectgroupnamespacerole"."organization_id", "projectgroupnamespacerole"."partner_id", "projectgroupnamespacerole"."role_id", "projectgroupnamespacerole"."group_id", "projectgroupnamespacerole"."project_id", "projectgroupnamespacerole"."namespace", "projectgroupnamespacerole"."active" FROM "authsrv_projectgroupnamespacerole" AS "projectgroupnamespacerole" WHERE \(project_id = '` + puuid.String() + `'\) AND \(trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"namespace"}).AddRow("namespace2"))

	ns := NewNamespaceService(db)
	nl, err := ns.GetProjectNamespaces(context.Background(), puuid)
	if err != nil {
		t.Fatal("unable to get namespaces", err)
	}
	if len(nl) != 2 {
		t.Errorf("incorrect number of namespaces; expected '%v', got '%v'", 2, len(nl))
	}
	if nl[0] != "namespace1" {
		t.Errorf("incorrect namespace name; expected '%v', got '%v'", "namespace1", nl[0])
	}
	if nl[1] != "namespace2" {
		t.Errorf("incorrect namespace name; expected '%v', got '%v'", "namespace2", nl[1])
	}
}

func TestGetAccountProjectNamespaces(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	puuid := uuid.New()
	uuuid := uuid.New()
	mock.ExpectQuery(`SELECT "projectaccountnamespacerole"."id", "projectaccountnamespacerole"."name", "projectaccountnamespacerole"."description", "projectaccountnamespacerole"."created_at", "projectaccountnamespacerole"."modified_at", "projectaccountnamespacerole"."trash", "projectaccountnamespacerole"."organization_id", "projectaccountnamespacerole"."partner_id", "projectaccountnamespacerole"."role_id", "projectaccountnamespacerole"."account_id", "projectaccountnamespacerole"."project_id", "projectaccountnamespacerole"."namespace", "projectaccountnamespacerole"."active" FROM "authsrv_projectaccountnamespacerole" AS "projectaccountnamespacerole" WHERE \(project_id = '` + puuid.String() + `'\) AND \(account_id = '` + uuuid.String() + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"namespace"}).AddRow("namespace1"))

	ns := NewNamespaceService(db)
	nl, err := ns.GetAccountProjectNamespaces(context.Background(), puuid, uuuid)
	if err != nil {
		t.Fatal("unable to get namespaces", err)
	}
	if len(nl) != 1 {
		t.Errorf("incorrect number of namespaces; expected '%v', got '%v'", 1, len(nl))
	}
	if nl[0] != "namespace1" {
		t.Errorf("incorrect namespace name; expected '%v', got '%v'", "namespace1", nl[0])
	}
}

func TestGetGroupProjectNamespaces(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	puuid := uuid.New()
	uuuid := uuid.New()
	mock.ExpectQuery(`SELECT "projectgroupnamespacerole"."id", "projectgroupnamespacerole"."name", "projectgroupnamespacerole"."description", "projectgroupnamespacerole"."created_at", "projectgroupnamespacerole"."modified_at", "projectgroupnamespacerole"."trash", "projectgroupnamespacerole"."organization_id", "projectgroupnamespacerole"."partner_id", "projectgroupnamespacerole"."role_id", "projectgroupnamespacerole"."group_id", "projectgroupnamespacerole"."project_id", "projectgroupnamespacerole"."namespace", "projectgroupnamespacerole"."active" FROM "authsrv_projectgroupnamespacerole" AS "projectgroupnamespacerole" JOIN authsrv_groupaccount ON projectgroupnamespacerole.group_id=authsrv_groupaccount.group_id WHERE \(project_id = '+` + puuid.String() + `+'\) AND \(authsrv_groupaccount.account_id = '` + uuuid.String() + `'\) AND \(projectgroupnamespacerole.trash = FALSE\) AND \(authsrv_groupaccount.trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"namespace"}).AddRow("namespace1"))

	ns := NewNamespaceService(db)
	nl, err := ns.GetGroupProjectNamespaces(context.Background(), puuid, uuuid)
	if err != nil {
		t.Fatal("unable to get namespaces", err)
	}
	if len(nl) != 1 {
		t.Errorf("incorrect number of namespaces; expected '%v', got '%v'", 1, len(nl))
	}
	if nl[0] != "namespace1" {
		t.Errorf("incorrect namespace name; expected '%v', got '%v'", "namespace1", nl[0])
	}
}
