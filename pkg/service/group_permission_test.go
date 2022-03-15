package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetGroupPermissions(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewGroupPermissionService(db)

	groupNames := []string{"mygroup", "admin"}
	gid := uuid.New()
	oid := uuid.New().String()
	pid := uuid.New().String()

	mock.ExpectQuery(`SELECT "sgp"."group_id", "sgp"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"group_id"}).AddRow(gid))

	_, err := ps.GetGroupPermissions(context.Background(), groupNames, oid, pid)
	if err != nil {
		t.Fatal("could not get GetGroupPermissions:", err)
	}
}

func TestGetGroupProjectsByPermission(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewGroupPermissionService(db)

	groupNames := []string{"mygroup", "admin"}
	gid := uuid.New()
	oid := uuid.New().String()
	pid := uuid.New().String()

	mock.ExpectQuery(`SELECT "sgp"."group_id", "sgp"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"group_id"}).AddRow(gid))

	_, err := ps.GetGroupProjectsByPermission(context.Background(), groupNames, oid, pid, "")
	if err != nil {
		t.Fatal("could not get GetGroupProjectsByPermission:", err)
	}
}

func TestGetGroupPermissionsByProjectIDPermissions(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewGroupPermissionService(db)

	groupNames := []string{"mygroup", "admin"}
	projectNames := []string{"myproject"}
	permissionNames := []string{"read"}
	gid := uuid.New()
	oid := uuid.New().String()
	pid := uuid.New().String()

	mock.ExpectQuery(`SELECT "sgp"."group_id", "sgp"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"group_id"}).AddRow(gid))

	_, err := ps.GetGroupPermissionsByProjectIDPermissions(context.Background(), groupNames, oid, pid, projectNames, permissionNames)
	if err != nil {
		t.Fatal("could not get GetGroupProjectsByPermission:", err)
	}
}
