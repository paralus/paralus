package server

import (
	"context"

	"github.com/paralus/paralus/pkg/query"
	"github.com/paralus/paralus/pkg/service"
	rpcv3 "github.com/paralus/paralus/proto/rpc/role"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	rolepbv3 "github.com/paralus/paralus/proto/types/rolepb/v3"
)

type rolepermissionServer struct {
	service.RolepermissionService
}

// NewRolePermissionServer returns new role server implementation
func NewRolePermissionServer(ps service.RolepermissionService) rpcv3.RolepermissionServiceServer {
	return &rolepermissionServer{ps}
}

func (s *rolepermissionServer) GetRolepermissions(ctx context.Context, req *commonv3.QueryOptions) (*rolepbv3.RolePermissionList, error) {
	return s.List(ctx, query.WithOptions(req))
}
