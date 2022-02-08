package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
	"github.com/google/uuid"
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
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
	return db, mock
}

func performGroupBasicChecks(t *testing.T, group *userv3.Group, guuid string) {
	_, err := uuid.Parse(group.GetMetadata().GetOrganization())
	if err == nil {
		t.Error("org in metadata should be name not id")
	}
	_, err = uuid.Parse(group.GetMetadata().GetPartner())
	if err == nil {
		t.Error("partner in metadata should be name not id")
	}
	if group.GetMetadata().GetName() != "group-"+guuid {
		t.Error("invalid name returned")
	}
	if group.Status.ConditionStatus != v3.ConditionStatus_StatusOK {
		t.Error("group status is not OK")
	}
}

func TestCreateGroupNoUsersNoRoles(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewGroupService(db)
	defer gs.Close()

	guuid := uuid.New().String()
	puuid := uuid.New().String()
	fmt.Println("puuid:", puuid)
	ouuid := uuid.New().String()
	fmt.Println("ouuid:", ouuid)

	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	mock.ExpectQuery(`SELECT "group"."id" FROM "authsrv_group" AS "group" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'group-` + guuid + `'.`).WithArgs()
	mock.ExpectQuery(`INSERT INTO "authsrv_group"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))

	group := &userv3.Group{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
		Spec:     &userv3.GroupSpec{},
	}
	group, err := gs.Create(context.Background(), group)
	if err != nil {
		t.Fatal("could not create group:", err)
	}
	performGroupBasicChecks(t, group, guuid)
}

func TestCreateGroupDuplicate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewGroupService(db)
	defer gs.Close()

	guuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	group := &userv3.Group{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
		Spec:     &userv3.GroupSpec{},
	}

	// Try to recreate
	mock.ExpectQuery(`SELECT "group"."id" FROM "authsrv_group" AS "group"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))
	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	// TODO: more precise checks
	mock.ExpectQuery(`INSERT INTO "authsrv_group"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))
	_, err := gs.Create(context.Background(), group)
	if err == nil {
		t.Fatal("should not be able to recreate group with same name")
	}
}

func TestCreateGroupWithUsersNoRoles(t *testing.T) {
	tt := []struct {
		name  string
		users []string
	}{
		{"single user", []string{"users-" + uuid.New().String()}},
		{"multiple users", []string{"users-" + uuid.New().String(), "users-" + uuid.New().String()}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			gs := NewGroupService(db)
			defer gs.Close()

			guuid := uuid.New().String()
			puuid := uuid.New().String()
			ouuid := uuid.New().String()

			// TODO: more precise checks
			mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
			mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
			mock.ExpectQuery(`SELECT "group"."id" FROM "authsrv_group" AS "group" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'group-` + guuid + `'.`).WithArgs()
			mock.ExpectQuery(`INSERT INTO "authsrv_group"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))
			for _, u := range tc.users {
				mock.ExpectQuery(`SELECT "identities"."id" FROM "identities" WHERE .*traits ->> 'email' = '` + u + `'`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuid.New().String(), []byte(`{"email":"johndoe@provider.com"}`)))
			}
			mock.ExpectQuery(`INSERT INTO "authsrv_groupaccount"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

			group := &userv3.Group{
				Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
				Spec:     &userv3.GroupSpec{Users: tc.users},
			}
			group, err := gs.Create(context.Background(), group)
			if err != nil {
				t.Fatal("could not create group:", err)
			}
			performGroupBasicChecks(t, group, guuid)
			for i, ru := range group.Spec.Users {
				if ru != tc.users[i] {
					t.Errorf("user id '%v' not found in resource response", tc.users[i])
				}
			}
		})
	}
}

func TestCreateGroupNoUsersWithRoles(t *testing.T) {
	// projectid := uuid.New().String()
	// var namespaceid int64 = 7
	tt := []struct {
		name       string
		roles      []*userv3.ProjectNamespaceRole
		dbname     string
		shouldfail bool
	}{
		{"just role", []*userv3.ProjectNamespaceRole{{Role: uuid.New().String()}}, "authsrv_grouprole", false},
		// {"just project", []*userv3.ProjectNamespaceRole{{Project: &projectid}}, "authsrv_grouprole", true},                                   // no role creation without role
		// {"just namespace", []*userv3.ProjectNamespaceRole{{Namespace: &namespaceid}}, "authsrv_grouprole", true},                             // no role creation without role,
		// {"project and namespace", []*userv3.ProjectNamespaceRole{{Project: &projectid, Namespace: &namespaceid}}, "authsrv_grouprole", true}, // no role creation without role,
		// {"project and role", []*userv3.ProjectNamespaceRole{{Project: &projectid, Role: uuid.New().String()}}, "authsrv_projectgrouprole", false},
		// {"project role namespace", []*userv3.ProjectNamespaceRole{{Project: &projectid, Namespace: &namespaceid, Role: uuid.New().String()}}, "authsrv_projectgroupnamespacerole", false},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			gs := NewGroupService(db)
			defer gs.Close()

			guuid := uuid.New().String()
			puuid := uuid.New().String()
			ouuid := uuid.New().String()
			pruuid := uuid.New().String()

			// TODO: more precise checks
			mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
			mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
			mock.ExpectQuery(`SELECT "group"."id" FROM "authsrv_group" AS "group" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'group-` + guuid + `'.`).WithArgs()
			mock.ExpectQuery(`INSERT INTO "authsrv_group"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))
			mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
			if tc.roles[0].Project != nil {
				mock.ExpectQuery(`SELECT "project"."id" FROM "authsrv_project" AS "project"`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
			}
			mock.ExpectQuery(fmt.Sprintf(`INSERT INTO "%v"`, tc.dbname)).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

			group := &userv3.Group{
				Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
				Spec:     &userv3.GroupSpec{ProjectNamespaceRoles: tc.roles},
			}
			group, err := gs.Create(context.Background(), group)
			if tc.shouldfail {
				if err == nil {
					// TODO: check for proper error messages
					t.Fatal("expected group not to be created, but was created")
				} else {
					return
				}
			}
			if err != nil {
				t.Fatal("could not create group:", err)
			}
			performGroupBasicChecks(t, group, guuid)
			for i, rr := range group.Spec.ProjectNamespaceRoles {
				if rr != tc.roles[i] {
					t.Errorf("role '%v' not found in resource response", tc.roles[i])
				}
			}
		})
	}
}

