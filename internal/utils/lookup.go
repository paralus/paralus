package utils

import (
	"context"
	"fmt"

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// TODO: could use a better name
type lookup struct {
	dao pg.EntityDAO
}

type Lookup interface {
	GetPartnerId(context.Context, string) (uuid.UUID, error)
	GetOrganizationId(context.Context, string) (uuid.UUID, error)
	GetProjectId(context.Context, string) (uuid.UUID, error)

	GetPartnerName(context.Context, uuid.UUID) (string, error)
	GetOrganizationName(context.Context, uuid.UUID) (string, error)
	GetProjectName(context.Context, uuid.UUID) (string, error)
}

func NewLookup(db *bun.DB) Lookup {
	return &lookup{
		dao: pg.NewEntityDAO(db),
	}
}

func (l *lookup) GetPartnerId(ctx context.Context, name string) (uuid.UUID, error) {
	entity, err := l.dao.GetIdByName(ctx, name, &models.Partner{})
	if err != nil {
		return uuid.Nil, err
	}
	if prt, ok := entity.(*models.Partner); ok {
		return prt.ID, nil
	}
	return uuid.Nil, fmt.Errorf("no partner found with name %v", name)
}

func (l *lookup) GetOrganizationId(ctx context.Context, name string) (uuid.UUID, error) {
	entity, err := l.dao.GetIdByName(ctx, name, &models.Organization{})
	if err != nil {
		return uuid.Nil, err
	}
	if org, ok := entity.(*models.Organization); ok {
		return org.ID, nil
	}
	return uuid.Nil, fmt.Errorf("no organization found with name %v", name)
}

func (l *lookup) GetProjectId(ctx context.Context, name string) (uuid.UUID, error) {
	entity, err := l.dao.GetIdByName(ctx, name, &models.Project{})
	if err != nil {
		return uuid.Nil, err
	}
	if proj, ok := entity.(*models.Project); ok {
		return proj.ID, nil
	}
	return uuid.Nil, fmt.Errorf("no project found with name %v", name)
}

func (l *lookup) GetPartnerName(ctx context.Context, id uuid.UUID) (string, error) {
	entity, err := l.dao.GetNameById(ctx, id, &models.Partner{})
	if err != nil {
		return "", err
	}
	if prt, ok := entity.(*models.Partner); ok {
		return prt.Name, nil
	}
	return "", fmt.Errorf("no partner found with id %v", id)
}

func (l *lookup) GetOrganizationName(ctx context.Context, id uuid.UUID) (string, error) {
	entity, err := l.dao.GetNameById(ctx, id, &models.Organization{})
	if err != nil {
		return "", err
	}
	if org, ok := entity.(*models.Organization); ok {
		return org.Name, nil
	}
	return "", fmt.Errorf("no organization found with id %v", id)
}

func (l *lookup) GetProjectName(ctx context.Context, id uuid.UUID) (string, error) {
	entity, err := l.dao.GetNameById(ctx, id, &models.Project{})
	if err != nil {
		return "", err
	}
	if proj, ok := entity.(*models.Project); ok {
		return proj.Name, nil
	}
	return "", fmt.Errorf("no project found with id %v", id)
}
