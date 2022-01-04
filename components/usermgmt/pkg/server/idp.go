package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
)

type idpServer struct {
	service.IdpService
}

func NewIdpServer(is service.IdpService) rpcv3.IdpServer {
	return &idpServer{is}
}

func (s *idpServer) CreateIdp(ctx context.Context, idp *userv3.NewIdp) (*userv3.Idp, error) {
	return s.IdpService.CreateIdp(ctx, idp)
}

func (s *idpServer) UpdateIdp(ctx context.Context, idp *userv3.UpdateIdp) (*userv3.Idp, error) {
	return s.IdpService.UpdateIdp(ctx, idp)
}

func (s *idpServer) GetSpConfigById(ctx context.Context, idpID *userv3.IdpID) (*userv3.SpConfig, error) {
	return s.IdpService.GetSpConfigById(ctx, idpID)
}

func (s *idpServer) ListIdps(ctx context.Context, req *userv3.ListIdpsRequest) (*userv3.ListIdpsResponse, error) {
	return s.IdpService.ListIdps(ctx, req)
}
