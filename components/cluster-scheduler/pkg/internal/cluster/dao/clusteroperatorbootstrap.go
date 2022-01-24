package dao

import (
	"context"
	"database/sql"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/uptrace/bun"
)

var _log = log.GetLogger()

// ClusterOperatorBootstrapDao is the interface for cluster operator bootstrap
type ClusterOperatorBootstrapDao interface {
	// create edge operator bootstrap
	CreateOperatorBootstrap(ctx context.Context, bootstrap *models.ClusterOperatorBootstrap) error
	// GetOperatorBootstrap
	GetOperatorBootstrap(ctx context.Context, edgeID string) (*models.ClusterOperatorBootstrap, error)
}

// clusterOperatorBootstrapDao implements ClusterOperatorBootstrapDao
type clusterOperatorBootstrapDao struct {
	dao pg.EntityDAO
}

// ClusterOperatorBootstrapDao return new cluster credentials dao
func NewClusterOperatorBootstrapDao(dao pg.EntityDAO) ClusterOperatorBootstrapDao {
	return &clusterOperatorBootstrapDao{
		dao: dao,
	}
}

func (es *clusterOperatorBootstrapDao) CreateOperatorBootstrap(ctx context.Context, bootstrap *models.ClusterOperatorBootstrap) error {
	_log.Infow("CreateOperatorBootstrap: Creating operator bootstrap data", "cluster", bootstrap.ClusterId)

	err := es.dao.GetInstance().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var bstrap *models.ClusterOperatorBootstrap

		entity, err := es.dao.GetX(ctx, "edge_id", bootstrap.ClusterId, &bstrap)
		if err != nil {
			_log.Infow("CreateOperatorBootstrap: No existing boostrap data detected", "edge", bootstrap.ClusterId)
		} else {
			_log.Infow("CreateOperatorBootstrap: Removing existing boostrap data", "edge", bootstrap.ClusterId)

			bstrap = entity.(*models.ClusterOperatorBootstrap)
			err = es.dao.DeleteX(ctx, "edge_id", bstrap.ClusterId, bstrap)
			if err != nil {
				_log.Errorw("Error while deleting bootstrap data", "Error", err)
				return err
			}
			_log.Infow("CreateOperatorBootstrap: Deleted existing boostrap data", "cluster", bootstrap.ClusterId)
		}

		_, err = tx.NewInsert().Model(bootstrap).Exec(ctx)
		if err != nil {
			_log.Errorw("Error inserting bootstrap data", "Error", err)
			return err
		}

		_log.Infow("Inserted bootstrap data", "cluster", bootstrap.ClusterId)

		return nil
	})

	if err != nil {
		_log.Errorw("Exception while adding bootstrap data", "Error:", err)
	}
	return nil
}

func (es *clusterOperatorBootstrapDao) GetOperatorBootstrap(ctx context.Context, clusterid string) (*models.ClusterOperatorBootstrap, error) {

	var bootstrap models.ClusterOperatorBootstrap
	entity, err := es.dao.GetX(ctx, "clusterid", clusterid, bootstrap)
	if err != nil {
		_log.Errorw("Error while fetching bootstrap data using tx ", "Error", err)
		return nil, err
	}

	return entity.(*models.ClusterOperatorBootstrap), err

}
