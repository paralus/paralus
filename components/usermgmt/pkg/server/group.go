package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
	userpbv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
)

type groupServer struct {
	service.GroupService
}

// NewGroupServer returns new group server implementation
func NewGroupServer(ps service.GroupService) rpcv3.GroupServer {
	return &groupServer{ps}
}

func (s *groupServer) CreateGroup(ctx context.Context, p *userpbv3.Group) (*userpbv3.Group, error) {
	return s.Create(ctx, p)
}

func (s *groupServer) GetGroups(ctx context.Context, p *userpbv3.Group) (*userpbv3.GroupList, error) {
	return s.List(ctx, p)
}

func (s *groupServer) GetGroup(ctx context.Context, p *userpbv3.Group) (*userpbv3.Group, error) {
	return s.GetByName(ctx, p)
}

func (s *groupServer) DeleteGroup(ctx context.Context, p *userpbv3.Group) (*userpbv3.Group, error) {
	return s.Delete(ctx, p)
}

func (s *groupServer) UpdateGroup(ctx context.Context, p *userpbv3.Group) (*userpbv3.Group, error) {
	return s.Update(ctx, p)
}
