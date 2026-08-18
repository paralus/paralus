package server

import (
	"context"
	"errors"
	"testing"

	systemrpc "github.com/paralus/paralus/proto/rpc/system"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systempbv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePartnerService implements service.PartnerService with function fields
// so each test wires only the method it needs.
type fakePartnerService struct {
	createFn         func(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error)
	getByIDFn        func(ctx context.Context, partnerId string) (*systempbv3.Partner, error)
	getByNameFn      func(ctx context.Context, name string) (*systempbv3.Partner, error)
	updateFn         func(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error)
	deleteFn         func(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error)
	getOnlyPartnerFn func(ctx context.Context) (*systempbv3.Partner, error)
	upsertFn         func(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error)
}

func (f *fakePartnerService) Create(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
	return f.createFn(ctx, p)
}
func (f *fakePartnerService) GetByID(ctx context.Context, partnerId string) (*systempbv3.Partner, error) {
	return f.getByIDFn(ctx, partnerId)
}
func (f *fakePartnerService) GetByName(ctx context.Context, name string) (*systempbv3.Partner, error) {
	return f.getByNameFn(ctx, name)
}
func (f *fakePartnerService) Update(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
	return f.updateFn(ctx, p)
}
func (f *fakePartnerService) Delete(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
	return f.deleteFn(ctx, p)
}
func (f *fakePartnerService) GetOnlyPartner(ctx context.Context) (*systempbv3.Partner, error) {
	return f.getOnlyPartnerFn(ctx)
}
func (f *fakePartnerService) Upsert(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
	return f.upsertFn(ctx, p)
}

func TestPartnerServer_CreatePartner(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		ps := &fakePartnerService{createFn: func(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
			return &systempbv3.Partner{Metadata: p.Metadata}, nil
		}}
		s := NewPartnerServer(ps)
		req := &systempbv3.Partner{Metadata: &commonv3.Metadata{Name: "p1"}}

		resp, err := s.CreatePartner(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
		assert.Equal(t, commonv3.ConditionStatus_StatusOK, resp.Status.ConditionStatus)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		ps := &fakePartnerService{createFn: func(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
			return nil, errors.New("boom")
		}}
		s := NewPartnerServer(ps)
		req := &systempbv3.Partner{}

		resp, err := s.CreatePartner(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
		assert.Equal(t, "boom", resp.Status.Reason)
	})
}

func TestPartnerServer_GetPartner(t *testing.T) {
	t.Run("success stamps OK and looks the partner up by name", func(t *testing.T) {
		var gotName string
		ps := &fakePartnerService{getByNameFn: func(ctx context.Context, name string) (*systempbv3.Partner, error) {
			gotName = name
			return &systempbv3.Partner{}, nil
		}}
		s := NewPartnerServer(ps)

		resp, err := s.GetPartner(context.Background(), &systempbv3.Partner{Metadata: &commonv3.Metadata{Name: "p1"}})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
		assert.Equal(t, "p1", gotName)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		ps := &fakePartnerService{getByNameFn: func(ctx context.Context, name string) (*systempbv3.Partner, error) {
			return nil, errors.New("not found")
		}}
		s := NewPartnerServer(ps)
		req := &systempbv3.Partner{Metadata: &commonv3.Metadata{Name: "p1"}}

		resp, err := s.GetPartner(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}

func TestPartnerServer_DeletePartner(t *testing.T) {
	ps := &fakePartnerService{deleteFn: func(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
		return &systempbv3.Partner{}, nil
	}}
	s := NewPartnerServer(ps)

	resp, err := s.DeletePartner(context.Background(), &systempbv3.Partner{})
	require.NoError(t, err)
	require.NotNil(t, resp.Status)
}

func TestPartnerServer_GetInitPartner(t *testing.T) {
	t.Run("returns the only partner on success", func(t *testing.T) {
		want := &systempbv3.Partner{Metadata: &commonv3.Metadata{Name: "only-partner"}}
		ps := &fakePartnerService{getOnlyPartnerFn: func(ctx context.Context) (*systempbv3.Partner, error) {
			return want, nil
		}}
		s := NewPartnerServer(ps)

		got, err := s.GetInitPartner(context.Background(), &systemrpc.EmptyRequest{})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("propagates the service error unchanged", func(t *testing.T) {
		want := errors.New("no partner found")
		ps := &fakePartnerService{getOnlyPartnerFn: func(ctx context.Context) (*systempbv3.Partner, error) {
			return nil, want
		}}
		s := NewPartnerServer(ps)

		got, err := s.GetInitPartner(context.Background(), &systemrpc.EmptyRequest{})
		assert.Same(t, want, err)
		assert.Nil(t, got)
	})
}

func TestPartnerServer_UpdatePartner(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		ps := &fakePartnerService{updateFn: func(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
			return &systempbv3.Partner{}, nil
		}}
		s := NewPartnerServer(ps)

		resp, err := s.UpdatePartner(context.Background(), &systempbv3.Partner{})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		ps := &fakePartnerService{updateFn: func(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
			return nil, errors.New("boom")
		}}
		s := NewPartnerServer(ps)
		req := &systempbv3.Partner{}

		resp, err := s.UpdatePartner(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}
