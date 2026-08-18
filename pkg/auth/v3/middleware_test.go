package authv3

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/common"
	rpcv3 "github.com/paralus/paralus/proto/rpc/v3"
	commonpbv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func allowIsRequestAllowed(ctx context.Context, req *commonpbv3.IsRequestAllowedRequest) (*commonpbv3.IsRequestAllowedResponse, error) {
	return &commonpbv3.IsRequestAllowedResponse{
		Status:      commonpbv3.RequestStatus_RequestAllowed,
		SessionData: &commonpbv3.SessionData{Username: "bob"},
	}, nil
}

func TestServeHTTP_ExcludedURLPassesThrough(t *testing.T) {
	var nextCalled bool
	next := func(rw http.ResponseWriter, r *http.Request) { nextCalled = true }

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	serveHTTP(Option{ExcludeURLs: []string{"^/healthz$"}}, nil, allowIsRequestAllowed, rw, r, next)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestServeHTTP_NonPromptPathIsForbidden(t *testing.T) {
	var nextCalled bool
	next := func(rw http.ResponseWriter, r *http.Request) { nextCalled = true }

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v2/something/else", nil)

	serveHTTP(Option{}, nil, allowIsRequestAllowed, rw, r, next)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusForbidden, rw.Code)
}

func TestServeHTTP_PromptPathResolvesProjectAndAllows(t *testing.T) {
	db, mock := newMockBunDB(t)

	orgID := uuid.New()
	rows := sqlmock.NewRows([]string{"project", "organization", "project_id", "organization_id", "partner_id"}).
		AddRow("my-proj", "my-org", "proj-id", orgID.String(), "partner-id")
	mock.ExpectQuery(".*").WillReturnRows(rows)

	var nextCalled bool
	var gotCtx context.Context
	next := func(rw http.ResponseWriter, r *http.Request) {
		nextCalled = true
		gotCtx = r.Context()
	}

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v2/debug/prompt/project/my-proj/cluster/my-cluster", nil)

	serveHTTP(Option{}, db, allowIsRequestAllowed, rw, r, next)

	require.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rw.Code)

	sd, ok := gotCtx.Value(common.SessionDataKey).(*commonpbv3.SessionData)
	require.True(t, ok)
	// serveHTTP assigns the organization *ID* (not the name) into
	// SessionData.Organization -- existing behavior, asserted as-is.
	assert.Equal(t, orgID.String(), sd.Organization)
	require.NotNil(t, sd.Project)
	require.Len(t, sd.Project.List, 1)
	assert.Equal(t, "my-proj", sd.Project.List[0].Project)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServeHTTP_PromptPathProjectLookupFailureIsForbidden(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("no rows"))

	var nextCalled bool
	next := func(rw http.ResponseWriter, r *http.Request) { nextCalled = true }

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v2/debug/prompt/project/missing/cluster/my-cluster", nil)

	serveHTTP(Option{}, db, allowIsRequestAllowed, rw, r, next)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusForbidden, rw.Code)
}

func TestServeHTTP_IsRequestAllowedErrorIsInternalServerError(t *testing.T) {
	db, mock := newMockBunDB(t)
	rows := sqlmock.NewRows([]string{"project", "organization", "project_id", "organization_id", "partner_id"}).
		AddRow("my-proj", "my-org", "proj-id", uuid.New().String(), "partner-id")
	mock.ExpectQuery(".*").WillReturnRows(rows)

	failing := func(ctx context.Context, req *commonpbv3.IsRequestAllowedRequest) (*commonpbv3.IsRequestAllowedResponse, error) {
		return nil, errors.New("boom")
	}

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v2/debug/prompt/project/my-proj/cluster/my-cluster", nil)

	serveHTTP(Option{}, db, failing, rw, r, func(http.ResponseWriter, *http.Request) {})

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestServeHTTP_StatusMapping(t *testing.T) {
	tests := []struct {
		name       string
		status     commonpbv3.RequestStatus
		reason     string
		wantStatus int
	}{
		{
			name:       "not allowed maps to forbidden",
			status:     commonpbv3.RequestStatus_RequestMethodOrURLNotAllowed,
			reason:     "nope",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "not authenticated maps to unauthorized",
			status:     commonpbv3.RequestStatus_RequestNotAuthenticated,
			reason:     "no session",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unknown status maps to internal server error",
			status:     commonpbv3.RequestStatus_Unknown,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockBunDB(t)
			rows := sqlmock.NewRows([]string{"project", "organization", "project_id", "organization_id", "partner_id"}).
				AddRow("my-proj", "my-org", "proj-id", uuid.New().String(), "partner-id")
			mock.ExpectQuery(".*").WillReturnRows(rows)

			isRequestAllowed := func(ctx context.Context, req *commonpbv3.IsRequestAllowedRequest) (*commonpbv3.IsRequestAllowedResponse, error) {
				return &commonpbv3.IsRequestAllowedResponse{Status: tt.status, Reason: tt.reason}, nil
			}

			rw := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/v2/debug/prompt/project/my-proj/cluster/my-cluster", nil)

			serveHTTP(Option{}, db, isRequestAllowed, rw, r, func(http.ResponseWriter, *http.Request) {})
			assert.Equal(t, tt.wantStatus, rw.Code)
		})
	}
}

func TestAuthMiddleware_ServeHTTP_DelegatesToServeHTTP(t *testing.T) {
	am := &authMiddleware{
		ac:  authContext{},
		opt: Option{ExcludeURLs: []string{"^/healthz$"}},
	}

	var nextCalled bool
	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	am.ServeHTTP(rw, r, func(http.ResponseWriter, *http.Request) { nextCalled = true })

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rw.Code)
}

// fakeAuthServiceClient implements rpcv3.AuthServiceClient for
// remoteAuthMiddleware tests, avoiding a real grpc.ClientConn.
type fakeAuthServiceClient struct {
	isRequestAllowedFn func(ctx context.Context, in *commonpbv3.IsRequestAllowedRequest, opts ...grpc.CallOption) (*commonpbv3.IsRequestAllowedResponse, error)
}

func (f *fakeAuthServiceClient) IsRequestAllowed(ctx context.Context, in *commonpbv3.IsRequestAllowedRequest, opts ...grpc.CallOption) (*commonpbv3.IsRequestAllowedResponse, error) {
	return f.isRequestAllowedFn(ctx, in, opts...)
}

var _ rpcv3.AuthServiceClient = (*fakeAuthServiceClient)(nil)

func TestRemoteAuthMiddleware_ServeHTTP_DelegatesToRemoteClient(t *testing.T) {
	as := &fakeAuthServiceClient{
		isRequestAllowedFn: func(ctx context.Context, in *commonpbv3.IsRequestAllowedRequest, opts ...grpc.CallOption) (*commonpbv3.IsRequestAllowedResponse, error) {
			return &commonpbv3.IsRequestAllowedResponse{
				Status:      commonpbv3.RequestStatus_RequestAllowed,
				SessionData: &commonpbv3.SessionData{Username: "remote-bob"},
			}, nil
		},
	}
	am := &remoteAuthMiddleware{
		as:  as,
		opt: Option{ExcludeURLs: []string{"^/healthz$"}},
	}

	var nextCalled bool
	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	am.ServeHTTP(rw, r, func(http.ResponseWriter, *http.Request) { nextCalled = true })

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rw.Code)
}
