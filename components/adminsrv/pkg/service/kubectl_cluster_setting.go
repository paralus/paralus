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

// KubectlClusterSettingsService is the interface for kubectl cluster setting operations
type KubectlClusterSettingsService interface {
	Close() error
	Get(ctx context.Context, orgID string, clusterID string) (*sentry.KubectlClusterSettings, error)
	Patch(ctx context.Context, kc *sentry.KubectlClusterSettings) error
}

// kubectlClusterSettingsService implements KubectlClusterSettingsService
type kubectlClusterSettingsService struct {
	dao  pg.EntityDAO
	kdao dao.KubeconfigDao
}

// NewKubectlClusterSettingsService return new kubectl cluster setting service
func NewkubectlClusterSettingsService(db *bun.DB) KubectlClusterSettingsService {
	edao := pg.NewEntityDAO(db)
	return &kubectlClusterSettingsService{
		dao:  edao,
		kdao: dao.NewKubeconfigDao(edao),
	}
}

func (kcs *kubectlClusterSettingsService) Close() error {
	return kcs.dao.Close()
}

func (kcs *kubectlClusterSettingsService) Get(ctx context.Context, orgID string, clusterID string) (*sentry.KubectlClusterSettings, error) {
	kc, err := kcs.kdao.GetkubectlClusterSettings(ctx, uuid.MustParse(orgID), clusterID)
	if err == sql.ErrNoRows {
		return nil, constants.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return prepareKubectlSettingResponse(kc), nil
}

func (kcs *kubectlClusterSettingsService) Patch(ctx context.Context, kc *sentry.KubectlClusterSettings) error {
	err := kcs.dao.GetInstance().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := kcs.kdao.GetkubectlClusterSettings(ctx, uuid.MustParse(kc.OrganizationID), kc.Name)
		if err != nil {
			if err == sql.ErrNoRows {
				kcsdb := convertToKubeCtlSettingModel(kc)
				kcsdb.CreatedAt = time.Now()
				kcs.kdao.CreatekubectlClusterSettings(ctx, kcsdb)
			}
			return err
		}
		kcsdb := convertToKubeCtlSettingModel(kc)
		kcsdb.ModifiedAt = time.Now()
		return kcs.kdao.UpdatekubectlClusterSettings(ctx, kcsdb)
	})
	return err
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
