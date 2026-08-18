package server

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/common"
	"github.com/paralus/paralus/pkg/query"
	rpcv3 "github.com/paralus/paralus/proto/rpc/user"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	userpbv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUserService implements service.UserService with function fields so
// each test wires only the method it needs; everything else panics to
// surface accidental extra calls.
type fakeUserService struct {
	createFn               func(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error)
	getByNameFn            func(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error)
	getUserInfoFn          func(ctx context.Context, u *userpbv3.User) (*userpbv3.UserInfo, error)
	updateFn               func(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error)
	updateForceResetFlagFn func(ctx context.Context, username string) error
	deleteFn               func(ctx context.Context, u *userpbv3.User) (*rpcv3.UserDeleteApiKeysResponse, error)
	listFn                 func(ctx context.Context, opts ...query.Option) (*userpbv3.UserList, error)
	retrieveCliConfigFn    func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*common.CliConfigDownloadData, error)
	forgotPasswordFn       func(ctx context.Context, req *rpcv3.UserForgotPasswordRequest) (*rpcv3.UserForgotPasswordResponse, error)
	createLoginAuditLogFn  func(ctx context.Context, req *rpcv3.UserLoginAuditRequest) (*rpcv3.UserLoginAuditResponse, error)
}

func (f *fakeUserService) Create(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error) {
	return f.createFn(ctx, u)
}
func (f *fakeUserService) GetByID(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error) {
	panic("not implemented")
}
func (f *fakeUserService) GetByName(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error) {
	return f.getByNameFn(ctx, u)
}
func (f *fakeUserService) GetUserInfo(ctx context.Context, u *userpbv3.User) (*userpbv3.UserInfo, error) {
	return f.getUserInfoFn(ctx, u)
}
func (f *fakeUserService) Update(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error) {
	return f.updateFn(ctx, u)
}
func (f *fakeUserService) UpdateForceResetFlag(ctx context.Context, username string) error {
	return f.updateForceResetFlagFn(ctx, username)
}
func (f *fakeUserService) Delete(ctx context.Context, u *userpbv3.User) (*rpcv3.UserDeleteApiKeysResponse, error) {
	return f.deleteFn(ctx, u)
}
func (f *fakeUserService) List(ctx context.Context, opts ...query.Option) (*userpbv3.UserList, error) {
	return f.listFn(ctx, opts...)
}
func (f *fakeUserService) RetrieveCliConfig(ctx context.Context, req *rpcv3.ApiKeyRequest) (*common.CliConfigDownloadData, error) {
	return f.retrieveCliConfigFn(ctx, req)
}
func (f *fakeUserService) UpdateIdpUserGroupPolicy(ctx context.Context, a, b, c string) error {
	panic("not implemented")
}
func (f *fakeUserService) ForgotPassword(ctx context.Context, req *rpcv3.UserForgotPasswordRequest) (*rpcv3.UserForgotPasswordResponse, error) {
	return f.forgotPasswordFn(ctx, req)
}
func (f *fakeUserService) CreateLoginAuditLog(ctx context.Context, req *rpcv3.UserLoginAuditRequest) (*rpcv3.UserLoginAuditResponse, error) {
	return f.createLoginAuditLogFn(ctx, req)
}

// fakeApiKeyService implements service.ApiKeyService with function fields.
type fakeApiKeyService struct {
	deleteFn func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserDeleteApiKeysResponse, error)
	listFn   func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserListApiKeysResponse, error)
}

func (f *fakeApiKeyService) Create(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
	panic("not implemented")
}
func (f *fakeApiKeyService) Get(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
	panic("not implemented")
}
func (f *fakeApiKeyService) GetByKey(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
	panic("not implemented")
}
func (f *fakeApiKeyService) Delete(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserDeleteApiKeysResponse, error) {
	return f.deleteFn(ctx, req)
}
func (f *fakeApiKeyService) List(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserListApiKeysResponse, error) {
	return f.listFn(ctx, req)
}

func TestUserServer_CreateUser(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		us := &fakeUserService{createFn: func(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error) {
			return &userpbv3.User{Metadata: u.Metadata}, nil
		}}
		s := NewUserServer(us, &fakeApiKeyService{})
		req := &userpbv3.User{Metadata: &commonv3.Metadata{Name: "u1"}}

		resp, err := s.CreateUser(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
		assert.Equal(t, commonv3.ConditionStatus_StatusOK, resp.Status.ConditionStatus)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		us := &fakeUserService{createFn: func(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error) {
			return nil, errors.New("boom")
		}}
		s := NewUserServer(us, &fakeApiKeyService{})
		req := &userpbv3.User{}

		resp, err := s.CreateUser(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
		assert.Equal(t, "boom", resp.Status.Reason)
	})
}

