package dao

import (
	"context"
	"strings"
	"time"

	"github.com/RafaySystems/rcloud-base/internal/cluster/constants"
	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	commonv3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func CreateCluster(ctx context.Context, tx bun.IDB, cluster *models.Cluster) error {

	clstrToken := &models.ClusterToken{
		OrganizationId: cluster.OrganizationId,
		PartnerId:      cluster.PartnerId,
		ProjectId:      cluster.ProjectId,
		CreatedAt:      time.Now(),
	}
	err := CreateToken(ctx, tx, clstrToken)
	if err != nil {
		return err
	}

	if cluster.OverrideSelector == "" {
		cluster.OverrideSelector = strings.Join([]string{constants.OverrideCluster, cluster.Name}, "=")
	}
	cluster.Token = clstrToken.Name

	_, err = tx.NewInsert().Model(cluster).Exec(ctx)
	if err != nil {
		return err
	}

	// set cluster id label
	_, err = tx.NewUpdate().Model(cluster).
		Set("labels = labels || ?", map[string]string{
			constants.ClusterID: cluster.ID.String(),
		}).Where("id = ?", cluster.ID).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func UpdateCluster(ctx context.Context, db bun.IDB, c *models.Cluster) error {

	_, err := pg.Update(ctx, db, c.ID, c)
	if err != nil {
		return err
	}

	return nil
}

func UpdateClusterAnnotations(ctx context.Context, db bun.IDB, c *models.Cluster) error {

	_, err := db.NewUpdate().Model((*models.Cluster)(nil)).
		Set("annotations = ?", c.Annotations).
		Where("id = ?", c.ID).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func GetCluster(ctx context.Context, db bun.IDB, cluster *models.Cluster) (*models.Cluster, error) {

	if cluster.ID != uuid.Nil {
		_, err := pg.GetByID(ctx, db, cluster.ID, cluster)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := pg.GetByName(ctx, db, cluster.Name, cluster)
		if err != nil {
			return nil, err
		}
	}

	return cluster, nil
}

func DeleteCluster(ctx context.Context, db bun.IDB, c *models.Cluster) error {
	_, err := db.
		NewUpdate().Model(c).
		Set("trash = ?", true).
		Set("deleted_at = ?", time.Now()).
		Where("id = ?", c.ID).Exec(ctx)
	return err
}

func ListClusters(ctx context.Context, db bun.IDB, qo commonv3.QueryOptions) (clusters []models.Cluster, err error) {

	pid := uuid.NullUUID{UUID: uuid.MustParse(qo.Partner), Valid: true}
	oid := uuid.NullUUID{UUID: uuid.MustParse(qo.Organization), Valid: true}
	prid := uuid.NullUUID{UUID: uuid.MustParse(qo.Project), Valid: true}

	err = pg.ListByProject(ctx, db, pid, oid, prid, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters, err
}

func GetClusterForToken(ctx context.Context, db bun.IDB, token string) (cluster *models.Cluster, err error) {
	entity, err := pg.GetX(ctx, db, "token", token, &models.Cluster{})
	if err != nil {
		return nil, err
	}
	return entity.(*models.Cluster), err
}

func Notify(db *bun.DB, chanName string, value string) error {
	_, err := db.Exec("NOTIFY ?, ?", bun.Ident(chanName), value)
	return err
}
