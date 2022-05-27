package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	v3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	systemv3 "github.com/RafayLabs/rcloud-base/proto/types/systempb/v3"
	"github.com/google/uuid"
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
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "organization"."id", "organization"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))
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

func TestProjectDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id", "project"."name", .* FROM "authsrv_project" AS "project" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "project-"+puuid))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "authsrv_projectgrouprole" AS "projectgrouprole" SET trash = TRUE WHERE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "authsrv_projectaccountresourcerole" AS "projectaccountresourcerole" SET trash = TRUE WHERE`).WillReturnResult(sqlmock.NewResult(1, 1))

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

func TestProjectDeleteNonExist(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	mazc := mockAuthzClient{}
	ps := NewProjectService(db, &mazc, getLogger(), true)

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id", "project"."name", .* FROM "authsrv_project" AS "project" WHERE`).
		WithArgs().WillReturnError(fmt.Errorf("No data available"))

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

	partuuid := uuid.New().String()
	ouuid := uuid.New().String()
	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id", "project"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "partner_id"}).AddRow(puuid, ouuid, partuuid))

	mock.ExpectQuery(`SELECT "organization"."id", "organization"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ouuid))

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(partuuid))

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

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id", "project"."name", .* FROM "authsrv_project" AS "project" WHERE .*id = '` + puuid + `'`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "project-"+puuid))

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

	puuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id", "project"."name", .* FROM "authsrv_project" AS "project" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "project-"+puuid))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "authsrv_projectgrouprole" AS "projectgrouprole" SET trash = TRUE WHERE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "authsrv_projectaccountresourcerole" AS "projectaccountresourcerole" SET trash = TRUE WHERE`).WillReturnResult(sqlmock.NewResult(1, 1))
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