func TestCreateGroupWithUsersWithRoles(t *testing.T) {
	projectid := uuid.New().String()
	var namespaceid int64 = 7
	tt := []struct {
		name       string
		users      []string
		roles      []*userv3.ProjectNamespaceRole
		dbname     string
		shouldfail bool
	}{
		{"just role", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Role: uuid.New().String()}}, "authsrv_grouprole", false},
		{"just project", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Project: &projectid}}, "authsrv_grouprole", true},                                   // no role creation without role
		{"just namespace", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Namespace: &namespaceid}}, "authsrv_grouprole", true},                             // no role creation without role,
		{"project and namespace", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Project: &projectid, Namespace: &namespaceid}}, "authsrv_grouprole", true}, // no role creation without role,
		{"project and role", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Project: &projectid, Role: uuid.New().String()}}, "authsrv_projectgrouprole", false},
		{"project role namespace", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Project: &projectid, Namespace: &namespaceid, Role: uuid.New().String()}}, "authsrv_projectgroupnamespacerole", false},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			gs := NewGroupService(db)
			defer gs.Close()

			guuid := uuid.New().String()
			puuid := uuid.New().String()
			ouuid := uuid.New().String()
			pruuid := uuid.New().String()

			mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
			mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
			mock.ExpectQuery(`SELECT "group"."id" FROM "authsrv_group" AS "group" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'group-` + guuid + `'.`).WithArgs()

			// TODO: more precise checks
			mock.ExpectQuery(`INSERT INTO "authsrv_group"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))
			for _, u := range tc.users {
				mock.ExpectQuery(`SELECT "identities"."id" FROM "identities" WHERE .*traits ->> 'email' = '` + u + `'`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuid.New().String(), []byte(`{"email":"johndoe@provider.com"}`)))
			}
			mock.ExpectQuery(`INSERT INTO "authsrv_groupaccount"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

			mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
			if tc.roles[0].Project != nil {
				mock.ExpectQuery(`SELECT "project"."id" FROM "authsrv_project" AS "project"`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
			}
			mock.ExpectQuery(fmt.Sprintf(`INSERT INTO "%v"`, tc.dbname)).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

			group := &userv3.Group{
				Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
				Spec:     &userv3.GroupSpec{ProjectNamespaceRoles: tc.roles, Users: tc.users},
			}
			group, err := gs.Create(context.Background(), group)
			if tc.shouldfail {
				if err == nil {
					// TODO: check for proper error messages
					t.Fatal("expected group not to be created, but was created")
				} else {
					return
				}
			}
			if err != nil {
				t.Fatal("could not create group:", err)
			}
			performGroupBasicChecks(t, group, guuid)
			for i, ru := range group.Spec.Users {
				if ru != tc.users[i] {
					t.Errorf("user id '%v' not found in resource response", tc.users[i])
				}
			}
			for i, rr := range group.Spec.ProjectNamespaceRoles {
				if rr != tc.roles[i] {
					t.Errorf("role '%v' not found in resource response", tc.roles[i])
				}
			}
		})
	}
}

