package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
	userpbv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
)

type rolepermissionServer struct {
	service.RolepermissionService
}

// NewRolePermissionServer returns new role server implementation
func NewRolePermissionServer(ps service.RolepermissionService) rpcv3.RolepermissionServer {
	return &rolepermissionServer{ps}
}

func (s *rolepermissionServer) GetRolepermissions(ctx context.Context, p *userpbv3.RolePermission) (*userpbv3.RolePermissionList, error) {
	return s.List(ctx, p)
}
