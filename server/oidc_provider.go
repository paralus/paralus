package server

import (
	"context"

	"github.com/RafayLabs/rcloud-base/pkg/service"
	rpcv3 "github.com/RafayLabs/rcloud-base/proto/rpc/system"
	systemv3 "github.com/RafayLabs/rcloud-base/proto/types/systempb/v3"
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
