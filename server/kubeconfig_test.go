package server

import (
	"context"
	"errors"
	"testing"

	"github.com/paralus/paralus/internal/constants"
	"github.com/paralus/paralus/pkg/service"
	sentryrpc "github.com/paralus/paralus/proto/rpc/sentry"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeKubeconfigSettingService implements service.KubeconfigSettingService.
type fakeKubeconfigSettingService struct {
	getFn   func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error)
	patchFn func(ctx context.Context, ks *sentry.KubeconfigSetting) error
}

func (f *fakeKubeconfigSettingService) Get(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
	return f.getFn(ctx, orgID, accountID, isSSO)
}
func (f *fakeKubeconfigSettingService) Patch(ctx context.Context, ks *sentry.KubeconfigSetting) error {
	return f.patchFn(ctx, ks)
}

// fakeKubeconfigRevocationService implements service.KubeconfigRevocationService.
type fakeKubeconfigRevocationService struct {
	getFn   func(ctx context.Context, orgID, accountID string, isSSOUser bool) (*sentry.KubeconfigRevocation, error)
	patchFn func(ctx context.Context, kr *sentry.KubeconfigRevocation) error
}

func (f *fakeKubeconfigRevocationService) Get(ctx context.Context, orgID, accountID string, isSSOUser bool) (*sentry.KubeconfigRevocation, error) {
	return f.getFn(ctx, orgID, accountID, isSSOUser)
}
func (f *fakeKubeconfigRevocationService) Patch(ctx context.Context, kr *sentry.KubeconfigRevocation) error {
	return f.patchFn(ctx, kr)
}

// newTestKubeConfigServer builds a *kubeConfigServer directly (rather than
// going through the NewKubeConfigServer constructor, which only returns the
// sentryrpc.KubeConfigServiceServer interface) so tests can also reach
// RevokeKubeconfigSSO, which is implemented on the concrete type but is not
// part of that interface.
func newTestKubeConfigServer(kss *fakeKubeconfigSettingService, krs *fakeKubeconfigRevocationService) *kubeConfigServer {
	var kssIface service.KubeconfigSettingService
	if kss != nil {
		kssIface = kss
	}
	var krsIface service.KubeconfigRevocationService
	if krs != nil {
		krsIface = krs
	}
	return &kubeConfigServer{kss: kssIface, krs: krsIface}
}

func TestKubeConfigServer_RevokeKubeconfig(t *testing.T) {
	t.Run("revokes for the current session user (non-SSO) when there is no url scope", func(t *testing.T) {
		var patched *sentry.KubeconfigRevocation
		krs := &fakeKubeconfigRevocationService{patchFn: func(ctx context.Context, kr *sentry.KubeconfigRevocation) error {
			patched = kr
			return nil
		}}
		s := newTestKubeConfigServer(nil, krs)

		req := &sentryrpc.RevokeKubeconfigRequest{Opts: &commonv3.QueryOptions{
			Organization: "org1", Partner: "p1", Account: "acc1", IsSSOUser: true,
		}}
		resp, err := s.RevokeKubeconfig(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, patched)
		assert.Equal(t, "acc1", patched.AccountID)
		assert.True(t, patched.IsSSOUser, "empty UrlScope should carry through opts.IsSSOUser")
	})

	t.Run("forces isSSOUser false when a url scope is present, regardless of opts.IsSSOUser", func(t *testing.T) {
		var patched *sentry.KubeconfigRevocation
		krs := &fakeKubeconfigRevocationService{patchFn: func(ctx context.Context, kr *sentry.KubeconfigRevocation) error {
			patched = kr
			return nil
		}}
		s := newTestKubeConfigServer(nil, krs)

		req := &sentryrpc.RevokeKubeconfigRequest{Opts: &commonv3.QueryOptions{
			Organization: "org1", Account: "acc1", IsSSOUser: true, UrlScope: "user/acc1",
		}}
		_, err := s.RevokeKubeconfig(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, patched)
		assert.False(t, patched.IsSSOUser)
	})

	t.Run("propagates a patch error", func(t *testing.T) {
		want := errors.New("boom")
		krs := &fakeKubeconfigRevocationService{patchFn: func(ctx context.Context, kr *sentry.KubeconfigRevocation) error {
			return want
		}}
		s := newTestKubeConfigServer(nil, krs)

		_, err := s.RevokeKubeconfig(context.Background(), &sentryrpc.RevokeKubeconfigRequest{Opts: &commonv3.QueryOptions{}})
		assert.Same(t, want, err)
	})
}

