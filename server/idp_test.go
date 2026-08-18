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

// fakeIdpService implements service.IdpService with function fields so each
// test wires only the method it needs.
type fakeIdpService struct {
	createFn    func(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error)
	getByIDFn   func(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error)
	getByNameFn func(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error)
	listFn      func(ctx context.Context) (*systemv3.IdpList, error)
	updateFn    func(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error)
	deleteFn    func(ctx context.Context, idp *systemv3.Idp) error
}

func (f *fakeIdpService) Create(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	return f.createFn(ctx, idp)
}
func (f *fakeIdpService) GetByID(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	return f.getByIDFn(ctx, idp)
}
func (f *fakeIdpService) GetByName(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	return f.getByNameFn(ctx, idp)
}
func (f *fakeIdpService) List(ctx context.Context) (*systemv3.IdpList, error) {
	return f.listFn(ctx)
}
func (f *fakeIdpService) Update(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	return f.updateFn(ctx, idp)
}
func (f *fakeIdpService) Delete(ctx context.Context, idp *systemv3.Idp) error {
	return f.deleteFn(ctx, idp)
}

func TestIdpServer_CreateIdp(t *testing.T) {
	t.Run("returns the service response on success", func(t *testing.T) {
		want := &systemv3.Idp{Metadata: &commonv3.Metadata{Name: "idp1"}}
		is := &fakeIdpService{createFn: func(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
			return want, nil
		}}
		s := NewIdpServer(is)

		got, err := s.CreateIdp(context.Background(), &systemv3.Idp{})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("propagates the service error unchanged", func(t *testing.T) {
		want := errors.New("boom")
		is := &fakeIdpService{createFn: func(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
			return nil, want
		}}
		s := NewIdpServer(is)

		_, err := s.CreateIdp(context.Background(), &systemv3.Idp{})
		assert.Same(t, want, err)
	})
}

func TestIdpServer_GetIdp(t *testing.T) {
	want := &systemv3.Idp{}
	is := &fakeIdpService{getByNameFn: func(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
		return want, nil
	}}
	s := NewIdpServer(is)

	got, err := s.GetIdp(context.Background(), &systemv3.Idp{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestIdpServer_ListIdps(t *testing.T) {
	want := &systemv3.IdpList{}
	is := &fakeIdpService{listFn: func(ctx context.Context) (*systemv3.IdpList, error) {
		return want, nil
	}}
	s := NewIdpServer(is)

	got, err := s.ListIdps(context.Background(), &commonv3.Empty{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestIdpServer_UpdateIdp(t *testing.T) {
	want := &systemv3.Idp{}
	is := &fakeIdpService{updateFn: func(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
		return want, nil
	}}
	s := NewIdpServer(is)

	got, err := s.UpdateIdp(context.Background(), &systemv3.Idp{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestIdpServer_DeleteIdp(t *testing.T) {
	t.Run("returns an empty response on success", func(t *testing.T) {
		is := &fakeIdpService{deleteFn: func(ctx context.Context, idp *systemv3.Idp) error {
			return nil
		}}
		s := NewIdpServer(is)

		resp, err := s.DeleteIdp(context.Background(), &systemv3.Idp{})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("propagates the service error but still returns an empty response", func(t *testing.T) {
		want := errors.New("boom")
		is := &fakeIdpService{deleteFn: func(ctx context.Context, idp *systemv3.Idp) error {
			return want
		}}
		s := NewIdpServer(is)

		resp, err := s.DeleteIdp(context.Background(), &systemv3.Idp{})
		assert.Same(t, want, err)
		assert.NotNil(t, resp)
	})
}
