package authv3

import (
	"context"

	"github.com/paralus/paralus/internal/models"
	rpcv3 "github.com/paralus/paralus/proto/rpc/user"
	authzv1 "github.com/paralus/paralus/proto/types/authz"
)

// fakeApiKeyService implements service.ApiKeyService with only GetByKey
// wired; other methods panic if called since none of the code under test
// exercises them.
type fakeApiKeyService struct {
	getByKeyFn func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error)
}

func (f *fakeApiKeyService) Create(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
	panic("not implemented")
}
func (f *fakeApiKeyService) Get(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
	panic("not implemented")
}
func (f *fakeApiKeyService) GetByKey(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
	return f.getByKeyFn(ctx, req)
}
func (f *fakeApiKeyService) Delete(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserDeleteApiKeysResponse, error) {
	panic("not implemented")
}
func (f *fakeApiKeyService) List(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserListApiKeysResponse, error) {
	panic("not implemented")
}

// fakeAuthzService implements service.AuthzService with only Enforce wired.
type fakeAuthzService struct {
	enforceFn func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error)
}

func (f *fakeAuthzService) Enforce(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
	return f.enforceFn(ctx, req)
}
func (f *fakeAuthzService) ListPolicies(ctx context.Context, p *authzv1.Policy) (*authzv1.Policies, error) {
	panic("not implemented")
}
func (f *fakeAuthzService) CreatePolicies(ctx context.Context, p *authzv1.Policies) (*authzv1.BoolReply, error) {
	panic("not implemented")
}
func (f *fakeAuthzService) DeletePolicies(ctx context.Context, p *authzv1.Policy) (*authzv1.BoolReply, error) {
	panic("not implemented")
}
func (f *fakeAuthzService) ListUserGroups(ctx context.Context, p *authzv1.UserGroup) (*authzv1.UserGroups, error) {
	panic("not implemented")
}
func (f *fakeAuthzService) CreateUserGroups(ctx context.Context, p *authzv1.UserGroups) (*authzv1.BoolReply, error) {
	panic("not implemented")
}
func (f *fakeAuthzService) DeleteUserGroups(ctx context.Context, p *authzv1.UserGroup) (*authzv1.BoolReply, error) {
	panic("not implemented")
}
func (f *fakeAuthzService) ListRolePermissionMappings(ctx context.Context, p *authzv1.FilteredRolePermissionMapping) (*authzv1.RolePermissionMappingList, error) {
	panic("not implemented")
}
func (f *fakeAuthzService) CreateRolePermissionMappings(ctx context.Context, p *authzv1.RolePermissionMappingList) (*authzv1.BoolReply, error) {
	panic("not implemented")
}
func (f *fakeAuthzService) DeleteRolePermissionMappings(ctx context.Context, p *authzv1.FilteredRolePermissionMapping) (*authzv1.BoolReply, error) {
	panic("not implemented")
}