func TestKubeConfigServer_RevokeKubeconfigSSO(t *testing.T) {
	t.Run("always revokes with IsSSOUser true", func(t *testing.T) {
		var patched *sentry.KubeconfigRevocation
		krs := &fakeKubeconfigRevocationService{patchFn: func(ctx context.Context, kr *sentry.KubeconfigRevocation) error {
			patched = kr
			return nil
		}}
		s := newTestKubeConfigServer(nil, krs)

		req := &sentryrpc.RevokeKubeconfigRequest{Opts: &commonv3.QueryOptions{Account: "acc1"}}
		resp, err := s.RevokeKubeconfigSSO(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, patched)
		assert.True(t, patched.IsSSOUser)
	})

	t.Run("propagates a patch error", func(t *testing.T) {
		want := errors.New("boom")
		krs := &fakeKubeconfigRevocationService{patchFn: func(ctx context.Context, kr *sentry.KubeconfigRevocation) error {
			return want
		}}
		s := newTestKubeConfigServer(nil, krs)

		_, err := s.RevokeKubeconfigSSO(context.Background(), &sentryrpc.RevokeKubeconfigRequest{Opts: &commonv3.QueryOptions{}})
		assert.Same(t, want, err)
	})
}

func TestKubeConfigServer_GetOrganizationSetting(t *testing.T) {
	t.Run("rejects an invalid organization url scope", func(t *testing.T) {
		s := newTestKubeConfigServer(nil, nil)

		req := &sentryrpc.GetKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "bogus"}}
		_, err := s.GetOrganizationSetting(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("returns 8-hour default validities when no settings row exists", func(t *testing.T) {
		kss := &fakeKubeconfigSettingService{getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
			return nil, constants.ErrNotFound
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.GetKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "organization/org1", Organization: "org1"}}
		resp, err := s.GetOrganizationSetting(context.Background(), req)
		require.NoError(t, err)
		assert.EqualValues(t, 28800, resp.ValiditySeconds)
		assert.EqualValues(t, 28800, resp.SaValiditySeconds)
	})

	t.Run("propagates a non-ErrNotFound lookup error", func(t *testing.T) {
		want := errors.New("db is down")
		kss := &fakeKubeconfigSettingService{getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
			return nil, want
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.GetKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "organization/org1", Organization: "org1"}}
		_, err := s.GetOrganizationSetting(context.Background(), req)
		assert.Same(t, want, err)
	})

	t.Run("maps the stored settings including SaValiditySeconds when a row exists", func(t *testing.T) {
		kss := &fakeKubeconfigSettingService{getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
			assert.Equal(t, "org1", orgID)
			return &sentry.KubeconfigSetting{
				ValiditySeconds:   100,
				SaValiditySeconds: 200,
				DisableWebKubectl: true,
			}, nil
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.GetKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "organization/org1", Organization: "different-org"}}
		resp, err := s.GetOrganizationSetting(context.Background(), req)
		require.NoError(t, err)
		assert.EqualValues(t, 100, resp.ValiditySeconds)
		assert.EqualValues(t, 200, resp.SaValiditySeconds)
		assert.True(t, resp.DisableWebKubectl)
	})
}

