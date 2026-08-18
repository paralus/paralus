package authv3

import (
	"context"
	"testing"

	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/common"
	rpcv3 "github.com/paralus/paralus/proto/rpc/user"
	authzv1 "github.com/paralus/paralus/proto/types/authz"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// fakeMetaRequest implements hasMetadata so it can stand in for a real
// generated protobuf request.
type fakeMetaRequest struct {
	meta *commonv3.Metadata
}

func (f *fakeMetaRequest) GetMetadata() *commonv3.Metadata { return f.meta }

func echoHandler(called *bool, gotCtx *context.Context) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		*called = true
		*gotCtx = ctx
		return "handled", nil
	}
}

func TestNewAuthUnaryInterceptor_ExcludedMethodBypassesAuth(t *testing.T) {
	ac := authContext{}
	interceptor := ac.NewAuthUnaryInterceptor(Option{ExcludeRPCMethods: []string{"/svc/Excluded"}})

	var called bool
	var gotCtx context.Context
	resp, err := interceptor(context.Background(), &fakeMetaRequest{}, &grpc.UnaryServerInfo{FullMethod: "/svc/Excluded"}, echoHandler(&called, &gotCtx))

	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "handled", resp)
}

func TestNewAuthUnaryInterceptor_MissingGRPCMetadata(t *testing.T) {
	ac := authContext{}
	interceptor := ac.NewAuthUnaryInterceptor(Option{})

	var called bool
	var gotCtx context.Context
	_, err := interceptor(context.Background(), &fakeMetaRequest{}, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, echoHandler(&called, &gotCtx))

	require.Error(t, err)
	assert.False(t, called)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func apiKeyMD(apiKeyID, secret, url, method string) context.Context {
	md := metadata.New(map[string]string{
		"X-API-KEYID":      apiKeyID,
		"X-API-TOKEN":      getTokenCheckSum([]byte(secret)),
		"x-gateway-url":    url,
		"x-gateway-method": method,
	})
	return metadata.NewIncomingContext(context.Background(), md)
}

func TestNewAuthUnaryInterceptor_AllowedPathPlumbsSessionData(t *testing.T) {
	ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
		return &models.ApiKey{Secret: "s", Name: "svc-a"}, nil
	}}
	as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
		return &authzv1.BoolReply{Res: true}, nil
	}}
	ac := authContext{ks: ks, as: as}
	interceptor := ac.NewAuthUnaryInterceptor(Option{})

	ctx := apiKeyMD("key-1", "s", "/v3/clusters", "GET")
	md, _ := metadata.FromIncomingContext(ctx)
	md.Set("x-gateway-remote-addr", "1.2.3.4")
	md.Set("x-gateway-host", "api.paralus.dev")
	md.Set("x-gateway-user-agent", "test-agent")
	ctx = metadata.NewIncomingContext(context.Background(), md)

	req := &fakeMetaRequest{meta: &commonv3.Metadata{Organization: "org-1", Project: "proj-1"}}

	var called bool
	var gotCtx context.Context
	_, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, echoHandler(&called, &gotCtx))

	require.NoError(t, err)
	require.True(t, called)

	sd, ok := gotCtx.Value(common.SessionDataKey).(*commonv3.SessionData)
	require.True(t, ok)
	assert.Equal(t, "1.2.3.4", sd.ClientIp)
	assert.Equal(t, "api.paralus.dev", sd.ClientHost)
	assert.Equal(t, "test-agent", sd.ClientUa)
	assert.Equal(t, "svc-a", sd.Username)
}

func TestNewAuthUnaryInterceptor_MetadataOverrideForProjectAndOrgServices(t *testing.T) {
	ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
		return &models.ApiKey{Secret: "s"}, nil
	}}

	var gotParams []string
	as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
		gotParams = req.Params
		return &authzv1.BoolReply{Res: true}, nil
	}}
	ac := authContext{ks: ks, as: as}
	interceptor := ac.NewAuthUnaryInterceptor(Option{})

	ctx := apiKeyMD("key-1", "s", "/v3/projects/my-proj", "GET")
	req := &fakeMetaRequest{meta: &commonv3.Metadata{Name: "my-proj"}}

	var called bool
	var gotCtx context.Context
	_, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/paralus.dev.rpc.v3.Project/Get"}, echoHandler(&called, &gotCtx))

	require.NoError(t, err)
	require.True(t, called)
	// params: [u:sub, ns, proj, org, url, method] -- proj should be meta.Name
	require.Len(t, gotParams, 6)
	assert.Equal(t, "my-proj", gotParams[2])
}

func TestNewAuthUnaryInterceptor_PermissionDenied(t *testing.T) {
	ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
		return &models.ApiKey{Secret: "s"}, nil
	}}
	as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
		return &authzv1.BoolReply{Res: false}, nil
	}}
	ac := authContext{ks: ks, as: as}
	interceptor := ac.NewAuthUnaryInterceptor(Option{})

	ctx := apiKeyMD("key-1", "s", "/v3/clusters", "GET")
	var called bool
	var gotCtx context.Context
	_, err := interceptor(ctx, &fakeMetaRequest{}, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, echoHandler(&called, &gotCtx))

	require.Error(t, err)
	assert.False(t, called)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestNewAuthUnaryInterceptor_InternalErrorOnAuthenticateFailure(t *testing.T) {
	ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
		return &models.ApiKey{Secret: "s"}, nil
	}}
	ac := authContext{ks: ks}
	interceptor := ac.NewAuthUnaryInterceptor(Option{})

	// wrong token -> ErrInvalidSignature -> IsRequestAllowed returns an
	// error, which the interceptor maps to codes.Internal.
	ctx := apiKeyMD("key-1", "different-secret", "/v3/clusters", "GET")
	var called bool
	var gotCtx context.Context
	_, err := interceptor(ctx, &fakeMetaRequest{}, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, echoHandler(&called, &gotCtx))

	require.Error(t, err)
	assert.False(t, called)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestNewAuthUnaryInterceptor_UnauthenticatedStatusMapping(t *testing.T) {
	ac := authContext{kc: newKratosStub(t, kratosUnauthorizedHandler)}
	interceptor := ac.NewAuthUnaryInterceptor(Option{})

	md := metadata.New(map[string]string{"x-gateway-url": "/v3/clusters", "x-gateway-method": "GET"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var called bool
	var gotCtx context.Context
	_, err := interceptor(ctx, &fakeMetaRequest{}, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, echoHandler(&called, &gotCtx))

	require.Error(t, err)
	assert.False(t, called)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestNewAuthUnaryInterceptor_NoAuthzExcludedMethod(t *testing.T) {
	ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
		return &models.ApiKey{Secret: "s"}, nil
	}}
	authzCalled := false
	as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
		authzCalled = true
		return &authzv1.BoolReply{Res: true}, nil
	}}
	ac := authContext{ks: ks, as: as}
	interceptor := ac.NewAuthUnaryInterceptor(Option{ExcludeAuthzMethods: []string{"/svc/Method"}})

	ctx := apiKeyMD("key-1", "s", "/v3/clusters", "GET")
	var called bool
	var gotCtx context.Context
	_, err := interceptor(ctx, &fakeMetaRequest{}, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, echoHandler(&called, &gotCtx))

	require.NoError(t, err)
	assert.True(t, called)
	assert.False(t, authzCalled, "ExcludeAuthzMethods should skip authorization")
}
