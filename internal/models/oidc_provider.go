package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type OIDCProvider struct {
	bun.BaseModel `bun:"table:authsrv_oidc_provider,alias:oidcprovider"`

	Id             uuid.UUID `bun:"id,type:uuid,pk,default:uuid_generate_v4()"`
	Name           string    `bun:"name,notnull,unique"` // unique constraint on id,name cols group
	Description    string    `bun:"description"`
	OrganizationId uuid.UUID `bun:"organization_id,type:uuid"`
	PartnerId      uuid.UUID `bun:"partner_id,type:uuid"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt     time.Time `bun:"modified_at,notnull,default:current_timestamp"`

	ProviderName    string                 `bun:"provider_name,notnull"`
	MapperURL       string                 `bun:"mapper_url"`
	MapperFilename  string                 `bun:"mapper_filename"`
	ClientId        string                 `bun:"client_id,notnull"`
	ClientSecret    string                 `bun:"client_secret,notnull"`
	Scopes          []string               `bun:"scopes,array,notnull"`
	IssuerURL       string                 `bun:"issuer_url,unique,notnull"`
	AuthURL         string                 `bun:"auth_url"`
	TokenURL        string                 `bun:"token_url"`
	RequestedClaims map[string]interface{} `bun:"requested_claims,type:jsonb"`
	Predefined      bool                   `bun:"predefined,notnull"`
	Trash           bool                   `bun:"trash,default:false"`
}
