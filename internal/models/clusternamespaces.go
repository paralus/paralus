package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ClusterNamespace struct {
	bun.BaseModel `bun:"table:cluster_namespaces,alias:cns"`

	ClusterId  uuid.UUID       `bun:"cluster_id,type:uuid"`
	Name       string          `bun:"name,notnull"`
	Hash       string          `bun:"hash,notnull"`
	DeletedAt  time.Time       `bun:"deleted_at"`
	Type       string          `bun:"type,notnull"`
	Namespace  json.RawMessage `bun:"namespace,type:jsonb,notnull"`
	Conditions json.RawMessage `bun:"conditions,type:jsonb,notnull,default:'[]'"`
	Status     json.RawMessage `bun:"status,type:jsonb,notnull,default:'{}'"`
}
