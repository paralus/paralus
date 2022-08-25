package server

import (
	"context"

	"github.com/paralus/paralus/pkg/service"
	rpcv3 "github.com/paralus/paralus/proto/rpc/role"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	rolepbv3 "github.com/paralus/paralus/proto/types/rolepb/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type roleServer struct {
	service.RoleService
}

// NewRoleServer returns new role server implementation
func NewRoleServer(ps service.RoleService) rpcv3.RoleServiceServer {
	return &roleServer{ps}
}

func updateRoleStatus(req *rolepbv3.Role, resp *rolepbv3.Role, err error) *rolepbv3.Role {
	if err != nil {
		req.Status = &v3.Status{
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return req
	}
	resp.Status = &v3.Status{ConditionStatus: v3.ConditionStatus_StatusOK}
	return resp
}

func (s *roleServer) CreateRole(ctx context.Context, req *rolepbv3.Role) (*rolepbv3.Role, error) {
	resp, err := s.Create(ctx, req)
	return updateRoleStatus(req, resp, err), err
}

func (s *roleServer) GetRoles(ctx context.Context, req *rolepbv3.Role) (*rolepbv3.RoleList, error) {
	return s.List(ctx, req)
}

func (s *roleServer) GetRole(ctx context.Context, req *rolepbv3.Role) (*rolepbv3.Role, error) {
	resp, err := s.GetByName(ctx, req)
	return updateRoleStatus(req, resp, err), err
}

func (s *roleServer) DeleteRole(ctx context.Context, req *rolepbv3.Role) (*rolepbv3.Role, error) {
	resp, err := s.Delete(ctx, req)
	return updateRoleStatus(req, resp, err), err
}

func (s *roleServer) UpdateRole(ctx context.Context, req *rolepbv3.Role) (*rolepbv3.Role, error) {
	resp, err := s.Update(ctx, req)
	return updateRoleStatus(req, resp, err), err
}
