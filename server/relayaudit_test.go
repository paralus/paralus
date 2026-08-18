package server

import (
	"context"
	"errors"
	"testing"

	ec "github.com/paralus/paralus/pkg/common"
	v1 "github.com/paralus/paralus/proto/rpc/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// fakeRelayAuditService implements service.RelayAuditService with function
// fields so each test wires only the method it needs.
type fakeRelayAuditService struct {
	getRelayAuditFn           func(ctx context.Context, req *v1.RelayAuditRequest) (*v1.RelayAuditResponse, error)
	getRelayAuditByProjectsFn func(ctx context.Context, req *v1.RelayAuditRequest) (*v1.RelayAuditResponse, error)
}

func (f *fakeRelayAuditService) GetRelayAudit(ctx context.Context, req *v1.RelayAuditRequest) (*v1.RelayAuditResponse, error) {
	return f.getRelayAuditFn(ctx, req)
}
func (f *fakeRelayAuditService) GetRelayAuditByProjects(ctx context.Context, req *v1.RelayAuditRequest) (*v1.RelayAuditResponse, error) {
	return f.getRelayAuditByProjectsFn(ctx, req)
}

func TestRelayAuditServer_GetRelayAudit(t *testing.T) {
	t.Run("RelayAPI audit type delegates to the relay audit service and stamps the audit type", func(t *testing.T) {
		rs := &fakeRelayAuditService{getRelayAuditFn: func(ctx context.Context, req *v1.RelayAuditRequest) (*v1.RelayAuditResponse, error) {
			return &v1.RelayAuditResponse{}, nil
		}}
		s, err := NewRelayAuditServer(rs, &fakeAuditLogService{})
		require.NoError(t, err)

		resp, err := s.GetRelayAudit(context.Background(), &v1.RelayAuditRequest{AuditType: ec.RelayAPIAuditType})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, ec.RelayAPIAuditType, resp.AuditType)
	})

	t.Run("propagates a relay audit service error", func(t *testing.T) {
		want := errors.New("boom")
		rs := &fakeRelayAuditService{getRelayAuditFn: func(ctx context.Context, req *v1.RelayAuditRequest) (*v1.RelayAuditResponse, error) {
			return nil, want
		}}
		s, err := NewRelayAuditServer(rs, &fakeAuditLogService{})
		require.NoError(t, err)

		_, err = s.GetRelayAudit(context.Background(), &v1.RelayAuditRequest{AuditType: ec.RelayAPIAuditType})
		assert.Same(t, want, err)
	})

	t.Run("non-RelayAPI audit type is converted and delegated to the audit log service", func(t *testing.T) {
		result, _ := structpb.NewStruct(map[string]interface{}{"hits": "1"})
		var gotReq *v1.GetAuditLogSearchRequest
		al := &fakeAuditLogService{getAuditLogFn: func(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error) {
			gotReq = req
			return &v1.GetAuditLogSearchResponse{Result: result}, nil
		}}
		s, err := NewRelayAuditServer(&fakeRelayAuditService{}, al)
		require.NoError(t, err)

		resp, err := s.GetRelayAudit(context.Background(), &v1.RelayAuditRequest{
			AuditType: ec.RelayCommandsAuditType,
			Filter:    &v1.RelayAuditQueryFilter{Cluster: "c1"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, ec.RelayCommandsAuditType, resp.AuditType)
		assert.Same(t, result, resp.Result)
		require.NotNil(t, gotReq)
		require.NotNil(t, gotReq.Filter)
		assert.Equal(t, "c1", gotReq.Filter.Cluster)
	})

	t.Run("propagates an audit log service error for a non-RelayAPI audit type", func(t *testing.T) {
		want := errors.New("boom")
		al := &fakeAuditLogService{getAuditLogFn: func(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error) {
			return nil, want
		}}
		s, err := NewRelayAuditServer(&fakeRelayAuditService{}, al)
		require.NoError(t, err)

		_, err = s.GetRelayAudit(context.Background(), &v1.RelayAuditRequest{AuditType: ec.RelayCommandsAuditType})
		assert.Same(t, want, err)
	})
}

func TestRelayAuditServer_GetRelayAuditByProjects(t *testing.T) {
	t.Run("RelayAPI audit type delegates to the relay audit service and stamps the audit type", func(t *testing.T) {
		rs := &fakeRelayAuditService{getRelayAuditByProjectsFn: func(ctx context.Context, req *v1.RelayAuditRequest) (*v1.RelayAuditResponse, error) {
			return &v1.RelayAuditResponse{}, nil
		}}
		s, err := NewRelayAuditServer(rs, &fakeAuditLogService{})
		require.NoError(t, err)

		resp, err := s.GetRelayAuditByProjects(context.Background(), &v1.RelayAuditRequest{AuditType: ec.RelayAPIAuditType})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, ec.RelayAPIAuditType, resp.AuditType)
	})

	t.Run("propagates a relay audit service error", func(t *testing.T) {
		want := errors.New("boom")
		rs := &fakeRelayAuditService{getRelayAuditByProjectsFn: func(ctx context.Context, req *v1.RelayAuditRequest) (*v1.RelayAuditResponse, error) {
			return nil, want
		}}
		s, err := NewRelayAuditServer(rs, &fakeAuditLogService{})
		require.NoError(t, err)

		_, err = s.GetRelayAuditByProjects(context.Background(), &v1.RelayAuditRequest{AuditType: ec.RelayAPIAuditType})
		assert.Same(t, want, err)
	})

	t.Run("non-RelayAPI audit type is converted and delegated to the audit log service", func(t *testing.T) {
		result, _ := structpb.NewStruct(map[string]interface{}{"hits": "2"})
		al := &fakeAuditLogService{getAuditLogByProjectsFn: func(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error) {
			return &v1.GetAuditLogSearchResponse{Result: result}, nil
		}}
		s, err := NewRelayAuditServer(&fakeRelayAuditService{}, al)
		require.NoError(t, err)

		resp, err := s.GetRelayAuditByProjects(context.Background(), &v1.RelayAuditRequest{AuditType: ec.RelayCommandsAuditType})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, ec.RelayCommandsAuditType, resp.AuditType)
		assert.Same(t, result, resp.Result)
	})

	t.Run("propagates an audit log service error for a non-RelayAPI audit type", func(t *testing.T) {
		want := errors.New("boom")
		al := &fakeAuditLogService{getAuditLogByProjectsFn: func(ctx context.Context, req *v1.GetAuditLogSearchRequest) (*v1.GetAuditLogSearchResponse, error) {
			return nil, want
		}}
		s, err := NewRelayAuditServer(&fakeRelayAuditService{}, al)
		require.NoError(t, err)

		_, err = s.GetRelayAuditByProjects(context.Background(), &v1.RelayAuditRequest{AuditType: ec.RelayCommandsAuditType})
		assert.Same(t, want, err)
	})
}
