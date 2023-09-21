package service

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/audit"
	"go.uber.org/zap"
)

func getLogger() *zap.Logger {
	ao := audit.AuditOptions{
		LogPath:    "/dev/stdout",
		MaxSizeMB:  1,
		MaxBackups: 10, // Should we let sidecar do rotation?
		MaxAgeDays: 10, // Make these configurable via env
	}
	return audit.GetAuditLogger(&ao)
}

func performBasicAuthzChecks(t *testing.T, mazc mockAuthzClient, cpCount, dpCount, cugCount, dugCount, crpmCount, drpmCount int) {
	if len(mazc.cp) != cpCount {
		t.Errorf("unexpected number of calls to Authz CreatePolicies; expctex '%v', got '%v'", cpCount, len(mazc.cp))
	}
	if len(mazc.dp) != dpCount {
		t.Errorf("unexpected number of calls to Authz DeletePolicies; expctex '%v', got '%v'", dpCount, len(mazc.dp))
	}
	if len(mazc.cug) != cugCount {
		t.Errorf("unexpected number of calls to Authz CreateUserGroups; expctex '%v', got '%v'", cugCount, len(mazc.cug))
	}
	if len(mazc.dug) != dugCount {
		t.Errorf("unexpected number of calls to Authz DeleteUserGroups; expctex '%v', got '%v'", dugCount, len(mazc.dug))
	}
	if len(mazc.crpm) != crpmCount {
		t.Errorf("unexpected number of calls to Authz CreateRolePermissionMapping; expctex '%v', got '%v'", crpmCount, len(mazc.crpm))
	}
	if len(mazc.drpm) != drpmCount {
		t.Errorf("unexpected number of calls to Authz DeleteRolePermissionMapping; expctex '%v', got '%v'", drpmCount, len(mazc.drpm))
	}
}

func performBasicAuthProviderChecks(t *testing.T, ma mockAuthProvider, cCount, uCount, rCount, dCount int) {
	if len(ma.c) != cCount {
		t.Errorf("unexpected number of calls to Auth Provider Create; expctex '%v', got '%v'", cCount, len(ma.c))
	}
	if len(ma.u) != uCount {
		t.Errorf("unexpected number of calls to Auth Provider Update; expctex '%v', got '%v'", uCount, len(ma.u))
	}
	if len(ma.r) != rCount {
		t.Errorf("unexpected number of calls to Auth Provider GetRecoveryLink; expctex '%v', got '%v'", rCount, len(ma.r))
	}
	if len(ma.d) != dCount {
		t.Errorf("unexpected number of calls to Auth Provider Delete; expctex '%v', got '%v'", dCount, len(ma.d))
	}
}

func idname(uid string, resource string) string {
	return resource + "-" + uid
}

func idnamea(uid string, resource string) *string {
	name := resource + "-" + uid
	return &name
}

func addFetchEmptyExpecteation(mock sqlmock.Sqlmock, resource string) {
	mock.ExpectQuery(`SELECT "` + resource + `"."id" FROM "authsrv_` + resource + `" AS "` + resource + `"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}))
}

func addFailingFetchExpecteation(mock sqlmock.Sqlmock, resource string) {
	mock.ExpectQuery(`SELECT "` + resource + `"."id" FROM "authsrv_` + resource + `" AS "` + resource + `"`).
		WithArgs().WillReturnError(fmt.Errorf("no data available"))
}

func addFetchIdExpectation(mock sqlmock.Sqlmock, resource string) string {
	uid := uuid.New().String()
	mock.ExpectQuery(`SELECT "` + resource + `"."id" FROM "authsrv_` + resource + `" AS "` + resource + `"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uid))
	return uid
}

func addFetchByIdExpectation(mock sqlmock.Sqlmock, resource, uid string) {
	mock.ExpectQuery(`SELECT "` + resource + `"."id".* FROM "authsrv_` + resource + `" AS "` + resource + `" WHERE .id = '` + uid + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uid, resource+"-"+uid))
}

func addFetchIdByNameExpectation(mock sqlmock.Sqlmock, resource, name string) string {
	uid := uuid.NewString()
	mock.ExpectQuery(`SELECT "` + resource + `"."id" FROM "authsrv_` + resource + `" AS "` + resource + `" WHERE .name = '` + name + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uid))
	return uid
}

