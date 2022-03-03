package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ResourceRolePermission struct {
	bun.BaseModel `bun:"table:authsrv_resourcerolepermission,alias:resourcerolepermission"`

	ID                   uuid.UUID `bun:"id,type:uuid,pk,default:uuid_generate_v4()"`
	Name                 string    `bun:"name,notnull" json:"name"`
	Description          string    `bun:"description,notnull"`
	CreatedAt            time.Time `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt           time.Time `bun:"modified_at,notnull,default:current_timestamp"`
	Trash                bool      `bun:"trash,notnull,default:false"`
	ResourcePermissionId uuid.UUID `bun:"resource_permission_id,type:uuid"`
	ResourceRoleId       uuid.UUID `bun:"resource_role_id,type:uuid"`
}