func TestKubeConfigServer_GetUserSetting(t *testing.T) {
	t.Run("rejects an invalid user url scope", func(t *testing.T) {
		s := newTestKubeConfigServer(nil, nil)

		req := &sentryrpc.GetKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "bogus"}}
		_, err := s.GetUserSetting(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("falls back to organization settings when no per-user row exists", func(t *testing.T) {
		var gotIsSSO bool
		kss := &fakeKubeconfigSettingService{getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
			gotIsSSO = isSSO
			if accountID == "acc1" {
				return nil, constants.ErrNotFound
			}
			return &sentry.KubeconfigSetting{ValiditySeconds: 42}, nil
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.GetKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "user/acc1", Organization: "org1"}}
		resp, err := s.GetUserSetting(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, gotIsSSO)
		assert.EqualValues(t, 42, resp.ValiditySeconds)
	})

	t.Run("propagates a non-ErrNotFound lookup error", func(t *testing.T) {
		want := errors.New("boom")
		kss := &fakeKubeconfigSettingService{getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
			return nil, want
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.GetKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "user/acc1", Organization: "org1"}}
		_, err := s.GetUserSetting(context.Background(), req)
		assert.Same(t, want, err)
	})

	t.Run("maps the stored per-user settings when a row exists", func(t *testing.T) {
		kss := &fakeKubeconfigSettingService{getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
			assert.Equal(t, "acc1", accountID)
			return &sentry.KubeconfigSetting{ValiditySeconds: 55, DisableCLIKubectl: true}, nil
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.GetKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "user/acc1", Organization: "org1"}}
		resp, err := s.GetUserSetting(context.Background(), req)
		require.NoError(t, err)
		assert.EqualValues(t, 55, resp.ValiditySeconds)
		assert.True(t, resp.DisableCLIKubectl)
	})
}

func TestKubeConfigServer_GetSSOUserSetting(t *testing.T) {
	t.Run("looks up settings with isSSO true and falls back to organization settings on ErrNotFound", func(t *testing.T) {
		var gotIsSSOForUserLookup bool
		kss := &fakeKubeconfigSettingService{getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
			if accountID == "acc1" {
				// This is the per-user lookup GetSSOUserSetting itself performs.
				gotIsSSOForUserLookup = isSSO
				return nil, constants.ErrNotFound
			}
			// This is the fallback org-level lookup made by
			// GetOrganizationSetting, which always passes isSSO=false.
			return &sentry.KubeconfigSetting{ValiditySeconds: 7}, nil
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.GetKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "user/acc1", Organization: "org1"}}
		resp, err := s.GetSSOUserSetting(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, gotIsSSOForUserLookup)
		assert.EqualValues(t, 7, resp.ValiditySeconds)
	})

	t.Run("rejects an invalid user url scope", func(t *testing.T) {
		s := newTestKubeConfigServer(nil, nil)

		_, err := s.GetSSOUserSetting(context.Background(), &sentryrpc.GetKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "bogus"}})
		assert.Error(t, err)
	})
}

