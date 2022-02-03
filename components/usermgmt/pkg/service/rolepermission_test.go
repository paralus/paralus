package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
	"github.com/google/uuid"
)

func performRolePermissionBasicChecks(t *testing.T, role *userv3.RolePermission, ruuid string) {
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
	if role.Status.ConditionStatus != v3.ConditionStatus_StatusOK {
		t.Error("group status is not OK")
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

	role := &userv3.RolePermission{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid},
	}
	rolelist, err := rs.List(context.Background(), role)
	if err != nil {
		t.Fatal("could not list rolepermissions:", err)
	}
	if rolelist.Metadata.Count != 2 {
		t.Errorf("incorrect number of rolepermissions returned, expected 2; got %v", rolelist.Metadata.Count)
	}
	if rolelist.Items[0].Metadata.Id != ruuid1 || rolelist.Items[1].Metadata.Id != ruuid2 {
		t.Errorf("incorrect role ids returned when listing")
	}
}
