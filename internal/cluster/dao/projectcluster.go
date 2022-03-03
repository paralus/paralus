package dao

import (
	"context"

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/pkg/query"
	commonv3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	"github.com/google/uuid"
)

// ProjectClusterDao is the interface for project cluster operations
type ProjectClusterDao interface {
	// create project cluster
	CreateProjectCluster(ctx context.Context, pc *models.ProjectCluster) error
	// get projects for cluster
	GetProjectsForCluster(ctx context.Context, clusterID uuid.UUID) ([]models.ProjectCluster, error)
	// delete projects for cluster
	DeleteProjectsForCluster(ctx context.Context, clusterID uuid.UUID) error
	// Validate if the project in scope is owner of the cluster
	ValidateClusterAccess(ctx context.Context, opts commonv3.QueryOptions) (bool, error)
}

// projectClusterDao implements ProjectClusterDao
type projectClusterDao struct {
	dao pg.EntityDAO
}

// ProjectClusterDao return new project cluster dao
func NewProjectClusterDao(dao pg.EntityDAO) ProjectClusterDao {
	return &projectClusterDao{
		dao: dao,
	}
}

func (s *projectClusterDao) CreateProjectCluster(ctx context.Context, pc *models.ProjectCluster) error {
	_, err := s.dao.Create(ctx, pc)
	if err != nil {
		return err
	}
	return nil
}

func (s *projectClusterDao) GetProjectsForCluster(ctx context.Context, clusterID uuid.UUID) ([]models.ProjectCluster, error) {
	var projectClusters []models.ProjectCluster
	err := s.dao.GetInstance().NewSelect().Model(&projectClusters).Where("cluster_id = ?", clusterID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return projectClusters, nil
}

func (s *projectClusterDao) DeleteProjectsForCluster(ctx context.Context, clusterID uuid.UUID) error {
	return s.dao.DeleteX(ctx, "cluster_id", clusterID, &models.ProjectCluster{})
}

// Check if the project in scope is owner of the cluster
func (s *projectClusterDao) ValidateClusterAccess(ctx context.Context, opts commonv3.QueryOptions) (bool, error) {
	var _c models.Cluster
	q, err := query.Select(s.dao.GetInstance().NewSelect().Model(&_c), &opts)
	if err != nil {
		return false, err
	}

	count, err := q.ScanAndCount(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
