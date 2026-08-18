package server

// GetUserAuthorization is a thin one-line wrapper around
// authz.GetAuthorization, which lives in pkg/sentry/authz and has its own
// test scope covering its many branches (org/user kubectl settings,
// partner/super-admin bypass, per-cluster role resolution, namespace
// provisioning, etc.) across seven service dependencies. Exercising those
// branches through this wrapper would mean re-implementing that package's
// test suite here. Per the task scope, only a minimal smoke test is
// provided: an error-propagation case (the first dependency call,
// kss.Get, failing) and a success case that reaches a real, distinct
// return path (the system-user short-circuit in
// authz.GetAuthorization/getSystemUserAuthz) without needing to wire the
// other six dependencies at all, since that path returns before they are
// ever called.

import (
	"context"
	"testing"

	"github.com/paralus/paralus/internal/constants"
	sentryrpc "github.com/paralus/paralus/proto/rpc/sentry"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterAuthzServer_GetUserAuthorization(t *testing.T) {
	t.Run("propagates a non-ErrNotFound error from the first dependency call (kss.Get)", func(t *testing.T) {
		kss := &fakeKubeconfigSettingService{getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
			return nil, assert.AnError
		}}
		s := NewClusterAuthzServer(nil, nil, nil, nil, nil, kss, nil)

		resp, err := s.GetUserAuthorization(context.Background(), &sentryrpc.GetUserAuthorizationRequest{
			UserCN: "a=acc1/o=org1/p=part1",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("system user CN short-circuits to getSystemUserAuthz without needing the other six dependencies", func(t *testing.T) {
		kss := &fakeKubeconfigSettingService{getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
			return nil, constants.ErrNotFound
		}}
		s := NewClusterAuthzServer(nil, nil, nil, nil, nil, kss, nil)

		resp, err := s.GetUserAuthorization(context.Background(), &sentryrpc.GetUserAuthorizationRequest{
			UserCN: "su=true/u=sysuser/a=acc1/o=org1/p=part1",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "sysuser", resp.UserName)
		require.NotNil(t, resp.ServiceAccount)
	})
}
