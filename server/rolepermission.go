package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/proto/rpc/role"
	rolepbv3 "github.com/RafaySystems/rcloud-base/proto/types/rolepb/v3"
)

type rolepermissionServer struct {
	service.RolepermissionService
}

// NewRolePermissionServer returns new role server implementation
func NewRolePermissionServer(ps service.RolepermissionService) rpcv3.RolepermissionServer {
	return &rolepermissionServer{ps}
}

func (s *rolepermissionServer) GetRolepermissions(ctx context.Context, p *rolepbv3.RolePermission) (*rolepbv3.RolePermissionList, error) {
	return s.List(ctx, p)
}
