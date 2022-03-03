package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	v3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	rolev3 "github.com/RafaySystems/rcloud-base/proto/types/rolepb/v3"
	"github.com/google/uuid"
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
	defer rs.Close()

	ruuid1 := uuid.New().String()
	ruuid2 := uuid.New().String()
	puuid := uuid.New().String()
	ouuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "resourcepermission"."id", "resourcepermission"."name".* FROM "authsrv_resourcepermission" AS "resourcepermission"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(ruuid1, "role-"+ruuid1).AddRow(ruuid2, "role-"+ruuid2))

	role := &rolev3.RolePermission{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid},
	}
	rolelist, err := rs.List(context.Background(), role)
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
