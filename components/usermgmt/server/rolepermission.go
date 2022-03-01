package server

import (
	"context"

	rpcv3 "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/role"
	rolepbv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/rolepb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/service"
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
