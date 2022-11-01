package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/common"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
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

func IsInternalRequest(ctx context.Context) bool {
	v := ctx.Value(common.SessionInternalKey)
	b, ok := v.(bool)
	return ok && b
}

// getLastLoginTime return latest authenticated time from sessions.
func getLastLoginTime(sessions []models.KratosSessions) time.Time {
	var auths []int64
	for _, s := range sessions {
		auths = append(auths, s.AuthenticatedAt.UnixMilli())
	}
	latest := auths[0]
	for _, auth := range auths {
		if auth > latest {
			latest = auth
		}
	}
	return time.UnixMilli(latest)
}
