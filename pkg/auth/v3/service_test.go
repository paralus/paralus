package authv3

import (
	"context"
	"testing"

	"github.com/paralus/paralus/internal/models"
	rpcv3 "github.com/paralus/paralus/proto/rpc/user"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_IsRequestAllowed_DelegatesToContext(t *testing.T) {
	ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
		return &models.ApiKey{Secret: "s", Name: "svc"}, nil
	}}
	ac := NewAuthContext(nil, nil, ks, nil)
	svc := NewAuthService(ac)

	req := &commonv3.IsRequestAllowedRequest{
		XApiKey:   "k",
		XApiToken: getTokenCheckSum([]byte("s")),
		NoAuthz:   true, // avoid needing an AuthzService fake
	}

	res, err := svc.IsRequestAllowed(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, commonv3.RequestStatus_RequestAllowed, res.Status)
	assert.Equal(t, "svc", res.SessionData.Username)
}
