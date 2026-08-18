package server

import (
	"context"
	"errors"
	"testing"

	"github.com/paralus/paralus/internal/constants"
	sentryrpc "github.com/paralus/paralus/proto/rpc/sentry"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeKubectlClusterSettingsService implements service.KubectlClusterSettingsService.
type fakeKubectlClusterSettingsService struct {
	getFn   func(ctx context.Context, orgID, clusterID string) (*sentry.KubectlClusterSettings, error)
	patchFn func(ctx context.Context, kc *sentry.KubectlClusterSettings) error
}

func (f *fakeKubectlClusterSettingsService) Get(ctx context.Context, orgID, clusterID string) (*sentry.KubectlClusterSettings, error) {
	return f.getFn(ctx, orgID, clusterID)
}
func (f *fakeKubectlClusterSettingsService) Patch(ctx context.Context, kc *sentry.KubectlClusterSettings) error {
	return f.patchFn(ctx, kc)
}

// bootstrapCountService only wires the two BootstrapService methods this
// server file actually calls; everything else panics.
type bootstrapCountService struct {
	fakeBootstrapService
	countFn           func(ctx context.Context, clusterID, orgID string) (int, error)
	agentForClusterFn func(ctx context.Context, clusterID, orgID string) (*sentry.BootstrapAgent, error)
}

func (f *bootstrapCountService) GetBootstrapAgentCountForClusterID(ctx context.Context, clusterID, orgID string) (int, error) {
	return f.countFn(ctx, clusterID, orgID)
}
func (f *bootstrapCountService) GetBootstrapAgentForClusterID(ctx context.Context, clusterID, orgID string) (*sentry.BootstrapAgent, error) {
	return f.agentForClusterFn(ctx, clusterID, orgID)
}

func TestKubectlClusterSettingsServer_GetKubectlClusterSettings(t *testing.T) {
	t.Run("invalid cluster scope is rejected before touching any service", func(t *testing.T) {
		bs := &bootstrapCountService{}
		kcs := &fakeKubectlClusterSettingsService{}
		s := NewKubectlClusterSettingsServer(bs, kcs)

		req := &sentryrpc.GetKubectlClusterSettingsRequest{Opts: &commonv3.QueryOptions{UrlScope: "not-a-cluster-scope"}}
		_, err := s.GetKubectlClusterSettings(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("propagates a bootstrap-agent-count error", func(t *testing.T) {
		want := errors.New("count failed")
		bs := &bootstrapCountService{countFn: func(ctx context.Context, clusterID, orgID string) (int, error) {
			return 0, want
		}}
		kcs := &fakeKubectlClusterSettingsService{}
		s := NewKubectlClusterSettingsServer(bs, kcs)

		req := &sentryrpc.GetKubectlClusterSettingsRequest{Opts: &commonv3.QueryOptions{UrlScope: "cluster/c1", Organization: "org1"}}
		_, err := s.GetKubectlClusterSettings(context.Background(), req)
		assert.Same(t, want, err)
	})

	t.Run("returns both kubectl flags disabled when no settings row exists yet", func(t *testing.T) {
		bs := &bootstrapCountService{countFn: func(ctx context.Context, clusterID, orgID string) (int, error) {
			return 1, nil
		}}
		kcs := &fakeKubectlClusterSettingsService{getFn: func(ctx context.Context, orgID, clusterID string) (*sentry.KubectlClusterSettings, error) {
			return nil, constants.ErrNotFound
		}}
		s := NewKubectlClusterSettingsServer(bs, kcs)

		req := &sentryrpc.GetKubectlClusterSettingsRequest{Opts: &commonv3.QueryOptions{UrlScope: "cluster/c1", Organization: "org1"}}
		resp, err := s.GetKubectlClusterSettings(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, resp.DisableWebKubectl)
		assert.False(t, resp.DisableCLIKubectl)
	})

	t.Run("propagates a non-ErrNotFound settings-lookup error", func(t *testing.T) {
		want := errors.New("db is down")
		bs := &bootstrapCountService{countFn: func(ctx context.Context, clusterID, orgID string) (int, error) {
			return 1, nil
		}}
		kcs := &fakeKubectlClusterSettingsService{getFn: func(ctx context.Context, orgID, clusterID string) (*sentry.KubectlClusterSettings, error) {
			return nil, want
		}}
		s := NewKubectlClusterSettingsServer(bs, kcs)

		req := &sentryrpc.GetKubectlClusterSettingsRequest{Opts: &commonv3.QueryOptions{UrlScope: "cluster/c1", Organization: "org1"}}
		_, err := s.GetKubectlClusterSettings(context.Background(), req)
		assert.Same(t, want, err)
	})

	t.Run("returns the stored flags when a settings row exists", func(t *testing.T) {
		bs := &bootstrapCountService{countFn: func(ctx context.Context, clusterID, orgID string) (int, error) {
			return 1, nil
		}}
		kcs := &fakeKubectlClusterSettingsService{getFn: func(ctx context.Context, orgID, clusterID string) (*sentry.KubectlClusterSettings, error) {
			return &sentry.KubectlClusterSettings{DisableWebKubectl: true, DisableCLIKubectl: false}, nil
		}}
		s := NewKubectlClusterSettingsServer(bs, kcs)

		req := &sentryrpc.GetKubectlClusterSettingsRequest{Opts: &commonv3.QueryOptions{UrlScope: "cluster/c1", Organization: "org1"}}
		resp, err := s.GetKubectlClusterSettings(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.DisableWebKubectl)
		assert.False(t, resp.DisableCLIKubectl)
	})
}

func TestKubectlClusterSettingsServer_UpdateKubectlClusterSettings(t *testing.T) {
	t.Run("invalid cluster scope is rejected before touching any service", func(t *testing.T) {
		bs := &bootstrapCountService{}
		kcs := &fakeKubectlClusterSettingsService{}
		s := NewKubectlClusterSettingsServer(bs, kcs)

		req := &sentryrpc.UpdateKubectlClusterSettingsRequest{Opts: &commonv3.QueryOptions{UrlScope: "bogus"}}
		_, err := s.UpdateKubectlClusterSettings(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("propagates a bootstrap-agent-count error", func(t *testing.T) {
		want := errors.New("count failed")
		bs := &bootstrapCountService{countFn: func(ctx context.Context, clusterID, orgID string) (int, error) {
			return 0, want
		}}
		kcs := &fakeKubectlClusterSettingsService{}
		s := NewKubectlClusterSettingsServer(bs, kcs)

		req := &sentryrpc.UpdateKubectlClusterSettingsRequest{Opts: &commonv3.QueryOptions{UrlScope: "cluster/c1", Organization: "org1"}}
		_, err := s.UpdateKubectlClusterSettings(context.Background(), req)
		assert.Same(t, want, err)
	})

	t.Run("patches settings with the resolved cluster ID and requested flags, tolerating a missing agent", func(t *testing.T) {
		var patched *sentry.KubectlClusterSettings
		bs := &bootstrapCountService{
			countFn: func(ctx context.Context, clusterID, orgID string) (int, error) { return 1, nil },
			agentForClusterFn: func(ctx context.Context, clusterID, orgID string) (*sentry.BootstrapAgent, error) {
				return nil, errors.New("no agent yet")
			},
		}
		kcs := &fakeKubectlClusterSettingsService{patchFn: func(ctx context.Context, kc *sentry.KubectlClusterSettings) error {
			patched = kc
			return nil
		}}
		s := NewKubectlClusterSettingsServer(bs, kcs)

		req := &sentryrpc.UpdateKubectlClusterSettingsRequest{
			Opts:              &commonv3.QueryOptions{UrlScope: "cluster/c1", Organization: "org1", Partner: "p1"},
			DisableWebKubectl: true,
			DisableCLIKubectl: true,
		}
		resp, err := s.UpdateKubectlClusterSettings(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, patched)
		assert.Equal(t, "c1", patched.Name)
		assert.Equal(t, "org1", patched.OrganizationID)
		assert.Equal(t, "p1", patched.PartnerID)
		assert.True(t, patched.DisableWebKubectl)
		assert.True(t, patched.DisableCLIKubectl)
	})

	t.Run("propagates a patch error", func(t *testing.T) {
		want := errors.New("patch failed")
		bs := &bootstrapCountService{
			countFn:           func(ctx context.Context, clusterID, orgID string) (int, error) { return 1, nil },
			agentForClusterFn: func(ctx context.Context, clusterID, orgID string) (*sentry.BootstrapAgent, error) { return nil, nil },
		}
		kcs := &fakeKubectlClusterSettingsService{patchFn: func(ctx context.Context, kc *sentry.KubectlClusterSettings) error {
			return want
		}}
		s := NewKubectlClusterSettingsServer(bs, kcs)

		req := &sentryrpc.UpdateKubectlClusterSettingsRequest{Opts: &commonv3.QueryOptions{UrlScope: "cluster/c1", Organization: "org1"}}
		_, err := s.UpdateKubectlClusterSettings(context.Background(), req)
		assert.Same(t, want, err)
	})
}
