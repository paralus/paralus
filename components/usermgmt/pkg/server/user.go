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
	user, err := s.Create(ctx, p)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (s *userServer) GetUsers(ctx context.Context, p *userpbv3.User) (*userpbv3.UserList, error) {
	users, err := s.List(ctx, p)
	if err != nil {
		return users, err
	}
	return users, nil
}

func (s *userServer) GetUser(ctx context.Context, p *userpbv3.User) (*userpbv3.User, error) {
	// user, err := s.GetByName(ctx, p.Metadata.Name)
	// if err != nil {
	user, err := s.GetByID(ctx, p.Metadata.Id)
	if err != nil {
		return user, err
	}
	// }

	return user, nil
}

func (s *userServer) DeleteUser(ctx context.Context, p *userpbv3.User) (*rpcv3.DeleteUserResponse, error) {
	_, err := s.Delete(ctx, p)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *userServer) UpdateUser(ctx context.Context, p *userpbv3.User) (*userpbv3.User, error) {
	user, err := s.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return user, nil
}
