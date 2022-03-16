package dao

import (
	"context"
	"errors"

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	infrav3 "github.com/RafaySystems/rcloud-base/proto/types/infrapb/v3"
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
	_, err := pg.Create(ctx, db, token)
	return err
}

// registerToken registers the cluster token
func RegisterToken(ctx context.Context, db bun.IDB, token string) (*models.ClusterToken, error) {

	entity, err := pg.GetX(ctx, db, "name", token, &models.ClusterToken{})
	if err != nil {
		return nil, ErrInvalidToken
	}
	ct := entity.(*models.ClusterToken)
	ct.State = infrav3.ClusterTokenState_TokenUsed.String()

	pg.Update(ctx, db, ct.ID, ct)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return ct, nil
}