func addFetchExpectation(mock sqlmock.Sqlmock, resource string) string {
	uid := uuid.New().String()
	mock.ExpectQuery(`SELECT "` + resource + `"."id".* FROM "authsrv_` + resource + `" AS "` + resource + `"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uid, resource+"-name"))
	return uid
}

func addFetchByNameExpectation(mock sqlmock.Sqlmock, resource string, uid string) {
	mock.ExpectQuery(`SELECT "` + resource + `"."id", "` + resource + `"."name", .* FROM "authsrv_` + resource + `" AS "` + resource + `" WHERE .*name = '` + resource + `-` + uid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uid, resource+"-"+uid))
}

func addDeleteExpectation(mock sqlmock.Sqlmock, resource string, uid string) {
	mock.ExpectExec(`UPDATE "authsrv_` + resource + `" AS "` + resource + `" SET trash = TRUE WHERE .id = '` + uid).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func addUnavailableExpectation(mock sqlmock.Sqlmock, resource, partner, org, uid string) {
	mock.ExpectQuery(`SELECT "` + resource + `"."id" FROM "authsrv_` + resource + `" AS "` + resource + `" WHERE .organization_id = '` + org + `'. AND .partner_id = '` + partner + `'. AND .name = '` + resource + `-` + uid + `'.`).
		WillReturnError(fmt.Errorf("no data available"))
}

func addParterOrgFetchExpectation(mock sqlmock.Sqlmock) (string, string) {
	pid := uuid.New().String()
	oid := uuid.New().String()
	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(oid))
	return pid, oid
}

func addOrgParterFetchExpectation(mock sqlmock.Sqlmock) (string, string) {
	pid := uuid.New().String()
	oid := uuid.New().String()
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(oid))
	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pid))
	return pid, oid
}

func addResourceRoleFetchExpectation(mock sqlmock.Sqlmock, scope string) string {
	uid := uuid.New().String()
	mock.ExpectQuery(`SELECT "resourcerole"."id".* FROM "authsrv_resourcerole" AS "resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "scope"}).AddRow(uid, "role-"+uid, scope))
	return uid
}

func addGroupUserMappingsUpdateExpectation(mock sqlmock.Sqlmock, group string) {
	mock.ExpectQuery(`UPDATE "authsrv_groupaccount" AS "groupaccount" SET trash = TRUE WHERE ."group_id" = '` + group + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
}

func addUserGroupMappingsUpdateExpectation(mock sqlmock.Sqlmock, account string) {
	mock.ExpectQuery(`UPDATE "authsrv_groupaccount" AS "groupaccount" SET trash = TRUE WHERE ."account_id" = '` + account + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
}

func addUserIdFetchExpectation(mock sqlmock.Sqlmock) string {
	uid := uuid.New().String()
	mock.ExpectQuery(`SELECT "identities"."id" FROM "identities" WHERE .*traits ->> 'email' = 'user-` + uid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uid, []byte(`{"email":"user-`+uid+`", "first_name": "John", "last_name": "Doe", "description": "The OG user."}`)))
	return uid
}

func addUserFetchExpectation(mock sqlmock.Sqlmock) string {
	uid := uuid.New().String()
	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" WHERE .*traits ->> 'email' = 'user-` + uid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uid, []byte(`{"email":"user-`+uid+`", "first_name": "John", "last_name": "Doe", "description": "The OG user."}`)))
	return uid
}

func addUserFullFetchExpectation(mock sqlmock.Sqlmock) string {
	uid := uuid.New().String()
	mock.ExpectQuery(`SELECT "identities"."id", "identities"."schema_id", "identities"."traits", "identities"."created_at", "identities"."updated_at", "identities"."state", "identities"."state_changed_at", "identities"."nid", "identities"."metadata_public", "identity_credential"."id" AS "identity_credential__id", "identity_credential"."identity_id" AS "identity_credential__identity_id", "identity_credential"."identity_credential_type_id" AS "identity_credential__identity_credential_type_id", "identity_credential__identity_credential_type"."id" AS "identity_credential__identity_credential_type__id", "identity_credential__identity_credential_type"."name" AS "identity_credential__identity_credential_type__name" FROM "identities" LEFT JOIN "identity_credentials" AS "identity_credential" ON ."identity_credential"."identity_id" = "identities"."id". LEFT JOIN "identity_credential_types" AS "identity_credential__identity_credential_type" ON ."identity_credential__identity_credential_type"."id" = "identity_credential"."identity_credential_type_id". WHERE .traits ->> 'email' = 'user-` + uid + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits", "identity_credential__identity_credential_type__name"}).AddRow(uid, []byte(`{"email":"user-`+uid+`", "first_name": "John", "last_name": "Doe", "description": "The OG user."}`), "password"))
	return uid
}

func addUserFullFetchExpectationWithIdpGroups(mock sqlmock.Sqlmock) string {
	uid := uuid.New().String()
	mock.ExpectQuery(`SELECT "identities"."id", "identities"."schema_id", "identities"."traits", "identities"."created_at", "identities"."updated_at", "identities"."state", "identities"."state_changed_at", "identities"."nid", "identities"."metadata_public", "identity_credential"."id" AS "identity_credential__id", "identity_credential"."identity_id" AS "identity_credential__identity_id", "identity_credential"."identity_credential_type_id" AS "identity_credential__identity_credential_type_id", "identity_credential__identity_credential_type"."id" AS "identity_credential__identity_credential_type__id", "identity_credential__identity_credential_type"."name" AS "identity_credential__identity_credential_type__name" FROM "identities" LEFT JOIN "identity_credentials" AS "identity_credential" ON ."identity_credential"."identity_id" = "identities"."id". LEFT JOIN "identity_credential_types" AS "identity_credential__identity_credential_type" ON ."identity_credential__identity_credential_type"."id" = "identity_credential"."identity_credential_type_id". WHERE .traits ->> 'email' = 'user-` + uid + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits", "identity_credential__identity_credential_type__name"}).AddRow(uid, []byte(`{"email":"user-`+uid+`", "first_name": "John", "last_name": "Doe", "description": "The OG user.", "idp_groups": ["BigShot"]}`), "password"))
	return uid
}

func addUsersGroupFetchExpectation(mock sqlmock.Sqlmock, user string) string {
	uid := uuid.New().String()
	mock.ExpectQuery(`SELECT "group"."id".* FROM "authsrv_group" AS "group" JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id WHERE .authsrv_groupaccount.account_id = '` + user + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("group-" + uid))
	return uid
}

func addUserRoleMappingsFetchExpectation(mock sqlmock.Sqlmock, user string, project string) {
	uid := uuid.New().String()
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_accountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id WHERE .authsrv_accountresourcerole.account_id = '` + user + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + uid))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id WHERE .authsrv_projectaccountresourcerole.account_id = '` + user + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+uid, "project-"+project))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id WHERE .authsrv_projectaccountnamespacerole.account_id = '` + user + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+uid, "project-"+project, "ns"))
}

func addUserRoleMappingsUpdateExpectation(mock sqlmock.Sqlmock, uuuid string) string {
	uid := uuid.New().String()
	mock.ExpectQuery(`UPDATE "authsrv_accountresourcerole" AS "accountresourcerole" SET trash = TRUE WHERE ."account_id" = '` + uuuid + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uid))
	mock.ExpectQuery(`UPDATE "authsrv_projectaccountresourcerole" AS "projectaccountresourcerole" SET trash = TRUE WHERE ."account_id" = '` + uuuid + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uid))
	mock.ExpectQuery(`UPDATE "authsrv_projectaccountnamespacerole" AS "projectaccountnamespacerole" SET trash = TRUE WHERE ."account_id" = '` + uuuid + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uid))
	return uid
}

