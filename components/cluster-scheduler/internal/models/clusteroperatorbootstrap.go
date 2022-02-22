package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ClusterOperatorBootstrap struct {
	bun.BaseModel `bun:"table:cluster_operator_bootstrap,alias:operator_bootstrap"`

	ClusterId   uuid.UUID `bun:"cluster_id,type:uuid,notnull"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt  time.Time `bun:"modified_at,notnull,default:current_timestamp"`
	Trash       bool      `bun:"trash,notnull,default:false"`
	YamlContent string    `bun:"yaml_content"`
}
