package authv3

import (
	"context"

	"github.com/RafayLabs/rcloud-base/pkg/common"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
)

func NewSessionContext(ctx context.Context, s *commonv3.SessionData) context.Context {
	return context.WithValue(ctx, common.SessionDataKey, s)
}

func GetSession(ctx context.Context) (*commonv3.SessionData, bool) {
	s, ok := ctx.Value(common.SessionDataKey).(*commonv3.SessionData)
	return s, ok
}
