package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/common"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	rolev3 "github.com/paralus/paralus/proto/types/rolepb/v3"
)

func performRoleBasicChecks(t *testing.T, role *rolev3.Role, ruuid string) {
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
}

func performRoleBasicAuthzChecks(t *testing.T, mazc mockAuthzClient, ruuid string) {
	if len(mazc.drpm) > 0 {
		if mazc.drpm[len(mazc.drpm)-1].Role != "role-"+ruuid {
			t.Errorf("incorrect role sent to authz; expected '%v' got '%v'", "role-"+ruuid, mazc.drpm[len(mazc.drpm)-1].Role)
		}
	}
	if len(mazc.crpm) > 0 {
		if len(mazc.crpm[len(mazc.crpm)-1].RolePermissionMappingList) != 1 {
			t.Errorf("invalid number of roles sent to authz; expected 1, got '%v'", len(mazc.crpm[len(mazc.crpm)-1].RolePermissionMappingList))
		}
		if mazc.crpm[len(mazc.crpm)-1].RolePermissionMappingList[0].Role != "role-"+ruuid {
			t.Errorf("incorrect role sent to authz; expected '%v' got '%v'", "role-"+ruuid, mazc.crpm[len(mazc.crpm)-1].RolePermissionMappingList[0].Role)
		}
		if len(mazc.crpm[len(mazc.crpm)-1].RolePermissionMappingList[0].Permission) != 1 {
			t.Errorf("incorrect number of permissions sent to authz; expected '1', got '%v'", len(mazc.crpm[len(mazc.crpm)-1].RolePermissionMappingList[0].Permission))
		}
		if mazc.crpm[len(mazc.crpm)-1].RolePermissionMappingList[0].Permission[0] != "ops_star.all" {
			t.Errorf("incorrect permissions sent to authz; expected 'ops_star.all', got '%v'", mazc.crpm[len(mazc.crpm)-1].RolePermissionMappingList[0].Permission[0])
		}
	}
}

func TestCreateRole(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'role-` + ruuid + `'.`).
		WillReturnError(fmt.Errorf("no data available"))

	mock.ExpectBegin()
	// TODO: more precise checks
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))
	mock.ExpectCommit()

	role := &rolev3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
		Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: "system"},
	}
	role, err := rs.Create(context.Background(), role)
	if err != nil {
		t.Fatal("could not create group:", err)
	}
	performRoleBasicChecks(t, role, ruuid)
}
func TestCreateRoleNoBuiltinOverride(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'role-` + ruuid + `'.`).
		WillReturnError(fmt.Errorf("no data available"))

	mock.ExpectBegin()
	// TODO: more precise checks
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerole".* TRUE, FALSE, 'system'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))
	mock.ExpectCommit()

	role := &rolev3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
		Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: "system", Builtin: true},
	}
	role, err := rs.Create(context.Background(), role)
	if err != nil {
		t.Fatal("could not create group:", err)
	}
	performRoleBasicChecks(t, role, ruuid)
}

func TestCreateRoleBuiltinOverride(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'role-` + ruuid + `'.`).
		WillReturnError(fmt.Errorf("no data available"))

	mock.ExpectBegin()
	// TODO: more precise checks
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerole".* TRUE, TRUE, 'system'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))
	mock.ExpectCommit()

	role := &rolev3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
		Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: "system", Builtin: true},
	}

	internalCtx := context.WithValue(context.Background(), common.SessionInternalKey, true)
	role, err := rs.Create(internalCtx, role)
	if err != nil {
		t.Fatal("could not create group:", err)
	}
	performRoleBasicChecks(t, role, ruuid)
}

func TestCreateRoleWithOpsAllPermission(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'role-` + ruuid + `'.`).
		WillReturnError(fmt.Errorf("no data available"))

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))
	mock.ExpectQuery(`SELECT "resourcepermission"."id" FROM "authsrv_resourcepermission" AS "resourcepermission" WHERE .name = 'ops_star.all'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerolepermission"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectCommit()

	role := &rolev3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
		Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: "system", Rolepermissions: []string{"ops_star.all"}},
	}
	role, err := rs.Create(context.Background(), role)
	if err != nil {
		t.Fatal("could not create group:", err)
	}
	performRoleBasicChecks(t, role, ruuid)

}

func TestCreateRoleWithPermissions(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'role-` + ruuid + `'.`).
		WillReturnError(fmt.Errorf("no data available"))

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))
	mock.ExpectQuery(`SELECT "resourcepermission"."id" FROM "authsrv_resourcepermission" AS "resourcepermission" WHERE .name = 'partner.read'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectQuery(`SELECT "resourcepermission"."id" FROM "authsrv_resourcepermission" AS "resourcepermission" WHERE .name = 'organization.read'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerolepermission"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectCommit()

	role := &rolev3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
		Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: "system", Rolepermissions: []string{"partner.read", "organization.read"}},
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

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectQuery(` SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'role-` + ruuid + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))

	mock.ExpectBegin()
	// TODO: more precise checks
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))
	mock.ExpectCommit()

	role := &rolev3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
		Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: "system"},
	}
	_, err := rs.Create(context.Background(), role)
	if err == nil {
		t.Error("should not be able to recreate group with same name")
	}
}

