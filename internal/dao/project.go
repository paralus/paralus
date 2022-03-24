package dao

import (
	"context"

	userv3 "github.com/RafayLabs/rcloud-base/proto/types/userpb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func GetProjectGroupRoles(ctx context.Context, db bun.IDB, id uuid.UUID) ([]*userv3.ProjectNamespaceRole, error) {
	var pr = []*userv3.ProjectNamespaceRole{}
	err := db.NewSelect().Table("authsrv_projectgrouprole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id`).
		Join(`JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id`).
		Where("authsrv_projectgrouprole.project_id = ?", id).
		Scan(ctx, &pr)
	if err != nil {
		return nil, err
	}

	var pnr = []*userv3.ProjectNamespaceRole{}
	err = db.NewSelect().Table("authsrv_projectgroupnamespacerole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group, namespace_id as namespace").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id`).
		Join(`JOIN authsrv_group ON authsrv_group.id=authsrv_projectgroupnamespacerole.group_id`). // also need a namespace join
		Where("authsrv_projectgroupnamespacerole.project_id = ?", id).
		Scan(ctx, &pnr)
	if err != nil {
		return nil, err
	}

	return append(pr, pnr...), err
}

func GetProjectUserRoles(ctx context.Context, db bun.IDB, id uuid.UUID) ([]*userv3.UserRole, error) {

	var pr = []*userv3.UserRole{}
	err := db.NewSelect().Table("authsrv_projectaccountresourcerole").
		ColumnExpr("distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id`).
		Join(`JOIN identities ON identities.id=authsrv_projectaccountresourcerole.account_id`).
		Where("authsrv_projectaccountresourcerole.project_id = ?", id).
		Scan(ctx, &pr)
	if err != nil {
		return nil, err
	}

	return pr, err
}
