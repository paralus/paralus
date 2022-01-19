package dao

import (
	"context"
	"errors"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/rs/xid"
)

var (
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrUsedToken is returned when token has been registered
	ErrUsedToken = errors.New("used token")
)

// ClusterTokenDao is the interface for cluster token operations
type ClusterTokenDao interface {
	// create cluster token
	CreateToken(ctx context.Context, c *models.ClusterToken) error
	//register the token
	RegisterToken(ctx context.Context, token string) (*models.ClusterToken, error)
}

// clusterTokenDao implements ClusterTokenDao
type clusterTokenDao struct {
	dao pg.EntityDAO
}

// ClusterDao return new cluster dao
func NewClusterTokenDao(dao pg.EntityDAO) ClusterTokenDao {
	return &clusterTokenDao{
		dao: dao,
	}
}

// CreateToken creates a token for given cluster name
func (s *clusterTokenDao) CreateToken(ctx context.Context, token *models.ClusterToken) error {
	token.Name = xid.New().String()
	_, err := s.dao.Create(ctx, token)
	return err
}

// registerToken registers the cluster token
func (s *clusterTokenDao) RegisterToken(ctx context.Context, token string) (*models.ClusterToken, error) {

	entity, err := s.dao.GetX(ctx, "name", token, models.ClusterToken{})
	if err != nil {
		return nil, ErrInvalidToken
	}
	ct := entity.(models.ClusterToken)
	ct.State = infrav3.ClusterTokenState_TokenUsed.String()

	s.dao.Update(ctx, ct.ID, ct)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return &ct, nil
}
