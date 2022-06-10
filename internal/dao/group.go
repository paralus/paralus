package dao

import (
	"context"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/uptrace/bun"
)

// GetUsers gets the list of users in a given group
func GetUsers(ctx context.Context, db bun.IDB, id uuid.UUID) ([]models.KratosIdentities, error) {
	var entities = []models.KratosIdentities{}
	err := db.NewSelect().Model(&entities).
		Join(`JOIN authsrv_groupaccount ON identities.id=authsrv_groupaccount.account_id`).
		Where(`authsrv_groupaccount.group_id = ?`, id).
		Where("authsrv_groupaccount.trash = ?", false).
		Scan(ctx)
	return entities, err
}

func GetGroupRoles(ctx context.Context, db bun.IDB, id uuid.UUID) ([]*userv3.ProjectNamespaceRole, error) {
	// Could possibly union them later for some speedup
	var r = []*userv3.ProjectNamespaceRole{}
	err := db.NewSelect().Table("authsrv_grouprole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_group.name as group").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_grouprole.role_id`).
		Join(`JOIN authsrv_group ON authsrv_group.id=authsrv_grouprole.group_id`).
		Where("authsrv_grouprole.group_id = ?", id).
		Where("authsrv_resourcerole.trash = ?", false).
		Where("authsrv_grouprole.trash = ?", false).
		Scan(ctx, &r)
	if err != nil {
		return nil, err
	}

	var pr = []*userv3.ProjectNamespaceRole{}
	err = db.NewSelect().Table("authsrv_projectgrouprole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id`).
		Join(`JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id`).
		Where("authsrv_projectgrouprole.group_id = ?", id).
		Where("authsrv_project.trash = ?", false).
		Where("authsrv_projectgrouprole.trash = ?", false).
		Where("authsrv_resourcerole.trash = ?", false).
		Scan(ctx, &pr)
	if err != nil {
		return nil, err
	}

	var pnr = []*userv3.ProjectNamespaceRole{}
	err = db.NewSelect().Table("authsrv_projectgroupnamespacerole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project, namespace, authsrv_group.name as group").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id`). // also need a namespace join
		Join(`JOIN authsrv_group ON authsrv_group.id=authsrv_projectgroupnamespacerole.group_id`).
		Where("authsrv_projectgroupnamespacerole.group_id = ?", id).
		Where("authsrv_project.trash = ?", false).
		Where("authsrv_projectgroupnamespacerole.trash = ?", false).
		Where("authsrv_resourcerole.trash = ?", false).
		Scan(ctx, &pnr)
	if err != nil {
		return nil, err
	}

	return append(append(r, pr...), pnr...), err
}
