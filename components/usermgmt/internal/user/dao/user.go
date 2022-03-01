package dao

import (
	"context"

	userv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/userpb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/internal/models"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type userDAO struct {
	db *bun.DB
}

// User specific db access
type UserDAO interface {
	Close() error
	// get groups for user
	GetGroups(context.Context, uuid.UUID) ([]models.Group, error)
	// get roles for user
	GetRoles(context.Context, uuid.UUID) ([]*userv3.ProjectNamespaceRole, error)
}

// NewUserDao return new user dao
func NewUserDAO(db *bun.DB) *userDAO {
	return &userDAO{db}
}

func (dao *userDAO) Close() error {
	// XXX: if one dao closes the db connections, won't other have issues?
	return dao.db.Close()
}

func (dao *userDAO) GetGroups(ctx context.Context, id uuid.UUID) ([]models.Group, error) {
	var entities = []models.Group{}
	err := dao.db.NewSelect().Model(&entities).
		Join(`JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id`).
		Where("authsrv_groupaccount.account_id = ?", id).
		Scan(ctx)
	return entities, err
}

func (dao *userDAO) GetRoles(ctx context.Context, id uuid.UUID) ([]*userv3.ProjectNamespaceRole, error) {
	// Could possibily union them later for some speedup
	var r = []*userv3.ProjectNamespaceRole{}
	err := dao.db.NewSelect().Table("authsrv_accountresourcerole").
		ColumnExpr("authsrv_resourcerole.name as role").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id`).
		Where("authsrv_accountresourcerole.account_id = ?", id).
		Scan(ctx, &r)
	if err != nil {
		return nil, err
	}

	var pr = []*userv3.ProjectNamespaceRole{}
	err = dao.db.NewSelect().Table("authsrv_projectaccountresourcerole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id`).
		Where("authsrv_projectaccountresourcerole.account_id = ?", id).
		Scan(ctx, &pr)
	if err != nil {
		return nil, err
	}

	var pnr = []*userv3.ProjectNamespaceRole{}
	err = dao.db.NewSelect().Table("authsrv_projectaccountnamespacerole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id`). // also need a namespace join
		Where("authsrv_projectaccountnamespacerole.account_id = ?", id).
		Scan(ctx, &pnr)
	if err != nil {
		return nil, err
	}

	return append(append(r, pr...), pnr...), err
}
