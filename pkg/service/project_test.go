package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/common"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
)

func performProjectBasicChecks(t *testing.T, project *systemv3.Project, puuid string) {
	if project.GetMetadata().GetName() != "project-"+puuid {
		t.Error("invalid name returned")
	}
}

func TestCreateProject(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	puuid := uuid.New().String()

	addFetchExpectation(mock, "organization")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "authsrv_project"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectCommit()

	project := &systemv3.Project{
		Metadata: &v3.Metadata{Id: puuid, Name: "project-" + puuid, Organization: "orgname"},
		Spec:     &systemv3.ProjectSpec{},
	}
	project, err := ps.Create(context.Background(), project)
	if err != nil {
		t.Fatal("could not create project:", err)
	}
	performProjectBasicChecks(t, project, puuid)
}

func TestCreateProjectDuplicate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	gs := NewProjectService(db, &mazc, getLogger(), true)

	puuid := uuid.New().String()

	project := &systemv3.Project{
		Metadata: &v3.Metadata{Id: puuid, Name: "project-" + puuid},
		Spec:     &systemv3.ProjectSpec{},
	}

	// Try to recreate
	mock.ExpectQuery(`INSERT INTO "authsrv_project"`).
		WithArgs().WillReturnError(fmt.Errorf("unique constraint violation"))
	_, err := gs.Create(context.Background(), project)
	if err == nil {
		t.Fatal("should not be able to recreate project with same name")
	}
}

func TestCreateProjectFull(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	puuid := uuid.New().String()

	addFetchExpectation(mock, "organization")
	addFetchEmptyExpecteation(mock, "project") // not existing project
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "authsrv_project"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", "resourcerole"."description", "resourcerole"."created_at", "resourcerole"."modified_at", "resourcerole"."trash", "resourcerole"."organization_id", "resourcerole"."partner_id", "resourcerole"."is_global", "resourcerole"."builtin", "resourcerole"."scope" FROM "authsrv_resourcerole" AS "resourcerole" WHERE \(name = 'test-role'\) AND \(trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "scope"}).AddRow(puuid, "resourcerole-"+puuid, "namespace"))
	addFetchExpectation(mock, "group")
	mock.ExpectQuery(`INSERT INTO "authsrv_projectgroupnamespacerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.NewString()))
	mock.ExpectQuery(`SELECT "identities"."id".* FROM "identities" WHERE .*traits ->> 'email' = 'test-user'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "traits"}).AddRow(uuid.NewString(), []byte(`{"email":"test-user", "first_name": "John", "last_name": "Doe", "description": "The OG user."}`)))
	mock.ExpectQuery(`SELECT "resourcerole"."id", "resourcerole"."name", "resourcerole"."description", "resourcerole"."created_at", "resourcerole"."modified_at", "resourcerole"."trash", "resourcerole"."organization_id", "resourcerole"."partner_id", "resourcerole"."is_global", "resourcerole"."builtin", "resourcerole"."scope" FROM "authsrv_resourcerole" AS "resourcerole" WHERE \(name = 'test-role'\) AND \(trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "scope"}).AddRow(puuid, "resourcerole-"+puuid, "namespace"))
	mock.ExpectQuery(`INSERT INTO "authsrv_projectaccountnamespacerole"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.NewString()))
	mock.ExpectCommit()

	te := []string{"test-project", "test-namespace", "test-group", "test-role", "test-user"}
	project := &systemv3.Project{
		Metadata: &v3.Metadata{Id: puuid, Name: "project-" + puuid, Organization: "orgname"},
		Spec: &systemv3.ProjectSpec{
			ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{
				{Project: &te[0], Namespace: &te[1], Group: &te[2], Role: te[3]},
			},
			UserRoles: []*userv3.UserRole{
				{User: te[4], Role: te[3]},
			},
		},
	}
	project, err := ps.Create(context.Background(), project)
	if err != nil {
		t.Fatal("could not create project:", err)
	}
	performProjectBasicChecks(t, project, puuid)
}

func TestProjectDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id", "project"."name", .* FROM "authsrv_project" AS "project" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "project-"+puuid))
	mock.ExpectBegin()
	// return empty rows
	mock.ExpectQuery(`SELECT "cluster"."id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectExec(`UPDATE "authsrv_projectgrouprole" AS "projectgrouprole" SET trash = TRUE WHERE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`UPDATE "authsrv_projectgroupnamespacerole" AS "projectgroupnamespacerole" SET trash = TRUE WHERE ."project_id" = '` + puuid + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectExec(`UPDATE "authsrv_projectaccountresourcerole" AS "projectaccountresourcerole" SET trash = TRUE WHERE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`UPDATE "authsrv_projectaccountnamespacerole" AS "projectaccountnamespacerole" SET trash = TRUE WHERE ."project_id" = '` + puuid + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))

	mock.ExpectExec(`UPDATE "authsrv_project"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	project := &systemv3.Project{
		Metadata: &v3.Metadata{Id: puuid, Name: "project-" + puuid},
	}
	_, err := ps.Delete(context.Background(), project)
	if err != nil {
		t.Fatal("could not delete project:", err)
	}
}

func TestNonEmptyProjectDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	puuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id", "project"."name", .* FROM "authsrv_project" AS "project" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "project-"+puuid))

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT "cluster"."id", "cluster"."organization_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cuuid))

	mock.ExpectRollback()
	mock.ExpectQuery(`SELECT "projectcluster"."cluster_id", "projectcluster"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cuuid))

	project := &systemv3.Project{
		Metadata: &v3.Metadata{Id: puuid, Name: "project-" + puuid},
	}
	_, err := ps.Delete(context.Background(), project)
	if err == nil {
		t.Fatal("non empty project deleted:", err)
	}
}

func TestProjectDeleteNonExist(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	puuid := uuid.New().String()

	addFailingFetchExpecteation(mock, "project")

	project := &systemv3.Project{
		Metadata: &v3.Metadata{Id: puuid, Name: "project-" + puuid},
	}
	_, err := ps.Delete(context.Background(), project)
	if err == nil {
		t.Fatal("deleted non existent project")
	}
}

func TestProjectGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	puuid := uuid.New().String()
	addFetchExpectation(mock, "project")
	addFetchExpectation(mock, "organization")
	addFetchExpectation(mock, "partner")

	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group 
		FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id 
		JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id 
		JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id WHERE`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("ADMIN"))

	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group, 
		namespace FROM "authsrv_projectgroupnamespacerole" JOIN authsrv_resourcerole 
		ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id JOIN authsrv_project 
		ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id JOIN authsrv_group 
		ON authsrv_group.id=authsrv_projectgroupnamespacerole.group_id WHERE`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("ADMIN"))

	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user 
		FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id 
		JOIN identities ON identities.id=authsrv_projectaccountresourcerole.account_id WHERE`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("ADMIN"))

	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user, namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole 
		ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN identities 
		ON identities.id=authsrv_projectaccountnamespacerole.account_id WHERE`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("ADMIN"))

	project := &systemv3.Project{
		Metadata: &v3.Metadata{Id: puuid, Name: "project-" + puuid},
	}
	_, err := ps.GetByName(context.Background(), project.GetMetadata().Name)
	if err != nil {
		t.Fatal("could not get project:", err)
	}
}

func TestProjectGetById(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	puuid := uuid.NewString()

	addFetchByIdExpectation(mock, "project", puuid)

	project := &systemv3.Project{
		Metadata: &v3.Metadata{Id: puuid, Name: "project-" + puuid},
	}
	project, err := ps.GetByID(context.Background(), project.Metadata.Id)
	if err != nil {
		t.Fatal("could not get project:", err)
	}
	performProjectBasicChecks(t, project, puuid)
}

