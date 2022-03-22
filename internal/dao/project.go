package dao

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func GetProjectOrganization(ctx context.Context, db bun.IDB, id uuid.UUID) (string, string, error) {
	// Could possibly union them later for some speedup
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
