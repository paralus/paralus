package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Partner struct {
	bun.BaseModel `bun:"table:authsrv_partner,alias:partner"`

	ID                        uuid.UUID       `bun:"id,type:uuid,pk,default:uuid_generate_v4()"`
	Name                      string          `bun:"name,notnull"`
	Description               string          `bun:"description,notnull"`
	CreatedAt                 time.Time       `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt                time.Time       `bun:"modified_at,notnull,default:current_timestamp"`
	Trash                     bool            `bun:"trash,notnull"`
	Settings                  json.RawMessage `bun:"settings,notnull,type:jsonb"`
	Host                      string          `bun:"host,notnull"`
	Domain                    string          `bun:"domain,notnull"`
	TosLink                   string          `bun:"tos_link,notnull"`
	LogoLink                  string          `bun:"logo_link,notnull"`
	NotificationEmail         string          `bun:"notification_email,notnull"`
	ParentId                  uuid.UUID       `bun:"parent_id,type:uuid"`
	HelpdeskEmail             string          `bun:"partner_helpdesk_email,notnull"`
	ProductName               string          `bun:"partner_product_name"`
	SupportTeamName           string          `bun:"support_team_name,notnull"`
	OpsHost                   string          `bun:"ops_host,notnull"`
	FavIconLink               string          `bun:"fav_icon_link,notnull"`
	IsTOTPEnabled             bool            `bun:"is_totp_enabled,notnull"`
	IsSyntheticPartnerEnabled bool            `bun:"is_synthetic_partner_enabled,notnull"`
}
