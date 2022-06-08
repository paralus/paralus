package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/pkg/utils"
	"github.com/uptrace/bun"
)

// NamespaceService is the interface for namespace operations
type NamespaceService interface {
	// GetProjectNamespaces
	GetProjectNamespaces(ctx context.Context, projectID uuid.UUID) ([]string, error)
	GetAccountProjectNamespaces(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID) ([]string, error)
	GetGroupProjectNamespaces(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID) ([]string, error)
}

// namespaceService implements NamespaceService
type namespaceService struct {
	db *bun.DB
}

// NewNamespaceService return new namespace service
func NewNamespaceService(db *bun.DB) NamespaceService {
	return &namespaceService{db}
}

func (s *namespaceService) GetProjectNamespaces(ctx context.Context, projectID uuid.UUID) ([]string, error) {

	cns, err := dao.GetProjectNamespaces(ctx, s.db, projectID)
	if err != nil {
		return nil, err
	}

	return utils.Unique(cns), nil
}

func (s *namespaceService) GetAccountProjectNamespaces(ctx context.Context, projectID, accountID uuid.UUID) ([]string, error) {
	cns, err := dao.GetAccountProjectNamespaces(ctx, s.db, projectID, accountID)
	if err != nil {
		return nil, err
	}

	return utils.Unique(cns), nil
}

func (s *namespaceService) GetGroupProjectNamespaces(ctx context.Context, projectID, accountID uuid.UUID) ([]string, error) {
	cns, err := dao.GetGroupProjectNamespaces(ctx, s.db, projectID, accountID)
	if err != nil {
		return nil, err
	}

	return utils.Unique(cns), nil
}
