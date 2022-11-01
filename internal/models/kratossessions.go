package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type KratosSessions struct {
	bun.BaseModel `bun:"table:sessions,alias:sessions"`

	ID              uuid.UUID `bun:"id,type:uuid,pk"`
	AuthenticatedAt time.Time `bun:"authenticated_at,notnull"`
	IdentityId      uuid.UUID `bun:"identity_id,notnull"`
	// Fill other columns of sessions table when necessary
}
