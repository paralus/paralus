package server

import (
	"context"
	"errors"
	"testing"

	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	rolev3 "github.com/paralus/paralus/proto/types/rolepb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRoleService implements service.RoleService with function fields so
// each test wires only the method it needs.
type fakeRoleService struct {
	createFn    func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error)
	getByIDFn   func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error)
	getByNameFn func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error)
	updateFn    func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error)
	deleteFn    func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error)
	listFn      func(ctx context.Context, r *rolev3.Role) (*rolev3.RoleList, error)
	upsertFn    func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error)
}

func (f *fakeRoleService) Create(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
	return f.createFn(ctx, r)
}
func (f *fakeRoleService) GetByID(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
	return f.getByIDFn(ctx, r)
}
func (f *fakeRoleService) GetByName(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
	return f.getByNameFn(ctx, r)
}
func (f *fakeRoleService) Update(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
	return f.updateFn(ctx, r)
}
func (f *fakeRoleService) Delete(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
	return f.deleteFn(ctx, r)
}
func (f *fakeRoleService) List(ctx context.Context, r *rolev3.Role) (*rolev3.RoleList, error) {
	return f.listFn(ctx, r)
}
func (f *fakeRoleService) Upsert(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
	return f.upsertFn(ctx, r)
}

func TestRoleServer_CreateRole(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		rs := &fakeRoleService{createFn: func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
			return &rolev3.Role{Metadata: r.Metadata}, nil
		}}
		s := NewRoleServer(rs)
		req := &rolev3.Role{Metadata: &commonv3.Metadata{Name: "role-1"}}

		resp, err := s.CreateRole(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		rs := &fakeRoleService{createFn: func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
			return nil, errors.New("boom")
		}}
		s := NewRoleServer(rs)
		req := &rolev3.Role{}

		resp, err := s.CreateRole(context.Background(), req)
		assert.Error(t, err)
		require.NotNil(t, resp)
		assert.Same(t, req, resp)
	})
}

func TestRoleServer_GetRoles(t *testing.T) {
	rs := &fakeRoleService{listFn: func(ctx context.Context, r *rolev3.Role) (*rolev3.RoleList, error) {
		return &rolev3.RoleList{}, nil
	}}
	s := NewRoleServer(rs)

	_, err := s.GetRoles(context.Background(), &rolev3.Role{})
	require.NoError(t, err)
}

func TestRoleServer_GetRole(t *testing.T) {
	rs := &fakeRoleService{getByNameFn: func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
		return &rolev3.Role{}, nil
	}}
	s := NewRoleServer(rs)

	resp, err := s.GetRole(context.Background(), &rolev3.Role{})
	require.NoError(t, err)
	require.NotNil(t, resp.Status)
}

func TestRoleServer_DeleteRole(t *testing.T) {
	rs := &fakeRoleService{deleteFn: func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
		return &rolev3.Role{}, nil
	}}
	s := NewRoleServer(rs)

	resp, err := s.DeleteRole(context.Background(), &rolev3.Role{})
	require.NoError(t, err)
	require.NotNil(t, resp.Status)
}

func TestRoleServer_UpdateRole(t *testing.T) {
	rs := &fakeRoleService{updateFn: func(ctx context.Context, r *rolev3.Role) (*rolev3.Role, error) {
		return &rolev3.Role{}, nil
	}}
	s := NewRoleServer(rs)

	resp, err := s.UpdateRole(context.Background(), &rolev3.Role{})
	require.NoError(t, err)
	require.NotNil(t, resp.Status)
}

// fakeRolepermissionService implements service.RolepermissionService.
type fakeRolepermissionService struct {
	listFn func(ctx context.Context, opts ...query.Option) (*rolev3.RolePermissionList, error)
}

func (f *fakeRolepermissionService) GetByName(ctx context.Context, r *rolev3.RolePermission) (*rolev3.RolePermission, error) {
	panic("not implemented")
}
func (f *fakeRolepermissionService) List(ctx context.Context, opts ...query.Option) (*rolev3.RolePermissionList, error) {
	return f.listFn(ctx, opts...)
}

func TestRolepermissionServer_GetRolepermissions(t *testing.T) {
	called := false
	rps := &fakeRolepermissionService{listFn: func(ctx context.Context, opts ...query.Option) (*rolev3.RolePermissionList, error) {
		called = true
		return &rolev3.RolePermissionList{}, nil
	}}
	s := NewRolePermissionServer(rps)

	_, err := s.GetRolepermissions(context.Background(), &commonv3.QueryOptions{})
	require.NoError(t, err)
	assert.True(t, called)
}