func TestUpdateGroupWithUsersWithRoles(t *testing.T) {
	tt := []struct {
		name   string
		users  []string
		roles  []*userv3.ProjectNamespaceRole
		dbname string
	}{
		{"user role udpate", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Role: uuid.New().String()}}, "authsrv_grouprole"},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			gs := NewGroupService(db)
			defer gs.Close()

			guuid := uuid.New().String()
			puuid := uuid.New().String()
			ouuid := uuid.New().String()
			pruuid := uuid.New().String()

			// performing update
			mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
			mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
			mock.ExpectQuery(`SELECT "group"."id", "group"."name",.* FROM "authsrv_group" AS "group" WHERE .*name = 'group-` + guuid + `'`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(guuid, "group-"+guuid))

			// TODO: more precise checks
			mock.ExpectExec(`DELETE FROM "authsrv_groupaccount" AS "groupaccount" WHERE ."group_id" = '` + guuid).
				WillReturnResult(sqlmock.NewResult(1, 1))
			for _, u := range tc.users {
				mock.ExpectQuery(`SELECT "identities"."id" FROM "identities" WHERE .*traits ->> 'email' = '` + u + `'`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuid.New().String(), []byte(`{"email":"johndoe@provider.com"}`)))
			}
			mock.ExpectQuery(`INSERT INTO "authsrv_groupaccount"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
			mock.ExpectExec(`DELETE FROM "authsrv_grouprole" AS "grouprole" WHERE ."group_id" = '` + guuid).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(`DELETE FROM "authsrv_projectgrouprole" AS "projectgrouprole" WHERE ."group_id" = '` + guuid).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(`DELETE FROM "authsrv_projectgroupnamespacerole" AS "projectgroupnamespacerole" WHERE ."group_id" = '` + guuid).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(`SELECT "resourcerole"."id" FROM "authsrv_resourcerole" AS "resourcerole"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
			if tc.roles[0].Project != nil {
				mock.ExpectQuery(`SELECT "project"."id" FROM "authsrv_project" AS "project"`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
			}
			mock.ExpectQuery(fmt.Sprintf(`INSERT INTO "%v"`, tc.dbname)).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
			mock.ExpectExec(`UPDATE "authsrv_group"`).
				WillReturnResult(sqlmock.NewResult(1, 1))

			group := &userv3.Group{
				Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
				Spec:     &userv3.GroupSpec{ProjectNamespaceRoles: tc.roles, Users: tc.users},
			}
			group, err := gs.Update(context.Background(), group)
			if err != nil {
				t.Fatal("could not update group:", err)
			}
			performGroupBasicChecks(t, group, guuid)

			for i, ru := range group.Spec.Users {
				if ru != tc.users[i] {
					t.Errorf("user id '%v' not found in resource response", tc.users[i])
				}
			}
			for i, rr := range group.Spec.ProjectNamespaceRoles {
				if rr != tc.roles[i] {
					t.Errorf("role '%v' not found in resource response", tc.roles[i])
				}
			}
		})
	}
}

func TestGroupDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewGroupService(db)
	defer gs.Close()

	guuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	mock.ExpectQuery(`SELECT "group"."id", "group"."name", .* FROM "authsrv_group" AS "group" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(guuid, "group-"+guuid))

	mock.ExpectExec(`DELETE FROM "authsrv_grouprole" AS "grouprole" WHERE ."group_id" = '` + guuid).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "authsrv_projectgrouprole" AS "projectgrouprole" WHERE ."group_id" = '` + guuid).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "authsrv_projectgroupnamespacerole" AS "projectgroupnamespacerole" WHERE ."group_id" = '` + guuid).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "authsrv_groupaccount" AS "groupaccount" WHERE ."group_id" = '` + guuid).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "authsrv_group" AS "group" WHERE .id = '` + guuid).
		WillReturnResult(sqlmock.NewResult(1, 1))

	group := &userv3.Group{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
	}
	_, err := gs.Delete(context.Background(), group)
	if err != nil {
		t.Fatal("could not delete group:", err)
	}
}

func TestGroupDeleteNonExist(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewGroupService(db)
	defer gs.Close()

	guuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "group"."id", "group"."name", .* FROM "authsrv_group" AS "group" WHERE`).
		WithArgs().WillReturnError(fmt.Errorf("No data available"))

	group := &userv3.Group{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
	}
	_, err := gs.Delete(context.Background(), group)
	if err == nil {
		t.Fatal("deleted non existant group")
	}
}

func TestGroupGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewGroupService(db)
	defer gs.Close()

	guuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	uuuid := uuid.New().String()
	ruuid := uuid.New().String()
	pruuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	mock.ExpectQuery(`SELECT "group"."id", "group"."name", .* FROM "authsrv_group" AS "group" WHERE .*name = 'group-` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(guuid, "group-"+guuid))
	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id WHERE .authsrv_groupaccount.group_id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))

	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_grouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_grouprole.role_id WHERE .authsrv_grouprole.group_id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id WHERE .authsrv_projectgrouprole.group_id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace FROM "authsrv_projectgroupnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id WHERE .authsrv_projectgroupnamespacerole.group_id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, 9))

	group := &userv3.Group{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
	}
	group, err := gs.GetByName(context.Background(), group)
	if err != nil {
		t.Fatal("could not get group:", err)
	}
	performGroupBasicChecks(t, group, guuid)
	if group.GetSpec().GetUsers()[0] != "johndoe@provider.com" {
		t.Errorf("incorrect username in for group, expected johndoe@provider.com ; got '%v'", group.GetSpec().GetUsers()[0])
	}

	if len(group.GetSpec().GetProjectNamespaceRoles()) != 3 {
		t.Errorf("invalid number of roles returned for user, expected 3; got '%v'", len(group.GetSpec().GetProjectNamespaceRoles()))
	}
	if group.GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != 9 {
		t.Errorf("invalid namespace in role returned for user, expected 9; got '%v'", group.GetSpec().GetProjectNamespaceRoles()[2].Namespace)
	}
}

func TestGroupGetById(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewGroupService(db)
	defer gs.Close()

	guuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	uuuid := uuid.New().String()
	ruuid := uuid.New().String()
	pruuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "group"."id", "group"."name", .* FROM "authsrv_group" AS "group" WHERE .*id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(guuid, "group-"+guuid))
	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id WHERE .authsrv_groupaccount.group_id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))

	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_grouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_grouprole.role_id WHERE .authsrv_grouprole.group_id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id WHERE .authsrv_projectgrouprole.group_id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace FROM "authsrv_projectgroupnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id WHERE .authsrv_projectgroupnamespacerole.group_id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, 9))

	group := &userv3.Group{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Id: guuid},
	}
	group, err := gs.GetByID(context.Background(), group)
	if err != nil {
		t.Fatal("could not get group:", err)
	}
	performGroupBasicChecks(t, group, guuid)
	if group.GetSpec().GetUsers()[0] != "johndoe@provider.com" {
		t.Errorf("incorrect username in for group, expected johndoe@provider.com ; got '%v'", group.GetSpec().GetUsers()[0])
	}

	if len(group.GetSpec().GetProjectNamespaceRoles()) != 3 {
		t.Errorf("invalid number of roles returned for user, expected 3; got '%v'", len(group.GetSpec().GetProjectNamespaceRoles()))
	}
	if group.GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != 9 {
		t.Errorf("invalid namespace in role returned for user, expected 9; got '%v'", group.GetSpec().GetProjectNamespaceRoles()[2].Namespace)
	}
}

func TestGroupList(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	gs := NewGroupService(db)
	defer gs.Close()

	guuid1 := uuid.New().String()
	guuid2 := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	uuuid := uuid.New().String()
	ruuid := uuid.New().String()
	pruuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "group"."id", "group"."name", .* FROM "authsrv_group" AS "group"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(guuid1, "group-"+guuid1).AddRow(guuid2, "group-"+guuid2))

	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id WHERE .authsrv_groupaccount.group_id = '` + guuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_grouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_grouprole.role_id WHERE .authsrv_grouprole.group_id = '` + guuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id WHERE .authsrv_projectgrouprole.group_id = '` + guuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace FROM "authsrv_projectgroupnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id WHERE .authsrv_projectgroupnamespacerole.group_id = '` + guuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, 9))

	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id WHERE .authsrv_groupaccount.group_id = '` + guuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_grouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_grouprole.role_id WHERE .authsrv_grouprole.group_id = '` + guuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id WHERE .authsrv_projectgrouprole.group_id = '` + guuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace FROM "authsrv_projectgroupnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id WHERE .authsrv_projectgroupnamespacerole.group_id = '` + guuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, 9))

	group := &userv3.Group{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid},
	}
	grouplist, err := gs.List(context.Background(), group)
	if err != nil {
		t.Fatal("could not list groups:", err)
	}
	if grouplist.Metadata.Count != 2 {
		t.Errorf("incorrect number of groups returned, expected 2; got %v", grouplist.Metadata.Count)
	}
	if grouplist.Items[0].Metadata.Name != "group-"+guuid1 || grouplist.Items[1].Metadata.Name != "group-"+guuid2 {
		t.Errorf("incorrect group ids returned when listing")
	}
	if grouplist.Items[0].GetSpec().GetUsers()[0] != "johndoe@provider.com" {
		t.Errorf("incorrect username in for group, expected johndoe@provider.com ; got '%v'", grouplist.Items[0].GetSpec().GetUsers()[0])
	}
}
