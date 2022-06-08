package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/constants"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KubectlClusterSettingsService is the interface for kubectl cluster setting operations
type KubectlClusterSettingsService interface {
	Get(ctx context.Context, orgID string, clusterID string) (*sentry.KubectlClusterSettings, error)
	Patch(ctx context.Context, kc *sentry.KubectlClusterSettings) error
}

// kubectlClusterSettingsService implements KubectlClusterSettingsService
type kubectlClusterSettingsService struct {
	db *bun.DB
}

// NewKubectlClusterSettingsService return new kubectl cluster setting service
func NewkubectlClusterSettingsService(db *bun.DB) KubectlClusterSettingsService {
	return &kubectlClusterSettingsService{db}
}

func (kcs *kubectlClusterSettingsService) Get(ctx context.Context, orgID string, clusterID string) (*sentry.KubectlClusterSettings, error) {
	kc, err := dao.GetkubectlClusterSettings(ctx, kcs.db, uuid.MustParse(orgID), clusterID)
	if err == sql.ErrNoRows {
		return nil, constants.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return prepareKubectlSettingResponse(kc), nil
}

func (kcs *kubectlClusterSettingsService) Patch(ctx context.Context, kc *sentry.KubectlClusterSettings) error {
	return kcs.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := dao.GetkubectlClusterSettings(ctx, tx, uuid.MustParse(kc.OrganizationID), kc.Name)
		if err != nil {
			if err == sql.ErrNoRows {
				kcsdb := convertToKubeCtlSettingModel(kc)
				kcsdb.CreatedAt = time.Now()
				return dao.CreatekubectlClusterSettings(ctx, tx, kcsdb)
			}
			return err
		}
		kcsdb := convertToKubeCtlSettingModel(kc)
		kcsdb.ModifiedAt = time.Now()
		return dao.UpdatekubectlClusterSettings(ctx, tx, kcsdb)
	})
}

func convertToKubeCtlSettingModel(kcs *sentry.KubectlClusterSettings) *models.KubectlClusterSetting {
	kcsm := &models.KubectlClusterSetting{
		Name:              kcs.Name,
		OrganizationId:    uuid.MustParse(kcs.OrganizationID),
		DisableWebKubectl: kcs.DisableWebKubectl,
		DisableCliKubectl: kcs.DisableCLIKubectl,
	}
	if kcs.PartnerID != "" {
		kcsm.PartnerId, _ = uuid.Parse(kcs.PartnerID)
	}
	return kcsm
}

func prepareKubectlSettingResponse(kcs *models.KubectlClusterSetting) *sentry.KubectlClusterSettings {
	kc := &sentry.KubectlClusterSettings{
		Name:              kcs.Name,
		OrganizationID:    kcs.OrganizationId.String(),
		DisableWebKubectl: kcs.DisableWebKubectl,
		DisableCLIKubectl: kcs.DisableCliKubectl,
		CreatedAt:         timestamppb.New(kcs.CreatedAt),
		ModifiedAt:        timestamppb.New(kcs.ModifiedAt),
	}
	if kcs.PartnerId != uuid.Nil {
		kc.PartnerID = kcs.PartnerId.String()
	}
	return kc
}
