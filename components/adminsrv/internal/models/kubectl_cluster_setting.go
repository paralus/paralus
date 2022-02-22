package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type KubectlClusterSetting struct {
	bun.BaseModel `bun:"table:sentry_kubectl_cluster_settings,alias:kc"`

	Name              string    `bun:"name,pk,notnull"`
	OrganizationId    uuid.UUID `bun:"organization_id,notnull,type:uuid"`
	PartnerId         uuid.UUID `bun:"partner_id,type:uuid,notnull"`
	DisableWebKubectl bool      `bun:"disable_web_kubectl,notnull,default:false"`
	DisableCliKubectl bool      `bun:"disable_cli_kubectl,notnull,default:false"`
	CreatedAt         time.Time `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt        time.Time `bun:"modified_at"`
	DeletedAt         time.Time `bun:"deleted_at"`
}
