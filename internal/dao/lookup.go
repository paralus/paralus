package dao

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/uptrace/bun"
)

func GetPartnerId(ctx context.Context, db bun.IDB, name string) (uuid.UUID, error) {
	entity, err := GetIdByName(ctx, db, name, &models.Partner{})
	if err != nil {
		return uuid.Nil, err
	}
	if prt, ok := entity.(*models.Partner); ok {
		return prt.ID, nil
	}
	return uuid.Nil, fmt.Errorf("no partner found with name %v", name)
}

func GetOrganizationId(ctx context.Context, db bun.IDB, name string) (uuid.UUID, error) {
	entity, err := GetIdByName(ctx, db, name, &models.Organization{})
	if err != nil {
		return uuid.Nil, err
	}
	if org, ok := entity.(*models.Organization); ok {
		return org.ID, nil
	}
	return uuid.Nil, fmt.Errorf("no organization found with name %v", name)
}

func GetProjectId(ctx context.Context, db bun.IDB, name string) (uuid.UUID, error) {
	entity, err := GetIdByName(ctx, db, name, &models.Project{})
	if err != nil {
		return uuid.Nil, err
	}
	if proj, ok := entity.(*models.Project); ok {
		return proj.ID, nil
	}
	return uuid.Nil, fmt.Errorf("no project found with name %v", name)
}

func GetPartnerName(ctx context.Context, db bun.IDB, id uuid.UUID) (string, error) {
	entity, err := GetNameById(ctx, db, id, &models.Partner{})
	if err != nil {
		return "", err
	}
	if prt, ok := entity.(*models.Partner); ok {
		return prt.Name, nil
	}
	return "", fmt.Errorf("no partner found with id %v", id)
}

func GetOrganizationName(ctx context.Context, db bun.IDB, id uuid.UUID) (string, error) {
	entity, err := GetNameById(ctx, db, id, &models.Organization{})
	if err != nil {
		return "", err
	}
	if org, ok := entity.(*models.Organization); ok {
		return org.Name, nil
	}
	return "", fmt.Errorf("no organization found with id %v", id)
}

func GetProjectName(ctx context.Context, db bun.IDB, id uuid.UUID) (string, error) {
	entity, err := GetNameById(ctx, db, id, &models.Project{})
	if err != nil {
		return "", err
	}
	if proj, ok := entity.(*models.Project); ok {
		return proj.Name, nil
	}
	return "", fmt.Errorf("no project found with id %v", id)
}
