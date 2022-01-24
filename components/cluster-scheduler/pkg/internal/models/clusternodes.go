package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ClusterNodes struct {
	bun.BaseModel `bun:"table:cluster_nodes,alias:clusternodes"`

	ID             uuid.UUID       `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	ClusterId      uuid.UUID       `bun:"cluster_id,type:uuid,notnull"`
	OrganizationId uuid.UUID       `bun:"organization_id,type:uuid,notnull"`
	PartnerId      uuid.UUID       `bun:"partner_id,type:uuid,notnull"`
	ProjectId      uuid.UUID       `bun:"project_id,type:uuid,notnull"`
	Name           string          `bun:"name,notnull"`
	DisplayName    string          `bun:"display_name,notnull"`
	CreatedAt      time.Time       `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt     time.Time       `bun:"modified_at"`
	DeletedAt      time.Time       `bun:"deleted_at"`
	Labels         json.RawMessage `bun:"labels,type:jsonb,notnull,default:'{}'"`
	Annotations    json.RawMessage `bun:"annotations,type:jsonb,notnull,default:'{}'"`
	Unschedulable  bool            `bun:"unschedulable,notnull,default:false"`
	Taints         json.RawMessage `bun:"taints,type:jsonb,notnull,default:'{}'"`
	Conditions     json.RawMessage `bun:"conditions,type:jsonb,notnull,default:'{}'"`
	NodeInfo       json.RawMessage `bun:"node_info,type:jsonb,notnull,default:'{}'"`
	State          string          `bun:"state,notnull"`
	Capacity       json.RawMessage `bun:"capacity,type:jsonb,notnull,default:'{}'"`
	Allocatable    json.RawMessage `bun:"allocatable,type:jsonb,notnull,default:'{}'"`
	Allocated      json.RawMessage `bun:"allocated,type:jsonb,notnull,default:'{}'"`
	Ips            json.RawMessage `bun:"ips,type:jsonb,notnull,default:'[]'"`
}
