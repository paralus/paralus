package server

import (
	"context"

	"github.com/paralus/paralus/pkg/service"
	rpcv3 "github.com/paralus/paralus/proto/rpc/system"
	"github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
)

type oidcProvider struct {
	service.OIDCProviderService
}

func NewOIDCServer(providerSvc service.OIDCProviderService) rpcv3.OIDCProviderServiceServer {
	return &oidcProvider{providerSvc}
}

func (s *oidcProvider) CreateOIDCProvider(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	return s.Create(ctx, p)
}
func (s *oidcProvider) GetOIDCProvider(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	return s.GetByName(ctx, p)
}
func (s *oidcProvider) ListOIDCProvider(ctx context.Context, p *commonv3.Empty) (*systemv3.OIDCProviderList, error) {
	return s.List(ctx)
}
func (s *oidcProvider) UpdateOIDCProvider(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	return s.Update(ctx, p)
}
func (s *oidcProvider) DeleteOIDCProvider(ctx context.Context, p *systemv3.OIDCProvider) (*commonv3.Empty, error) {
	// TODO: if successful return 204 NO CONTENT
	return &commonv3.Empty{}, s.Delete(ctx, p)
}
