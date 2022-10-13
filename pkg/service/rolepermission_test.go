package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	rolev3 "github.com/paralus/paralus/proto/types/rolepb/v3"
)

func performRolePermissionBasicChecks(t *testing.T, role *rolev3.RolePermission, ruuid string) {
	_, err := uuid.Parse(role.GetMetadata().GetOrganization())
	if err == nil {
		t.Error("org in metadata should be name not id")
	}
	_, err = uuid.Parse(role.GetMetadata().GetPartner())
	if err == nil {
		t.Error("partner in metadata should be name not id")
	}
	if role.GetMetadata().GetId() != ruuid {
		t.Error("invalid uuid returned")
	}
}

func TestRolePermissionList(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	rs := NewRolepermissionService(db)

	ruuid1 := uuid.New().String()
	ruuid2 := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "resourcepermission"."id", "resourcepermission"."name".* FROM "authsrv_resourcepermission" AS "resourcepermission"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(ruuid1, "role-"+ruuid1).AddRow(ruuid2, "role-"+ruuid2))

	req := &commonv3.QueryOptions{
		Partner:      "partner-" + puuid,
		Organization: "org-" + ouuid,
	}
	rolelist, err := rs.List(context.Background(), query.WithOptions(req))
	if err != nil {
		t.Fatal("could not list rolepermissions:", err)
	}
	if rolelist.Metadata.Count != 2 {
		t.Errorf("incorrect number of rolepermissions returned, expected 2; got %v", rolelist.Metadata.Count)
	}
	if rolelist.Items[0].Metadata.Name != "role-"+ruuid1 || rolelist.Items[1].Metadata.Name != "role-"+ruuid2 {
		t.Errorf("incorrect role ids returned when listing")
	}
}

func TestRolePermissionListWithSelectors(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	rs := NewRolepermissionService(db)

	ruuid1 := uuid.New().String()
	ruuid2 := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT authsrv_resourcepermission.name as name, authsrv_resourcepermission.scope as scope FROM "authsrv_resourcepermission" WHERE \(scope = 'NAMESPACE'\) AND \(authsrv_resourcepermission.trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(ruuid1, "role-"+ruuid1))

	mock.ExpectQuery(`SELECT authsrv_resourcepermission.name as name, authsrv_resourcepermission.scope as scope FROM "authsrv_resourcepermission" WHERE \(scope = 'PROJECT'\) AND \(authsrv_resourcepermission.trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(ruuid2, "role-"+ruuid2))

	mock.ExpectQuery(`SELECT authsrv_resourcepermission.name as name, authsrv_resourcepermission.description as description, authsrv_resourcepermission.scope as scope FROM "authsrv_resourcepermission" WHERE \(name IN \('partner.read', 'organization.read'\)\) AND \(authsrv_resourcepermission.trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(ruuid2, "role-"+ruuid2).AddRow(ruuid1, "role-"+ruuid1))

	req := &commonv3.QueryOptions{
		Partner:      "partner-" + puuid,
		Organization: "org-" + ouuid,
		Selector:     "namespace",
	}
	rolelist, err := rs.List(context.Background(), query.WithOptions(req))
	if err != nil {
		t.Fatal("could not list rolepermissions:", err)
	}
	if rolelist.Metadata.Count != 4 {
		t.Errorf("incorrect number of rolepermissions returned, expected 2; got %v", rolelist.Metadata.Count)
	}
	if rolelist.Items[0].Metadata.Name != "role-"+ruuid1 || rolelist.Items[1].Metadata.Name != "role-"+ruuid2 {
		t.Errorf("incorrect role ids returned when listing")
	}
}

func TestRolePermissionGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	rs := NewRolepermissionService(db)

	ruuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "resourcepermission"."id", "resourcepermission"."name".* FROM "authsrv_resourcepermission" AS "resourcepermission"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(ruuid, "role-"+ruuid))

	req := &rolev3.RolePermission{
		Metadata: &v3.Metadata{},
	}
	role, err := rs.GetByName(context.Background(), req)
	if err != nil {
		t.Fatal("could not list rolepermissions:", err)
	}
	if role.Metadata.Name != "role-"+ruuid {
		t.Errorf("incorrect role name; expected '%v', got '%v'", "role-"+ruuid, role.Metadata.Name)
	}
}
