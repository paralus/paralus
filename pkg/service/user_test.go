package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/common"
	"github.com/paralus/paralus/pkg/query"
	userrpcv3 "github.com/paralus/paralus/proto/rpc/user"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
)

func performUserBasicChecks(t *testing.T, user *userv3.User, uuuid string) {
	_, err := uuid.Parse(user.GetMetadata().GetOrganization())
	if err == nil {
		t.Error("org in metadata should be name not id")
	}
	_, err = uuid.Parse(user.GetMetadata().GetPartner())
	if err == nil {
		t.Error("partner in metadata should be name not id")
	}
}

func performUserBasicAuthzChecks(t *testing.T, mazc mockAuthzClient, uuuid string, roles []*userv3.ProjectNamespaceRole) {
	if len(mazc.cp) > 0 {
		for i, u := range mazc.cp[len(mazc.cp)-1].Policies {
			if u.Sub != "u:user-"+uuuid {
				t.Errorf("invalid sub in policy sent to authz; expected '%v', got '%v'", "u:user-"+uuuid, u.Sub)
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
}

func TestCreateUser(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	uuuid := uuid.New().String()
	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectBegin()
	mock.ExpectCommit()

	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
		Spec:     &userv3.UserSpec{},
	}
	user, err := us.Create(context.Background(), user)
	if err != nil {
		t.Fatal("could not create user:", err)
	}
	performUserBasicChecks(t, user, uuuid)
	if user.GetMetadata().GetName() != "user-"+uuuid {
		t.Errorf("expected name 'user-%v'; got '%v'", uuuid, user.GetMetadata().GetName())
	}
	performBasicAuthProviderChecks(t, *ap, 1, 0, 0, 0)
}

func TestCreateUserWithRole(t *testing.T) {
	tt := []struct {
		name       string
		role       bool
		project    bool
		namespace  bool
		dbname     string
		scope      string
		shouldfail bool
	}{
		{"just role", true, false, false, "authsrv_accountresourcerole", "system", false},
		{"just role org scope", true, false, false, "authsrv_accountresourcerole", "organization", false},
		{"just project", false, true, false, "authsrv_accountrole", "system", true},         // no role creation without role
		{"just namespace", false, false, true, "authsrv_accountrole", "system", true},       // no role creation without role,
		{"project and namespace", false, true, true, "authsrv_accountrole", "system", true}, // no role creation without role,
		{"project and role", true, true, false, "authsrv_projectaccountresourcerole", "project", false},
		{"project role namespace", true, true, true, "authsrv_projectaccountresourcerole", "project", false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			ap := &mockAuthProvider{}
			mazc := mockAuthzClient{}
			us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

			uuuid := uuid.New().String()

			puuid, ouuid := addParterOrgFetchExpectation(mock)

			mock.ExpectBegin()
			ruuid := addResourceRoleFetchExpectation(mock, tc.scope)
			role := &userv3.ProjectNamespaceRole{}
			if tc.role {
				role.Role = idname(ruuid, "role")
			}
			if tc.project {
				pruuid := addFetchIdExpectation(mock, "project")
				role.Project = &pruuid
			}
			if tc.namespace {
				var ns = "ns"
				role.Namespace = &ns
			}
			mock.ExpectQuery(fmt.Sprintf(`INSERT INTO "%v"`, tc.dbname)).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
			mock.ExpectCommit()

			user := &userv3.User{
				Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
				Spec:     &userv3.UserSpec{ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{role}},
			}

			user, err := us.Create(context.Background(), user)
			if tc.shouldfail {
				if err == nil {
					// TODO: check for proper error messages
					t.Fatal("expected user not to be created, but was created")
				} else {
					return
				}
			}
			if err != nil {
				t.Fatal("could not create user:", err)
			}
			performUserBasicChecks(t, user, uuuid)
			if user.GetMetadata().GetName() != "user-"+uuuid {
				t.Errorf("expected name 'user-%v'; got '%v'", uuuid, user.GetMetadata().GetName())
			}

			performBasicAuthProviderChecks(t, *ap, 1, 0, 0, 0)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	// performing update
	uuuid := addUserFullFetchExpectation(mock)
	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectBegin()
	_ = addUserRoleMappingsUpdateExpectation(mock, uuuid)
	addUserGroupMappingsUpdateExpectation(mock, uuuid)
	ruuid := addResourceRoleFetchExpectation(mock, "project")
	pruuid := addFetchExpectation(mock, "project")
	mock.ExpectQuery(`INSERT INTO "authsrv_projectaccountresourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectCommit()

	var ns string = "ns"
	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
		Spec:     &userv3.UserSpec{ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{{Project: idnamea(pruuid, "project"), Namespace: &ns, Role: idname(ruuid, "role")}}},
	}
	user, err := us.Update(context.Background(), user)
	if err != nil {
		t.Fatal("could not create user:", err)
	}
	performUserBasicChecks(t, user, uuuid)
	if user.GetMetadata().GetName() != "user-"+uuuid {
		t.Errorf("expected name 'user-%v'; got '%v'", uuuid, user.GetMetadata().GetName())
	}
	performBasicAuthProviderChecks(t, *ap, 0, 1, 0, 0)
}

func TestUpdateUserWithGroup(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	// performing update
	uuuid := addUserFullFetchExpectation(mock)
	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectBegin()
	_ = addUserRoleMappingsUpdateExpectation(mock, uuuid)
	addUserGroupMappingsUpdateExpectation(mock, uuuid)
	ruuid := addResourceRoleFetchExpectation(mock, "project")
	pruuid := addFetchExpectation(mock, "project")
	mock.ExpectQuery(`INSERT INTO "authsrv_projectaccountresourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	addFetchExpectation(mock, "group")
	mock.ExpectQuery(`INSERT INTO "authsrv_groupaccount"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectCommit()

	var ns string = "ns"
	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
		Spec: &userv3.UserSpec{
			Groups:                []string{"group"},
			ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{{Project: idnamea(pruuid, "project"), Namespace: &ns, Role: idname(ruuid, "role")}},
		},
	}
	user, err := us.Update(context.Background(), user)
	if err != nil {
		t.Fatal("could not create user:", err)
	}
	performUserBasicChecks(t, user, uuuid)
	if user.GetMetadata().GetName() != "user-"+uuuid {
		t.Errorf("expected name 'user-%v'; got '%v'", uuuid, user.GetMetadata().GetName())
	}
	performBasicAuthProviderChecks(t, *ap, 0, 1, 0, 0)
}

func TestUpdateUserWithIdpGroupPassed(t *testing.T) {
	// Having idp groups passed down should not affect, it should come from db
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	// performing update
	uuuid := addUserFullFetchExpectation(mock)
	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectBegin()
	_ = addUserRoleMappingsUpdateExpectation(mock, uuuid)
	addUserGroupMappingsUpdateExpectation(mock, uuuid)
	ruuid := addResourceRoleFetchExpectation(mock, "project")
	pruuid := addFetchExpectation(mock, "project")
	mock.ExpectQuery(`INSERT INTO "authsrv_projectaccountresourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	// addFetchExpectation(mock, "group")
	// mock.ExpectQuery(`INSERT INTO "authsrv_groupaccount"`).
	// 	WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectCommit()

	var ns = "ns"
	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
		Spec: &userv3.UserSpec{
			IdpGroups:             []string{"group"},
			ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{{Project: idnamea(pruuid, "project"), Namespace: &ns, Role: idname(ruuid, "role")}},
		},
	}
	user, err := us.Update(context.Background(), user)
	if err != nil {
		t.Fatal("could not create user:", err)
	}
	performUserBasicChecks(t, user, uuuid)
	if user.GetMetadata().GetName() != "user-"+uuuid {
		t.Errorf("expected name 'user-%v'; got '%v'", uuuid, user.GetMetadata().GetName())
	}
	performBasicAuthProviderChecks(t, *ap, 0, 1, 0, 0)
}

func TestUpdateUserWithIdpGroupFetched(t *testing.T) {
	// Having idp groups passed down should not affect, it should come from db
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	// performing update
	uuuid := addUserFullFetchExpectationWithIdpGroups(mock)
	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectBegin()
	_ = addUserRoleMappingsUpdateExpectation(mock, uuuid)
	addUserGroupMappingsUpdateExpectation(mock, uuuid)
	ruuid := addResourceRoleFetchExpectation(mock, "project")
	pruuid := addFetchExpectation(mock, "project")
	mock.ExpectQuery(`INSERT INTO "authsrv_projectaccountresourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	addFetchExpectation(mock, "group")
	mock.ExpectQuery(`INSERT INTO "authsrv_groupaccount"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectCommit()

	var ns = "ns"
	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
		Spec: &userv3.UserSpec{
			IdpGroups:             []string{"group"},
			ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{{Project: idnamea(pruuid, "project"), Namespace: &ns, Role: idname(ruuid, "role")}},
		},
	}
	user, err := us.Update(context.Background(), user)
	if err != nil {
		t.Fatal("could not create user:", err)
	}
	performUserBasicChecks(t, user, uuuid)
	if user.GetMetadata().GetName() != "user-"+uuuid {
		t.Errorf("expected name 'user-%v'; got '%v'", uuuid, user.GetMetadata().GetName())
	}
	performBasicAuthProviderChecks(t, *ap, 0, 1, 0, 0)
}

func TestUpdateUserInvalid(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	// performing update
	uuuid := addUserFullFetchExpectation(mock)
	puuid, ouuid := addParterOrgFetchExpectation(mock)
	mock.ExpectBegin()
	_ = addUserRoleMappingsUpdateExpectation(mock, uuuid)
	addUserGroupMappingsUpdateExpectation(mock, uuuid)
	ruuid := addResourceRoleFetchExpectation(mock, "project")
	pruuid := addFetchExpectation(mock, "project")
	mock.ExpectQuery(`INSERT INTO "authsrv_projectaccountresourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectCommit()

	var ns string = "ns"
	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
		Spec: &userv3.UserSpec{
			IdpGroups:             []string{"unnecessary"},
			ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{{Project: idnamea(pruuid, "project"), Namespace: &ns, Role: idname(ruuid, "role")}},
		},
	}
	user, err := us.Update(context.Background(), user)
	if err != nil {
		t.Fatal("could not create user:", err)
	}
	performUserBasicChecks(t, user, uuuid)
	if len(user.Spec.IdpGroups) != 0 {
		t.Errorf("Idp groups added to local user")
	}
	if user.GetMetadata().GetName() != "user-"+uuuid {
		t.Errorf("expected name 'user-%v'; got '%v'", uuuid, user.GetMetadata().GetName())
	}
	performBasicAuthProviderChecks(t, *ap, 0, 1, 0, 0)
}

func TestUserGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	guuid := uuid.New().String()
	ruuid := uuid.New().String()
	pruuid := uuid.New().String()
	authenticated := time.Date(2022, 11, 1, 17, 49, 0, 0, time.UTC)

	uuuid := addUserFetchExpectation(mock)
	mock.ExpectQuery(`SELECT "group"."id".* FROM "authsrv_group" AS "group" JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id WHERE .authsrv_groupaccount.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("group-" + guuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_group.name as group FROM "authsrv_grouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_grouprole.role_id JOIN authsrv_group ON authsrv_group.id=authsrv_grouprole.group_id WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "group"}).AddRow("role-"+ruuid, "group-"+guuid))

	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+puuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace, authsrv_group.name as group FROM "authsrv_projectgroupnamespacerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+puuid))

	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_accountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id WHERE .authsrv_accountresourcerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))

	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id WHERE .authsrv_projectaccountresourcerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id WHERE .authsrv_projectaccountnamespacerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, "ns"))
	mock.ExpectQuery(`select .* from sessions where .*`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"max"}).
		AddRow(authenticated))

	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
	}
	user, err := us.GetByName(context.Background(), user)
	if err != nil {
		t.Fatal("could not get user:", err)
	}
	performUserBasicChecks(t, user, uuuid)
	if user.GetMetadata().GetName() != "user-"+uuuid {
		t.Errorf("invalid email for user, expected '%v'; got '%v'", "user-"+uuuid, user.GetMetadata().GetName())
	}
	if len(user.GetSpec().GetGroups()) != 1 {
		t.Errorf("invalid number of groups returned for user, expected 2; got '%v'", len(user.GetSpec().GetGroups()))
	}
	if len(user.GetSpec().GetProjectNamespaceRoles()) != 6 {
		t.Errorf("invalid number of roles returned for user, expected 3; got '%v'", len(user.GetSpec().GetProjectNamespaceRoles()))
	}
	if user.GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != "ns" {
		t.Errorf("invalid namespace in role returned for user, expected ns; got '%v'", user.GetSpec().GetProjectNamespaceRoles()[2].Namespace)
	}
	performBasicAuthProviderChecks(t, *ap, 0, 0, 0, 0)
}

