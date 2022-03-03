package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SSOAccountGroupProjectRole struct {
	bun.BaseModel `bun:"table:sentry_ssoaccount_group_project_roles,alias:acc"`

	Id                    uuid.UUID `bun:"id,type:uuid"`
	Username              string    `bun:"username"`
	RoleName              string    `bun:"role_name"`
	ProjectId             string    `bun:"project_id"`
	ProjectName           string    `bun:"project_name"`
	GroupName             string    `bun:"group_name"`
	AccountOrganizationId uuid.UUID `bun:"account_organization_id,type:uuid"`
	OrganizationId        uuid.UUID `bun:"organization_id,type:uuid"`
	PartnerId             uuid.UUID `bun:"partner_id,type:uuid"`
	Scope                 string    `bun:"scope"`
	LastLogin             time.Time `bun:"last_login"`
	CreatedAt             time.Time `bun:"created_at"`
	FirstName             string    `bun:"first_name"`
	LastName              string    `bun:"last_name"`
	Phone                 string    `bun:"phone"`
	Name                  string    `bun:"name"`
	LastLogout            time.Time `bun:"last_logout"`
}
