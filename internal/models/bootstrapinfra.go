package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type BootstrapInfra struct {
	bun.BaseModel `bun:"table:sentry_bootstrap_infra,alias:bi"`

	Name           string          `bun:"name,pk,notnull"`
	OrganizationId uuid.UUID       `bun:"organization_id,type:uuid"`
	PartnerId      uuid.UUID       `bun:"partner_id,type:uuid"`
	ProjectId      uuid.UUID       `bun:"project_id,type:uuid"`
	DisplayName    string          `bun:"display_name,notnull"`
	CreatedAt      time.Time       `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt     time.Time       `bun:"modified_at"`
	DeletedAt      time.Time       `bun:"deleted_at"`
	Labels         json.RawMessage `bun:"labels,type:jsonb,notnull,default:'{}'"`
	Annotations    json.RawMessage `bun:"annotations,type:jsonb,notnull,default:'{}'"`
	CaCert         string          `bun:"ca_cert,notnull"`
	CaKey          string          `bun:"ca_key,notnull"`
}
