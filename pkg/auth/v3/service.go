package authv3

import (
	"context"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
)

type authService struct {
	ac authContext
}

type AuthService interface {
	IsRequestAllowed(context.Context, *commonv3.IsRequestAllowedRequest) (*commonv3.IsRequestAllowedResponse, error)
}

func NewAuthService(ac authContext) AuthService {
	return &authService{ac}
}

// Auth is exposed as an external service so that other modules like
// prompt can call into this inorder to authenticate. This will be
// made use of using `remoteAuthMiddleware` in other services.
func (s *authService) IsRequestAllowed(ctx context.Context, req *commonv3.IsRequestAllowedRequest) (*commonv3.IsRequestAllowedResponse, error) {
	return s.ac.IsRequestAllowed(ctx, req)
}