func TestUserGetInfo(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), false)

	uuuid := uuid.New().String()
	fakeuuuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	guuid := uuid.New().String()
	ruuid := uuid.New().String()
	pruuid := uuid.New().String()
	fakescope := uuid.New().String()

	mock.ExpectQuery(`SELECT "identities"."id", "identities"."schema_id", .*WHERE .traits ->> 'email' = 'user-` + uuuid + `'.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)))
	mock.ExpectQuery(`SELECT "group"."id".* FROM "authsrv_group" AS "group" JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id WHERE .authsrv_groupaccount.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("group-" + guuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_group.name as group FROM "authsrv_grouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_grouprole.role_id JOIN authsrv_group ON authsrv_group.id=authsrv_grouprole.group_id WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "group"}).AddRow("role-"+ruuid, "group-"+guuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace, authsrv_group.name as group FROM "authsrv_projectgroupnamespacerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_accountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id WHERE .authsrv_accountresourcerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id WHERE .authsrv_projectaccountresourcerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id WHERE .authsrv_projectaccountnamespacerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, "ns"))
	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."scope" FROM "authsrv_resourcerole" AS "resourcerole" WHERE .name = 'role-` + ruuid + `'. AND .trash = FALSE.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "scope", "name"}).AddRow(ruuid, fakescope, "role-"+ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcepermission.name as name FROM "authsrv_resourcepermission" JOIN authsrv_resourcerolepermission ON authsrv_resourcerolepermission.resource_permission_id=authsrv_resourcepermission.id WHERE .authsrv_resourcerolepermission.resource_role_id = '` + ruuid + `'. AND .authsrv_resourcepermission.trash = FALSE. AND .authsrv_resourcerolepermission.trash = FALSE.`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("account.read").AddRow("account.write"))

	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + fakeuuuid},
	}
	ctx := context.WithValue(context.Background(), common.SessionDataKey, &v3.SessionData{Username: "user-" + uuuid})
	userinfo, err := us.GetUserInfo(ctx, user)

	if err != nil {
		t.Fatal("could not get user:", err)
	}

	if userinfo.Metadata.Name != "johndoe@provider.com" {
		t.Errorf("incorrect username; expected '%v', got '%v'", "johndoe@provider.com", userinfo.Metadata.Name)
	}
	if userinfo.Spec.FirstName != "John" {
		t.Errorf("incorrect first name; expected '%v', got '%v'", "John", userinfo.Spec.FirstName)
	}
	if userinfo.Spec.LastName != "Doe" {
		t.Errorf("incorrect last name; expected '%v', got '%v'", "Doe", userinfo.Spec.LastName)
	}
	if len(userinfo.Spec.Groups) != 1 {
		t.Errorf("incorrect number of groups; expected '%v', got '%v'", 1, len(userinfo.Spec.Groups))
	}
	if userinfo.Spec.Groups[0] != "group-"+guuid {
		t.Errorf("incorrect group name; expected '%v', got '%v'", "group-"+guuid, userinfo.Spec.Groups[0])
	}
	if len(userinfo.Spec.Permissions) != 6 {
		t.Errorf("incorrect number of permissions; expected '%v', got '%v'", 6, len(userinfo.Spec.Permissions))
	}
	if len(userinfo.Spec.Permissions[0].Permissions) != 2 {
		t.Errorf("incorrect number of permissions; expected '%v', got '%v'", 2, len(userinfo.Spec.Permissions[0].Permissions))
	}
	if len(*userinfo.Spec.Permissions[0].Scope) == 0 {
		t.Errorf("incorrect scope for permissions; expected '%v', got '%v'", fakescope, *userinfo.Spec.Permissions[0].Scope)
	}

}

func TestUserGetById(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	uuuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	pruuid := uuid.New().String()

	// lookup by id
	mock.ExpectQuery(`SELECT "identities"."id",.* FROM "identities" WHERE .*id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)))

	guuid := addUsersGroupFetchExpectation(mock, uuuid)
	addGroupRoleMappingsFetchExpectation(mock, guuid, pruuid)
	addUserRoleMappingsFetchExpectation(mock, uuuid, pruuid)

	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Id: uuuid},
	}
	user, err := us.GetByID(context.Background(), user)
	if err != nil {
		t.Fatal("could not get user:", err)
	}
	performUserBasicChecks(t, user, uuuid)
	if len(user.GetSpec().GetGroups()) != 1 {
		t.Errorf("invalid number of groups returned for user, expected 1; got '%v'", len(user.GetSpec().GetGroups()))
	}
	if len(user.GetSpec().GetProjectNamespaceRoles()) != 6 {
		t.Errorf("invalid number of roles returned for user, expected 6; got '%v'", len(user.GetSpec().GetProjectNamespaceRoles()))
	}
	if user.GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != "ns" {
		t.Errorf("invalid namespace in role returned for user, expected ns; got '%v'", user.GetSpec().GetProjectNamespaceRoles()[2].Namespace)
	}

	performBasicAuthProviderChecks(t, *ap, 0, 0, 0, 0)
}