func TestProjectUpdate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	puuid := addFetchExpectation(mock, "project")
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "authsrv_projectgrouprole" AS "projectgrouprole" SET trash = TRUE WHERE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`UPDATE "authsrv_projectgroupnamespacerole" AS "projectgroupnamespacerole" SET trash = TRUE WHERE ."project_id" = '` + puuid + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectExec(`UPDATE "authsrv_projectaccountresourcerole" AS "projectaccountresourcerole" SET trash = TRUE WHERE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`UPDATE "authsrv_projectaccountnamespacerole" AS "projectaccountnamespacerole" SET trash = TRUE WHERE ."project_id" = '` + puuid + `'. AND .trash = false. RETURNING *`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))
	mock.ExpectExec(`UPDATE "authsrv_project"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group 
		FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id 
		JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id 
		JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id WHERE`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("ADMIN", "project-"+puuid))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group, 
		namespace FROM "authsrv_projectgroupnamespacerole" 
		JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id 
		JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id 
		JOIN authsrv_group ON authsrv_group.id=authsrv_projectgroupnamespacerole.group_id WHERE`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project"}).AddRow("ADMIN", "project-"+puuid))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user 
		FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id 
		JOIN identities ON identities.id=authsrv_projectaccountresourcerole.account_id WHERE`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "user"}).AddRow("ADMIN", "user@email.com"))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user, namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole 
		ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN identities 
		ON identities.id=authsrv_projectaccountnamespacerole.account_id WHERE`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("ADMIN"))
	mock.ExpectCommit()

	project := &systemv3.Project{
		Metadata: &v3.Metadata{Id: puuid, Name: "project-" + puuid},
		Spec:     &systemv3.ProjectSpec{},
	}
	_, err := ps.Update(context.Background(), project)
	if err != nil {
		t.Fatal("could not update project:", err)
	}
}

func TestProjectList(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	uid := uuid.NewString()
	ouuid := addFetchExpectation(mock, "organization")
	puuid := addFetchExpectation(mock, "partner")
	mock.ExpectQuery(`SELECT "project"."id", "project"."name", .* FROM "authsrv_project" AS "project" WHERE \(partner_id = '` + puuid + `'\) AND \(organization_id = '` + ouuid + `'\) AND \(trash = false\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uid, "project-"+uid))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id WHERE \(authsrv_projectgrouprole.project_id = '` + uid + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "group"}).AddRow("test-role1", "test-project1", "test-group1"))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group, namespace FROM "authsrv_projectgroupnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id JOIN authsrv_group ON authsrv_group.id=authsrv_projectgroupnamespacerole.group_id WHERE \(authsrv_projectgroupnamespacerole.project_id = '` + uid + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "group", "namespace"}).AddRow("test-role2", "test-project2", "test-group2", "test-namespace2"))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN identities ON identities.id=authsrv_projectaccountresourcerole.account_id WHERE \(authsrv_projectaccountresourcerole.project_id = '` + uid + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "user"}).AddRow("test-role3", "test-user3"))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user, namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN identities ON identities.id=authsrv_projectaccountnamespacerole.account_id WHERE \(authsrv_projectaccountnamespacerole.project_id = '` + uid + `'\) AND \(authsrv_projectaccountnamespacerole.trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "user", "namespace"}).AddRow("test-role4", "test-user4", "test-namespace4"))

	project := &systemv3.Project{
		Metadata: &v3.Metadata{
			Organization: ouuid,
		},
	}
	pl, err := ps.List(context.Background(), project)
	if err != nil {
		t.Fatal("could not get project:", err)
	}
	if pl.Metadata.Count != 1 {
		t.Errorf("incorrect number of projects returned; expected '%v', got '%v'", 1, pl.Metadata.Count)
	}
	if pl.Items[0].Metadata.Name != "project-"+uid {
		t.Errorf("incorrect project name; expected '%v', got '%v'", "project-"+uid, pl.Items[0].Metadata.Name)
	}
	if *pl.Items[0].Spec.ProjectNamespaceRoles[0].Project != "test-project1" {
		t.Errorf("incorrect projectnamespacerole; expected '%v', got '%v'", "test-project1", *pl.Items[0].Spec.ProjectNamespaceRoles[0].Project)
	}
	if *pl.Items[0].Spec.ProjectNamespaceRoles[1].Namespace != "test-namespace2" {
		t.Errorf("incorrect projectnamespacerole; expected '%v', got '%v'", "test-namespace2", *pl.Items[0].Spec.ProjectNamespaceRoles[1].Namespace)
	}
	if pl.Items[0].Spec.UserRoles[0].User != "test-user3" {
		t.Errorf("incorrect userrole; expected '%v', got '%v'", "test-user3", pl.Items[0].Spec.UserRoles[0].User)
	}
	if pl.Items[0].Spec.UserRoles[1].Namespace != "test-namespace4" {
		t.Errorf("incorrect userrole; expected '%v', got '%v'", "test-namespace4", pl.Items[0].Spec.UserRoles[1].Namespace)
	}
}

