package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/RafayLabs/rcloud-base/pkg/common"
	"github.com/RafayLabs/rcloud-base/pkg/query"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	v3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	userv3 "github.com/RafayLabs/rcloud-base/proto/types/userpb/v3"
	"github.com/google/uuid"
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
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{})

	uuuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))

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
	performBasicAuthProviderChecks(t, *ap, 1, 0, 1, 0)
}

func TestCreateUserWithRole(t *testing.T) {
	pruuid := uuid.New().String()
	prname := "project-" + pruuid
	ruuid := uuid.New().String()
	rname := "project-" + ruuid
	var namespaceid int64 = 7
	tt := []struct {
		name       string
		roles      []*userv3.ProjectNamespaceRole
		dbname     string
		scope      string
		shouldfail bool
	}{
		{"just role", []*userv3.ProjectNamespaceRole{{Role: rname}}, "authsrv_accountresourcerole", "system", false},
		{"just role org scope", []*userv3.ProjectNamespaceRole{{Role: rname}}, "authsrv_accountresourcerole", "organization", false},
		{"just project", []*userv3.ProjectNamespaceRole{{Project: &prname}}, "authsrv_accountrole", "system", true},                                   // no role creation without role
		{"just namespace", []*userv3.ProjectNamespaceRole{{Namespace: &namespaceid}}, "authsrv_accountrole", "system", true},                          // no role creation without role,
		{"project and namespace", []*userv3.ProjectNamespaceRole{{Project: &prname, Namespace: &namespaceid}}, "authsrv_accountrole", "system", true}, // no role creation without role,
		{"project and role", []*userv3.ProjectNamespaceRole{{Project: &prname, Role: rname}}, "authsrv_projectaccountresourcerole", "project", false},
		{"project role namespace", []*userv3.ProjectNamespaceRole{{Project: &prname, Namespace: &namespaceid, Role: rname}}, "authsrv_projectaccountresourcerole", "project", false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			ap := &mockAuthProvider{}
			mazc := mockAuthzClient{}
			us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{})

			uuuid := uuid.New().String()
			puuid := uuid.New().String()
			pruuid := uuid.New().String()
			ouuid := uuid.New().String()

			mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
			mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))

			mock.ExpectBegin()
			mock.ExpectQuery(`SELECT "resourcerole"."id".* FROM "authsrv_resourcerole" AS "resourcerole"`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "scope"}).AddRow(pruuid, "role-name", tc.scope))
			if tc.roles[0].Project != nil {
				mock.ExpectQuery(`SELECT "project"."id" FROM "authsrv_project" AS "project"`).
					WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
			}
			mock.ExpectQuery(fmt.Sprintf(`INSERT INTO "%v"`, tc.dbname)).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
			mock.ExpectCommit()

			user := &userv3.User{
				Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
				Spec:     &userv3.UserSpec{ProjectNamespaceRoles: tc.roles},
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

			performBasicAuthProviderChecks(t, *ap, 1, 0, 1, 0)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{})

	uuuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	pruuid := uuid.New().String()
	prname := "project-" + pruuid
	ruuid := uuid.New().String()
	rname := "project-" + ruuid
	var namespaceid int64 = 7

	// performing update
	mock.ExpectQuery(`SELECT "identities"."id" FROM "identities" WHERE .*traits ->> 'email' = 'user-` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	mock.ExpectQuery(`SELECT "partner"."id" FROM "authsrv_partner" AS "partner"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "organization"."id" FROM "authsrv_organization" AS "organization"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "authsrv_accountresourcerole" AS "accountresourcerole" SET trash = TRUE WHERE`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "authsrv_projectaccountresourcerole" AS "projectaccountresourcerole" SET trash = TRUE WHERE`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "authsrv_projectaccountnamespacerole" AS "projectaccountnamespacerole" SET trash = TRUE WHERE`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT "resourcerole"."id".* FROM "authsrv_resourcerole" AS "resourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "scope"}).AddRow(pruuid, "role-name", "project"))
	mock.ExpectQuery(`SELECT "project"."id" FROM "authsrv_project" AS "project"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pruuid))
	mock.ExpectQuery(`INSERT INTO "authsrv_projectaccountresourcerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectCommit()

	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
		Spec:     &userv3.UserSpec{ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{{Project: &prname, Namespace: &namespaceid, Role: rname}}},
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

func TestUserGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{})

	uuuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	guuid := uuid.New().String()
	ruuid := uuid.New().String()
	pruuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "identities"."id", .* FROM "identities" WHERE .*traits ->> 'email' = 'user-` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)))
	mock.ExpectQuery(`SELECT "group"."id".* FROM "authsrv_group" AS "group" JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id WHERE .authsrv_groupaccount.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("group-" + guuid).AddRow("group2-" + guuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_accountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id WHERE .authsrv_accountresourcerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id WHERE .authsrv_projectaccountresourcerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id WHERE .authsrv_projectaccountnamespacerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, 9))

	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
	}
	user, err := us.GetByName(context.Background(), user)
	if err != nil {
		t.Fatal("could not get user:", err)
	}
	performUserBasicChecks(t, user, uuuid)
	if user.GetMetadata().GetName() != "johndoe@provider.com" {
		t.Errorf("invalid email for user, expected johndoe@provider.com; got '%v'", user.GetMetadata().GetName())
	}
	if len(user.GetSpec().GetGroups()) != 2 {
		t.Errorf("invalid number of groups returned for user, expected 2; got '%v'", len(user.GetSpec().GetGroups()))
	}
	if len(user.GetSpec().GetProjectNamespaceRoles()) != 3 {
		t.Errorf("invalid number of roles returned for user, expected 3; got '%v'", len(user.GetSpec().GetProjectNamespaceRoles()))
	}
	if user.GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != 9 {
		t.Errorf("invalid namespace in role returned for user, expected 9; got '%v'", user.GetSpec().GetProjectNamespaceRoles()[2].Namespace)
	}
	performBasicAuthProviderChecks(t, *ap, 0, 0, 0, 0)
}

func TestUserGetById(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{})

	uuuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	guuid := uuid.New().String()
	ruuid := uuid.New().String()
	pruuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "identities"."id",.* FROM "identities" WHERE .*id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)))
	mock.ExpectQuery(`SELECT "group"."id".* FROM "authsrv_group" AS "group" JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id WHERE .authsrv_groupaccount.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("group-" + guuid).AddRow("group2-" + guuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_accountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id WHERE .authsrv_accountresourcerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id WHERE .authsrv_projectaccountresourcerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id WHERE .authsrv_projectaccountnamespacerole.account_id = '` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, 9))

	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Id: uuuid},
	}
	user, err := us.GetByID(context.Background(), user)
	if err != nil {
		t.Fatal("could not get user:", err)
	}
	performUserBasicChecks(t, user, uuuid)
	if len(user.GetSpec().GetGroups()) != 2 {
		t.Errorf("invalid number of groups returned for user, expected 2; got '%v'", len(user.GetSpec().GetGroups()))
	}
	if len(user.GetSpec().GetProjectNamespaceRoles()) != 3 {
		t.Errorf("invalid number of roles returned for user, expected 3; got '%v'", len(user.GetSpec().GetProjectNamespaceRoles()))
	}
	if user.GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != 9 {
		t.Errorf("invalid namespace in role returned for user, expected 9; got '%v'", user.GetSpec().GetProjectNamespaceRoles()[2].Namespace)
	}

	performBasicAuthProviderChecks(t, *ap, 0, 0, 0, 0)
}

func TestUserList(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{})

	uuuid1 := uuid.New().String()
	uuuid2 := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	guuid := uuid.New().String()
	ruuid := uuid.New().String()
	pruuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "identities"."id",.* FROM "identities"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).
		AddRow(uuuid1, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)).
		AddRow(uuuid2, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)))

	mock.ExpectQuery(`SELECT "group"."id".* FROM "authsrv_group" AS "group" JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id WHERE .authsrv_groupaccount.account_id = '` + uuuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("group-" + guuid).AddRow("group2-" + guuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_accountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id WHERE .authsrv_accountresourcerole.account_id = '` + uuuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id WHERE .authsrv_projectaccountresourcerole.account_id = '` + uuuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id WHERE .authsrv_projectaccountnamespacerole.account_id = '` + uuuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, 9))

	mock.ExpectQuery(`SELECT "group"."id".* FROM "authsrv_group" AS "group" JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id WHERE .authsrv_groupaccount.account_id = '` + uuuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("group-" + guuid).AddRow("group2-" + guuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_accountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id WHERE .authsrv_accountresourcerole.account_id = '` + uuuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id WHERE .authsrv_projectaccountresourcerole.account_id = '` + uuuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id WHERE .authsrv_projectaccountnamespacerole.account_id = '` + uuuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, 9))

	qo := &commonv3.QueryOptions{}
	userlist, err := us.List(context.Background(), query.WithOptions(qo))
	if err != nil {
		t.Fatal("could not list users:", err)
	}
	if userlist.Metadata.Count != 2 {
		t.Errorf("incorrect number of users returned, expected 2; got %v", userlist.Metadata.Count)
	}
	if userlist.Items[0].Metadata.Name != "johndoe@provider.com" || userlist.Items[1].Metadata.Name != "johndoe@provider.com" {
		t.Errorf("incorrect user names returned when listing; expected '%v' and '%v'; got '%v' and '%v'", "johndoe@provider.com", "johndoe@provider.com", userlist.Items[0].Metadata.Name, userlist.Items[1].Metadata.Name)
	}
	if len(userlist.Items[0].GetSpec().GetGroups()) != 2 {
		t.Errorf("invalid number of groups returned for user, expected 2; got '%v'", len(userlist.Items[0].GetSpec().GetGroups()))
	}

	if len(userlist.Items[0].GetSpec().GetProjectNamespaceRoles()) != 3 {
		t.Errorf("invalid number of roles returned for user, expected 3; got '%v'", len(userlist.Items[0].GetSpec().GetProjectNamespaceRoles()))
	}
	if userlist.Items[0].GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != 9 {
		t.Errorf("invalid namespace in role returned for user, expected 9; got '%v'", userlist.Items[0].GetSpec().GetProjectNamespaceRoles()[2].Namespace)
	}

	performBasicAuthProviderChecks(t, *ap, 0, 0, 0, 0)
}
func TestUserFiletered(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{})

	uuuid1 := uuid.New().String()
	uuuid2 := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	guuid := uuid.New().String()
	ruuid := uuid.New().String()
	pruuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "identities"."id", .*WHERE .traits ->> 'email' ILIKE '%filter-query%'. OR .traits ->> 'first_name' ILIKE '%filter-query%'. OR .traits ->> 'last_name' ILIKE '%filter-query%'. ORDER BY "traits ->> 'email' asc" LIMIT 50 OFFSET 20`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).
		AddRow(uuuid1, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)).
		AddRow(uuuid2, []byte(`{"email":"johndoe@provider.com", "first_name": "John", "last_name": "Doe", "organization_id": "`+ouuid+`", "partner_id": "`+puuid+`", "description": "My awesome user"}`)))

	mock.ExpectQuery(`SELECT "group"."id".* FROM "authsrv_group" AS "group" JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id WHERE .authsrv_groupaccount.account_id = '` + uuuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("group-" + guuid).AddRow("group2-" + guuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_accountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id WHERE .authsrv_accountresourcerole.account_id = '` + uuuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id WHERE .authsrv_projectaccountresourcerole.account_id = '` + uuuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id WHERE .authsrv_projectaccountnamespacerole.account_id = '` + uuuid1 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, 9))

	mock.ExpectQuery(`SELECT "group"."id".* FROM "authsrv_group" AS "group" JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id WHERE .authsrv_groupaccount.account_id = '` + uuuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"name"}).
		AddRow("group-" + guuid).AddRow("group2-" + guuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role FROM "authsrv_accountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id WHERE .authsrv_accountresourcerole.account_id = '` + uuuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("role-" + ruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id WHERE .authsrv_projectaccountresourcerole.account_id = '` + uuuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("role-"+ruuid, "project-"+pruuid))
	mock.ExpectQuery(`SELECT authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id WHERE .authsrv_projectaccountnamespacerole.account_id = '` + uuuid2 + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "namespace"}).AddRow("role-"+ruuid, "project-"+pruuid, 9))

	qo := &commonv3.QueryOptions{Q: "filter-query", Limit: 50, Offset: 20, OrderBy: "email", Order: "asc"}
	userlist, err := us.List(context.Background(), query.WithOptions(qo))
	if err != nil {
		t.Fatal("could not list users:", err)
	}
	if userlist.Metadata.Count != 2 {
		t.Errorf("incorrect number of users returned, expected 2; got %v", userlist.Metadata.Count)
	}

	if userlist.Items[0].Metadata.Name != "johndoe@provider.com" {
		t.Errorf("incorrect user names returned when listing; expected '%v' and '%v'; got '%v' and '%v'", "johndoe@provider.com", "johndoe@provider.com", userlist.Items[0].Metadata.Name, userlist.Items[1].Metadata.Name)
	}

	if len(userlist.Items[0].GetSpec().GetGroups()) != 2 {
		t.Errorf("invalid number of groups returned for user, expected 2; got '%v'", len(userlist.Items[0].GetSpec().GetGroups()))
	}

	if len(userlist.Items[0].GetSpec().GetProjectNamespaceRoles()) != 3 {
		t.Errorf("invalid number of roles returned for user, expected 3; got '%v'", len(userlist.Items[0].GetSpec().GetProjectNamespaceRoles()))
	}
	if userlist.Items[0].GetSpec().GetProjectNamespaceRoles()[2].GetNamespace() != 9 {
		t.Errorf("invalid namespace in role returned for user, expected 9; got '%v'", userlist.Items[0].GetSpec().GetProjectNamespaceRoles()[2].Namespace)
	}

	performBasicAuthProviderChecks(t, *ap, 0, 0, 0, 0)
}

func TestUserDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ap := &mockAuthProvider{}
	mazc := mockAuthzClient{}
	us := NewUserService(ap, db, &mazc, nil, common.CliConfigDownloadData{})

	uuuid := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "identities"."id" FROM "identities" WHERE .*traits ->> 'email' = 'user-` + uuuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuuid, []byte(`{"email":"johndoe@provider.com"}`)))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "authsrv_accountresourcerole" AS "accountresourcerole" SET trash = TRUE WHERE`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "authsrv_projectaccountresourcerole" AS "projectaccountresourcerole" SET trash = TRUE WHERE`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "authsrv_projectaccountnamespacerole" AS "projectaccountnamespacerole" SET trash = TRUE WHERE`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// User delete is via kratos
	mock.ExpectExec(`UPDATE "authsrv_groupaccount" AS "groupaccount" SET trash = TRUE WHERE`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user := &userv3.User{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
	}
	_, err := us.Delete(context.Background(), user)
	if err != nil {
		t.Fatal("could not delete user:", err)
	}

	performBasicAuthProviderChecks(t, *ap, 0, 0, 0, 1)
}
