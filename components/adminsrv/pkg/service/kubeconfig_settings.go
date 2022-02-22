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

// KubeconfigSettingService is the interface for kube config setting operations
type KubeconfigSettingService interface {
	Close() error
	Get(ctx context.Context, orgID string, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error)
	Patch(ctx context.Context, ks *sentry.KubeconfigSetting) error
}

// kubeconfigSettingService implements KubeconfigSettingService
type kubeconfigSettingService struct {
	dao  pg.EntityDAO
	kdao dao.KubeconfigDao
}

// NewKubeconfigSettingService return new kubeconfig setting service
func NewKubeconfigSettingService(db *bun.DB) KubeconfigSettingService {
	edao := pg.NewEntityDAO(db)
	return &kubeconfigSettingService{
		dao:  edao,
		kdao: dao.NewKubeconfigDao(edao),
	}
}

func (krs *kubeconfigSettingService) Close() error {
	return krs.dao.Close()
}

func (kss *kubeconfigSettingService) Get(ctx context.Context, orgID string, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
	oid, err := uuid.Parse(orgID)
	if err != nil {
		_log.Info("organization identifier is empty")
	}
	aid, err := uuid.Parse(accountID)
	if err != nil {
		_log.Info("account identifier is empty")
	}

	kr, err := kss.kdao.GetKubeconfigSetting(ctx, oid, aid, isSSO)
	if err == sql.ErrNoRows {
		return nil, constants.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return prepareKubeCfgSettingResponse(kr), nil
}

func (kss *kubeconfigSettingService) Patch(ctx context.Context, ks *sentry.KubeconfigSetting) error {
	accId, err := uuid.Parse(ks.AccountID)
	if err != nil {
		accId = uuid.Nil
	}
	err = kss.dao.GetInstance().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := kss.kdao.GetKubeconfigSetting(ctx, uuid.MustParse(ks.OrganizationID), accId, ks.IsSSOUser)
		db := convertToKubeCfgSettingModel(ks)
		if err != nil && err == sql.ErrNoRows {
			db.CreatedAt = time.Now()
			return kss.kdao.CreateKubeconfigSetting(ctx, convertToKubeCfgSettingModel(ks))
		}
		db.ModifiedAt = time.Now()
		return kss.kdao.UpdateKubeconfigSetting(ctx, convertToKubeCfgSettingModel(ks))
	})
	return err
}

func prepareKubeCfgSettingResponse(ks *models.KubeconfigSetting) *sentry.KubeconfigSetting {
	return &sentry.KubeconfigSetting{
		Id:                          ks.ID.String(),
		OrganizationID:              ks.OrganizationId.String(),
		PartnerID:                   ks.PartnerId.String(),
		AccountID:                   ks.AccountId.String(),
		Scope:                       ks.Scope,
		ValiditySeconds:             ks.ValiditySeconds,
		CreatedAt:                   timestamppb.New(ks.CreatedAt),
		ModifiedAt:                  timestamppb.New(ks.ModifiedAt),
		EnableSessionCheck:          ks.EnforceRsId,
		IsSSOUser:                   ks.IsSSOUser,
		EnablePrivateRelay:          ks.EnablePrivateRelay,
		EnforceOrgAdminSecretAccess: ks.EnforceOrgAdminSecretAccess,
		DisableWebKubectl:           ks.DisableWebKubectl,
		DisableCLIKubectl:           ks.DisableCLIKubectl,
	}
}

func convertToKubeCfgSettingModel(ks *sentry.KubeconfigSetting) *models.KubeconfigSetting {
	kss := &models.KubeconfigSetting{
		OrganizationId:              uuid.MustParse(ks.OrganizationID),
		Scope:                       ks.Scope,
		ValiditySeconds:             ks.ValiditySeconds,
		EnforceRsId:                 ks.EnableSessionCheck,
		IsSSOUser:                   ks.IsSSOUser,
		DisableWebKubectl:           ks.DisableWebKubectl,
		DisableCLIKubectl:           ks.DisableCLIKubectl,
		EnablePrivateRelay:          ks.EnablePrivateRelay,
		EnforceOrgAdminSecretAccess: ks.EnforceOrgAdminSecretAccess,
	}
	if ks.AccountID != "" {
		kss.AccountId = uuid.MustParse(ks.AccountID)
	}
	if ks.PartnerID != "" {
		kss.PartnerId = uuid.MustParse(ks.PartnerID)
	}
	return kss
}
