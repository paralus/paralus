package dao

import (
	"context"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/uptrace/bun"
)

func CreateProjectCluster(ctx context.Context, db bun.IDB, pc *models.ProjectCluster) error {
	_, err := dao.Create(ctx, db, pc)
	if err != nil {
		return err
	}
	return nil
}

func GetProjectsForCluster(ctx context.Context, db bun.IDB, clusterID uuid.UUID) ([]models.ProjectCluster, error) {
	var projectClusters []models.ProjectCluster
	err := db.NewSelect().Model(&projectClusters).Where("cluster_id = ?", clusterID).Where("trash = ?", false).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return projectClusters, nil
}

func DeleteProjectsForCluster(ctx context.Context, db bun.IDB, clusterID uuid.UUID) error {
	return dao.DeleteX(ctx, db, "cluster_id", clusterID, &models.ProjectCluster{})
}

// Check if the project in scope is owner of the cluster
func ValidateClusterAccess(ctx context.Context, db bun.IDB, opts commonv3.QueryOptions) (bool, error) {
	var _c models.Cluster
	q, err := query.Select(db.NewSelect().Model(&_c), &opts)
	if err != nil {
		return false, err
	}

	count, err := q.ScanAndCount(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
