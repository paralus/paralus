package dao

import (
	"context"
	"errors"

	"github.com/RafayLabs/rcloud-base/internal/dao"
	"github.com/RafayLabs/rcloud-base/internal/models"
	infrav3 "github.com/RafayLabs/rcloud-base/proto/types/infrapb/v3"
	"github.com/rs/xid"
	"github.com/uptrace/bun"
)

var (
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrUsedToken is returned when token has been registered
	ErrUsedToken = errors.New("used token")
)

// CreateToken creates a token for given cluster name
func CreateToken(ctx context.Context, db bun.IDB, token *models.ClusterToken) error {
	token.Name = xid.New().String()
	_, err := dao.Create(ctx, db, token)
	return err
}

// registerToken registers the cluster token
func RegisterToken(ctx context.Context, db bun.IDB, token string) (*models.ClusterToken, error) {

	entity, err := dao.GetX(ctx, db, "name", token, &models.ClusterToken{})
	if err != nil {
		return nil, ErrInvalidToken
	}
	ct := entity.(*models.ClusterToken)
	ct.State = infrav3.ClusterTokenState_TokenUsed.String()

	dao.Update(ctx, db, ct.ID, ct)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return ct, nil
}
