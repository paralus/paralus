package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/authz/pkg/service"
	rpcv1 "github.com/RafaySystems/rcloud-base/components/authz/proto/rpc/v1"
	authzpbv1 "github.com/RafaySystems/rcloud-base/components/authz/proto/types"
)

type authzServer struct {
	as service.AuthzService
}

func NewAuthzServer(as service.AuthzService) rpcv1.AuthzServer {
	return &authzServer{as}
}

func (s *authzServer) Enforce(ctx context.Context, p *authzpbv1.EnforceRequest) (*authzpbv1.BoolReply, error) {
	return s.as.Enforce(ctx, p)
}

func (s *authzServer) ListPolicies(ctx context.Context, p *authzpbv1.Policy) (*authzpbv1.Policies, error) {
	return s.as.ListPolicies(ctx, p)
}

func (s *authzServer) CreatePolicies(ctx context.Context, p *authzpbv1.Policies) (*authzpbv1.BoolReply, error) {
	return s.as.CreatePolicies(ctx, p)
}

func (s *authzServer) DeletePolicies(ctx context.Context, p *authzpbv1.Policy) (*authzpbv1.BoolReply, error) {
	return s.as.DeletePolicies(ctx, p)
}

func (s *authzServer) ListUserGroups(ctx context.Context, p *authzpbv1.UserGroup) (*authzpbv1.UserGroups, error) {
	return s.as.ListUserGroups(ctx, p)
}

func (s *authzServer) CreateUserGroups(ctx context.Context, p *authzpbv1.UserGroups) (*authzpbv1.BoolReply, error) {
	return s.as.CreateUserGroups(ctx, p)
}

func (s *authzServer) DeleteUserGroups(ctx context.Context, p *authzpbv1.UserGroup) (*authzpbv1.BoolReply, error) {
	return s.as.DeleteUserGroups(ctx, p)
}

func (s *authzServer) ListRolePermissionMappings(ctx context.Context, p *authzpbv1.FilteredRolePermissionMapping) (*authzpbv1.RolePermissionMappingList, error) {
	return s.as.ListRolePermissionMappings(ctx, p)
}

func (s *authzServer) CreateRolePermissionMappings(ctx context.Context, p *authzpbv1.RolePermissionMappingList) (*authzpbv1.BoolReply, error) {
	return s.as.CreateRolePermissionMappings(ctx, p)
}

func (s *authzServer) DeleteRolePermissionMappings(ctx context.Context, p *authzpbv1.FilteredRolePermissionMapping) (*authzpbv1.BoolReply, error) {
	return s.as.DeleteRolePermissionMappings(ctx, p)
}
