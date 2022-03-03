package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type KratosIdentities struct {
	bun.BaseModel `bun:"table:identities,alias:identities"`

	ID             uuid.UUID              `bun:"id,type:uuid,pk"`
	SchemaId       string                 `bun:"schema_id,notnull"`
	Traits         map[string]interface{} `bun:"traits,type:jsonb"`
	CreatedAt      time.Time              `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time              `bun:"updated_at,notnull,default:current_timestamp"`
	State          string                 `bun:"state,notnull"`
	StateChangedAt time.Time              `bun:"state_changed_at,notnull,default:current_timestamp"`
	NId            uuid.UUID              `bun:"nid,type:uuid,pk"`
}
