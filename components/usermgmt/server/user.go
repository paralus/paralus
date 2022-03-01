package server

import (
	"context"

	rpcv3 "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/user"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	userpbv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/userpb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userServer struct {
	service.UserService
}

// NewUserServer returns new user server implementation
func NewUserServer(ps service.UserService) rpcv3.UserServer {
	return &userServer{ps}
}
func updateUserStatus(req *userpbv3.User, resp *userpbv3.User, err error) *userpbv3.User {
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

func (s *userServer) CreateUser(ctx context.Context, req *userpbv3.User) (*userpbv3.User, error) {
	resp, err := s.Create(ctx, req)
	return updateUserStatus(req, resp, err), err
}

func (s *userServer) GetUsers(ctx context.Context, req *userpbv3.User) (*userpbv3.UserList, error) {
	return s.List(ctx, req)
}

func (s *userServer) GetUser(ctx context.Context, req *userpbv3.User) (*userpbv3.User, error) {
	resp, err := s.GetByName(ctx, req)
	return updateUserStatus(req, resp, err), err
}

func (s *userServer) DeleteUser(ctx context.Context, req *userpbv3.User) (*rpcv3.DeleteUserResponse, error) {
	return s.Delete(ctx, req)
}

func (s *userServer) UpdateUser(ctx context.Context, req *userpbv3.User) (*userpbv3.User, error) {
	resp, err := s.Update(ctx, req)
	return updateUserStatus(req, resp, err), err
}
