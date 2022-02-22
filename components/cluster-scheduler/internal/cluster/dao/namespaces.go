package dao

import (
	"context"
	"time"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/internal/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/RafaySystems/rcloud-base/components/common/proto/types/scheduler"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ClusterNamespacesDao is the interface for cluster namespaces operations
type ClusterNamespacesDao interface {
	// Get Namespace
	GetNamespace(ctx context.Context, clusterID uuid.UUID, name string) (models.ClusterNamespace, error)
	// GetNamespaces
	GetNamespaces(ctx context.Context, clusterID uuid.UUID) ([]models.ClusterNamespace, error)
	// GetNamespacesForConditions
	GetNamespacesForConditions(ctx context.Context, clusterID uuid.UUID, conditions []scheduler.ClusterNamespaceCondition) ([]models.ClusterNamespace, int, error)
	// UpdateNamespaceStatus
	UpdateNamespaceStatus(ctx context.Context, updated *models.ClusterNamespace) error
	// GetNamespaceHashes
	GetNamespaceHashes(ctx context.Context, clusterID uuid.UUID) ([]infrav3.NameHash, error)
}

// clusterNamespacesDao implements ClusterNamespacesDao
type clusterNamespacesDao struct {
	dao pg.EntityDAO
}

// ClusterNamespacesDao return new cluster namespaces dao
func NewClusterNamespacesDao(dao pg.EntityDAO) ClusterNamespacesDao {
	return &clusterNamespacesDao{
		dao: dao,
	}
}

func (s clusterNamespacesDao) GetNamespace(ctx context.Context, clusterID uuid.UUID, name string) (models.ClusterNamespace, error) {

	var cn models.ClusterNamespace

	err := s.dao.GetInstance().NewSelect().Model(&cn).
		Where("cluster_id = ?", clusterID).
		Where("name = ?", name).
		Scan(ctx)

	if err != nil {
		return cn, err
	}

	return cn, nil
}

func (s clusterNamespacesDao) GetNamespaces(ctx context.Context, clusterID uuid.UUID) ([]models.ClusterNamespace, error) {
	var cns []models.ClusterNamespace

	_, err := s.dao.GetX(ctx, "cluster_id", clusterID, &cns)
	return cns, err
}

func (s clusterNamespacesDao) GetNamespacesForConditions(ctx context.Context, clusterID uuid.UUID, conditions []scheduler.ClusterNamespaceCondition) ([]models.ClusterNamespace, int, error) {
	var cns []models.ClusterNamespace

	q := s.dao.GetInstance().NewSelect().Model(&cns).Where("cluster_id = ?", clusterID)

	for _, condition := range conditions {
		q.WhereGroup("", func(sq *bun.SelectQuery) *bun.SelectQuery {
			sq = sq.Where(conditionStatusQ, int(condition.Type), map[string]string{
				"status": condition.Status.String(),
			})
			since := time.Now().Add(-time.Minute)
			if !condition.LastUpdated.IsValid() {
				since = condition.LastUpdated.AsTime().Add(-time.Minute)
			}

			sq = sq.Where(conditionLastUpdatedQ, int(condition.Type), since)

			return sq
		})
	}

	count, err := q.ScanAndCount(ctx)
	return cns, count, err
}

func (s clusterNamespacesDao) UpdateNamespaceStatus(ctx context.Context, updated *models.ClusterNamespace) error {

	_, err := s.dao.GetInstance().NewUpdate().Model(updated).
		Set("conditions = ?", updated.Conditions).
		Set("status = ?", updated.Status).
		Where("cluster_id = ?", updated.ClusterId).
		Where("name = ?", updated.Name).
		Exec(ctx, updated)

	return err
}

func (s clusterNamespacesDao) GetNamespaceHashes(ctx context.Context, clusterID uuid.UUID) ([]infrav3.NameHash, error) {

	var nameHashes []infrav3.NameHash

	err := s.dao.GetInstance().NewSelect().
		Model((*models.ClusterNamespace)(nil)).
		Column("name", "hash").
		//TODO: to be changed to ClusterTaskDeleted later once task is supported
		ColumnExpr(deletingExpr, 3, map[string]string{"status": commonv3.RafayConditionStatus_NotSet.String()}).
		Where("cluster_id = ?", clusterID).
		Scan(ctx, &nameHashes)

	if err != nil {
		return nil, err
	}

	return nameHashes, nil
}
