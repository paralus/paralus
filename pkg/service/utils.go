package service

import (
	"context"

	"github.com/RafayLabs/rcloud-base/internal/dao"
	"github.com/RafayLabs/rcloud-base/pkg/common"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func getPartnerOrganization(ctx context.Context, db bun.IDB, partner, org string) (uuid.UUID, uuid.UUID, error) {
	partnerId, err := dao.GetPartnerId(ctx, db, partner)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	organizationId, err := dao.GetOrganizationId(ctx, db, org)
	if err != nil {
		return partnerId, uuid.Nil, err
	}
	return partnerId, organizationId, nil

}

func GetSessionDataFromContext(ctx context.Context) (*commonv3.SessionData, bool) {
	s, ok := ctx.Value(common.SessionDataKey).(*commonv3.SessionData)
	return s, ok
}
