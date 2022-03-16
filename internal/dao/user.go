package dao

import (
	"context"

	"github.com/RafaySystems/rcloud-base/internal/models"
	userv3 "github.com/RafaySystems/rcloud-base/proto/types/userpb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func GetGroups(ctx context.Context, db bun.IDB, id uuid.UUID) ([]models.Group, error) {
	var entities = []models.Group{}
	err := db.NewSelect().Model(&entities).
		Join(`JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id`).
		Where("authsrv_groupaccount.account_id = ?", id).
		Scan(ctx)
	return entities, err
}

func GetUserRoles(ctx context.Context, db bun.IDB, id uuid.UUID) ([]*userv3.ProjectNamespaceRole, error) {
	// Could possibly union them later for some speedup
	var r = []*userv3.ProjectNamespaceRole{}
	err := db.NewSelect().Table("authsrv_accountresourcerole").
		ColumnExpr("authsrv_resourcerole.name as role").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id`).
		Where("authsrv_accountresourcerole.account_id = ?", id).
		Scan(ctx, &r)
	if err != nil {
		return nil, err
	}

	var pr = []*userv3.ProjectNamespaceRole{}
	err = db.NewSelect().Table("authsrv_projectaccountresourcerole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id`).
		Where("authsrv_projectaccountresourcerole.account_id = ?", id).
		Scan(ctx, &pr)
	if err != nil {
		return nil, err
	}

	var pnr = []*userv3.ProjectNamespaceRole{}
	err = db.NewSelect().Table("authsrv_projectaccountnamespacerole").
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
