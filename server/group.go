package server

import (
	"context"

	"github.com/paralus/paralus/pkg/query"
	"github.com/paralus/paralus/pkg/service"
	rpcv3 "github.com/paralus/paralus/proto/rpc/user"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	userpbv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type groupServer struct {
	service.GroupService
}

// NewGroupServer returns new group server implementation
func NewGroupServer(ps service.GroupService) rpcv3.GroupServiceServer {
	return &groupServer{ps}
}

func updateGroupStatus(req *userpbv3.Group, resp *userpbv3.Group, err error) *userpbv3.Group {
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

func (s *groupServer) CreateGroup(ctx context.Context, req *userpbv3.Group) (*userpbv3.Group, error) {
	resp, err := s.Create(ctx, req)
	return updateGroupStatus(req, resp, err), err
}

func (s *groupServer) GetGroups(ctx context.Context, req *commonv3.QueryOptions) (*userpbv3.GroupList, error) {
	return s.List(ctx, query.WithOptions(req))
}

func (s *groupServer) GetGroup(ctx context.Context, req *userpbv3.Group) (*userpbv3.Group, error) {
	resp, err := s.GetByName(ctx, req)
	return updateGroupStatus(req, resp, err), err
}

func (s *groupServer) DeleteGroup(ctx context.Context, req *userpbv3.Group) (*userpbv3.Group, error) {
	resp, err := s.Delete(ctx, req)
	return updateGroupStatus(req, resp, err), err
}

func (s *groupServer) UpdateGroup(ctx context.Context, req *userpbv3.Group) (*userpbv3.Group, error) {
	resp, err := s.Update(ctx, req)
	return updateGroupStatus(req, resp, err), err
}
