package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
)

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
}

func performGroupBasicAuthzChecks(t *testing.T, mazc mockAuthzClient, guuid string, users []string, roles []*userv3.ProjectNamespaceRole) {
	if len(mazc.cug) > 0 {
		for i, u := range mazc.cug[len(mazc.cug)-1].UserGroups {
			if u.User != "u:"+users[i] {
				t.Errorf("invalid user sent to authz; expected 'u:%v', got '%v'", users[i], u.User)
			}
			if u.Grp != "g:group-"+guuid {
				t.Errorf("invalid group sent to authz; expected 'g:group-%v', got '%v'", guuid, u.Grp)
			}
		}
	}
	if len(mazc.cp) > 0 {
		for i, u := range mazc.cp[len(mazc.cp)-1].Policies {
			if u.Sub != "g:group-"+guuid {
				t.Errorf("invalid sub in policy sent to authz; expected '%v', got '%v'", "g:group-"+guuid, u.Sub)
			}
			if u.Obj != roles[i].Role {
				t.Errorf("invalid obj in policy sent to authz; expected '%v', got '%v'", roles[i].Role, u.Obj)
			}
			if roles[i].Namespace != nil {
				if u.Ns != fmt.Sprint(*roles[i].Namespace) {
					t.Errorf("invalid ns in policy sent to authz; expected '%v', got '%v'", fmt.Sprint(roles[i].Namespace), u.Ns)
				}
			} else {
				if u.Ns != "*" {
					t.Errorf("invalid ns in policy sent to authz; expected '%v', got '%v'", "*", u.Ns)
				}
			}
			if roles[i].Project != nil {
				if u.Proj != *roles[i].Project {
					t.Errorf("invalid proj in policy sent to authz; expected '%v', got '%v'", roles[i].Project, u.Proj)
				}
			} else {
				if u.Proj != "*" {
					t.Errorf("invalid proj in policy sent to authz; expected '%v', got '%v'", "*", u.Proj)
				}
			}
		}
	}

	if len(mazc.dug) > 0 {
		if mazc.dug[len(mazc.dug)-1].Grp != "g:group-"+guuid {
			t.Errorf("invalid group sent to authz; expected 'g:group-%v', got '%v'", guuid, mazc.dug[len(mazc.dug)-1].Grp)
		}
	}
	if len(mazc.dp) > 0 {
		if mazc.dp[len(mazc.dp)-1].Sub != "g:group-"+guuid {
			t.Errorf("invalid sub in policy sent to authz; expected '%v', got '%v'", "g:group-"+guuid, mazc.dp[len(mazc.dp)-1].Sub)
		}
	}
}

