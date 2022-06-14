package dao

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/uptrace/bun"
)

func GetProjectNamespaces(ctx context.Context, db bun.IDB, projectID uuid.UUID) ([]string, error) {
	var cns []string

	var panr []models.ProjectAccountNamespaceRole
	err := db.NewSelect().Model(&panr).Where("project_id = ?", projectID).Where("trash = ?", false).Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	for _, nr := range panr {
		cns = append(cns, nr.Namespace)
	}

	var pgnr []models.ProjectGroupNamespaceRole
	err = db.NewSelect().Model(&pgnr).Where("project_id = ?", projectID).Where("trash = ?", false).Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	for _, nr := range pgnr {
		cns = append(cns, nr.Namespace)
	}

	return cns, err
}

func GetAccountProjectNamespaces(ctx context.Context, db bun.IDB, projectID uuid.UUID, accountID uuid.UUID) ([]string, error) {
	var cns []string

	var panr []models.ProjectAccountNamespaceRole
	err := db.NewSelect().Model(&panr).Where("project_id = ?", projectID).Where("account_id = ?", accountID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	for _, nr := range panr {
		cns = append(cns, nr.Namespace)
	}

	return cns, err
}

func GetGroupProjectNamespaces(ctx context.Context, db bun.IDB, projectID uuid.UUID, accountID uuid.UUID) ([]string, error) {
	var cns []string

	var pgnr []models.ProjectGroupNamespaceRole
	err := db.NewSelect().Model(&pgnr).Where("project_id = ?", projectID).
		Join(`JOIN authsrv_groupaccount ON projectgroupnamespacerole.group_id=authsrv_groupaccount.group_id`).
		Where("authsrv_groupaccount.account_id = ?", accountID).
		Where("projectgroupnamespacerole.trash = ?", false).
		Where("authsrv_groupaccount.trash = ?", false).Scan(ctx)
	if err != nil {
		return nil, err
	}
	for _, nr := range pgnr {
		cns = append(cns, nr.Namespace)
	}

	return cns, err
}
