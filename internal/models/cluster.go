package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Cluster struct {
	bun.BaseModel `bun:"table:cluster_clusters,alias:cluster"`

	ID                 uuid.UUID       `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationId     uuid.UUID       `bun:"organization_id,type:uuid"`
	PartnerId          uuid.UUID       `bun:"partner_id,type:uuid"`
	ProjectId          uuid.UUID       `bun:"project_id,type:uuid"`
	MetroId            uuid.UUID       `bun:"metro_id,type:uuid"`
	Name               string          `bun:"name,notnull"`
	Description        string          `bun:"description"`
	DisplayName        string          `bun:"display_name,notnull"`
	CreatedAt          time.Time       `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt         time.Time       `bun:"modified_at,default:current_timestamp"`
	DeletedAt          time.Time       `bun:"deleted_at"`
	Trash              bool            `bun:"trash,notnull,default:false"`
	Labels             json.RawMessage `bun:"labels,type:jsonb,notnull,default:'{}'"`
	Annotations        json.RawMessage `bun:"annotations,type:jsonb,notnull,default:'{}'"`
	BlueprintRef       string          `bun:"blueprint_ref,notnull,default:''"`
	ClusterType        string          `bun:"cluster_type"`
	OverrideSelector   string          `bun:"override_selector,notnull,default:''"`
	Token              string          `bun:"token"`
	Conditions         json.RawMessage `bun:"conditions,type:jsonb,notnull,default:'{}'"`
	PublishedBlueprint string          `bun:"published_blueprint,notnull,default:''"`
	SystemTaskCount    int64           `bun:"system_task_count,notnull,default:0"`
	CustomTaskCount    int64           `bun:"custom_task_count,notnull,default:0"`
	AuxiliaryTaskCount int64           `bun:"auxiliary_task_count,notnull,default:0"`
	Extra              json.RawMessage `bun:"extra,type:jsonb,notnull,default:'{}'"`
	ShareMode          string          `bun:"share_mode,default:'CUSTOM'"`
	ProxyConfig        json.RawMessage `bun:"proxy_config,type:jsonb"`
}
