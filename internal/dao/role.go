package dao

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/uptrace/bun"
)

func GetRolePermissions(ctx context.Context, db bun.IDB, id uuid.UUID) ([]models.ResourcePermission, error) {
	// TODO: filter by parter and org
	var r = []models.ResourcePermission{}
	err := db.NewSelect().Table("authsrv_resourcepermission").
		ColumnExpr("authsrv_resourcepermission.name as name").
		Join(`JOIN authsrv_resourcerolepermission ON authsrv_resourcerolepermission.resource_permission_id=authsrv_resourcepermission.id`).
		Where("authsrv_resourcerolepermission.resource_role_id = ?", id).
		Where("authsrv_resourcepermission.trash = ?", false).
		Where("authsrv_resourcerolepermission.trash = ?", false).
		Scan(ctx, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func GetRolePermissionsByScope(ctx context.Context, db bun.IDB, scope string) ([]models.ResourcePermission, error) {
	// Could possibly union them later for some speedup
	var r = []models.ResourcePermission{}
	err := db.NewSelect().Table("authsrv_resourcepermission").
		ColumnExpr("authsrv_resourcepermission.name as name, authsrv_resourcepermission.scope as scope").
		Where("scope = ?", strings.ToUpper(scope)).
		Where("authsrv_resourcepermission.trash = ?", false).
		Scan(ctx, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func GetRolePermissionsByNames(ctx context.Context, db bun.IDB, permissions ...string) ([]models.ResourcePermission, error) {
	var r = []models.ResourcePermission{}
	err := db.NewSelect().Table("authsrv_resourcepermission").
		ColumnExpr("authsrv_resourcepermission.name as name, authsrv_resourcepermission.description as description, authsrv_resourcepermission.scope as scope").
		Where("name IN (?)", bun.In(permissions)).
		Where("authsrv_resourcepermission.trash = ?", false).
		Scan(ctx, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
