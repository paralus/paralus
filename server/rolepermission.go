package server

import (
	"context"

	"github.com/RafayLabs/rcloud-base/pkg/query"
	"github.com/RafayLabs/rcloud-base/pkg/service"
	rpcv3 "github.com/RafayLabs/rcloud-base/proto/rpc/role"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	rolepbv3 "github.com/RafayLabs/rcloud-base/proto/types/rolepb/v3"
)

type rolepermissionServer struct {
	service.RolepermissionService
}

// NewRolePermissionServer returns new role server implementation
func NewRolePermissionServer(ps service.RolepermissionService) rpcv3.RolepermissionServer {
	return &rolepermissionServer{ps}
}

func (s *rolepermissionServer) GetRolepermissions(ctx context.Context, req *commonv3.QueryOptions) (*rolepbv3.RolePermissionList, error) {
	return s.List(ctx, query.WithOptions(req))
}