func TestUserServer_GetUsers(t *testing.T) {
	us := &fakeUserService{listFn: func(ctx context.Context, opts ...query.Option) (*userpbv3.UserList, error) {
		return &userpbv3.UserList{}, nil
	}}
	s := NewUserServer(us, &fakeApiKeyService{})

	_, err := s.GetUsers(context.Background(), &commonv3.QueryOptions{})
	require.NoError(t, err)
}

func TestUserServer_GetUser(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		us := &fakeUserService{getByNameFn: func(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error) {
			return &userpbv3.User{}, nil
		}}
		s := NewUserServer(us, &fakeApiKeyService{})

		resp, err := s.GetUser(context.Background(), &userpbv3.User{})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		us := &fakeUserService{getByNameFn: func(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error) {
			return nil, errors.New("not found")
		}}
		s := NewUserServer(us, &fakeApiKeyService{})
		req := &userpbv3.User{}

		resp, err := s.GetUser(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}

func TestUserServer_GetUserInfo(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		us := &fakeUserService{getUserInfoFn: func(ctx context.Context, u *userpbv3.User) (*userpbv3.UserInfo, error) {
			return &userpbv3.UserInfo{}, nil
		}}
		s := NewUserServer(us, &fakeApiKeyService{})

		resp, err := s.GetUserInfo(context.Background(), &userpbv3.User{})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
		assert.Equal(t, commonv3.ConditionStatus_StatusOK, resp.Status.ConditionStatus)
	})

	// BUG: unlike every other handler in this package, GetUserInfo's error
	// branch stamps the Failed status onto req (the input) but returns resp
	// (the service's return value) as the response, not req. Since resp is
	// nil on a typical service error, the stamped status is discarded and
	// callers get a nil UserInfo with no status information at all. Compare
	// with GetUser/CreateUser/UpdateUser above, which correctly return req.
	t.Run("propagates the service error but drops the stamped status because it returns resp, not req", func(t *testing.T) {
		us := &fakeUserService{getUserInfoFn: func(ctx context.Context, u *userpbv3.User) (*userpbv3.UserInfo, error) {
			return nil, errors.New("boom")
		}}
		s := NewUserServer(us, &fakeApiKeyService{})
		req := &userpbv3.User{}

		resp, err := s.GetUserInfo(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		require.NotNil(t, req.Status)
		assert.Equal(t, commonv3.ConditionStatus_StatusFailed, req.Status.ConditionStatus)
	})
}

