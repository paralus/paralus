package dao

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/cluster/constants"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/query"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ClusterDao is the interface for cluster operations
type ClusterDao interface {
	// create cluster
	CreateCluster(ctx context.Context, c *models.Cluster) error
	// create or update cluster
	UpdateCluster(ctx context.Context, c *models.Cluster) error
	//list clusters
	ListClusters(ctx context.Context, qo commonv3.QueryOptions) ([]models.Cluster, error)
	// delete cluster
	DeleteCluster(ctx context.Context, c *models.Cluster) error
	// get cluster
	GetCluster(ctx context.Context, c *models.Cluster) (*models.Cluster, error)
	//get cluster for token
	GetClusterForToken(ctx context.Context, token string) (cluster *models.Cluster, err error)
}

// clusterDao implements ClusterDao
type clusterDao struct {
	cdao  pg.EntityDAO
	ctdao ClusterTokenDao
	pcdao ProjectClusterDao
}

// ClusterDao return new cluster dao
func NewClusterDao(edao pg.EntityDAO) ClusterDao {
	return &clusterDao{
		cdao:  edao,
		ctdao: NewClusterTokenDao(edao),
		pcdao: NewProjectClusterDao(edao),
	}
}

func (s *clusterDao) CreateCluster(ctx context.Context, cluster *models.Cluster) error {

	err := s.cdao.GetInstance().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		err := s.createCluster(ctx, cluster, tx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *clusterDao) createCluster(ctx context.Context, cluster *models.Cluster, tx bun.Tx) error {

	clstrToken := &models.ClusterToken{
		OrganizationId: cluster.OrganizationId,
		PartnerId:      cluster.PartnerId,
		ProjectId:      cluster.ProjectId,
		CreatedAt:      time.Now(),
	}
	err := s.ctdao.CreateToken(ctx, clstrToken)
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

func (s *clusterDao) UpdateCluster(ctx context.Context, c *models.Cluster) error {

	_, err := s.cdao.Update(ctx, c.ID, c)
	if err != nil {
		return err
	}

	return nil
}

func (s *clusterDao) GetCluster(ctx context.Context, cluster *models.Cluster) (*models.Cluster, error) {
	entity, err := s.cdao.GetByID(ctx, cluster.ID, cluster)
	if err != nil {
		return nil, err
	}

	return entity.(*models.Cluster), nil
}

func (s *clusterDao) DeleteCluster(ctx context.Context, c *models.Cluster) error {
	return s.cdao.Delete(ctx, c.ID, c)
}

func (s *clusterDao) ListClusters(ctx context.Context, qo commonv3.QueryOptions) (clusters []models.Cluster, err error) {

	q := s.cdao.GetInstance().
		NewSelect().
		Model(&clusters)

	fmt.Printf("query before ", q)

	q, err = query.Select(q, &qo)
	if err != nil {
		_log.Errorw("error building query ", q)
		return nil, err
	}

	if qo.BlueprintRef != "" {
		q.Where("?TableAlias.blueprint_ref = ?", qo.BlueprintRef)
	}

	q = query.Paginate(q, &qo)

	if qo.ProjectID != "" {
		q.ColumnExpr("?TableAlias.*").
			Join("JOIN cluster_project_cluster AS pc ON pc.cluster_id = ?TableAlias.id").
			WhereGroup("grp", func(sq *bun.SelectQuery) *bun.SelectQuery {
				q = q.Where("pc.project_id = ?", uuid.MustParse(qo.ProjectID)).
					WhereOr("cluster.share_mode = ?", infrav3.ClusterShareMode_CUSTOM)
				return q
			})
	}

	fmt.Printf("query is ", q)

	err = q.Scan(ctx, clusters)
	return clusters, err
}

func (s *clusterDao) GetClusterForToken(ctx context.Context, token string) (cluster *models.Cluster, err error) {
	entity, err := s.cdao.GetX(ctx, "token", token, &cluster)
	return entity.(*models.Cluster), err
}
