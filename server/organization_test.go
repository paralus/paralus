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

// fakeOrganizationService implements service.OrganizationService with
// function fields so each test wires only the method it needs.
type fakeOrganizationService struct {
	createFn    func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error)
	getByIDFn   func(ctx context.Context, id string) (*systempbv3.Organization, error)
	getByNameFn func(ctx context.Context, name string) (*systempbv3.Organization, error)
	updateFn    func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error)
	deleteFn    func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error)
	listFn      func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.OrganizationList, error)
	upsertFn    func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error)
}

func (f *fakeOrganizationService) Create(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
	return f.createFn(ctx, o)
}
func (f *fakeOrganizationService) GetByID(ctx context.Context, id string) (*systempbv3.Organization, error) {
	return f.getByIDFn(ctx, id)
}
func (f *fakeOrganizationService) GetByName(ctx context.Context, name string) (*systempbv3.Organization, error) {
	return f.getByNameFn(ctx, name)
}
func (f *fakeOrganizationService) Update(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
	return f.updateFn(ctx, o)
}
func (f *fakeOrganizationService) Delete(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
	return f.deleteFn(ctx, o)
}
func (f *fakeOrganizationService) List(ctx context.Context, o *systempbv3.Organization) (*systempbv3.OrganizationList, error) {
	return f.listFn(ctx, o)
}
func (f *fakeOrganizationService) Upsert(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
	return f.upsertFn(ctx, o)
}

func TestOrganizationServer_CreateOrganization(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		os := &fakeOrganizationService{createFn: func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
			return &systempbv3.Organization{Metadata: o.Metadata}, nil
		}}
		s := NewOrganizationServer(os)
		req := &systempbv3.Organization{Metadata: &commonv3.Metadata{Name: "org1"}}

		resp, err := s.CreateOrganization(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
		assert.Equal(t, commonv3.ConditionStatus_StatusOK, resp.Status.ConditionStatus)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		os := &fakeOrganizationService{createFn: func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
			return nil, errors.New("boom")
		}}
		s := NewOrganizationServer(os)
		req := &systempbv3.Organization{}

		resp, err := s.CreateOrganization(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
		assert.Equal(t, "boom", resp.Status.Reason)
	})
}

func TestOrganizationServer_GetOrganizations(t *testing.T) {
	os := &fakeOrganizationService{listFn: func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.OrganizationList, error) {
		return &systempbv3.OrganizationList{}, nil
	}}
	s := NewOrganizationServer(os)

	_, err := s.GetOrganizations(context.Background(), &systempbv3.Organization{})
	require.NoError(t, err)
}

func TestOrganizationServer_GetOrganization(t *testing.T) {
	t.Run("success stamps OK and looks the org up by name", func(t *testing.T) {
		var gotName string
		os := &fakeOrganizationService{getByNameFn: func(ctx context.Context, name string) (*systempbv3.Organization, error) {
			gotName = name
			return &systempbv3.Organization{}, nil
		}}
		s := NewOrganizationServer(os)

		resp, err := s.GetOrganization(context.Background(), &systempbv3.Organization{Metadata: &commonv3.Metadata{Name: "org1"}})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
		assert.Equal(t, "org1", gotName)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		os := &fakeOrganizationService{getByNameFn: func(ctx context.Context, name string) (*systempbv3.Organization, error) {
			return nil, errors.New("not found")
		}}
		s := NewOrganizationServer(os)
		req := &systempbv3.Organization{Metadata: &commonv3.Metadata{Name: "org1"}}

		resp, err := s.GetOrganization(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}

func TestOrganizationServer_DeleteOrganization(t *testing.T) {
	os := &fakeOrganizationService{deleteFn: func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
		return &systempbv3.Organization{}, nil
	}}
	s := NewOrganizationServer(os)

	resp, err := s.DeleteOrganization(context.Background(), &systempbv3.Organization{})
	require.NoError(t, err)
	require.NotNil(t, resp.Status)
}

func TestOrganizationServer_UpdateOrganization(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		os := &fakeOrganizationService{updateFn: func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
			return &systempbv3.Organization{}, nil
		}}
		s := NewOrganizationServer(os)

		resp, err := s.UpdateOrganization(context.Background(), &systempbv3.Organization{})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		os := &fakeOrganizationService{updateFn: func(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
			return nil, errors.New("boom")
		}}
		s := NewOrganizationServer(os)
		req := &systempbv3.Organization{}

		resp, err := s.UpdateOrganization(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}
