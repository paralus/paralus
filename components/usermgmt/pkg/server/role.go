package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
	userpbv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
)

type roleServer struct {
	service.RoleService
}

// NewRoleServer returns new role server implementation
func NewRoleServer(ps service.RoleService) rpcv3.RoleServer {
	return &roleServer{ps}
}

func (s *roleServer) CreateRole(ctx context.Context, p *userpbv3.Role) (*userpbv3.Role, error) {
	return s.Create(ctx, p)
}

func (s *roleServer) GetRoles(ctx context.Context, p *userpbv3.Role) (*userpbv3.RoleList, error) {
	return s.List(ctx, p)
}

func (s *roleServer) GetRole(ctx context.Context, p *userpbv3.Role) (*userpbv3.Role, error) {
	return s.GetByName(ctx, p)
}

func (s *roleServer) DeleteRole(ctx context.Context, p *userpbv3.Role) (*userpbv3.Role, error) {
	return s.Delete(ctx, p)
}

func (s *roleServer) UpdateRole(ctx context.Context, p *userpbv3.Role) (*userpbv3.Role, error) {
	return s.Update(ctx, p)
}
