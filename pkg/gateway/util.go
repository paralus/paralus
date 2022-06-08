package gateway

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// IsGatewayRequest returns true if the request is originated from
// Paralus Gateway
func IsGatewayRequest(ctx context.Context) bool {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(GatewayRequest); vals != nil {
			return true
		}
		return false
	}
	return false
}