func TestUserServer_DeleteUser(t *testing.T) {
	want := &rpcv3.UserDeleteApiKeysResponse{}
	us := &fakeUserService{deleteFn: func(ctx context.Context, u *userpbv3.User) (*rpcv3.UserDeleteApiKeysResponse, error) {
		return want, nil
	}}
	s := NewUserServer(us, &fakeApiKeyService{})

	got, err := s.DeleteUser(context.Background(), &userpbv3.User{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestUserServer_UpdateUser(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		us := &fakeUserService{updateFn: func(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error) {
			return &userpbv3.User{}, nil
		}}
		s := NewUserServer(us, &fakeApiKeyService{})

		resp, err := s.UpdateUser(context.Background(), &userpbv3.User{})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		us := &fakeUserService{updateFn: func(ctx context.Context, u *userpbv3.User) (*userpbv3.User, error) {
			return nil, errors.New("boom")
		}}
		s := NewUserServer(us, &fakeApiKeyService{})
		req := &userpbv3.User{}

		resp, err := s.UpdateUser(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}

func TestUserServer_UpdateUserForceReset(t *testing.T) {
	t.Run("fails when no session data is present in the context", func(t *testing.T) {
		s := NewUserServer(&fakeUserService{}, &fakeApiKeyService{})

		_, err := s.UpdateUserForceReset(context.Background(), &rpcv3.UpdateForceResetRequest{})
		assert.Error(t, err)
	})

	t.Run("resets the flag for the session's username on success", func(t *testing.T) {
		var gotUsername string
		us := &fakeUserService{updateForceResetFlagFn: func(ctx context.Context, username string) error {
			gotUsername = username
			return nil
		}}
		s := NewUserServer(us, &fakeApiKeyService{})

		ctx := context.WithValue(context.Background(), common.SessionDataKey, &commonv3.SessionData{Username: "alice"})
		resp, err := s.UpdateUserForceReset(ctx, &rpcv3.UpdateForceResetRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "alice", gotUsername)
	})

	t.Run("propagates the service error", func(t *testing.T) {
		want := errors.New("boom")
		us := &fakeUserService{updateForceResetFlagFn: func(ctx context.Context, username string) error {
			return want
		}}
		s := NewUserServer(us, &fakeApiKeyService{})

		ctx := context.WithValue(context.Background(), common.SessionDataKey, &commonv3.SessionData{Username: "alice"})
		_, err := s.UpdateUserForceReset(ctx, &rpcv3.UpdateForceResetRequest{})
		assert.Same(t, want, err)
	})
}

func TestUserServer_DownloadCliConfig(t *testing.T) {
	t.Run("fails when no session data is present in the context", func(t *testing.T) {
		s := NewUserServer(&fakeUserService{}, &fakeApiKeyService{})

		_, err := s.DownloadCliConfig(context.Background(), &rpcv3.CliConfigRequest{})
		assert.Error(t, err)
	})

	t.Run("marshals the retrieved cli config as JSON on success", func(t *testing.T) {
		want := &common.CliConfigDownloadData{Profile: "default", Project: "proj1"}
		var gotReq *rpcv3.ApiKeyRequest
		us := &fakeUserService{retrieveCliConfigFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*common.CliConfigDownloadData, error) {
			gotReq = req
			return want, nil
		}}
		s := NewUserServer(us, &fakeApiKeyService{})

		ctx := context.WithValue(context.Background(), common.SessionDataKey, &commonv3.SessionData{
			Username: "alice", Account: "acct1", Organization: "org1", Partner: "p1",
		})
		resp, err := s.DownloadCliConfig(ctx, &rpcv3.CliConfigRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "application/json", resp.ContentType)

		require.NotNil(t, gotReq)
		assert.Equal(t, "alice", gotReq.Username)
		assert.Equal(t, "acct1", gotReq.Id)
		assert.Equal(t, "org1", gotReq.OrganizationId)
		assert.Equal(t, "p1", gotReq.PartnerId)

		var got common.CliConfigDownloadData
		require.NoError(t, json.Unmarshal(resp.Data, &got))
		assert.Equal(t, *want, got)
	})

	t.Run("propagates a RetrieveCliConfig error", func(t *testing.T) {
		want := errors.New("boom")
		us := &fakeUserService{retrieveCliConfigFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*common.CliConfigDownloadData, error) {
			return nil, want
		}}
		s := NewUserServer(us, &fakeApiKeyService{})

		ctx := context.WithValue(context.Background(), common.SessionDataKey, &commonv3.SessionData{Username: "alice"})
		resp, err := s.DownloadCliConfig(ctx, &rpcv3.CliConfigRequest{})
		assert.Same(t, want, err)
		assert.Nil(t, resp)
	})
}

func TestUserServer_UserListApiKeys(t *testing.T) {
	want := &rpcv3.UserListApiKeysResponse{}
	ks := &fakeApiKeyService{listFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserListApiKeysResponse, error) {
		return want, nil
	}}
	s := NewUserServer(&fakeUserService{}, ks)

	got, err := s.UserListApiKeys(context.Background(), &rpcv3.ApiKeyRequest{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestUserServer_UserDeleteApiKeys(t *testing.T) {
	t.Run("returns an empty response on success", func(t *testing.T) {
		ks := &fakeApiKeyService{deleteFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserDeleteApiKeysResponse, error) {
			return &rpcv3.UserDeleteApiKeysResponse{}, nil
		}}
		s := NewUserServer(&fakeUserService{}, ks)

		resp, err := s.UserDeleteApiKeys(context.Background(), &rpcv3.ApiKeyRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("propagates the service error and returns a nil response", func(t *testing.T) {
		want := errors.New("boom")
		ks := &fakeApiKeyService{deleteFn: func(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserDeleteApiKeysResponse, error) {
			return nil, want
		}}
		s := NewUserServer(&fakeUserService{}, ks)

		resp, err := s.UserDeleteApiKeys(context.Background(), &rpcv3.ApiKeyRequest{})
		assert.Same(t, want, err)
		assert.Nil(t, resp)
	})
}

func TestUserServer_UserForgotPassword(t *testing.T) {
	want := &rpcv3.UserForgotPasswordResponse{RecoveryLink: "http://example.com/reset"}
	us := &fakeUserService{forgotPasswordFn: func(ctx context.Context, req *rpcv3.UserForgotPasswordRequest) (*rpcv3.UserForgotPasswordResponse, error) {
		return want, nil
	}}
	s := NewUserServer(us, &fakeApiKeyService{})

	got, err := s.UserForgotPassword(context.Background(), &rpcv3.UserForgotPasswordRequest{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestUserServer_AuditLogWebhook(t *testing.T) {
	want := &rpcv3.UserLoginAuditResponse{}
	us := &fakeUserService{createLoginAuditLogFn: func(ctx context.Context, req *rpcv3.UserLoginAuditRequest) (*rpcv3.UserLoginAuditResponse, error) {
		return want, nil
	}}
	s := NewUserServer(us, &fakeApiKeyService{})

	got, err := s.AuditLogWebhook(context.Background(), &rpcv3.UserLoginAuditRequest{})
	require.NoError(t, err)
	assert.Same(t, want, got)
}
