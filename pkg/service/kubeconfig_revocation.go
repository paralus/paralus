package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/RafaySystems/rcloud-base/internal/constants"
	"github.com/RafaySystems/rcloud-base/internal/dao"
	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/proto/types/sentry"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KubeconfigRevocation is the interface for bootstrap operations
type KubeconfigRevocationService interface {
	Get(ctx context.Context, orgID string, accountID string, isSSOUser bool) (*sentry.KubeconfigRevocation, error)
	Patch(ctx context.Context, kr *sentry.KubeconfigRevocation) error
}

// bootstrapService implements BootstrapService
type kubeconfigRevocationService struct {
	db *bun.DB
}

// NewKubeconfigRevocation return new kubeconfig revocation service
func NewKubeconfigRevocationService(db *bun.DB) KubeconfigRevocationService {
	return &kubeconfigRevocationService{db}
}

func (krs *kubeconfigRevocationService) Get(ctx context.Context, orgID string, accountID string, isSSOUser bool) (*sentry.KubeconfigRevocation, error) {
	kr, err := dao.GetKubeconfigRevocation(ctx, krs.db, uuid.MustParse(orgID), uuid.MustParse(accountID), isSSOUser)
	if err == sql.ErrNoRows {
		return nil, constants.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return prepareKubeCfgRevocationResponse(kr), nil
}

func prepareKubeCfgRevocationResponse(kr *models.KubeconfigRevocation) *sentry.KubeconfigRevocation {
	return &sentry.KubeconfigRevocation{
		OrganizationID: kr.OrganizationId.String(),
		PartnerID:      kr.PartnerId.String(),
		AccountID:      kr.AccountId.String(),
		RevokedAt:      timestamppb.New(kr.RevokedAt),
		IsSSOUser:      kr.IsSSOUser,
		CreatedAt:      timestamppb.New(kr.CreatedAt),
	}
}

func (krs *kubeconfigRevocationService) Patch(ctx context.Context, kr *sentry.KubeconfigRevocation) error {
	return krs.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := dao.GetKubeconfigRevocation(ctx, tx, uuid.MustParse(kr.OrganizationID), uuid.MustParse(kr.AccountID), kr.IsSSOUser)
		if err != nil && err == sql.ErrNoRows {
			kcr := convertToModel(kr)
			kcr.CreatedAt = time.Now()
			return dao.CreateKubeconfigRevocation(ctx, tx, kcr)
		}
		return dao.UpdateKubeconfigRevocation(ctx, tx, convertToModel(kr))
	})
}

func convertToModel(kr *sentry.KubeconfigRevocation) *models.KubeconfigRevocation {
	return &models.KubeconfigRevocation{
		OrganizationId: uuid.MustParse(kr.OrganizationID),
		PartnerId:      uuid.MustParse(kr.PartnerID),
		AccountId:      uuid.MustParse(kr.AccountID),
		RevokedAt:      kr.RevokedAt.AsTime(),
		IsSSOUser:      kr.IsSSOUser,
	}
}
