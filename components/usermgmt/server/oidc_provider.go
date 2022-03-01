package server

import (
	"context"

	rpcv3 "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/system"
	systemv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/systempb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/service"
	"google.golang.org/protobuf/types/known/emptypb"
)

type oidcProvider struct {
	service.OIDCProviderService
}

func NewOIDCServer(providerSvc service.OIDCProviderService) rpcv3.OIDCProviderServer {
	return &oidcProvider{providerSvc}
}

func (s *oidcProvider) CreateOIDCProvider(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	return s.Create(ctx, p)
}
func (s *oidcProvider) GetOIDCProvider(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	return s.GetByName(ctx, p)
}
func (s *oidcProvider) ListOIDCProvider(ctx context.Context, p *emptypb.Empty) (*systemv3.OIDCProviderList, error) {
	return s.List(ctx)
}
func (s *oidcProvider) UpdateOIDCProvider(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	return s.Update(ctx, p)
}
func (s *oidcProvider) DeleteOIDCProvider(ctx context.Context, p *systemv3.OIDCProvider) (*emptypb.Empty, error) {
	// TODO: if successful return 204 NO CONTENT
	return &emptypb.Empty{}, s.Delete(ctx, p)
}
