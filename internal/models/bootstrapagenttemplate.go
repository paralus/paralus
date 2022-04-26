package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type BootstrapAgentTemplate struct {
	bun.BaseModel `bun:"table:sentry_bootstrap_agent_template,alias:bat"`

	Name                   string          `bun:"name,pk,notnull"`
	OrganizationId         uuid.UUID       `bun:"organization_id,type:uuid"`
	PartnerId              uuid.UUID       `bun:"partner_id,type:uuid"`
	ProjectId              uuid.UUID       `bun:"project_id,type:uuid"`
	InfraRef               string          `bun:"infra_ref,notnull"`
	DisplayName            string          `bun:"display_name,notnull"`
	CreatedAt              time.Time       `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt             time.Time       `bun:"modified_at"`
	DeletedAt              time.Time       `bun:"deleted_at"`
	Labels                 json.RawMessage `bun:"labels,type:jsonb,notnull,default:'{}'"`
	Annotations            json.RawMessage `bun:"annotations,type:jsonb,notnull,default:'{}'"`
	AutoRegister           bool            `bun:"auto_register,notnull,default:false"`
	IgnoreMultipleRegister bool            `bun:"ignore_multiple_register,notnull,default:false"`
	AutoApprove            bool            `bun:"auto_approve,notnull,default:false"`
	TemplateType           string          `bun:"template_type,notnull"`
	Hosts                  json.RawMessage `bun:"hosts,type:jsonb,notnull,default:'{}'"`
	Token                  string          `bun:"token,notnull"`
	InclusterTemplate      string          `bun:"incluster_template,notnull"`
	OutofclusterTemplate   string          `bun:"outofcluster_template,notnull"`
	Trash                  bool            `bun:"type:bool"`
}
