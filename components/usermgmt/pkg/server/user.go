package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
	userpbv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
)

type userServer struct {
	service.UserService
}

// NewUserServer returns new user server implementation
func NewUserServer(ps service.UserService) rpcv3.UserServer {
	return &userServer{ps}
}

func (s *userServer) CreateUser(ctx context.Context, p *userpbv3.User) (*userpbv3.User, error) {
	return s.Create(ctx, p)
}

func (s *userServer) GetUsers(ctx context.Context, p *userpbv3.User) (*userpbv3.UserList, error) {
	return s.List(ctx, p)
}

func (s *userServer) GetUser(ctx context.Context, p *userpbv3.User) (*userpbv3.User, error) {
	// user, err := s.GetByName(ctx, p.Metadata.Name)
	// if err != nil {
	return s.GetByID(ctx, p.Metadata.Id)
	// }
	// return user, nil
}

func (s *userServer) DeleteUser(ctx context.Context, p *userpbv3.User) (*rpcv3.DeleteUserResponse, error) {
	_, err := s.Delete(ctx, p)
	return nil, err
}

func (s *userServer) UpdateUser(ctx context.Context, p *userpbv3.User) (*userpbv3.User, error) {
	return s.Update(ctx, p)
}
