package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ApiKey struct {
	bun.BaseModel `bun:"table:authsrv_apikey,alias:apikey"`

	ID              uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name            string    `bun:"name,notnull"`
	Description     string    `bun:"description,notnull"`
	CreatedAt       time.Time `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt      time.Time `bun:"modified_at,default:current_timestamp"`
	Trash           bool      `bun:"trash,notnull,default:false"`
	Key             string    `bun:"key,notnull"`
	AccountID       uuid.UUID `bun:"account_id,type:uuid"`
	OrganizationID  uuid.UUID `bun:"organization_id,type:uuid"`
	PartnerID       uuid.UUID `bun:"partner_id,type:uuid"`
	SecretMigration string    `bun:"secret_migration"`
	Secret          string    `bun:"secret,notnull"`
}