func addGroupRoleMappingsFetchExpectation(mock sqlmock.Sqlmock, group string, project string) {
	uid := uuid.New().String()
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_group.name as group FROM "authsrv_grouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_grouprole.role_id JOIN authsrv_group ON authsrv_group.id=authsrv_grouprole.group_id WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "group"}).AddRow("role-"+uid, "group-"+group))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+uid, "project-"+project))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace, authsrv_group.name as group FROM "authsrv_projectgroupnamespacerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+uid, "project-"+project, "ns"))
}

func addGroupRoleMappingsUpdateExpectation(mock sqlmock.Sqlmock, group string) string {
	uid := uuid.New().String()
	mock.ExpectQuery(`UPDATE "authsrv_grouprole" AS "grouprole" SET trash = TRUE WHERE ."group_id" = '` + group + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uid))
	mock.ExpectQuery(`UPDATE "authsrv_projectgrouprole" AS "projectgrouprole" SET trash = TRUE WHERE ."group_id" = '` + group + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uid))
	mock.ExpectQuery(`UPDATE "authsrv_projectgroupnamespacerole" AS "projectgroupnamespacerole" SET trash = TRUE WHERE ."group_id" = '` + group + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uid))
	return uid
}

func addSentryLookupExpectation(mock sqlmock.Sqlmock, users []string, partner, org string) {
	rows := sqlmock.NewRows([]string{"account_id"})
	for _, u := range users {
		rows.AddRow(u)
	}
	mock.ExpectQuery(`SELECT DISTINCT account_id FROM "sentry_account_permission" AS "sap" WHERE .partner_id = '` + partner + `'. AND .organization_id = '` + org + `'`).
		WithArgs().WillReturnRows(rows)
}
