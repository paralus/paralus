package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Idp struct {
	bun.BaseModel `bun:"table:authsrv_idp,alias:idp"`

	Id          uuid.UUID `bun:"id,type:uuid,pk,default:uuid_generate_v4()"`
	Name        string    `bun:"name,notnull,unique"`
	Description string    `bun:"description"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt  time.Time `bun:"modified_at,notnull,default:current_timestamp"`

	IdpName string `bun:"idp_name,notnull"`
	Domain  string `bun:"domain,notnull,unique"`
	// Deprecated
	// AcsURL             string    `bun:"acs_url,notnull,unique"`
	OrganizationId     uuid.UUID `bun:"organization_id,type:uuid"`
	PartnerId          uuid.UUID `bun:"partner_id,type:uuid"`
	SsoURL             string    `bun:"sso_url"`
	IdpCert            string    `bun:"idp_cert"`
	SpCert             string    `bun:"sp_cert"`
	SpKey              string    `bun:"sp_key"`
	MetadataURL        string    `bun:"metadata_url"`
	MetadataFilename   string    `bun:"metadata_filename"`
	Metadata           []byte    `bun:"metadata"`
	GroupAttributeName string    `bun:"group_attribute_name"`
	SaeEnabled         bool      `bun:"is_sae_enabled"`
	Trash              bool      `bun:"trash,default:false"`
}
