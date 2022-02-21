package service

import (
	"context"
	"strings"

	types "github.com/RafaySystems/rcloud-base/components/authz/proto/types"
	"google.golang.org/grpc"
)

type ApUpdate struct {
	id     string
	traits map[string]interface{}
}
type mockAuthProvider struct {
	c []map[string]interface{}
	u []ApUpdate
	r []string
	d []string
}

func (m *mockAuthProvider) Create(ctx context.Context, traits map[string]interface{}) (string, error) {
	m.c = append(m.c, traits)
	return strings.Split(traits["email"].(string), "user-")[1], nil
}
func (m *mockAuthProvider) Update(ctx context.Context, id string, traits map[string]interface{}) error {
	m.u = append(m.u, ApUpdate{id: id, traits: traits})
	return nil
}
func (m *mockAuthProvider) GetRecoveryLink(ctx context.Context, id string) (string, error) {
	m.r = append(m.r, id)
	return "https://recoverme.testing/" + id, nil
}
func (m *mockAuthProvider) Delete(ctx context.Context, id string) error {
	m.d = append(m.d, id)
	return nil
}

type mockAuthzClient struct {
	cp   []*types.Policies
	dp   []*types.Policy
	cug  []*types.UserGroups
	dug  []*types.UserGroup
	crpm []*types.RolePermissionMappingList
	drpm []*types.FilteredRolePermissionMapping
}

func (c *mockAuthzClient) Enforce(ctx context.Context, in *types.EnforceRequest, opts ...grpc.CallOption) (*types.BoolReply, error) {
	return &types.BoolReply{Res: true}, nil
}
func (c *mockAuthzClient) ListPolicies(ctx context.Context, in *types.Policy, opts ...grpc.CallOption) (*types.Policies, error) {
	return &types.Policies{}, nil
}
func (c *mockAuthzClient) CreatePolicies(ctx context.Context, in *types.Policies, opts ...grpc.CallOption) (*types.BoolReply, error) {
	c.cp = append(c.cp, in)
	return &types.BoolReply{Res: true}, nil
}
func (c *mockAuthzClient) DeletePolicies(ctx context.Context, in *types.Policy, opts ...grpc.CallOption) (*types.BoolReply, error) {
	c.dp = append(c.dp, in)
	return &types.BoolReply{Res: true}, nil
}
func (c *mockAuthzClient) ListUserGroups(ctx context.Context, in *types.UserGroup, opts ...grpc.CallOption) (*types.UserGroups, error) {
	return &types.UserGroups{}, nil
}
func (c *mockAuthzClient) CreateUserGroups(ctx context.Context, in *types.UserGroups, opts ...grpc.CallOption) (*types.BoolReply, error) {
	c.cug = append(c.cug, in)
	return &types.BoolReply{Res: true}, nil
}
func (c *mockAuthzClient) DeleteUserGroups(ctx context.Context, in *types.UserGroup, opts ...grpc.CallOption) (*types.BoolReply, error) {
	c.dug = append(c.dug, in)
	return &types.BoolReply{Res: true}, nil
}
func (c *mockAuthzClient) ListRolePermissionMappings(ctx context.Context, in *types.FilteredRolePermissionMapping, opts ...grpc.CallOption) (*types.RolePermissionMappingList, error) {
	return &types.RolePermissionMappingList{}, nil
}
func (c *mockAuthzClient) CreateRolePermissionMappings(ctx context.Context, in *types.RolePermissionMappingList, opts ...grpc.CallOption) (*types.BoolReply, error) {
	c.crpm = append(c.crpm, in)
	return &types.BoolReply{Res: true}, nil
}
func (c *mockAuthzClient) DeleteRolePermissionMappings(ctx context.Context, in *types.FilteredRolePermissionMapping, opts ...grpc.CallOption) (*types.BoolReply, error) {
	c.drpm = append(c.drpm, in)
	return &types.BoolReply{Res: true}, nil
}
