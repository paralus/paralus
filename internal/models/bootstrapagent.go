package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type BootstrapAgent struct {
	bun.BaseModel `bun:"table:sentry_bootstrap_agent,alias:ba"`

	ID             uuid.UUID       `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name           string          `bun:"name,notnull"`
	OrganizationId uuid.UUID       `bun:"organization_id,type:uuid"`
	PartnerId      uuid.UUID       `bun:"partner_id,type:uuid"`
	ProjectId      uuid.UUID       `bun:"project_id,type:uuid"`
	TemplateRef    string          `bun:"template_ref,notnull"`
	AgentMode      string          `bun:"agent_mode,notnull"`
	DisplayName    string          `bun:"display_name,notnull"`
	CreatedAt      time.Time       `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt     time.Time       `bun:"modified_at"`
	DeletedAt      time.Time       `bun:"deleted_at"`
	Labels         json.RawMessage `bun:"labels,type:jsonb,notnull,default:'{}'"`
	Annotations    json.RawMessage `bun:"annotations,type:jsonb,notnull,default:'{}'"`
	Token          string          `bun:"token,notnull"`
	TokenState     string          `bun:"token_state,notnull"`
	IPAddress      string          `bun:"ip_address,notnull"`
	LastCheckedIn  time.Time       `bun:"last_checked_in"`
	Fingerprint    string          `bun:"fingerprint,notnull"`
}
