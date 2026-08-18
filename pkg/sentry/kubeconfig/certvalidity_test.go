package kubeconfig

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paralus/paralus/internal/constants"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeKubeconfigSettingService implements service.KubeconfigSettingService
// with Get wired via a function field; Patch is unused here.
type fakeKubeconfigSettingService struct {
	getFn func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error)
}

func (f *fakeKubeconfigSettingService) Get(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
	return f.getFn(ctx, orgID, accountID, isSSO)
}
func (f *fakeKubeconfigSettingService) Patch(ctx context.Context, ks *sentry.KubeconfigSetting) error {
	panic("not implemented")
}

func TestGetCertValidity(t *testing.T) {
	t.Run("uses the user-level setting when present", func(t *testing.T) {
		kss := &fakeKubeconfigSettingService{
			getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
				if accountID != "" {
					return &sentry.KubeconfigSetting{ValiditySeconds: 3600}, nil
				}
				return nil, constants.ErrNotFound
			},
		}
		d, err := getCertValidity(context.Background(), "org1", "acct1", false, kss)
		require.NoError(t, err)
		assert.Equal(t, time.Hour, d)
	})

	t.Run("falls back to org-level setting when user-level is not found", func(t *testing.T) {
		kss := &fakeKubeconfigSettingService{
			getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
				if accountID == "" {
					return &sentry.KubeconfigSetting{ValiditySeconds: 7200}, nil
				}
				return nil, constants.ErrNotFound
			},
		}
		d, err := getCertValidity(context.Background(), "org1", "acct1", false, kss)
		require.NoError(t, err)
		assert.Equal(t, 2*time.Hour, d)
	})

	t.Run("defaults to 360 days when neither setting is found", func(t *testing.T) {
		kss := &fakeKubeconfigSettingService{
			getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
				return nil, constants.ErrNotFound
			},
		}
		d, err := getCertValidity(context.Background(), "org1", "acct1", false, kss)
		require.NoError(t, err)
		assert.Equal(t, 360*24*time.Hour, d)
	})

	t.Run("propagates a non-not-found error from the user-level lookup", func(t *testing.T) {
		kss := &fakeKubeconfigSettingService{
			getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
				return nil, errors.New("boom")
			},
		}
		_, err := getCertValidity(context.Background(), "org1", "acct1", false, kss)
		assert.Error(t, err)
	})

	t.Run("propagates a non-not-found error from the org-level lookup", func(t *testing.T) {
		kss := &fakeKubeconfigSettingService{
			getFn: func(ctx context.Context, orgID, accountID string, isSSO bool) (*sentry.KubeconfigSetting, error) {
				if accountID != "" {
					return nil, constants.ErrNotFound
				}
				return nil, errors.New("boom")
			},
		}
		_, err := getCertValidity(context.Background(), "org1", "acct1", false, kss)
		assert.Error(t, err)
	})
}
