package server

import (
	"context"
	"errors"
	"testing"

	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	userpbv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeGroupService implements service.GroupService with function fields so
// each test wires only the method it needs.
type fakeGroupService struct {
	createFn    func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error)
	getByIDFn   func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error)
	getByNameFn func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error)
	updateFn    func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error)
	deleteFn    func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error)
	listFn      func(ctx context.Context, opts ...query.Option) (*userpbv3.GroupList, error)
}

func (f *fakeGroupService) Create(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
	return f.createFn(ctx, g)
}
func (f *fakeGroupService) GetByID(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
	return f.getByIDFn(ctx, g)
}
func (f *fakeGroupService) GetByName(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
	return f.getByNameFn(ctx, g)
}
func (f *fakeGroupService) Update(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
	return f.updateFn(ctx, g)
}
func (f *fakeGroupService) Delete(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
	return f.deleteFn(ctx, g)
}
func (f *fakeGroupService) List(ctx context.Context, opts ...query.Option) (*userpbv3.GroupList, error) {
	return f.listFn(ctx, opts...)
}

func TestGroupServer_CreateGroup(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		gs := &fakeGroupService{createFn: func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
			return &userpbv3.Group{Metadata: g.Metadata}, nil
		}}
		s := NewGroupServer(gs)
		req := &userpbv3.Group{Metadata: &commonv3.Metadata{Name: "g1"}}

		resp, err := s.CreateGroup(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
		assert.Equal(t, commonv3.ConditionStatus_StatusOK, resp.Status.ConditionStatus)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		gs := &fakeGroupService{createFn: func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
			return nil, errors.New("boom")
		}}
		s := NewGroupServer(gs)
		req := &userpbv3.Group{}

		resp, err := s.CreateGroup(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
		assert.Equal(t, "boom", resp.Status.Reason)
	})
}

func TestGroupServer_GetGroups(t *testing.T) {
	gs := &fakeGroupService{listFn: func(ctx context.Context, opts ...query.Option) (*userpbv3.GroupList, error) {
		return &userpbv3.GroupList{}, nil
	}}
	s := NewGroupServer(gs)

	_, err := s.GetGroups(context.Background(), &commonv3.QueryOptions{})
	require.NoError(t, err)
}

func TestGroupServer_GetGroup(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		gs := &fakeGroupService{getByNameFn: func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
			return &userpbv3.Group{}, nil
		}}
		s := NewGroupServer(gs)

		resp, err := s.GetGroup(context.Background(), &userpbv3.Group{})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		gs := &fakeGroupService{getByNameFn: func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
			return nil, errors.New("not found")
		}}
		s := NewGroupServer(gs)
		req := &userpbv3.Group{}

		resp, err := s.GetGroup(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}

func TestGroupServer_DeleteGroup(t *testing.T) {
	gs := &fakeGroupService{deleteFn: func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
		return &userpbv3.Group{}, nil
	}}
	s := NewGroupServer(gs)

	resp, err := s.DeleteGroup(context.Background(), &userpbv3.Group{})
	require.NoError(t, err)
	require.NotNil(t, resp.Status)
}

func TestGroupServer_UpdateGroup(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		gs := &fakeGroupService{updateFn: func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
			return &userpbv3.Group{}, nil
		}}
		s := NewGroupServer(gs)

		resp, err := s.UpdateGroup(context.Background(), &userpbv3.Group{})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		gs := &fakeGroupService{updateFn: func(ctx context.Context, g *userpbv3.Group) (*userpbv3.Group, error) {
			return nil, errors.New("boom")
		}}
		s := NewGroupServer(gs)
		req := &userpbv3.Group{}

		resp, err := s.UpdateGroup(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}