func TestCreateGroupNoUsersNoRoles(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	gs := NewGroupService(db, &mazc, getLogger())

	guuid := uuid.New().String()

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "group"."id" FROM "authsrv_group" AS "group" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'group-` + guuid + `'.`).
		WillReturnError(fmt.Errorf("no data available"))

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "authsrv_group"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))
	mock.ExpectCommit()

	group := &userv3.Group{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
		Spec:     &userv3.GroupSpec{},
	}
	group, err := gs.Create(context.Background(), group)
	if err != nil {
		t.Fatal("could not create group:", err)
	}
	performGroupBasicChecks(t, group, guuid)
	performBasicAuthzChecks(t, mazc, 0, 0, 0, 0, 0, 0)
	performGroupBasicAuthzChecks(t, mazc, guuid, []string{}, []*userv3.ProjectNamespaceRole{})
}

func TestCreateGroupDuplicate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	gs := NewGroupService(db, &mazc, getLogger())

	guuid := uuid.New().String()

	// Try to recreate
	addFetchExpectation(mock, "group")
	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "authsrv_group"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))
	mock.ExpectCommit()

	group := &userv3.Group{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
		Spec:     &userv3.GroupSpec{},
	}
	_, err := gs.Create(context.Background(), group)
	if err == nil {
		t.Fatal("should not be able to recreate group with same name")
	}
	performBasicAuthzChecks(t, mazc, 0, 0, 0, 0, 0, 0)
	performGroupBasicAuthzChecks(t, mazc, guuid, []string{}, []*userv3.ProjectNamespaceRole{})
}

func TestCreateGroupWithUsersNoRoles(t *testing.T) {
	tt := []struct {
		name  string
		users int
	}{
		{"single user", 1},
		{"multiple users", 2},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			mazc := mockAuthzClient{}
			gs := NewGroupService(db, &mazc, getLogger())

			guuid := uuid.New().String()

			puuid, ouuid := addParterOrgFetchExpectation(mock)
			addUnavailableExpectation(mock, "group", puuid, ouuid, guuid)

			mock.ExpectBegin()
			mock.ExpectQuery(`INSERT INTO "authsrv_group"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))

			users := []string{}
			var i int
			for i = 0; i < tc.users; i++ {
				user := addUserFetchExpectation(mock)
				users = append(users, "user-"+user)
			}
			mock.ExpectQuery(`INSERT INTO "authsrv_groupaccount"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
			mock.ExpectCommit()

			group := &userv3.Group{
				Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
				Spec:     &userv3.GroupSpec{Users: users},
			}
			group, err := gs.Create(context.Background(), group)
			if err != nil {
				t.Fatal("could not create group:", err)
			}
			performGroupBasicChecks(t, group, guuid)
			for i, ru := range group.Spec.Users {
				if ru != users[i] {
					t.Errorf("user id '%v' not found in resource response", users[i])
				}
			}
			performBasicAuthzChecks(t, mazc, 0, 0, 1, 0, 0, 0)
			performGroupBasicAuthzChecks(t, mazc, guuid, users, []*userv3.ProjectNamespaceRole{})
		})
	}
}

func TestCreateGroupNoUsersWithRoles(t *testing.T) {
	tt := []struct {
		name       string
		roles      []*userv3.ProjectNamespaceRole
		dbname     string
		scope      string
		shouldfail bool
	}{
		{"just role", []*userv3.ProjectNamespaceRole{{Role: uuid.New().String()}}, "authsrv_grouprole", "system", false},
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

			mazc := mockAuthzClient{}
			gs := NewGroupService(db, &mazc, getLogger())

			guuid := uuid.New().String()
			pruuid := uuid.New().String()

			puuid, ouuid := addParterOrgFetchExpectation(mock)
			mock.ExpectQuery(`SELECT "group"."id" FROM "authsrv_group" AS "group" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'group-` + guuid + `'.`).
				WillReturnError(fmt.Errorf("no data available"))

			mock.ExpectBegin()
			mock.ExpectQuery(`INSERT INTO "authsrv_group"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))
			mock.ExpectQuery(`SELECT "resourcerole"."id".* FROM "authsrv_resourcerole" AS "resourcerole"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "scope"}).AddRow(pruuid, "role-name", tc.scope))
			if tc.roles[0].Project != nil {
				mock.ExpectQuery(`SELECT "project"."id" FROM "authsrv_project" AS "project"`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
			}
			mock.ExpectQuery(fmt.Sprintf(`INSERT INTO "%v"`, tc.dbname)).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
			mock.ExpectCommit()

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
			performBasicAuthzChecks(t, mazc, 1, 0, 0, 0, 0, 0)
			performGroupBasicAuthzChecks(t, mazc, guuid, []string{}, tc.roles)
		})
	}
}

func TestCreateGroupWithUsersWithRoles(t *testing.T) {
	projectid := uuid.New().String()
	var namespaceid string = "ns"
	tt := []struct {
		name       string
		users      []string
		roles      []*userv3.ProjectNamespaceRole
		dbname     string
		scope      string
		shouldfail bool
	}{
		{"just role", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Role: uuid.New().String()}}, "authsrv_grouprole", "system", false},
		{"just project", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Project: &projectid}}, "authsrv_grouprole", "system", true},                                    // no role creation without role
		{"just namespace", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Namespace: &namespaceid}}, "authsrv_projectgrouprole", "project", true},                      // no role creation without role,
		{"project and namespace", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Project: &projectid, Namespace: &namespaceid}}, "authsrv_grouprole", "project", true}, // no role creation without role,
		{"project and role", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Project: &projectid, Role: uuid.New().String()}}, "authsrv_projectgrouprole", "project", false},
		// {"project role namespace", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Project: &projectid, Namespace: &namespaceid, Role: uuid.New().String()}}, "authsrv_projectgroupnamespacerole", false},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			mazc := mockAuthzClient{}
			gs := NewGroupService(db, &mazc, getLogger())

			guuid := uuid.New().String()
			pruuid := uuid.New().String()

			puuid, ouuid := addParterOrgFetchExpectation(mock)
			mock.ExpectQuery(`SELECT "group"."id" FROM "authsrv_group" AS "group" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'group-` + guuid + `'.`).WithArgs()

			mock.ExpectBegin()
			// TODO: more precise checks
			mock.ExpectQuery(`INSERT INTO "authsrv_group"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guuid))
			for _, u := range tc.users {
				mock.ExpectQuery(`SELECT "identities"."id" FROM "identities" WHERE .*traits ->> 'email' = '` + u + `'`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuid.New().String(), []byte(`{"email":"johndoe@provider.com"}`)))
			}
			mock.ExpectQuery(`INSERT INTO "authsrv_groupaccount"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

			mock.ExpectQuery(`SELECT "resourcerole"."id".* FROM "authsrv_resourcerole" AS "resourcerole"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "scope"}).AddRow(pruuid, "role-name", tc.scope))
			if tc.roles[0].Project != nil {
				mock.ExpectQuery(`SELECT "project"."id" FROM "authsrv_project" AS "project"`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
			}
			mock.ExpectQuery(fmt.Sprintf(`INSERT INTO "%v"`, tc.dbname)).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
			mock.ExpectCommit()

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
			performBasicAuthzChecks(t, mazc, 1, 0, 1, 0, 0, 0)
			performGroupBasicAuthzChecks(t, mazc, guuid, tc.users, tc.roles)
		})
	}
}

