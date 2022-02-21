package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
	"google.golang.org/protobuf/types/known/emptypb"
)

type oidcProvider struct {
	service.OIDCProviderService
}

func NewOIDCServer(providerSvc service.OIDCProviderService) rpcv3.OIDCProviderServer {
	return &oidcProvider{providerSvc}
}

func (s *oidcProvider) CreateOIDCProvider(ctx context.Context, p *userv3.OIDCProvider) (*userv3.OIDCProvider, error) {
	return s.Create(ctx, p)
}
func (s *oidcProvider) GetOIDCProvider(ctx context.Context, p *userv3.OIDCProvider) (*userv3.OIDCProvider, error) {
	return s.GetByName(ctx, p)
}
func (s *oidcProvider) ListOIDCProvider(ctx context.Context, p *emptypb.Empty) (*userv3.OIDCProviderList, error) {
	return s.List(ctx)
}
func (s *oidcProvider) UpdateOIDCProvider(ctx context.Context, p *userv3.OIDCProvider) (*userv3.OIDCProvider, error) {
	return s.Update(ctx, p)
}
func (s *oidcProvider) DeleteOIDCProvider(ctx context.Context, p *userv3.OIDCProvider) (*emptypb.Empty, error) {
	// TODO: if successful return 204 NO CONTENT
	return &emptypb.Empty{}, s.Delete(ctx, p)
}
