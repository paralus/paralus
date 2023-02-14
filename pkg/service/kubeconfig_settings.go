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
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KubeconfigSettingService is the interface for kube config setting operations
type KubeconfigSettingService interface {
	Get(ctx context.Context, orgID string, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error)
	Patch(ctx context.Context, ks *sentry.KubeconfigSetting) error
}

// kubeconfigSettingService implements KubeconfigSettingService
type kubeconfigSettingService struct {
	db *bun.DB
}

// NewKubeconfigSettingService return new kubeconfig setting service
func NewKubeconfigSettingService(db *bun.DB) KubeconfigSettingService {
	return &kubeconfigSettingService{db}
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

	kr, err := dao.GetKubeconfigSetting(ctx, kss.db, oid, aid, isSSO)
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
	const maxSeconds = 30 * 24 * 60 * 60
	const minSeconds = 10 * 60

	minTimeDuration := time.Second * time.Duration(minSeconds) // min. 10 mins
	maxTimeDuration := time.Second * time.Duration(maxSeconds) // max. 30 days

	saValidityDuration := time.Second * time.Duration(ks.SaValiditySeconds)
	if saValidityDuration < minTimeDuration || saValidityDuration > maxTimeDuration {
		maxTimeDisplay, _ := time.ParseDuration(fmt.Sprintf("%ds", maxSeconds))
		minTimeDisplay, _ := time.ParseDuration(fmt.Sprintf("%ds", minSeconds))
		return fmt.Errorf("invalid sa validity duration. should be between %.0f mins and %.0f hours", minTimeDisplay.Minutes(), maxTimeDisplay.Hours())
	}

	return kss.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := dao.GetKubeconfigSetting(ctx, tx, uuid.MustParse(ks.OrganizationID), accId, ks.IsSSOUser)
		db := convertToKubeCfgSettingModel(ks)
		if err != nil && err == sql.ErrNoRows {
			db.CreatedAt = time.Now()
			return dao.CreateKubeconfigSetting(ctx, tx, convertToKubeCfgSettingModel(ks))
		}
		db.ModifiedAt = time.Now()
		return dao.UpdateKubeconfigSetting(ctx, tx, convertToKubeCfgSettingModel(ks))
	})
}

func prepareKubeCfgSettingResponse(ks *models.KubeconfigSetting) *sentry.KubeconfigSetting {
	return &sentry.KubeconfigSetting{
		Id:                          ks.ID.String(),
		OrganizationID:              ks.OrganizationId.String(),
		PartnerID:                   ks.PartnerId.String(),
		AccountID:                   ks.AccountId.String(),
		Scope:                       ks.Scope,
		ValiditySeconds:             ks.ValiditySeconds,
		SaValiditySeconds:           ks.SaValiditySeconds,
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
		SaValiditySeconds:           ks.SaValiditySeconds,
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