func TestUserList(t *testing.T) {
	authenticated := time.Date(2022, 11, 1, 17, 49, 0, 0, time.UTC)
	tests := []struct {
		name     string
		q        string
		limit    int64
		offset   int64
		orderBy  string
		order    string
		role     string
		group    string
		projects []string
		utype    string
	}{
		{"simple list", "", 50, 20, "", "", "", "", []string{}, ""},
		{"simple list with type", "", 50, 20, "", "", "", "", []string{}, "password"},
		{"sorted list", "", 50, 20, "email", "asc", "", "", []string{}, ""},
		{"sorted list with ALL projects", "", 50, 20, "email", "asc", "", "", []string{"ALL"}, ""},
		{"sorted list with single project", "", 50, 20, "email", "asc", "", "", []string{"project1"}, ""},
		{"sorted list with projects", "", 50, 20, "email", "asc", "", "", []string{"project1", "project2"}, ""},
		{"sorted list without dir", "", 50, 20, "email", "", "", "", []string{}, ""},
		{"sorted list with q", "filter-query", 50, 20, "email", "asc", "", "", []string{}, ""},
		{"sorted list with role", "", 50, 20, "email", "asc", "role-name", "", []string{}, ""},
		{"sorted list with role and group", "", 50, 20, "email", "asc", "role-name", "group-name", []string{}, ""},
		{"sorted list with q and role", "filter-query", 50, 20, "email", "asc", "role-name", "", []string{}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			ap := &mockAuthProvider{}
			mazc := mockAuthzClient{}
			us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

			uuuid1 := uuid.New().String()
			uuuid2 := uuid.New().String()
			pruuid := uuid.New().String()

			puuid, ouuid := addParterOrgFetchExpectation(mock)
			q := ""
			if tc.q != "" {
				q = ` AND ..traits ->> 'email' ILIKE '%` + tc.q + `%'. OR .traits ->> 'first_name' ILIKE '%` + tc.q + `%'. OR .traits ->> 'last_name' ILIKE '%` + tc.q + `%'.. `
			}
			order := ""
			if tc.orderBy != "" {
				order = `ORDER BY "traits ->> '` + tc.orderBy + `' `
			}
			if tc.order != "" {
				order = order + tc.order + `" `
			}
			if tc.role != "" {
				addFetchExpectation(mock, "resourcerole")
			}
			if tc.group != "" {
				addFetchExpectation(mock, "group")
			}
			for _, p := range tc.projects {
				if p == "ALL" {
					continue
				}
				addFetchIdByNameExpectation(mock, "project", p)
			}
			if tc.role != "" || tc.group != "" || len(tc.projects) != 0 {
				addSentryLookupExpectation(mock, []string{uuuid1, uuuid2}, puuid, ouuid)
				mock.ExpectQuery(`SELECT "identities"."id", .*WHERE .identities.id IN .'` + uuuid1 + `', '` + uuuid2 + `'..  ` + q + order + `LIMIT ` + fmt.Sprint(tc.limit) + ` OFFSET ` + fmt.Sprint(tc.offset)).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).
					AddRow(uuuid1, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)).
					AddRow(uuuid2, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)))
			} else {
				if tc.utype != "" {
					mock.ExpectQuery(`SELECT "identities"."id", .*, .*FROM "identities" LEFT JOIN "identity_credentials" AS "identity_credential" ON ."identity_credential"."identity_id" = "identities"."id". LEFT JOIN "identity_credential_types" AS "identity_credential__identity_credential_type" ON ."identity_credential__identity_credential_type"."id" = "identity_credential"."identity_credential_type_id". WHERE .name = '` + tc.utype + `'. LIMIT 50 OFFSET 20 *`).
						WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).
						AddRow(uuuid1, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)).
						AddRow(uuuid2, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)))
				} else {
					mock.ExpectQuery(`SELECT "identities"."id".* LIMIT 50 OFFSET 20$`).
						WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).
						AddRow(uuuid1, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)).
						AddRow(uuuid2, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)))
				}
			}

			guuid := addUsersGroupFetchExpectation(mock, uuuid1)
			addGroupRoleMappingsFetchExpectation(mock, guuid, pruuid)
			addUserRoleMappingsFetchExpectation(mock, uuuid1, pruuid)
			mock.ExpectQuery(`select .* from sessions where .*`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"max"}).
				AddRow(authenticated))

			guuid = addUsersGroupFetchExpectation(mock, uuuid2)
			addGroupRoleMappingsFetchExpectation(mock, guuid, pruuid)
			addUserRoleMappingsFetchExpectation(mock, uuuid2, pruuid)
			mock.ExpectQuery(`select .* from sessions where .*`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"max"}).
				AddRow(authenticated))

			qo := &v3.QueryOptions{
				Q:            tc.q,
				Limit:        tc.limit,
				Offset:       tc.offset,
				OrderBy:      tc.orderBy,
				Order:        tc.order,
				Organization: ouuid,
				Partner:      puuid,
				Role:         tc.role,
				Group:        tc.group,
				Type:         tc.utype,
				Project:      strings.Join(tc.projects, ","),
			}

			userlist, err := us.List(context.Background(), query.WithOptions(qo))
			if err != nil {
				t.Fatal("could not list users:", err)
			}
			if userlist.Metadata.Count != 2 {
				t.Fatalf("incorrect number of users returned, expected 2; got %v", userlist.Metadata.Count)
			}
			if userlist.Items[0].Metadata.Name != "johndoe@provider.com" || userlist.Items[1].Metadata.Name != "johndoe@provider.com" {
				t.Errorf("incorrect user names returned when listing; expected '%v' and '%v'; got '%v' and '%v'", "johndoe@provider.com", "johndoe@provider.com", userlist.Items[0].Metadata.Name, userlist.Items[1].Metadata.Name)
			}
			if len(userlist.Items[0].GetSpec().GetGroups()) != 1 {
				t.Errorf("invalid number of groups returned for user, expected 1; got '%v'", len(userlist.Items[0].GetSpec().GetGroups()))
			}

			if len(userlist.Items[0].GetSpec().GetProjectNamespaceRoles()) != 6 {
				t.Errorf("invalid number of roles returned for user, expected 6; got '%v'", len(userlist.Items[0].GetSpec().GetProjectNamespaceRoles()))
			}
			if userlist.Items[0].GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != "ns" {
				t.Errorf("invalid namespace in role returned for user, expected ns; got '%v'", userlist.Items[0].GetSpec().GetProjectNamespaceRoles()[2].Namespace)
			}

			performBasicAuthProviderChecks(t, *ap, 0, 0, 0, 0)

		})
	}
}

func TestUserDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	uuuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "identities"."id" FROM "identities" WHERE .*traits ->> 'email' = 'user-` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	mock.ExpectBegin()
	_ = addUserRoleMappingsUpdateExpectation(mock, uuuid)
	// User delete is via kratos
	addUserGroupMappingsUpdateExpectation(mock, uuuid)
	mock.ExpectCommit()

	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
	}
	ctx := context.WithValue(context.Background(), common.SessionDataKey, &v3.SessionData{Username: "not-user-" + uuuid})
	_, err := us.Delete(ctx, user)
	if err != nil {
		t.Fatal("could not delete user:", err)
	}

	performBasicAuthProviderChecks(t, *ap, 0, 0, 0, 1)
}
func TestUserDeleteSelf(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	uuuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "identities"."id" FROM "identities" WHERE .*traits ->> 'email' = 'user-` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	mock.ExpectBegin()
	_ = addUserRoleMappingsUpdateExpectation(mock, uuuid)
	// User delete is via kratos
	addUserGroupMappingsUpdateExpectation(mock, uuuid)
	mock.ExpectCommit()

	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
	}
	ctx := context.WithValue(context.Background(), common.SessionDataKey, &v3.SessionData{Username: "user-" + uuuid})
	_, err := us.Delete(ctx, user)
	if err == nil {
		t.Fatal("user able to delete their own account")
	}
}

