package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type KubeconfigSetting struct {
	bun.BaseModel `bun:"table:sentry_kubeconfig_setting,alias:ks"`

	ID                          uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationId              uuid.UUID `bun:"organization_id,notnull,type:uuid"`
	PartnerId                   uuid.UUID `bun:"partner_id,type:uuid,notnull"`
	AccountId                   uuid.UUID `bun:"account_id,type:uuid,notnull"`
	Scope                       string    `bun:"scope,notnull"`
	ValiditySeconds             int64     `bun:"validity_seconds,notnull,default:0"`
	SaValiditySeconds           int64     `bun:"sa_validity_seconds,notnull,default:28800"`
	CreatedAt                   time.Time `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt                  time.Time `bun:"modified_at"`
	DeletedAt                   time.Time `bun:"deleted_at"`
	EnforceRsId                 bool      `bun:"enforce_rsid,default:false"`
	DisableAllAudit             bool      `bun:"disable_all_audit,default:false"`
	DisableCmdAudit             bool      `bun:"disable_cmd_audit,default:false"`
	IsSSOUser                   bool      `bun:"is_sso_user,default:false"`
	DisableWebKubectl           bool      `bun:"disable_web_kubectl,default:false"`
	DisableCLIKubectl           bool      `bun:"disable_cli_kubectl,default:false"`
	EnablePrivateRelay          bool      `bun:"enable_privaterelay,default:false"`
	EnforceOrgAdminSecretAccess bool      `bun:"enforce_orgadmin_secret_access,default:false"`
}
