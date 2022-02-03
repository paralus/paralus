package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Organization struct {
	bun.BaseModel `bun:"table:authsrv_organization,alias:organization"`

	ID                       uuid.UUID       `bun:"id,type:uuid,pk,default:uuid_generate_v4()"`
	Name                     string          `bun:"name,notnull"`
	Description              string          `bun:"description,notnull"`
	CreatedAt                time.Time       `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt               time.Time       `bun:"modified_at,notnull,default:current_timestamp"`
	Trash                    bool            `bun:"trash,notnull"`
	Settings                 json.RawMessage `bun:"settings,notnull,type:jsonb"`
	BillingAddress           string          `bun:"billing_address,notnull,type:text"`
	PartnerId                uuid.UUID       `bun:"partner_id,notnull,type:uuid"`
	Active                   bool            `bun:"active,notnull"`
	Approved                 bool            `bun:"approved,notnull"`
	Type                     string          `bun:"type,notnull"`
	AddressLine1             string          `bun:"address_line1,notnull,type:text"`
	AddressLine2             string          `bun:"address_line2,notnull,type:text"`
	City                     string          `bun:"city,notnull,type:text"`
	Country                  string          `bun:"country,notnull,type:text"`
	Phone                    string          `bun:"phone,notnull,type:text"`
	State                    string          `bun:"state,notnull,type:text"`
	Zipcode                  string          `bun:"zipcode,notnull,type:text"`
	DeletedName              string          `bun:"deleted_name"`
	IsPrivate                bool            `bun:"is_private"`
	IsTOTPEnabled            bool            `bun:"is_totp_enabled,notnull"`
	AreClustersShared        bool            `bun:"are_clusters_shared,notnull"`
	PspsEnabled              bool            `bun:"psps_enabled,default:true"`
	CustomPspsEnabled        bool            `bun:"custom_psps_enabled"`
	DefaultBlueprintsEnabled bool            `bun:"default_blueprints_enabled,default:true"`
	Referer                  string          `bun:"referer"`
}
