package authv3

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	kclient "github.com/ory/kratos-client-go"
	rpcv3 "github.com/paralus/paralus/proto/rpc/user"
	authzv1 "github.com/paralus/paralus/proto/types/authz"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
)

func TestGetTokenCheckSum(t *testing.T) {
	sum := md5.Sum([]byte("secret"))
	want := base64.StdEncoding.EncodeToString(sum[:])

	assert.Equal(t, want, getTokenCheckSum([]byte("secret")))
	assert.NotEqual(t, want, getTokenCheckSum([]byte("other")))
}

// newMockBunDB returns a *bun.DB backed by go-sqlmock, for exercising
// dao.GetGroups without a real Postgres instance.
func newMockBunDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqldb.Close() })
	return bun.NewDB(sqldb, pgdialect.New()), mock
}

// newKratosStub starts an httptest server that fakes Kratos's
// GET /sessions/whoami endpoint, and returns an *kclient.APIClient pointed
// at it, matching how SetupAuthContext wires the real client.
func newKratosStub(t *testing.T, handler http.HandlerFunc) *kclient.APIClient {
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := kclient.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	return kclient.NewAPIClient(cfg)
}

func kratosUnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":{"message":"no credentials"}}`))
}

func kratosServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func kratosActiveSessionHandler(identityID, email, org, partner string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		body := `{
			"id": "session-1",
			"active": true,
			"identity": {
				"id": "` + identityID + `",
				"schema_id": "default",
				"schema_url": "http://kratos.test/schema.json",
				"traits": {"email": "` + email + `"},
				"metadata_public": {"Organization": "` + org + `", "Partner": "` + partner + `"}
			}
		}`
		_, _ = w.Write([]byte(body))
	}
}

func kratosInactiveSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"id": "session-1", "active": false}`))
}

func TestAuthenticate_ApiKeyBranch(t *testing.T) {
	t.Run("valid key and matching token succeeds", func(t *testing.T) {
		secret := "shh"
		ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
			return &models.ApiKey{Name: "svc-account", Secret: secret}, nil
		}}
		ac := &authContext{ks: ks}

		req := &commonv3.IsRequestAllowedRequest{
			XApiKey:   "key-1",
			XApiToken: getTokenCheckSum([]byte(secret)),
		}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{}}

		ok, err := ac.authenticate(context.Background(), req, res)
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, commonv3.RequestStatus_RequestAllowed, res.Status)
		assert.Equal(t, "svc-account", res.SessionData.Username)
	})

	t.Run("key lookup failure returns ErrInvalidAPIKey", func(t *testing.T) {
		ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
			return nil, errors.New("not found")
		}}
		ac := &authContext{ks: ks}

		req := &commonv3.IsRequestAllowedRequest{XApiKey: "missing"}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{}}

		ok, err := ac.authenticate(context.Background(), req, res)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrInvalidAPIKey)
	})

	t.Run("mismatched token returns ErrInvalidSignature", func(t *testing.T) {
		ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
			return &models.ApiKey{Secret: "correct"}, nil
		}}
		ac := &authContext{ks: ks}

		req := &commonv3.IsRequestAllowedRequest{XApiKey: "key-1", XApiToken: "wrong"}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{}}

		ok, err := ac.authenticate(context.Background(), req, res)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrInvalidSignature)
	})
}

func TestAuthenticate_SessionTokenBranch(t *testing.T) {
	t.Run("401 from kratos means not authenticated, no error", func(t *testing.T) {
		ac := &authContext{kc: newKratosStub(t, kratosUnauthorizedHandler)}
		req := &commonv3.IsRequestAllowedRequest{XSessionToken: "tok"}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{}}

		ok, err := ac.authenticate(context.Background(), req, res)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, commonv3.RequestStatus_RequestNotAuthenticated, res.Status)
	})

	t.Run("non-401 transport error is propagated", func(t *testing.T) {
		ac := &authContext{kc: newKratosStub(t, kratosServerErrorHandler)}
		req := &commonv3.IsRequestAllowedRequest{XSessionToken: "tok"}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{}}

		ok, err := ac.authenticate(context.Background(), req, res)
		assert.False(t, ok)
		assert.Error(t, err)
	})

	t.Run("inactive session is not authenticated", func(t *testing.T) {
		ac := &authContext{kc: newKratosStub(t, kratosInactiveSessionHandler)}
		req := &commonv3.IsRequestAllowedRequest{XSessionToken: "tok"}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{}}

		ok, err := ac.authenticate(context.Background(), req, res)
		require.NoError(t, err)
		assert.True(t, ok) // authenticate returns true (no transport error); status reflects the real state
		assert.Equal(t, commonv3.RequestStatus_RequestNotAuthenticated, res.Status)
	})

	t.Run("active session populates session data and groups", func(t *testing.T) {
		identityID := uuid.New()
		db, mock := newMockBunDB(t)

		rows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "modified_at", "trash", "organization_id", "partner_id", "type"}).
			AddRow(uuid.New(), "team-a", "", time.Now(), time.Now(), false, uuid.New(), uuid.New(), "")
		mock.ExpectQuery(".*").WillReturnRows(rows)

		ac := &authContext{
			kc: newKratosStub(t, kratosActiveSessionHandler(identityID.String(), "user@paralus.dev", "org-1", "partner-1")),
			db: db,
		}
		req := &commonv3.IsRequestAllowedRequest{XSessionToken: "tok"}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{}}

		ok, err := ac.authenticate(context.Background(), req, res)
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, commonv3.RequestStatus_RequestAllowed, res.Status)
		assert.Equal(t, "user@paralus.dev", res.SessionData.Username)
		assert.Equal(t, "org-1", res.SessionData.Organization)
		assert.Equal(t, "partner-1", res.SessionData.Partner)
		assert.Equal(t, identityID.String(), res.SessionData.Account)
		require.Len(t, res.SessionData.Groups, 1)
		assert.Equal(t, "team-a", res.SessionData.Groups[0])
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthorize(t *testing.T) {
	t.Run("defaults empty project/org to wildcard and allows", func(t *testing.T) {
		var gotParams []string
		as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
			gotParams = req.Params
			return &authzv1.BoolReply{Res: true}, nil
		}}
		ac := &authContext{as: as}

		req := &commonv3.IsRequestAllowedRequest{Url: "/api", Method: "GET"}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{Username: "bob"}}

		err := ac.authorize(context.Background(), req, res)
		require.NoError(t, err)
		assert.Equal(t, commonv3.RequestStatus_RequestAllowed, res.Status)
		assert.Equal(t, []string{"u:bob", "*", "*", "*", "/api", "GET"}, gotParams)
	})

	t.Run("uses explicit project/org when set", func(t *testing.T) {
		var gotParams []string
		as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
			gotParams = req.Params
			return &authzv1.BoolReply{Res: true}, nil
		}}
		ac := &authContext{as: as}

		req := &commonv3.IsRequestAllowedRequest{Project: "proj-1", Org: "org-1", Url: "/api", Method: "GET"}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{Username: "bob"}}

		require.NoError(t, ac.authorize(context.Background(), req, res))
		assert.Equal(t, []string{"u:bob", "*", "proj-1", "org-1", "/api", "GET"}, gotParams)
	})

	t.Run("denies when enforcer disallows", func(t *testing.T) {
		as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
			return &authzv1.BoolReply{Res: false}, nil
		}}
		ac := &authContext{as: as}

		req := &commonv3.IsRequestAllowedRequest{}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{Username: "bob"}}

		require.NoError(t, ac.authorize(context.Background(), req, res))
		assert.Equal(t, commonv3.RequestStatus_RequestMethodOrURLNotAllowed, res.Status)
		assert.NotEmpty(t, res.Reason)
	})

	t.Run("propagates enforcer error", func(t *testing.T) {
		as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
			return nil, errors.New("boom")
		}}
		ac := &authContext{as: as}

		req := &commonv3.IsRequestAllowedRequest{}
		res := &commonv3.IsRequestAllowedResponse{SessionData: &commonv3.SessionData{Username: "bob"}}

		err := ac.authorize(context.Background(), req, res)
		assert.Error(t, err)
	})
}

func TestIsRequestAllowed(t *testing.T) {
	t.Run("authentication failure short-circuits before authorization", func(t *testing.T) {
		authzCalled := false
		ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
			return nil, errors.New("bad key")
		}}
		as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
			authzCalled = true
			return &authzv1.BoolReply{Res: true}, nil
		}}
		ac := &authContext{ks: ks, as: as}

		_, err := ac.IsRequestAllowed(context.Background(), &commonv3.IsRequestAllowedRequest{XApiKey: "bad"})
		assert.Error(t, err)
		assert.False(t, authzCalled, "authorize must not run when authenticate failed")
	})

	t.Run("NoAuthz skips authorization even on successful authentication", func(t *testing.T) {
		authzCalled := false
		ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
			return &models.ApiKey{Secret: "s"}, nil
		}}
		as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
			authzCalled = true
			return &authzv1.BoolReply{Res: true}, nil
		}}
		ac := &authContext{ks: ks, as: as}

		req := &commonv3.IsRequestAllowedRequest{
			XApiKey:   "k",
			XApiToken: getTokenCheckSum([]byte("s")),
			NoAuthz:   true,
		}
		res, err := ac.IsRequestAllowed(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, authzCalled)
		assert.Equal(t, commonv3.RequestStatus_RequestAllowed, res.Status)
	})

	t.Run("full pass-through allows when both authn and authz succeed", func(t *testing.T) {
		ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
			return &models.ApiKey{Secret: "s"}, nil
		}}
		as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
			return &authzv1.BoolReply{Res: true}, nil
		}}
		ac := &authContext{ks: ks, as: as}

		req := &commonv3.IsRequestAllowedRequest{
			XApiKey:   "k",
			XApiToken: getTokenCheckSum([]byte("s")),
		}
		res, err := ac.IsRequestAllowed(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, commonv3.RequestStatus_RequestAllowed, res.Status)
	})

	t.Run("authorize error is propagated", func(t *testing.T) {
		ks := &fakeApiKeyService{getByKeyFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
			return &models.ApiKey{Secret: "s"}, nil
		}}
		as := &fakeAuthzService{enforceFn: func(ctx context.Context, req *authzv1.EnforceRequest) (*authzv1.BoolReply, error) {
			return nil, errors.New("enforcer down")
		}}
		ac := &authContext{ks: ks, as: as}

		req := &commonv3.IsRequestAllowedRequest{
			XApiKey:   "k",
			XApiToken: getTokenCheckSum([]byte("s")),
		}
		_, err := ac.IsRequestAllowed(context.Background(), req)
		assert.Error(t, err)
	})
}
