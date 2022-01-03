package service

import (
	"context"
	"fmt"

	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
	userrpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
)

type server struct{}

func (s *server) CreateUser(context.Context, *userv3.User) (*userv3.User, error) {
	fmt.Println("Called create user")
	return nil, nil
}
func (s *server) GetUsers(context.Context, *userrpcv3.GetUsersRequest) (*userrpcv3.GetUsersResponse, error) {
	return nil, nil
}
func (s *server) GetUser(context.Context, *userrpcv3.GetUserRequest) (*userv3.User, error) { return nil, nil }
func (s *server) UpdateUser(context.Context, *userrpcv3.PutUserRequest) (*userrpcv3.UserResponse, error) {
	return nil, nil
}
func (s *server) DeleteUser(context.Context, *userrpcv3.DeleteUserRequest) (*userrpcv3.UserResponse, error) {
	return nil, nil
}

func NewUserServer() userrpcv3.UserServer {
	return &server{}
}
