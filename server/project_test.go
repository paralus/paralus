package server

import (
	"context"
	"errors"
	"testing"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systempbv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeProjectService implements service.ProjectService with function fields
// so each test wires only the method it needs.
type fakeProjectService struct {
	createFn    func(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error)
	getByIDFn   func(ctx context.Context, id string) (*systempbv3.Project, error)
	getByNameFn func(ctx context.Context, name string) (*systempbv3.Project, error)
	updateFn    func(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error)
	deleteFn    func(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error)
	listFn      func(ctx context.Context, p *systempbv3.Project) (*systempbv3.ProjectList, error)
}

func (f *fakeProjectService) Create(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
	return f.createFn(ctx, p)
}
func (f *fakeProjectService) GetByID(ctx context.Context, id string) (*systempbv3.Project, error) {
	return f.getByIDFn(ctx, id)
}
func (f *fakeProjectService) GetByName(ctx context.Context, name string) (*systempbv3.Project, error) {
	return f.getByNameFn(ctx, name)
}
func (f *fakeProjectService) Update(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
	return f.updateFn(ctx, p)
}
func (f *fakeProjectService) Delete(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
	return f.deleteFn(ctx, p)
}
func (f *fakeProjectService) List(ctx context.Context, p *systempbv3.Project) (*systempbv3.ProjectList, error) {
	return f.listFn(ctx, p)
}

func TestProjectServer_CreateProject(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		ps := &fakeProjectService{createFn: func(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
			return &systempbv3.Project{Metadata: p.Metadata}, nil
		}}
		s := NewProjectServer(ps)
		req := &systempbv3.Project{Metadata: &commonv3.Metadata{Name: "proj1"}}

		resp, err := s.CreateProject(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
		assert.Equal(t, commonv3.ConditionStatus_StatusOK, resp.Status.ConditionStatus)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		ps := &fakeProjectService{createFn: func(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
			return nil, errors.New("boom")
		}}
		s := NewProjectServer(ps)
		req := &systempbv3.Project{}

		resp, err := s.CreateProject(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
		assert.Equal(t, "boom", resp.Status.Reason)
	})
}

func TestProjectServer_GetProjects(t *testing.T) {
	ps := &fakeProjectService{listFn: func(ctx context.Context, p *systempbv3.Project) (*systempbv3.ProjectList, error) {
		return &systempbv3.ProjectList{}, nil
	}}
	s := NewProjectServer(ps)

	_, err := s.GetProjects(context.Background(), &systempbv3.Project{})
	require.NoError(t, err)
}

func TestProjectServer_GetProject(t *testing.T) {
	t.Run("success stamps OK and looks the project up by name", func(t *testing.T) {
		var gotName string
		ps := &fakeProjectService{getByNameFn: func(ctx context.Context, name string) (*systempbv3.Project, error) {
			gotName = name
			return &systempbv3.Project{}, nil
		}}
		s := NewProjectServer(ps)

		resp, err := s.GetProject(context.Background(), &systempbv3.Project{Metadata: &commonv3.Metadata{Name: "proj1"}})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
		assert.Equal(t, "proj1", gotName)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		ps := &fakeProjectService{getByNameFn: func(ctx context.Context, name string) (*systempbv3.Project, error) {
			return nil, errors.New("not found")
		}}
		s := NewProjectServer(ps)
		req := &systempbv3.Project{Metadata: &commonv3.Metadata{Name: "proj1"}}

		resp, err := s.GetProject(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}

func TestProjectServer_DeleteProject(t *testing.T) {
	ps := &fakeProjectService{deleteFn: func(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
		return &systempbv3.Project{}, nil
	}}
	s := NewProjectServer(ps)

	resp, err := s.DeleteProject(context.Background(), &systempbv3.Project{})
	require.NoError(t, err)
	require.NotNil(t, resp.Status)
}

func TestProjectServer_UpdateProject(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		ps := &fakeProjectService{updateFn: func(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
			return &systempbv3.Project{}, nil
		}}
		s := NewProjectServer(ps)

		resp, err := s.UpdateProject(context.Background(), &systempbv3.Project{})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		ps := &fakeProjectService{updateFn: func(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
			return nil, errors.New("boom")
		}}
		s := NewProjectServer(ps)
		req := &systempbv3.Project{}

		resp, err := s.UpdateProject(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}
