package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Account struct {
	bun.BaseModel `bun:"table:identities,alias:identities"`

	ID        uuid.UUID              `bun:"id,type:uuid"`
	Traits    map[string]interface{} `bun:"traits,type:jsonb"`
	State     string                 `bun:"state,notnull"`
	Username  string                 `bun:"username"`
	LastLogin time.Time              `bun:"lastlogin"`
}
