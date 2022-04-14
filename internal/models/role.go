package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Role struct {
	bun.BaseModel `bun:"table:authsrv_resourcerole,alias:resourcerole"`

	ID             uuid.UUID `bun:"id,type:uuid,pk,default:uuid_generate_v4()"`
	Name           string    `bun:"name,notnull"`
	Description    string    `bun:"description,notnull"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt     time.Time `bun:"modified_at,notnull,default:current_timestamp"`
	Trash          bool      `bun:"trash,notnull,default:false"`
	OrganizationId uuid.UUID `bun:"organization_id,type:uuid"`
	PartnerId      uuid.UUID `bun:"partner_id,type:uuid"`
	IsGlobal       bool      `bun:"is_global,notnull,default:true"`
	Builtin        bool      `bun:"builtin,notnull,default:true"`
	Scope          string    `bun:"scope,notnull"`
}
