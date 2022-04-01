package models

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AccountPermission struct {
	bun.BaseModel `bun:"table:sentry_account_permission,alias:sap"`

	AccountId      uuid.UUID       `bun:"account_id,type:uuid"`
	ProjectId      uuid.UUID       `bun:"project_id,type:uuid"`
	GroupId        uuid.UUID       `bun:"group_id,type:uuid"`
	RoleId         uuid.UUID       `bun:"role_id,type:uuid"`
	RoleName       string          `bun:"role_name,type:string"`
	OrganizationId uuid.UUID       `bun:"organization_id,type:uuid"`
	PartnerId      uuid.UUID       `bun:"partner_id,type:uuid"`
	IsGlobal       bool            `bun:"is_global,notnull,default:true"` // does not matter
	Scope          string          `bun:"scope,notnull"`
	PermissionName string          `bun:"permission_name,type:string"`
	BaseUrl        string          `bun:"base_url,type:string"`
	Urls           json.RawMessage `bun:"urls,type:jsonb"`
}