func TestProjectListNonDev(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), false)

	uid := uuid.NewString()
	ouuid := addFetchExpectation(mock, "organization")
	puuid := addFetchExpectation(mock, "partner")
	uuuid := addUserFetchExpectation(mock)
	mock.ExpectQuery(`SELECT distinct account_id, project_id FROM "sentry_account_permission" AS "sap" WHERE \(sap.partner_id = '` + puuid + `'\) AND \(sap.organization_id = '` + ouuid + `'\) AND \(sap.account_id = '` + uuuid + `'\) AND \(sap.permission_name IN \('project.read', 'ops_star.all'\)\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"account_id", "project_id"}).AddRow(uuuid, uid))
	mock.ExpectQuery(`SELECT "project"."id", "project"."name", .* FROM "authsrv_project" AS "project" WHERE \(project.partner_id = '` + puuid + `'\) AND \(project.organization_id = '` + ouuid + `'\) AND \(project.trash = FALSE\) AND \(project.id IN \('` + uid + `'\)\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uid, "project-"+uid))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group FROM "authsrv_projectgrouprole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id WHERE \(authsrv_projectgrouprole.project_id = '` + uid + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "group"}).AddRow("test-role1", "test-project1", "test-group1"))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group, namespace FROM "authsrv_projectgroupnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id JOIN authsrv_group ON authsrv_group.id=authsrv_projectgroupnamespacerole.group_id WHERE \(authsrv_projectgroupnamespacerole.project_id = '` + uid + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "project", "group", "namespace"}).AddRow("test-role2", "test-project2", "test-group2", "test-namespace2"))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user FROM "authsrv_projectaccountresourcerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id JOIN identities ON identities.id=authsrv_projectaccountresourcerole.account_id WHERE \(authsrv_projectaccountresourcerole.project_id = '` + uid + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "user"}).AddRow("test-role3", "test-user3"))
	mock.ExpectQuery(`SELECT distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user, namespace FROM "authsrv_projectaccountnamespacerole" JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id JOIN identities ON identities.id=authsrv_projectaccountnamespacerole.account_id WHERE \(authsrv_projectaccountnamespacerole.project_id = '` + uid + `'\) AND \(authsrv_projectaccountnamespacerole.trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"role", "user", "namespace"}).AddRow("test-role4", "test-user4", "test-namespace4"))

	project := &systemv3.Project{
		Metadata: &v3.Metadata{
			Organization: ouuid,
		},
	}

	sd := v3.SessionData{Username: "user-" + uuuid}
	ctx := context.WithValue(context.Background(), common.SessionDataKey, &sd)
	pl, err := ps.List(ctx, project)
	if err != nil {
		t.Fatal("could not get project:", err)
	}

	if pl.Metadata.Count != 1 {
		t.Errorf("incorrect number of projects returned; expected '%v', got '%v'", 1, pl.Metadata.Count)
	}

	if pl.Items[0].Metadata.Name != "project-"+uid {
		t.Errorf("incorrect project name; expected '%v', got '%v'", "project-"+uid, pl.Items[0].Metadata.Name)
	}

	if *pl.Items[0].Spec.ProjectNamespaceRoles[0].Project != "test-project1" {
		t.Errorf("incorrect projectnamespacerole; expected '%v', got '%v'", "test-project1", *pl.Items[0].Spec.ProjectNamespaceRoles[0].Project)
	}

	if *pl.Items[0].Spec.ProjectNamespaceRoles[1].Namespace != "test-namespace2" {
		t.Errorf("incorrect projectnamespacerole; expected '%v', got '%v'", "test-namespace2", *pl.Items[0].Spec.ProjectNamespaceRoles[1].Namespace)
	}

	if pl.Items[0].Spec.UserRoles[0].User != "test-user3" {
		t.Errorf("incorrect userrole; expected '%v', got '%v'", "test-user3", pl.Items[0].Spec.UserRoles[0].User)
	}

	if pl.Items[0].Spec.UserRoles[1].Namespace != "test-namespace4" {
		t.Errorf("incorrect userrole; expected '%v', got '%v'", "test-namespace4", pl.Items[0].Spec.UserRoles[1].Namespace)
	}
}
