package dao

import (
	"context"
	"time"

	"github.com/RafaySystems/rcloud-base/internal/dao"
	"github.com/RafaySystems/rcloud-base/internal/models"
	commonv3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/proto/types/infrapb/v3"
	"github.com/RafaySystems/rcloud-base/proto/types/scheduler"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func GetNamespace(ctx context.Context, db bun.IDB, clusterID uuid.UUID, name string) (models.ClusterNamespace, error) {

	var cn models.ClusterNamespace

	err := db.NewSelect().Model(&cn).
		Where("cluster_id = ?", clusterID).
		Where("name = ?", name).
		Scan(ctx)

	if err != nil {
		return cn, err
	}

	return cn, nil
}

func GetNamespaces(ctx context.Context, db bun.IDB, clusterID uuid.UUID) ([]models.ClusterNamespace, error) {
	var cns []models.ClusterNamespace

	_, err := dao.GetX(ctx, db, "cluster_id", clusterID, &cns)
	return cns, err
}

func GetNamespacesForConditions(ctx context.Context, db bun.IDB, clusterID uuid.UUID, conditions []scheduler.ClusterNamespaceCondition) ([]models.ClusterNamespace, int, error) {
	var cns []models.ClusterNamespace

	q := db.NewSelect().Model(&cns).Where("cluster_id = ?", clusterID)

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

func UpdateNamespaceStatus(ctx context.Context, db bun.IDB, updated *models.ClusterNamespace) error {

	_, err := db.NewUpdate().Model(updated).
		Set("conditions = ?", updated.Conditions).
		Set("status = ?", updated.Status).
		Where("cluster_id = ?", updated.ClusterId).
		Where("name = ?", updated.Name).
		Exec(ctx, updated)

	return err
}

func GetNamespaceHashes(ctx context.Context, db bun.IDB, clusterID uuid.UUID) ([]infrav3.NameHash, error) {

	var nameHashes []infrav3.NameHash

	err := db.NewSelect().
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