func TestUpdateGroupWithUsersWithRoles(t *testing.T) {
	tt := []struct {
		name   string
		users  []string
		roles  []*userv3.ProjectNamespaceRole
		dbname string
		scope  string
	}{
		{"user role update", []string{"user-" + uuid.New().String()}, []*userv3.ProjectNamespaceRole{{Role: uuid.New().String()}}, "authsrv_grouprole", "system"},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			mazc := mockAuthzClient{}
			gs := NewGroupService(db, &mazc, getLogger())

			guuid := uuid.New().String()
			pruuid := uuid.New().String()

			// performing update
			puuid, ouuid := addParterOrgFetchExpectation(mock)
			mock.ExpectQuery(`SELECT "group"."id", "group"."name",.* FROM "authsrv_group" AS "group" WHERE .*name = 'group-` + guuid + `'`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(guuid, "group-"+guuid))

			mock.ExpectBegin()
			addGroupUserMappingsUpdateExpectation(mock, guuid)
			for _, u := range tc.users {
				// addUserFetchExpectation(mock) // TODO: look into this
				mock.ExpectQuery(`SELECT "identities"."id" FROM "identities" WHERE .*traits ->> 'email' = '` + u + `'`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuid.New().String(), []byte(`{"email":"johndoe@provider.com"}`)))
			}
			mock.ExpectQuery(`INSERT INTO "authsrv_groupaccount"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

			addGroupRoleMappingsUpdateExpectation(mock, guuid)
			mock.ExpectQuery(`SELECT "resourcerole"."id".* FROM "authsrv_resourcerole" AS "resourcerole"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "scope"}).AddRow(pruuid, "role-name", tc.scope))
			if tc.roles[0].Project != nil {
				mock.ExpectQuery(`SELECT "project"."id" FROM "authsrv_project" AS "project"`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
			}
			mock.ExpectQuery(fmt.Sprintf(`INSERT INTO "%v"`, tc.dbname)).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
			mock.ExpectExec(`UPDATE "authsrv_group"`).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

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
			performBasicAuthzChecks(t, mazc, 1, 1, 1, 1, 0, 0)
			performGroupBasicAuthzChecks(t, mazc, guuid, tc.users, tc.roles)
		})
	}
}

func TestGroupDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	gs := NewGroupService(db, &mazc, getLogger())

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	guuid := addFetchExpectation(mock, "group")
	mock.ExpectBegin()
	addGroupRoleMappingsUpdateExpectation(mock, guuid)
	addGroupUserMappingsUpdateExpectation(mock, guuid)
	addDeleteExpectation(mock, "group", guuid)
	mock.ExpectCommit()

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

	mazc := mockAuthzClient{}
	gs := NewGroupService(db, &mazc, getLogger())

	guuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	addUnavailableExpectation(mock, "group", puuid, ouuid, guuid)

	group := &userv3.Group{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "group-" + guuid},
	}
	_, err := gs.Delete(context.Background(), group)
	if err == nil {
		t.Fatal("deleted non existent group")
	}
}

func TestGroupGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	gs := NewGroupService(db, &mazc, getLogger())

	guuid := uuid.New().String()
	uuuid := uuid.New().String()

	puuid, ouuid := addParterOrgFetchExpectation(mock)
	addFetchByNameExpectation(mock, "group", guuid)
	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id WHERE .authsrv_groupaccount.group_id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	addGroupRoleMappingsFetchExpectation(mock, guuid, puuid)

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
	if group.GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != "ns" {
		t.Errorf("invalid namespace in role returned for user, expected ns; got '%v'", group.GetSpec().GetProjectNamespaceRoles()[2].Namespace)
	}
}

func TestGroupGetById(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	gs := NewGroupService(db, &mazc, getLogger())

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

	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_group.name as group FROM "authsrv_grouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_grouprole.role_id JOIN authsrv_group ON authsrv_group.id=authsrv_grouprole.group_id WHERE .authsrv_grouprole.group_id = '` + guuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace, authsrv_group.name as group FROM "authsrv_projectgroupnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id JOIN authsrv_group ON authsrv_group.id=authsrv_projectgroupnamespacerole.group_id WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, "ns"))

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
	if group.GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != "ns" {
		t.Errorf("invalid namespace in role returned for user, expected ns; got '%v'", group.GetSpec().GetProjectNamespaceRoles()[2].Namespace)
	}
}

func TestGroupList(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	gs := NewGroupService(db, &mazc, getLogger())

	guuid1 := uuid.New().String()
	guuid2 := uuid.New().String()
	uuuid := uuid.New().String()
	pruuid := uuid.New().String()

	_, _ = addOrgParterFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "group"."id", "group"."name", .* FROM "authsrv_group" AS "group"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(guuid1, "group-"+guuid1).AddRow(guuid2, "group-"+guuid2))

	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id WHERE .authsrv_groupaccount.group_id = '` + guuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	addGroupRoleMappingsFetchExpectation(mock, guuid1, pruuid)

	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id WHERE .authsrv_groupaccount.group_id = '` + guuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	addGroupRoleMappingsFetchExpectation(mock, guuid2, pruuid)

	qo := &commonv3.QueryOptions{}
	grouplist, err := gs.List(context.Background(), query.WithOptions(qo))
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

func TestGroupListFiltered(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	gs := NewGroupService(db, &mazc, getLogger())

	guuid1 := uuid.New().String()
	guuid2 := uuid.New().String()
	uuuid := uuid.New().String()
	pruuid := uuid.New().String()

	puuid, ouuid := addOrgParterFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "group"."id", "group"."name", .*WHERE .name ILIKE '%filter-query%'. AND .partner_id = '` + puuid + `'. AND .organization_id = '` + ouuid + `'. AND .trash = FALSE. ORDER BY "email" asc LIMIT 50 OFFSET 20`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(guuid1, "group-"+guuid1).AddRow(guuid2, "group-"+guuid2))

	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id WHERE .authsrv_groupaccount.group_id = '` + guuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	addGroupRoleMappingsFetchExpectation(mock, guuid1, pruuid)

	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id WHERE .authsrv_groupaccount.group_id = '` + guuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	addGroupRoleMappingsFetchExpectation(mock, guuid2, pruuid)

	qo := &commonv3.QueryOptions{Q: "filter-query", Limit: 50, Offset: 20, OrderBy: "email", Order: "asc"}
	grouplist, err := gs.List(context.Background(), query.WithOptions(qo))
	if err != nil {
		t.Fatal("could not list groups:", err)
	}
	if grouplist.Metadata.Count != 2 {
		t.Errorf("incorrect number of groups returned, expected 2; got %v", grouplist.Metadata.Count)
	}
	if grouplist.Items[0].Metadata.Name != "group-"+guuid1 {
		t.Errorf("incorrect group ids returned when listing")
	}
	if grouplist.Items[0].GetSpec().GetUsers()[0] != "johndoe@provider.com" {
		t.Errorf("incorrect username in for group, expected johndoe@provider.com ; got '%v'", grouplist.Items[0].GetSpec().GetUsers()[0])
	}
}
