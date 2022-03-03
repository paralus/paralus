package models

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AccountPermission struct {
	bun.BaseModel `bun:"table:sentry_account_permission,alias:sap"`

	AccountId      uuid.UUID       `bun:"account_id,type:uuid"`
	ProjecttId     string          `bun:"project_id"`
	OrganizationId uuid.UUID       `bun:"organization_id,type:uuid"`
	PartnerId      uuid.UUID       `bun:"partner_id,type:uuid"`
	RoleName       string          `bun:"role_name"`
	IsGlobal       bool            `bun:"is_global"`
	Scope          string          `bun:"scope"`
	PermissionName string          `bun:"permission_name"`
	BaseUrl        string          `bun:"base_url"`
	Urls           json.RawMessage `bun:"urls,type:jsonb"`
}
