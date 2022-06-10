package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/constants"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
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
	al *zap.Logger
}

// NewKubeconfigRevocation return new kubeconfig revocation service
func NewKubeconfigRevocationService(db *bun.DB, al *zap.Logger) KubeconfigRevocationService {
	return &kubeconfigRevocationService{db, al}
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
	accId := uuid.MustParse(kr.AccountID)
	entity, err := dao.GetM(ctx, krs.db, map[string]interface{}{"id": accId}, &models.KratosIdentities{})
	if err != nil {
		return err
	}
	if usr, ok := entity.(*models.KratosIdentities); ok {
		// We need user info inorder to add the audit log
		err = krs.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			_, err := dao.GetKubeconfigRevocation(ctx, tx, uuid.MustParse(kr.OrganizationID), accId, kr.IsSSOUser)
			if err != nil && err == sql.ErrNoRows {
				kcr := convertToModel(kr)
				kcr.CreatedAt = time.Now()
				return dao.CreateKubeconfigRevocation(ctx, tx, kcr)
			}
			return dao.UpdateKubeconfigRevocation(ctx, tx, convertToModel(kr))
		})
		if err != nil {
			return err
		}
		RevokeKubeconfigAuditEvent(ctx, krs.al, getUserTraits(usr.Traits).Email)
		return nil
	}

	return fmt.Errorf("unable to fetch user")
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
