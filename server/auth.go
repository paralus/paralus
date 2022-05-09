package server

import (
	"context"

	authv3 "github.com/RafayLabs/rcloud-base/pkg/auth/v3"
	rpcv3 "github.com/RafayLabs/rcloud-base/proto/rpc/v3"
	v3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
)

type authServer struct {
	as authv3.AuthService
}

// NewAuthServer returns new auth server implementation
func NewAuthServer(as authv3.AuthService) rpcv3.AuthServer {
	return &authServer{as}
}

func (s *authServer) IsRequestAllowed(ctx context.Context, ira *v3.IsRequestAllowedRequest) (*v3.IsRequestAllowedResponse, error) {
	return s.as.IsRequestAllowed(ctx, ira)
}
