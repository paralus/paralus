package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type KratosIdentityCredentialTypes struct {
	bun.BaseModel `bun:"table:identity_credential_types,alias:ict"`

	ID   uuid.UUID `bun:"id,type:uuid,pk"`
	Name string    `bun:"name,type:string"`
}

type KratosIdentityCredentials struct {
	bun.BaseModel `bun:"table:identity_credentials,alias:ic"`

	ID                       uuid.UUID                     `bun:"id,type:uuid,pk"`
	IdentityID               uuid.UUID                     `bun:"identity_id,type:uuid"`
	IdentityCredentialTypeID uuid.UUID                     `bun:"identity_credential_type_id,type:uuid"`
	IdentityCredentialType   KratosIdentityCredentialTypes `bun:"rel:has-one,join:identity_credential_type_id=id"`
}

type KratosIdentities struct {
	bun.BaseModel `bun:"table:identities,alias:identities"`

	ID                 uuid.UUID                 `bun:"id,type:uuid,pk"`
	SchemaId           string                    `bun:"schema_id,notnull"`
	Traits             map[string]interface{}    `bun:"traits,type:jsonb"`
	CreatedAt          time.Time                 `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt          time.Time                 `bun:"updated_at,notnull,default:current_timestamp"`
	State              string                    `bun:"state,notnull"`
	StateChangedAt     time.Time                 `bun:"state_changed_at,notnull,default:current_timestamp"`
	NId                uuid.UUID                 `bun:"nid,type:uuid,pk"`
	IdentityCredential KratosIdentityCredentials `bun:"rel:has-one,join:id=identity_id"`
	MetadataPublic     map[string]interface{}    `bun:"metadata_public,type:jsonb"`
}
