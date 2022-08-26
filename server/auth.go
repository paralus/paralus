package server

import (
	"context"

	authv3 "github.com/paralus/paralus/pkg/auth/v3"
	rpcv3 "github.com/paralus/paralus/proto/rpc/v3"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
)

type authServer struct {
	as authv3.AuthService
}

// NewAuthServer returns new auth server implementation
func NewAuthServer(as authv3.AuthService) rpcv3.AuthServiceServer {
	return &authServer{as}
}

func (s *authServer) IsRequestAllowed(ctx context.Context, ira *v3.IsRequestAllowedRequest) (*v3.IsRequestAllowedResponse, error) {
	return s.as.IsRequestAllowed(ctx, ira)
}
