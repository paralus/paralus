package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type KubeconfigRevocation struct {
	bun.BaseModel `bun:"table:sentry_kubeconfig_revocation,alias:kr"`

	ID             uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationId uuid.UUID `bun:"organization_id,notnull,type:uuid"`
	PartnerId      uuid.UUID `bun:"partner_id,type:uuid,notnull"`
	AccountId      uuid.UUID `bun:"account_id,type:uuid,notnull"`
	RevokedAt      time.Time `bun:"revoked_at"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
	IsSSOUser      bool      `bun:"is_sso_user,default:false"`
}
