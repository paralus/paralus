package server

import (
	"context"
	"encoding/json"
	"fmt"

	authv3 "github.com/RafaySystems/rcloud-base/pkg/auth/v3"
	"github.com/RafaySystems/rcloud-base/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/proto/rpc/user"
	commonv3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	v3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	userpbv3 "github.com/RafaySystems/rcloud-base/proto/types/userpb/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userServer struct {
	us service.UserService
	ks service.ApiKeyService
}

// NewUserServer returns new user server implementation
func NewUserServer(ps service.UserService, as service.ApiKeyService) rpcv3.UserServer {
	return &userServer{us: ps, ks: as}
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
	resp, err := s.us.Create(ctx, req)
	return updateUserStatus(req, resp, err), err
}

func (s *userServer) GetUsers(ctx context.Context, req *userpbv3.User) (*userpbv3.UserList, error) {
	return s.us.List(ctx, req)
}

func (s *userServer) GetUser(ctx context.Context, req *userpbv3.User) (*userpbv3.User, error) {
	resp, err := s.us.GetByName(ctx, req)
	return updateUserStatus(req, resp, err), err
}

func (s *userServer) DeleteUser(ctx context.Context, req *userpbv3.User) (*rpcv3.DeleteUserResponse, error) {
	return s.us.Delete(ctx, req)
}

func (s *userServer) UpdateUser(ctx context.Context, req *userpbv3.User) (*userpbv3.User, error) {
	resp, err := s.us.Update(ctx, req)
	return updateUserStatus(req, resp, err), err
}

func (s *userServer) DownloadCliConfig(ctx context.Context, req *rpcv3.CliConfigRequest) (*commonv3.HttpBody, error) {
	sessData, ok := authv3.GetSession(ctx)
	if !ok {
		return nil, fmt.Errorf("unable to retrieve session data")
	}
	request := &rpcv3.ApiKeyRequest{
		Username: sessData.Username,
		Id:       sessData.Account,
	}
	cliConfig, err := s.us.RetrieveCliConfig(ctx, request)
	if err != nil {
		return nil, err
	}

	bb, err := json.Marshal(cliConfig)
	if err != nil {
		return nil, err
	}

	return &commonv3.HttpBody{
		ContentType: "application/json",
		Data:        bb,
	}, nil
}

func (s *userServer) UserListApiKeys(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.ApiKeyResponseList, error) {
	return s.ks.List(ctx, req)
}

func (s *userServer) UserDeleteApiKeys(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.DeleteUserResponse, error) {
	_, err := s.ks.Delete(ctx, req)
	if err != nil {
		return nil, err
	}
	return &rpcv3.DeleteUserResponse{}, nil
}
