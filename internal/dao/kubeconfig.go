package dao

import (
	"context"
	"time"

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/proto/types/sentry"
	"github.com/google/uuid"
)

// KubeconfigDao is the interface for kubeconfig operations
type KubeconfigDao interface {
	GetKubeconfigRevocation(ctx context.Context, orgID, accountID uuid.UUID, isSSOUser bool) (*models.KubeconfigRevocation, error)
	CreateKubeconfigRevocation(ctx context.Context, kr *models.KubeconfigRevocation) error
	UpdateKubeconfigRevocation(ctx context.Context, kr *models.KubeconfigRevocation) error
	GetKubeconfigSetting(ctx context.Context, orgID, accountID uuid.UUID, issSSO bool) (*models.KubeconfigSetting, error)
	CreateKubeconfigSetting(ctx context.Context, ks *models.KubeconfigSetting) error
	UpdateKubeconfigSetting(ctx context.Context, ks *models.KubeconfigSetting) error
	GetkubectlClusterSettings(ctx context.Context, orgID uuid.UUID, name string) (*models.KubectlClusterSetting, error)
	CreatekubectlClusterSettings(ctx context.Context, kc *models.KubectlClusterSetting) error
	UpdatekubectlClusterSettings(ctx context.Context, kc *models.KubectlClusterSetting) error
}

// kubeconfigDao implements BootstrapDao
type kubeconfigDao struct {
	dao pg.EntityDAO
}

// KubeconfigDao return new kube config dao
func NewKubeconfigDao(edao pg.EntityDAO) KubeconfigDao {
	return &kubeconfigDao{
		dao: edao,
	}
}

func (s *kubeconfigDao) GetKubeconfigRevocation(ctx context.Context, orgID, accountID uuid.UUID, isSSOUser bool) (*models.KubeconfigRevocation, error) {
	var kr models.KubeconfigRevocation
	err := s.dao.GetInstance().NewSelect().Model(&kr).
		Where("organization_id = ?", orgID).
		Where("account_id = ?", accountID).
		Where("is_sso_user = ?", isSSOUser).
		Scan(ctx)
	return &kr, err
}

func (s *kubeconfigDao) CreateKubeconfigRevocation(ctx context.Context, kr *models.KubeconfigRevocation) error {
	_, err := s.dao.Create(ctx, kr)
	return err
}

func (s *kubeconfigDao) UpdateKubeconfigRevocation(ctx context.Context, kr *models.KubeconfigRevocation) error {
	q := s.dao.GetInstance().NewUpdate().Model(kr)

	q = q.Where("organization_id = ?", kr.OrganizationId).
		Where("account_id = ?", kr.AccountId).
		Where("is_sso_user = ?", kr.IsSSOUser)

	q = q.Set("revoked_at = ?", kr.RevokedAt)

	_, err := q.Exec(ctx)
	return err
}

func (s *kubeconfigDao) GetKubeconfigSetting(ctx context.Context, orgID, accountID uuid.UUID, issSSO bool) (*models.KubeconfigSetting, error) {
	var ks models.KubeconfigSetting
	err := s.dao.GetInstance().NewSelect().Model(&ks).
		Where("organization_id = ?", orgID).
		Where("account_id = ?", accountID).
		Where("is_sso_user= ?", issSSO).
		Scan(ctx)
	return &ks, err
}

func (s *kubeconfigDao) CreateKubeconfigSetting(ctx context.Context, ks *models.KubeconfigSetting) error {
	if ks.AccountId == uuid.Nil {
		ks.Scope = sentry.KubeconfigSettingOrganizationScope
	} else {
		ks.Scope = sentry.KubeconfigSettingUserScope
	}
	_, err := s.dao.Create(ctx, ks)
	return err
}

func (s *kubeconfigDao) UpdateKubeconfigSetting(ctx context.Context, ks *models.KubeconfigSetting) error {
	q := s.dao.GetInstance().NewUpdate().Model(ks)

	q = q.Where("organization_id = ?", ks.OrganizationId).
		Where("account_id = ?", ks.AccountId).
		Where("is_sso_user= ?", ks.IsSSOUser)

	q = q.Set("modified_at = ?", time.Now()).
		Set("validity_seconds = ?", ks.ValiditySeconds).
		Set("enforce_rsid = ?", ks.EnforceRsId).
		Set("is_sso_user = ?", ks.IsSSOUser).
		Set("disable_web_kubectl = ?", ks.DisableWebKubectl).
		Set("disable_cli_kubectl = ?", ks.DisableCLIKubectl).
		Set("enable_privaterelay = ?", ks.EnablePrivateRelay).
		Set("enforce_orgadmin_secret_access = ?", ks.EnforceOrgAdminSecretAccess) // allow only orgadmin to access secret API

	_, err := q.Exec(ctx)
	return err
}

func (s *kubeconfigDao) GetkubectlClusterSettings(ctx context.Context, orgID uuid.UUID, name string) (*models.KubectlClusterSetting, error) {
	var kc models.KubectlClusterSetting
	err := s.dao.GetInstance().NewSelect().Model(&kc).
		Where("organization_id = ?", orgID).
		Where("name = ?", name).Scan(ctx)
	return &kc, err
}

func (s *kubeconfigDao) CreatekubectlClusterSettings(ctx context.Context, kc *models.KubectlClusterSetting) error {
	_, err := s.dao.Create(ctx, kc)
	return err
}

func (s *kubeconfigDao) UpdatekubectlClusterSettings(ctx context.Context, kc *models.KubectlClusterSetting) error {
	q := s.dao.GetInstance().NewUpdate().Model(kc)

	q = q.Where("organization_id = ?", kc.OrganizationId).
		Where("name = ?", kc.Name)

	q = q.Set("disable_web_kubectl = ?", kc.DisableWebKubectl)
	q = q.Set("disable_cli_kubectl = ?", kc.DisableCliKubectl)

	_, err := q.Exec(ctx)
	return err
}
