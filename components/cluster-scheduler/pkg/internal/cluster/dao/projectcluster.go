package dao

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
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
