package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/internal/constants"
	"github.com/RafaySystems/rcloud-base/components/adminsrv/internal/dao"
	"github.com/RafaySystems/rcloud-base/components/adminsrv/internal/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/components/common/proto/types/sentry"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KubeconfigRevocation is the interface for bootstrap operations
type KubeconfigRevocationService interface {
	Close() error
	Get(ctx context.Context, orgID string, accountID string, isSSOUser bool) (*sentry.KubeconfigRevocation, error)
	Patch(ctx context.Context, kr *sentry.KubeconfigRevocation) error
}

// bootstrapService implements BootstrapService
type kubeconfigRevocationService struct {
	dao  pg.EntityDAO
	kdao dao.KubeconfigDao
}

// NewKubeconfigRevocation return new kubeconfig revocation service
func NewKubeconfigRevocationService(db *bun.DB) KubeconfigRevocationService {
	edao := pg.NewEntityDAO(db)
	return &kubeconfigRevocationService{
		dao:  edao,
		kdao: dao.NewKubeconfigDao(edao),
	}
}

func (krs *kubeconfigRevocationService) Close() error {
	return krs.dao.Close()
}

func (krs *kubeconfigRevocationService) Get(ctx context.Context, orgID string, accountID string, isSSOUser bool) (*sentry.KubeconfigRevocation, error) {
	kr, err := krs.kdao.GetKubeconfigRevocation(ctx, uuid.MustParse(orgID), uuid.MustParse(accountID), isSSOUser)
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
	err := krs.dao.GetInstance().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := krs.kdao.GetKubeconfigRevocation(ctx, uuid.MustParse(kr.OrganizationID), uuid.MustParse(kr.AccountID), kr.IsSSOUser)
		if err != nil && err == sql.ErrNoRows {
			kcr := convertToModel(kr)
			kcr.CreatedAt = time.Now()
			return krs.kdao.CreateKubeconfigRevocation(ctx, kcr)
		}
		return krs.kdao.UpdateKubeconfigRevocation(ctx, convertToModel(kr))
	})
	return err
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
