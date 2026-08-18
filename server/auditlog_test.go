package server

import (
	"context"
	"errors"
	"testing"

	v1 "github.com/paralus/paralus/proto/rpc/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAuditLogService implements service.AuditLogService with function
// fields so each test wires only the method it needs.
type fakeAuditLogService struct {
	getAuditLogFn           func(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error)
	getAuditLogByProjectsFn func(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error)
}

func (f *fakeAuditLogService) GetAuditLog(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error) {
	return f.getAuditLogFn(ctx, req)
}
func (f *fakeAuditLogService) GetAuditLogByProjects(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error) {
	return f.getAuditLogByProjectsFn(ctx, req)
}

func TestAuditLogServer_GetAuditLog(t *testing.T) {
	t.Run("returns the service response on success", func(t *testing.T) {
		want := &v1.GetAuditLogSearchResponse{}
		as := &fakeAuditLogService{getAuditLogFn: func(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error) {
			return want, nil
		}}
		s, err := NewAuditLogServer(as)
		require.NoError(t, err)

		got, err := s.GetAuditLog(context.Background(), &v1.GetAuditLogSearchRequest{})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("propagates the service error unchanged", func(t *testing.T) {
		want := errors.New("boom")
		as := &fakeAuditLogService{getAuditLogFn: func(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error) {
			return nil, want
		}}
		s, err := NewAuditLogServer(as)
		require.NoError(t, err)

		_, err = s.GetAuditLog(context.Background(), &v1.GetAuditLogSearchRequest{})
		assert.Same(t, want, err)
	})
}

func TestAuditLogServer_GetAuditLogByProjects(t *testing.T) {
	t.Run("returns the service response on success", func(t *testing.T) {
		want := &v1.GetAuditLogSearchResponse{}
		as := &fakeAuditLogService{getAuditLogByProjectsFn: func(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error) {
			return want, nil
		}}
		s, err := NewAuditLogServer(as)
		require.NoError(t, err)

		got, err := s.GetAuditLogByProjects(context.Background(), &v1.GetAuditLogSearchRequest{})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("propagates the service error unchanged", func(t *testing.T) {
		want := errors.New("boom")
		as := &fakeAuditLogService{getAuditLogByProjectsFn: func(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error) {
			return nil, want
		}}
		s, err := NewAuditLogServer(as)
		require.NoError(t, err)

		_, err = s.GetAuditLogByProjects(context.Background(), &v1.GetAuditLogSearchRequest{})
		assert.Same(t, want, err)
	})
}
