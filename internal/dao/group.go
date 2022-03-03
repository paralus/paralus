package dao

import (
	"context"

	"github.com/RafaySystems/rcloud-base/internal/models"
	userv3 "github.com/RafaySystems/rcloud-base/proto/types/userpb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type groupDAO struct {
	db *bun.DB
}

// Group specific db access
type GroupDAO interface {
	Close() error
	// get users for group
	GetUsers(context.Context, uuid.UUID) ([]models.KratosIdentities, error)
	// get roles for group
	GetRoles(context.Context, uuid.UUID) ([]*userv3.ProjectNamespaceRole, error)
}

// NewGroupDao return new group dao
func NewGroupDAO(db *bun.DB) *groupDAO {
	return &groupDAO{db}
}

func (dao *groupDAO) Close() error {
	return dao.db.Close()
}

// GetUsers gets the list of users in a given group
func (dao *groupDAO) GetUsers(ctx context.Context, id uuid.UUID) ([]models.KratosIdentities, error) {
	var entities = []models.KratosIdentities{}
	err := dao.db.NewSelect().Model(&entities).
		Join(`JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id`).
		Where(`authsrv_groupaccount.group_id = ?`, id).
		Scan(ctx)
	return entities, err
}

func (dao *groupDAO) GetRoles(ctx context.Context, id uuid.UUID) ([]*userv3.ProjectNamespaceRole, error) {
	// Could possibily union them later for some speedup
	var r = []*userv3.ProjectNamespaceRole{}
	err := dao.db.NewSelect().Table("authsrv_grouprole").
		ColumnExpr("authsrv_resourcerole.name as role").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_grouprole.role_id`).
		Where("authsrv_grouprole.group_id = ?", id).
		Scan(ctx, &r)
	if err != nil {
		return nil, err
	}

	var pr = []*userv3.ProjectNamespaceRole{}
	err = dao.db.NewSelect().Table("authsrv_projectgrouprole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id`).
		Where("authsrv_projectgrouprole.group_id = ?", id).
		Scan(ctx, &pr)
	if err != nil {
		return nil, err
	}

	var pnr = []*userv3.ProjectNamespaceRole{}
	err = dao.db.NewSelect().Table("authsrv_projectgroupnamespacerole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id`). // also need a namespace join
		Where("authsrv_projectgroupnamespacerole.group_id = ?", id).
		Scan(ctx, &pnr)
	if err != nil {
		return nil, err
	}

	return append(append(r, pr...), pnr...), err
}
