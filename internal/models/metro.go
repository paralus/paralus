package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Metro struct {
	bun.BaseModel `bun:"table:cluster_metro,alias:metro"`

	ID             uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name           string    `bun:"name,notnull"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt     time.Time `bun:"modified_at,notnull,default:current_timestamp"`
	Trash          bool      `bun:"trash,notnull,default:false"`
	Latitude       string    `bun:"latitude,notnull"`
	Longitude      string    `bun:"longitude,notnull"`
	City           string    `bun:"city"`
	State          string    `bun:"state"`
	Country        string    `bun:"country"`
	CountryCode    string    `bun:"cc"`
	StateCode      string    `bun:"st"`
	OrganizationId uuid.UUID `bun:"organization_id,type:uuid"`
	PartnerId      uuid.UUID `bun:"partner_id,type:uuid,notnull"`
}
