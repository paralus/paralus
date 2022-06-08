package dao

import (
	"context"

	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/log"
	"github.com/uptrace/bun"
)

var _log = log.GetLogger()

func CreateOperatorBootstrap(ctx context.Context, db bun.Tx, bootstrap *models.ClusterOperatorBootstrap) error {
	_log.Infow("CreateOperatorBootstrap: Creating operator bootstrap data", "cluster", bootstrap.ClusterId)

	var bstrap *models.ClusterOperatorBootstrap

	entity, err := dao.GetX(ctx, db, "edge_id", bootstrap.ClusterId, &bstrap)
	if err != nil {
		_log.Infow("CreateOperatorBootstrap: No existing bootstrap data detected", "edge", bootstrap.ClusterId)
	} else {
		_log.Infow("CreateOperatorBootstrap: Removing existing bootstrap data", "edge", bootstrap.ClusterId)

		bstrap = entity.(*models.ClusterOperatorBootstrap)
		err = dao.DeleteX(ctx, db, "edge_id", bstrap.ClusterId, bstrap)
		if err != nil {
			_log.Errorw("Error while deleting bootstrap data", "Error", err)
			return err
		}
		_log.Infow("CreateOperatorBootstrap: Deleted existing bootstrap data", "cluster", bootstrap.ClusterId)
	}

	_, err = db.NewInsert().Model(bootstrap).Exec(ctx)
	if err != nil {
		_log.Errorw("Error inserting bootstrap data", "Error", err)
		return err
	}

	_log.Infow("Inserted bootstrap data", "cluster", bootstrap.ClusterId)

	return nil
}

func GetOperatorBootstrap(ctx context.Context, db bun.IDB, clusterid string) (*models.ClusterOperatorBootstrap, error) {

	var bootstrap models.ClusterOperatorBootstrap
	entity, err := dao.GetX(ctx, db, "clusterid", clusterid, bootstrap)
	if err != nil {
		_log.Errorw("Error while fetching bootstrap data using tx ", "Error", err)
		return nil, err
	}

	return entity.(*models.ClusterOperatorBootstrap), err

}
