package models

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type Partner struct {
	bun.BaseModel `bun:"table:authsrv_partner,alias:partner"`

	ID                        int64           `bun:"id,pk,autoincrement"`
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
	ParentId                  int64           `bun:"parent_id"`
	HelpdeskEmail             string          `bun:"partner_helpdesk_email,notnull"`
	ProductName               string          `bun:"partner_product_name"`
	SupportTeamName           string          `bun:"support_team_name,notnull"`
	OpsHost                   string          `bun:"ops_host,notnull"`
	FavIconLink               string          `bun:"fav_icon_link,notnull"`
	IsTOTPEnabled             bool            `bun:"is_totp_enabled,notnull"`
	IsSyntheticPartnerEnabled bool            `bun:"is_synthetic_partner_enabled,notnull"`
}
