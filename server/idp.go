package server

import (
	"context"

	"github.com/paralus/paralus/pkg/service"
	rpcv3 "github.com/paralus/paralus/proto/rpc/system"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
)

type idpServer struct {
	service.IdpService
}

func NewIdpServer(is service.IdpService) rpcv3.IdpServiceServer {
	return &idpServer{is}
}

func (s *idpServer) CreateIdp(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	return s.IdpService.Create(ctx, idp)
}

func (s *idpServer) GetIdp(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	return s.IdpService.GetByName(ctx, idp)
}

func (s *idpServer) ListIdps(ctx context.Context, _ *commonv3.Empty) (*systemv3.IdpList, error) {
	return s.IdpService.List(ctx)
}

func (s *idpServer) UpdateIdp(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	return s.IdpService.Update(ctx, idp)
}

func (s *idpServer) DeleteIdp(ctx context.Context, idpID *systemv3.Idp) (*commonv3.Empty, error) {
	return &commonv3.Empty{}, s.IdpService.Delete(ctx, idpID)
}
