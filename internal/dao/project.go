package dao

import (
	"context"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/uptrace/bun"
)

type ProjectOrg struct {
	Project        string
	Organization   string
	ProjectId      string
	OrganizationId string
	PartnerId      string
}

func GetProjectOrganization(ctx context.Context, db bun.IDB, name string) (ProjectOrg, error) {
	var r ProjectOrg
	err := db.NewSelect().Table("authsrv_project").
		ColumnExpr("authsrv_project.name as project").
		ColumnExpr("authsrv_organization.name as organization").
		ColumnExpr("authsrv_project.id as project_id").
		ColumnExpr("authsrv_organization.id as organization_id").
		ColumnExpr("authsrv_organization.partner_id as partner_id").
		Join(`JOIN authsrv_organization ON authsrv_project.organization_id=authsrv_organization.id`).
		Where("authsrv_project.name = ?", name).
		Where("authsrv_project.trash = ?", false).
		Where("authsrv_organization.trash = ?", false).
		Scan(ctx, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func GetFileteredProjects(ctx context.Context, db bun.IDB, account, partner, org uuid.UUID) ([]models.Project, error) {
	ids := []uuid.UUID{}
	sp := []models.AccountPermission{}
	err := db.NewSelect().Model(&sp).
		ColumnExpr("distinct account_id, project_id").
		Where("sap.partner_id = ?", partner).
		Where("sap.organization_id = ?", org).
		Where("sap.account_id = ?", account).
		Where("sap.permission_name IN (?)", bun.In([]string{"project.read", "ops_star.all"})).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	all := false
	for _, p := range sp {
		if p.ProjectId == uuid.Nil {
			all = true
			break
		}
		ids = append(ids, p.ProjectId)
	}

	prjs := []models.Project{}
	if !all && len(ids) == 0 {
		return prjs, nil
	}
	q := db.NewSelect().Model(&prjs).
		Where("project.partner_id = ?", partner).
		Where("project.organization_id = ?", org).
		Where("project.trash = ?", false)
	if !all {
		q = q.Where("project.id IN (?)", bun.In(ids))
	}
	err = q.Scan(ctx)
	return prjs, err
}

func GetProjectGroupRoles(ctx context.Context, db bun.IDB, id uuid.UUID) ([]*userv3.ProjectNamespaceRole, error) {
	var pr = []*userv3.ProjectNamespaceRole{}
	err := db.NewSelect().Table("authsrv_projectgrouprole").
		ColumnExpr("distinct authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgrouprole.role_id`).
		Join(`JOIN authsrv_group ON authsrv_group.id=authsrv_projectgrouprole.group_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectgrouprole.project_id`).
		Where("authsrv_projectgrouprole.project_id = ?", id).
		Where("authsrv_projectgrouprole.trash = ?", false).
		Scan(ctx, &pr)
	if err != nil {
		return nil, err
	}

	var pnr = []*userv3.ProjectNamespaceRole{}
	err = db.NewSelect().Table("authsrv_projectgroupnamespacerole").
		ColumnExpr("distinct authsrv_resourcerole.name as role, authsrv_project.name as project, authsrv_group.name as group, namespace").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectgroupnamespacerole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectgroupnamespacerole.project_id`).
		Join(`JOIN authsrv_group ON authsrv_group.id=authsrv_projectgroupnamespacerole.group_id`).
		Where("authsrv_projectgroupnamespacerole.project_id = ?", id).
		Where("authsrv_projectgroupnamespacerole.trash = ?", false).
		Scan(ctx, &pnr)
	if err != nil {
		return nil, err
	}

	return append(pr, pnr...), err
}

func GetProjectUserRoles(ctx context.Context, db bun.IDB, id uuid.UUID) ([]*userv3.UserRole, error) {

	var ur = []*userv3.UserRole{}
	err := db.NewSelect().Table("authsrv_projectaccountresourcerole").
		ColumnExpr("distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id`).
		Join(`JOIN identities ON identities.id=authsrv_projectaccountresourcerole.account_id`).
		Where("authsrv_projectaccountresourcerole.project_id = ?", id).
		Where("authsrv_projectaccountresourcerole.trash = ?", false).
		Scan(ctx, &ur)
	if err != nil {
		return nil, err
	}

	var unr = []*userv3.UserRole{}
	err = db.NewSelect().Table("authsrv_projectaccountnamespacerole").
		ColumnExpr("distinct authsrv_resourcerole.name as role, identities.traits ->> 'email' as user, namespace").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id`).
		Join(`JOIN identities ON identities.id=authsrv_projectaccountnamespacerole.account_id`).
		Where("authsrv_projectaccountnamespacerole.project_id = ?", id).
		Where("authsrv_projectaccountnamespacerole.trash = ?", false).
		Scan(ctx, &unr)
	if err != nil {
		return nil, err
	}

	return append(ur, unr...), err
}