func TestKubeConfigServer_UpdateOrganizationSetting(t *testing.T) {
	t.Run("rejects an invalid organization url scope", func(t *testing.T) {
		s := newTestKubeConfigServer(nil, nil)

		req := &sentryrpc.UpdateKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "bogus"}}
		_, err := s.UpdateOrganizationSetting(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("rejects when the resolved scope disagrees with opts.Organization", func(t *testing.T) {
		s := newTestKubeConfigServer(nil, nil)

		req := &sentryrpc.UpdateKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "organization/org1", Organization: "org2"}}
		_, err := s.UpdateOrganizationSetting(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("patches the full setting including SaValiditySeconds and EnablePrivateRelay on success", func(t *testing.T) {
		var patched *sentry.KubeconfigSetting
		kss := &fakeKubeconfigSettingService{patchFn: func(ctx context.Context, ks *sentry.KubeconfigSetting) error {
			patched = ks
			return nil
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.UpdateKubeconfigSettingRequest{
			Opts:               &commonv3.QueryOptions{UrlScope: "organization/org1", Organization: "org1", Partner: "p1"},
			ValiditySeconds:    111,
			SaValiditySeconds:  222,
			EnablePrivateRelay: true,
		}
		resp, err := s.UpdateOrganizationSetting(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, patched)
		assert.EqualValues(t, 111, patched.ValiditySeconds)
		assert.EqualValues(t, 222, patched.SaValiditySeconds)
		assert.True(t, patched.EnablePrivateRelay)
	})

	t.Run("propagates a patch error", func(t *testing.T) {
		want := errors.New("boom")
		kss := &fakeKubeconfigSettingService{patchFn: func(ctx context.Context, ks *sentry.KubeconfigSetting) error {
			return want
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.UpdateKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "organization/org1", Organization: "org1"}}
		_, err := s.UpdateOrganizationSetting(context.Background(), req)
		assert.Same(t, want, err)
	})
}

func TestKubeConfigServer_UpdateUserSetting(t *testing.T) {
	t.Run("rejects an invalid user url scope", func(t *testing.T) {
		s := newTestKubeConfigServer(nil, nil)

		req := &sentryrpc.UpdateKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "bogus"}}
		_, err := s.UpdateUserSetting(context.Background(), req)
		assert.Error(t, err)
	})

	// BUG: unlike UpdateOrganizationSetting, UpdateUserSetting never copies
	// req.SaValiditySeconds or req.EnablePrivateRelay into the patched
	// sentry.KubeconfigSetting. A caller who sets either field on a
	// per-user update silently has it dropped -- the setting is patched
	// with those fields left at their zero value regardless of what was
	// requested. Same bug is present in UpdateSSOUserSetting below.
	t.Run("silently drops SaValiditySeconds and EnablePrivateRelay from the patched setting", func(t *testing.T) {
		var patched *sentry.KubeconfigSetting
		kss := &fakeKubeconfigSettingService{patchFn: func(ctx context.Context, ks *sentry.KubeconfigSetting) error {
			patched = ks
			return nil
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.UpdateKubeconfigSettingRequest{
			Opts:               &commonv3.QueryOptions{UrlScope: "user/acc1", Organization: "org1", Partner: "p1"},
			ValiditySeconds:    111,
			SaValiditySeconds:  222,
			EnablePrivateRelay: true,
		}
		resp, err := s.UpdateUserSetting(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, patched)
		assert.Equal(t, "acc1", patched.AccountID)
		assert.EqualValues(t, 111, patched.ValiditySeconds)
		assert.False(t, patched.IsSSOUser)
		assert.EqualValues(t, 0, patched.SaValiditySeconds, "BUG: SaValiditySeconds is dropped by UpdateUserSetting")
		assert.False(t, patched.EnablePrivateRelay, "BUG: EnablePrivateRelay is dropped by UpdateUserSetting")
	})

	t.Run("propagates a patch error", func(t *testing.T) {
		want := errors.New("boom")
		kss := &fakeKubeconfigSettingService{patchFn: func(ctx context.Context, ks *sentry.KubeconfigSetting) error {
			return want
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.UpdateKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "user/acc1", Organization: "org1"}}
		_, err := s.UpdateUserSetting(context.Background(), req)
		assert.Same(t, want, err)
	})
}

func TestKubeConfigServer_UpdateSSOUserSetting(t *testing.T) {
	t.Run("rejects an invalid user url scope", func(t *testing.T) {
		s := newTestKubeConfigServer(nil, nil)

		req := &sentryrpc.UpdateKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "bogus"}}
		_, err := s.UpdateSSOUserSetting(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("patches with IsSSOUser true", func(t *testing.T) {
		var patched *sentry.KubeconfigSetting
		kss := &fakeKubeconfigSettingService{patchFn: func(ctx context.Context, ks *sentry.KubeconfigSetting) error {
			patched = ks
			return nil
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.UpdateKubeconfigSettingRequest{
			Opts:            &commonv3.QueryOptions{UrlScope: "user/acc1", Organization: "org1"},
			ValiditySeconds: 99,
		}
		resp, err := s.UpdateSSOUserSetting(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, patched)
		assert.True(t, patched.IsSSOUser)
		assert.EqualValues(t, 99, patched.ValiditySeconds)
	})

	t.Run("propagates a patch error", func(t *testing.T) {
		want := errors.New("boom")
		kss := &fakeKubeconfigSettingService{patchFn: func(ctx context.Context, ks *sentry.KubeconfigSetting) error {
			return want
		}}
		s := newTestKubeConfigServer(kss, nil)

		req := &sentryrpc.UpdateKubeconfigSettingRequest{Opts: &commonv3.QueryOptions{UrlScope: "user/acc1", Organization: "org1"}}
		_, err := s.UpdateSSOUserSetting(context.Background(), req)
		assert.Same(t, want, err)
	})
}