func TestUpdateRole(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .*FROM "authsrv_resourcerole" AS "resourcerole" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'role-` + ruuid + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "organization_id", "partner_id"}).AddRow(ruuid, "role-"+ruuid, ouuid, puuid))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "authsrv_resourcerole" AS "resourcerole" SET "name" = 'role-` + ruuid + `', .*"organization_id" = '` + ouuid + `', "partner_id" = '` + puuid + `', "is_global" = TRUE, "builtin" = FALSE, "scope" = 'system' WHERE .id  = '` + ruuid + `'.`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "authsrv_resourcerolepermission" AS "resourcerolepermission" SET trash = TRUE WHERE ."resource_role_id" = '` + ruuid + `'.`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(`SELECT "resourcepermission"."id" FROM "authsrv_resourcepermission" AS "resourcepermission" WHERE .name = 'partner.read'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("partner.read"))
	mock.ExpectQuery(`SELECT "resourcepermission"."id" FROM "authsrv_resourcepermission" AS "resourcepermission" WHERE .name = 'organization.read'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("organization.read"))
	mock.ExpectQuery(`INSERT INTO "authsrv_resourcerolepermission"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruuid))
	mock.ExpectCommit()

	role := &rolev3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
		Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: "system", Rolepermissions: []string{"partner.read", "organization.read"}},
	}
	role, err := rs.Update(context.Background(), role)
	if err != nil {
		t.Fatal("could not update role:", err)
	}
	performRoleBasicChecks(t, role, ruuid)
}

func TestUpdateRoleBuiltin(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()
	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .*FROM "authsrv_resourcerole" AS "resourcerole" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'role-` + ruuid + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "organization_id", "partner_id", "builtin"}).AddRow(ruuid, "role-"+ruuid, ouuid, puuid, true))

	role := &rolev3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
		Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: "system", Rolepermissions: []string{"ops_star.all"}},
	}
	_, err := rs.Update(context.Background(), role)
	if err == nil {
		t.Fatal("builtin roles should not be updated")
	}
}

func TestRoleDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .* FROM "authsrv_resourcerole" AS "resourcerole" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ruuid, "role-"+ruuid))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "authsrv_resourcerolepermission" AS "resourcerolepermission" SET trash = TRUE WHERE ."resource_role_id" = '` + ruuid + `'.`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "authsrv_resourcerole" AS "resourcerole" SET trash = TRUE WHERE .id = '` + ruuid + `'.`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	role := &rolev3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
	}
	_, err := rs.Delete(context.Background(), role)
	if err != nil {
		t.Error("could not delete role:", err)
	}
}

func TestRoleDeleteNonExist(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .* FROM "authsrv_resourcerole" AS "resourcerole" WHERE`).
		WithArgs().WillReturnError(fmt.Errorf("no data available"))

	role := &rolev3.Role{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "role-" + ruuid},
	}
	_, err := rs.Delete(context.Background(), role)
	if err == nil {
		t.Error("deleted non existent role")
	}

}

func TestRoleGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

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

	role := &rolev3.Role{
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

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid := uuid.New().String()
	rruuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .* FROM "authsrv_resourcerole" AS "resourcerole" WHERE .*id = '` + ruuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ruuid, "role-"+ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcepermission.name as name FROM "authsrv_resourcepermission" JOIN authsrv_resourcerolepermission ON authsrv_resourcerolepermission.resource_permission_id=authsrv_resourcepermission.id WHERE .authsrv_resourcerolepermission.resource_role_id = '` + ruuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(rruuid, "resourcerole-"+rruuid))

	role := &rolev3.Role{
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

	mazc := mockAuthzClient{}
	rs := NewRoleService(db, &mazc, getLogger())

	ruuid1 := uuid.New().String()
	ruuid2 := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", .*FROM "authsrv_resourcerole" AS "resourcerole" WHERE .partner_id = '` + puuid + `'. AND .organization_id = '` + ouuid + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(ruuid1, "role-"+ruuid1).AddRow(ruuid2, "role-"+ruuid2))
	mock.ExpectQuery(`SELECT authsrv_resourcepermission.name as name FROM "authsrv_resourcepermission" JOIN authsrv_resourcerolepermission ON authsrv_resourcerolepermission.resource_permission_id=authsrv_resourcepermission.id WHERE .authsrv_resourcerolepermission.resource_role_id = '` + ruuid1 + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("ops_star.all"))
	mock.ExpectQuery(`SELECT authsrv_resourcepermission.name as name FROM "authsrv_resourcepermission" JOIN authsrv_resourcerolepermission ON authsrv_resourcerolepermission.resource_permission_id=authsrv_resourcepermission.id WHERE .authsrv_resourcerolepermission.resource_role_id = '` + ruuid2 + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("ops_star.all"))

	role := &rolev3.Role{
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
