package authv3

import (
	"context"

	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
)

type contextKey struct{}

var sessionDataKey contextKey

func NewSessionContext(ctx context.Context, s *commonv3.SessionData) context.Context {
	return context.WithValue(ctx, sessionDataKey, s)
}

func GetSession(ctx context.Context) (*commonv3.SessionData, bool) {
	s, ok := ctx.Value(sessionDataKey).(*commonv3.SessionData)
	return s, ok
}
