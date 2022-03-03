package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/proto/rpc/system"
	systemv3 "github.com/RafaySystems/rcloud-base/proto/types/systempb/v3"
	"google.golang.org/protobuf/types/known/emptypb"
)

type idpServer struct {
	service.IdpService
}

func NewIdpServer(is service.IdpService) rpcv3.IdpServer {
	return &idpServer{is}
}

func (s *idpServer) CreateIdp(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	return s.IdpService.Create(ctx, idp)
}

func (s *idpServer) GetIdp(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	return s.IdpService.GetByName(ctx, idp)
}

func (s *idpServer) ListIdps(ctx context.Context, _ *emptypb.Empty) (*systemv3.IdpList, error) {
	return s.IdpService.List(ctx)
}

func (s *idpServer) UpdateIdp(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	return s.IdpService.Update(ctx, idp)
}

func (s *idpServer) DeleteIdp(ctx context.Context, idpID *systemv3.Idp) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.IdpService.Delete(ctx, idpID)
}
