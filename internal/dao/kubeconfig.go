package dao

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/uptrace/bun"
)

func GetKubeconfigRevocation(ctx context.Context, db bun.IDB, orgID, accountID uuid.UUID, isSSOUser bool) (*models.KubeconfigRevocation, error) {
	var kr models.KubeconfigRevocation
	err := db.NewSelect().Model(&kr).
		Where("organization_id = ?", orgID).
		Where("account_id = ?", accountID).
		Where("is_sso_user = ?", isSSOUser).
		Scan(ctx)
	return &kr, err
}

func CreateKubeconfigRevocation(ctx context.Context, db bun.IDB, kr *models.KubeconfigRevocation) error {
	_, err := Create(ctx, db, kr)
	return err
}

func UpdateKubeconfigRevocation(ctx context.Context, db bun.IDB, kr *models.KubeconfigRevocation) error {
	q := db.NewUpdate().Model(kr)

	q = q.Where("organization_id = ?", kr.OrganizationId).
		Where("account_id = ?", kr.AccountId).
		Where("is_sso_user = ?", kr.IsSSOUser)

	q = q.Set("revoked_at = ?", kr.RevokedAt)

	_, err := q.Exec(ctx)
	return err
}

func GetKubeconfigSetting(ctx context.Context, db bun.IDB, orgID, accountID uuid.UUID, issSSO bool) (*models.KubeconfigSetting, error) {
	var ks models.KubeconfigSetting
	var err error = nil
	if len(accountID.String()) > 0 {
		err = db.NewSelect().Model(&ks).
			Where("organization_id = ?", orgID).
			Where("account_id = ?", accountID).
			Where("is_sso_user= ?", issSSO).
			Scan(ctx)
	} else {
		err = db.NewSelect().Model(&ks).
			Where("organization_id = ?", orgID).
			Where("is_sso_user= ?", issSSO).
			Scan(ctx)
	}
	return &ks, err
}

func CreateKubeconfigSetting(ctx context.Context, db bun.IDB, ks *models.KubeconfigSetting) error {
	if ks.AccountId == uuid.Nil {
		ks.Scope = sentry.KubeconfigSettingOrganizationScope
	} else {
		ks.Scope = sentry.KubeconfigSettingUserScope
	}
	_, err := Create(ctx, db, ks)
	return err
}

func UpdateKubeconfigSetting(ctx context.Context, db bun.IDB, ks *models.KubeconfigSetting) error {
	q := db.NewUpdate().Model(ks)

	q = q.Where("organization_id = ?", ks.OrganizationId).
		Where("account_id = ?", ks.AccountId).
		Where("is_sso_user= ?", ks.IsSSOUser)

	q = q.Set("modified_at = ?", time.Now()).
		Set("validity_seconds = ?", ks.ValiditySeconds).
		Set("sa_validity_seconds = ?", ks.SaValiditySeconds).
		Set("enforce_rsid = ?", ks.EnforceRsId).
		Set("is_sso_user = ?", ks.IsSSOUser).
		Set("disable_web_kubectl = ?", ks.DisableWebKubectl).
		Set("disable_cli_kubectl = ?", ks.DisableCLIKubectl).
		Set("enable_privaterelay = ?", ks.EnablePrivateRelay).
		Set("enforce_orgadmin_secret_access = ?", ks.EnforceOrgAdminSecretAccess) // allow only orgadmin to access secret API

	_, err := q.Exec(ctx)
	return err
}

func GetkubectlClusterSettings(ctx context.Context, db bun.IDB, orgID uuid.UUID, name string) (*models.KubectlClusterSetting, error) {
	var kc models.KubectlClusterSetting
	err := db.NewSelect().Model(&kc).
		Where("organization_id = ?", orgID).
		Where("name = ?", name).Scan(ctx)
	return &kc, err
}

func CreatekubectlClusterSettings(ctx context.Context, db bun.IDB, kc *models.KubectlClusterSetting) error {
	_, err := Create(ctx, db, kc)
	return err
}

func UpdatekubectlClusterSettings(ctx context.Context, db bun.IDB, kc *models.KubectlClusterSetting) error {
	q := db.NewUpdate().Model(kc)

	q = q.Where("organization_id = ?", kc.OrganizationId).
		Where("name = ?", kc.Name)

	q = q.Set("disable_web_kubectl = ?", kc.DisableWebKubectl)
	q = q.Set("disable_cli_kubectl = ?", kc.DisableCliKubectl)

	_, err := q.Exec(ctx)
	return err
}