func TestUserForgotPassword(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)

	uuuid := addUserFetchExpectation(mock)

	fpreq := &userrpcv3.UserForgotPasswordRequest{Username: "user-" + uuuid}
	fpresp, err := us.ForgotPassword(context.Background(), fpreq)
	if err != nil {
		t.Fatal("could not fetch password recovery link:", err)
	}
	if !strings.HasPrefix(fpresp.RecoveryLink, "https://recoverme.testing/") {
		t.Error("invalid recovery url generated")
	}
}

func TestUserRetrieveCliConfigGet(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	ks := NewApiKeyService(db, getLogger())
	us := NewUserService(ap, db, &mazc, ks, common.CliConfigDownloadData{}, getLogger(), true)

	uuuid := uuid.NewString()
	auuid := uuid.NewString()
	mock.ExpectQuery(`SELECT sap.* FROM "sentry_account_permission" AS "sap" JOIN authsrv_project as proj ON \(proj.id = sap.project_id\) AND \(proj.default = TRUE\) WHERE \(account_id = '` + uuuid + `'\) LIMIT 1`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(uuuid))
	_ = addFetchExpectation(mock, "project")
	_ = addFetchExpectation(mock, "organization")
	_ = addFetchExpectation(mock, "partner")

	mock.ExpectQuery(`SELECT "apikey"."id", "apikey"."name",.* FROM "authsrv_apikey" AS "apikey" WHERE \(name = 'user-` + uuuid + `'\) AND \(trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"key", "secret"}).AddRow(auuid, "apikey-"+auuid))

	req := &userrpcv3.ApiKeyRequest{Username: "user-" + uuuid, Id: uuuid}
	resp, err := us.RetrieveCliConfig(context.Background(), req)
	if err != nil {
		t.Fatal("could not fetch cli config:", err)
	}
	if resp.ApiKey != auuid {
		t.Error("incorrect apikey generated")
	}
	if resp.ApiSecret != "apikey-"+auuid {
		t.Error("incorrect apisecret generated")
	}
	if resp.Project != "project-name" {
		t.Error("invalid project name")
	}
	if resp.Organization != "organization-name" {
		t.Error("invalid organization name")
	}
	if resp.Partner != "partner-name" {
		t.Error("invalid partner name")
	}
}

func TestUserRetrieveCliConfigCreate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	ks := NewApiKeyService(db, getLogger())
	us := NewUserService(ap, db, &mazc, ks, common.CliConfigDownloadData{}, getLogger(), true)

	uuuid := uuid.NewString()
	auuid := uuid.NewString()
	mock.ExpectQuery(`SELECT sap.* FROM "sentry_account_permission" AS "sap" JOIN authsrv_project as proj ON \(proj.id = sap.project_id\) AND \(proj.default = TRUE\) WHERE \(account_id = '` + uuuid + `'\) LIMIT 1`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(uuuid))
	_ = addFetchExpectation(mock, "project")
	_ = addFetchExpectation(mock, "organization")
	_ = addFetchExpectation(mock, "partner")

	mock.ExpectQuery(`SELECT "apikey"."id", "apikey"."name",.* FROM "authsrv_apikey" AS "apikey" WHERE \(name = 'user-` + uuuid + `'\) AND \(trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	mock.ExpectQuery(`INSERT INTO "authsrv_apikey"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(auuid))

	req := &userrpcv3.ApiKeyRequest{Username: "user-" + uuuid, Id: uuuid}
	resp, err := us.RetrieveCliConfig(context.Background(), req)
	if err != nil {
		t.Fatal("could not fetch cli config:", err)
	}
	if len(resp.ApiKey) == 0 {
		t.Error("no apikey generated")
	}
	if len(resp.ApiSecret) == 0 {
		t.Error("no apisecret generated")
	}
	if resp.Project != "project-name" {
		t.Error("invalid project name")
	}
	if resp.Organization != "organization-name" {
		t.Error("invalid organization name")
	}
	if resp.Partner != "partner-name" {
		t.Error("invalid partner name")
	}
}

func TestCreateLoginAuditLog(t *testing.T) {
	tt := []struct {
		name            string
		uuid            string
		invalid         bool
		shouldHaveError bool
	}{
		{"invalid uid format", "user-" + uuid.New().String(), false, true},
		{"invalid user id", uuid.New().String(), true, true},
		{"valid user id", uuid.New().String(), false, false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			ap := &mockAuthProvider{}
			mazc := mockAuthzClient{}
			us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{}, getLogger(), true)
			if tc.invalid {

				uid := uuid.New().String()
				// without regexp QuoteMeta, getting mismatch actual and required SQL queries
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT traits ->> 'email' as name FROM "identities" WHERE (id = ('` + uid + `'))`)).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"traits"}).AddRow([]byte(`{"email":"johndoe@provider.com"}`)))

			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT traits ->> 'email' as name FROM "identities" WHERE (id = ('` + tc.uuid + `'))`)).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"traits"}).AddRow([]byte(`{"email":"johndoe@provider.com"}`)))

			}

			audreq := &userrpcv3.UserLoginAuditRequest{UserId: tc.uuid}
			_, err := us.CreateLoginAuditLog(context.TODO(), audreq)
			if tc.shouldHaveError && err == nil {

				t.Error("could not add audit log", err)
			}
		})
	}

}
