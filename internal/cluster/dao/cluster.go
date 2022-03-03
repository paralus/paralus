package dao

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/RafaySystems/rcloud-base/internal/cluster/constants"
	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	commonv3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
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
	// update relay config information
	UpdateClusterAnnotations(ctx context.Context, c *models.Cluster) error
	// Notify channel
	Notify(chanName, value string) error
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

func (s *clusterDao) UpdateClusterAnnotations(ctx context.Context, c *models.Cluster) error {

	_, err := s.cdao.GetInstance().NewUpdate().Model((*models.Cluster)(nil)).
		Set("annotations = ?", c.Annotations).
		Where("id = ?", c.ID).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *clusterDao) GetCluster(ctx context.Context, cluster *models.Cluster) (*models.Cluster, error) {

	if cluster.ID != uuid.Nil {
		_, err := s.cdao.GetByID(ctx, cluster.ID, cluster)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := s.cdao.GetByName(ctx, cluster.Name, cluster)
		if err != nil {
			return nil, err
		}
	}

	return cluster, nil
}

func (s *clusterDao) DeleteCluster(ctx context.Context, c *models.Cluster) error {
	_, err := s.cdao.GetInstance().
		NewUpdate().Model(c).
		Set("trash = ?", true).
		Set("deleted_at = ?", time.Now()).
		Where("id = ?", c.ID).Exec(ctx)
	return err
}

func (s *clusterDao) ListClusters(ctx context.Context, qo commonv3.QueryOptions) (clusters []models.Cluster, err error) {

	pid := uuid.NullUUID{UUID: uuid.MustParse(qo.Partner), Valid: true}
	oid := uuid.NullUUID{UUID: uuid.MustParse(qo.Organization), Valid: true}
	prid := uuid.NullUUID{UUID: uuid.MustParse(qo.Project), Valid: true}

	err = s.cdao.ListByProject(ctx, pid, oid, prid, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters, err
}

func (s *clusterDao) GetClusterForToken(ctx context.Context, token string) (cluster *models.Cluster, err error) {
	entity, err := s.cdao.GetX(ctx, "token", token, &models.Cluster{})
	if err != nil {
		return nil, err
	}
	return entity.(*models.Cluster), err
}

func (s *clusterDao) Notify(chanName, value string) error {
	_, err := s.cdao.GetInstance().Exec("NOTIFY ?, ?", bun.Ident(chanName), value)
	return err
}
