package server

import (
	"context"
	"errors"
	"testing"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeOIDCProviderService implements service.OIDCProviderService with
// function fields so each test wires only the method it needs.
type fakeOIDCProviderService struct {
	createFn    func(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error)
	getByIDFn   func(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error)
	getByNameFn func(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error)
	listFn      func(ctx context.Context) (*systemv3.OIDCProviderList, error)
	updateFn    func(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error)
	deleteFn    func(ctx context.Context, p *systemv3.OIDCProvider) error
}

func (f *fakeOIDCProviderService) Create(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	return f.createFn(ctx, p)
}
func (f *fakeOIDCProviderService) GetByID(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	return f.getByIDFn(ctx, p)
}
func (f *fakeOIDCProviderService) GetByName(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	return f.getByNameFn(ctx, p)
}
func (f *fakeOIDCProviderService) List(ctx context.Context) (*systemv3.OIDCProviderList, error) {
	return f.listFn(ctx)
}
func (f *fakeOIDCProviderService) Update(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	return f.updateFn(ctx, p)
}
func (f *fakeOIDCProviderService) Delete(ctx context.Context, p *systemv3.OIDCProvider) error {
	return f.deleteFn(ctx, p)
}

func TestOIDCProvider_CreateOIDCProvider(t *testing.T) {
	t.Run("returns the service response on success", func(t *testing.T) {
		want := &systemv3.OIDCProvider{Metadata: &commonv3.Metadata{Name: "provider1"}}
		os := &fakeOIDCProviderService{createFn: func(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
			return want, nil
		}}
		s := NewOIDCServer(os)

		got, err := s.CreateOIDCProvider(context.Background(), &systemv3.OIDCProvider{})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("propagates the service error unchanged", func(t *testing.T) {
		want := errors.New("boom")
		os := &fakeOIDCProviderService{createFn: func(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
			return nil, want
		}}
		s := NewOIDCServer(os)

		_, err := s.CreateOIDCProvider(context.Background(), &systemv3.OIDCProvider{})
		assert.Same(t, want, err)
	})
}

func TestOIDCProvider_GetOIDCProvider(t *testing.T) {
	want := &systemv3.OIDCProvider{}
	os := &fakeOIDCProviderService{getByNameFn: func(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
		return want, nil
	}}
	s := NewOIDCServer(os)

	got, err := s.GetOIDCProvider(context.Background(), &systemv3.OIDCProvider{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestOIDCProvider_ListOIDCProvider(t *testing.T) {
	want := &systemv3.OIDCProviderList{}
	os := &fakeOIDCProviderService{listFn: func(ctx context.Context) (*systemv3.OIDCProviderList, error) {
		return want, nil
	}}
	s := NewOIDCServer(os)

	got, err := s.ListOIDCProvider(context.Background(), &commonv3.Empty{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestOIDCProvider_UpdateOIDCProvider(t *testing.T) {
	want := &systemv3.OIDCProvider{}
	os := &fakeOIDCProviderService{updateFn: func(ctx context.Context, p *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
		return want, nil
	}}
	s := NewOIDCServer(os)

	got, err := s.UpdateOIDCProvider(context.Background(), &systemv3.OIDCProvider{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestOIDCProvider_DeleteOIDCProvider(t *testing.T) {
	t.Run("returns an empty response on success", func(t *testing.T) {
		os := &fakeOIDCProviderService{deleteFn: func(ctx context.Context, p *systemv3.OIDCProvider) error {
			return nil
		}}
		s := NewOIDCServer(os)

		resp, err := s.DeleteOIDCProvider(context.Background(), &systemv3.OIDCProvider{})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("propagates the service error but still returns an empty response", func(t *testing.T) {
		want := errors.New("boom")
		os := &fakeOIDCProviderService{deleteFn: func(ctx context.Context, p *systemv3.OIDCProvider) error {
			return want
		}}
		s := NewOIDCServer(os)

		resp, err := s.DeleteOIDCProvider(context.Background(), &systemv3.OIDCProvider{})
		assert.Same(t, want, err)
		assert.NotNil(t, resp)
	})
}
