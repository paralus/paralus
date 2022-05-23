package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type GroupAccount struct {
	bun.BaseModel `bun:"table:authsrv_groupaccount,alias:groupaccount"`

	ID          uuid.UUID `bun:"id,type:uuid,pk,default:uuid_generate_v4()"`
	Name        string    `bun:"name,notnull"`
	Description string    `bun:"description,notnull"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`
	ModifiedAt  time.Time `bun:"modified_at,notnull,default:current_timestamp"`
	Trash       bool      `bun:"trash,notnull,default:false"`
	AccountId   uuid.UUID `bun:"account_id,type:uuid"`
	GroupId     uuid.UUID `bun:"group_id,type:uuid"`
	Active      bool      `bun:"active,notnull"`

	Account *KratosIdentities `bun:"rel:has-one,join:account_id=id"`
	Group   *Group            `bun:"rel:has-one,join:group_id=id"`
}
