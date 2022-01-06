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
	role, err := s.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleServer) GetRoles(ctx context.Context, p *userpbv3.Role) (*userpbv3.RoleList, error) {
	roles, err := s.List(ctx, p)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *roleServer) GetRole(ctx context.Context, p *userpbv3.Role) (*userpbv3.Role, error) {
	role, err := s.GetByName(ctx, p.Metadata.Name)
	if err != nil {
		role, err = s.GetByID(ctx, p.Metadata.Id)
		if err != nil {
			return nil, err
		}
	}

	return role, nil
}

func (s *roleServer) DeleteRole(ctx context.Context, p *userpbv3.Role) (*userpbv3.Role, error) {
	_, err := s.Delete(ctx, p)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *roleServer) UpdateRole(ctx context.Context, p *userpbv3.Role) (*userpbv3.Role, error) {
	role, err := s.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return role, nil
}
