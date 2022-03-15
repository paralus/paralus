package dao

import (
	"context"

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type roleDAO struct {
	db *bun.DB
}

// Role specific db access
type RoleDAO interface {
	Close() error
	// get permissions for role
	GetRolePermissions(context.Context, uuid.UUID) ([]models.ResourcePermission, error)
}

// NewRoleDao return new group dao
func NewRoleDAO(db *bun.DB) *roleDAO {
	return &roleDAO{db}
}

func (dao *roleDAO) Close() error {
	return dao.db.Close()
}

func (dao *roleDAO) GetRolePermissions(ctx context.Context, id uuid.UUID) ([]models.ResourcePermission, error) {
	// Could possibly union them later for some speedup
	var r = []models.ResourcePermission{}
	err := dao.db.NewSelect().Table("authsrv_resourcepermission").
		ColumnExpr("authsrv_resourcepermission.name as name").
		Join(`JOIN authsrv_resourcerolepermission ON authsrv_resourcerolepermission.resource_permission_id=authsrv_resourcepermission.id`).
		Where("authsrv_resourcerolepermission.resource_role_id = ?", id).
		Scan(ctx, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
