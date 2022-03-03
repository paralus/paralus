package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ClusterToken struct {
	bun.BaseModel `bun:"table:cluster_tokens,alias:tokens"`

	ID             uuid.UUID       `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name           string          `bun:"name,notnull"`
	OrganizationId uuid.UUID       `bun:"organization_id,type:uuid"`
	PartnerId      uuid.UUID       `bun:"partner_id,type:uuid"`
	ProjectId      uuid.UUID       `bun:"project_id,type:uuid"`
	DisplayName    string          `bun:"display_name,notnull"`
	CreatedAt      time.Time       `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt     time.Time       `bun:"modified_at,default:current_timestamp"`
	DeletedAt      time.Time       `bun:"deleted_at"`
	Trash          bool            `bun:"trash,notnull,default:false"`
	Labels         json.RawMessage `bun:"labels,type:jsonb,notnull,default:'{}'"`
	Annotations    json.RawMessage `bun:"annotations,type:jsonb,notnull,default:'{}'"`
	TokenType      string          `bun:"token_type,notnull"`
	State          string          `bun:"state,notnull"`
}
