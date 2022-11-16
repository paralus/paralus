package models

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type AuditLog struct {
	bun.BaseModel `bun:"table:audit_logs,alias:auditlog"`

	Tag  string          `bun:"tag,notnull"`
	Time time.Time       `bun:"time,notnull"`
	Data json.RawMessage `bun:"data,type:jsonb,notnull"`
}

type AggregatorData struct {
	Count int64
	Key   string
}
