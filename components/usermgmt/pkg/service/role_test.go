package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
	"github.com/google/uuid"
)

func performRoleBasicChecks(t *testing.T, role *userv3.Role, ruuid string) {
	_, err := uuid.Parse(role.GetMetadata().GetOrganization())
	if err == nil {
		t.Error("org in metadata should be name not id")
	}
	_, err = uuid.Parse(role.GetMetadata().GetPartner())
	if err == nil {
		t.Error("partner in metadata should be name not id")
	}
	if role.GetMetadata().GetName() != "role-"+ruuid {
		t.Error("invalid name returned")
	}
	if role.Status.ConditionStatus != v3.ConditionStatus_StatusOK {
		t.Error("group status is not OK")
	}
}

func TestCreateRole(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	rs := NewRoleService(db)
	defer rs.Close()

	ruuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole"`).WithArgs()
	// TODO: more precise checks
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))

	role := &userv3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
		Spec:     &userv3.RoleSpec{IsGlobal: true, Scope: "cluster"},
	}
	role, err := rs.Create(context.Background(), role)
	if err != nil {
		t.Fatal("could not create group:", err)
	}
	performRoleBasicChecks(t, role, ruuid)
}

func TestCreateRoleDuplicate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	rs := NewRoleService(db)
	defer rs.Close()

	ruuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole"`).WithArgs()
	// TODO: more precise checks
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))

	role := &userv3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
		Spec:     &userv3.RoleSpec{IsGlobal: true, Scope: "cluster"},
	}
	role, err := rs.Create(context.Background(), role)
	if err != nil {
		t.Fatal("could not create group:", err)
	}
	performRoleBasicChecks(t, role, ruuid)

	mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))
	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	// TODO: more precise checks
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))

	_, err = rs.Create(context.Background(), role)
	if err == nil {
		t.Fatal("should not be able to recreate group with same name")
	}
}

func TestRoleDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	rs := NewRoleService(db)
	defer rs.Close()

	ruuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .* FROM "authsrv_resourcerole" AS "resourcerole" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ruuid, "role-"+ruuid))
	mock.ExpectExec(`DELETE FROM "authsrv_resourcerolepermission" AS "resourcerolepermission" WHERE ."resource_role_id" = '` + ruuid + `'.`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "authsrv_resourcerole" AS "resourcerole" WHERE .id = '` + ruuid + `'.`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	role := &userv3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
	}
	_, err := rs.Delete(context.Background(), role)
	if err != nil {
		t.Fatal("could not delete role:", err)
	}
}

func TestRoleDeleteNonExist(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	rs := NewRoleService(db)
	defer rs.Close()

	ruuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .* FROM "authsrv_resourcerole" AS "resourcerole" WHERE`).
		WithArgs().WillReturnError(fmt.Errorf("No data available"))

	role := &userv3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
	}
	_, err := rs.Delete(context.Background(), role)
	if err == nil {
		t.Fatal("deleted non existant role")
	}
}

func TestRoleGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	rs := NewRoleService(db)
	defer rs.Close()

	ruuid := uuid.New().String()
	rruuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .* FROM "authsrv_resourcerole" AS "resourcerole" WHERE .*name = 'role-` + ruuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ruuid, "role-"+ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcepermission.name as name FROM "authsrv_resourcepermission" JOIN authsrv_resourcerolepermission ON authsrv_resourcerolepermission.resource_permission_id=authsrv_resourcepermission.id WHERE .authsrv_resourcerolepermission.resource_role_id = '` + ruuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(rruuid, "resourcerole-"+rruuid))

	role := &userv3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
	}
	role, err := rs.GetByName(context.Background(), role)
	if err != nil {
		t.Fatal("could not get role:", err)
	}
	performRoleBasicChecks(t, role, ruuid)
}

func TestRoleGetById(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	rs := NewRoleService(db)
	defer rs.Close()

	ruuid := uuid.New().String()
	rruuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .* FROM "authsrv_resourcerole" AS "resourcerole" WHERE .*id = '` + ruuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ruuid, "role-"+ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcepermission.name as name FROM "authsrv_resourcepermission" JOIN authsrv_resourcerolepermission ON authsrv_resourcerolepermission.resource_permission_id=authsrv_resourcepermission.id WHERE .authsrv_resourcerolepermission.resource_role_id = '` + ruuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(rruuid, "resourcerole-"+rruuid))

	role := &userv3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Id: ruuid},
	}
	role, err := rs.GetByID(context.Background(), role)
	if err != nil {
		t.Fatal("could not get role:", err)
	}
	performRoleBasicChecks(t, role, ruuid)
}

func TestRoleList(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	rs := NewRoleService(db)
	defer rs.Close()

	ruuid1 := uuid.New().String()
	ruuid2 := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .* FROM "authsrv_resourcerole" AS "resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(ruuid1, "role-"+ruuid1).AddRow(ruuid2, "role-"+ruuid2))

	role := &userv3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid},
	}
	rolelist, err := rs.List(context.Background(), role)
	if err != nil {
		t.Fatal("could not list roles:", err)
	}
	if rolelist.Metadata.Count != 2 {
		t.Errorf("incorrect number of roles returned, expected 2; got %v", rolelist.Metadata.Count)
	}
	if rolelist.Items[0].Metadata.Name != "role-"+ruuid1 || rolelist.Items[1].Metadata.Name != "role-"+ruuid2 {
		t.Errorf("incorrect role names returned when listing")
	}
}
