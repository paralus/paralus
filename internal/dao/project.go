package dao

import (
	"context"

	"github.com/RafayLabs/rcloud-base/internal/models"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func GetProjectOrganization(ctx context.Context, db bun.IDB, id uuid.UUID) (string, string, error) {
	type projectOrg struct {
		Project      string
		Organization string
	}
	var r projectOrg
	err := db.NewSelect().Table("authsrv_project").
		ColumnExpr("authsrv_project.name as project").
		ColumnExpr("authsrv_organization.name as organization").
		Join(`JOIN authsrv_organization ON authsrv_project.organization_id=authsrv_organization.id`).
		Where("authsrv_project.id = ?", id).
		Where("authsrv_project.trash = ?", false).
		Where("authsrv_organization.trash = ?", false).
		Scan(ctx, &r)
	if err != nil {
		return "", "", err
	}
	return r.Project, r.Organization, nil
}

func GetFileteredProjects(ctx context.Context, db bun.IDB, account, partner, org uuid.UUID) ([]models.Project, error) {
	ids := []uuid.UUID{}
	sp := []models.SentryPermission{}
	err := db.NewSelect().Model(&sp).
		Where("sentry_permission.partner_id = ?", partner).
		Where("sentry_permission.organization_id = ?", org).
		Where("sentry_permission.account_id = ?", account).
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
