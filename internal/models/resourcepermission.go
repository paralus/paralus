package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ResourcePermission struct {
	bun.BaseModel `bun:"table:authsrv_resourcepermission,alias:resourcepermission"`

	ID                 uuid.UUID                `bun:"id,type:uuid,pk,default:uuid_generate_v4()"`
	Name               string                   `bun:"name,notnull" json:"name"`
	BaseUrl            string                   `bun:"base_url,notnull" json:"base_url"`
	Description        string                   `bun:"description,notnull"`
	CreatedAt          time.Time                `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt         time.Time                `bun:"modified_at,notnull,default:current_timestamp"`
	Trash              bool                     `bun:"trash,notnull,default:false"`
	ResourceUrls       []map[string]interface{} `bun:"resource_urls,type:jsonb" json:"resource_urls"`
	ResourceActionUrls []map[string]interface{} `bun:"resource_action_urls,type:jsonb" json:"resource_action_urls"`
	Scope              string                   `bun:"scope,notnull"`
}
