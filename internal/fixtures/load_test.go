package fixtures

import (
	"context"
	"testing"

	"github.com/paralus/paralus/internal/constants"
	"github.com/paralus/paralus/pkg/query"
	"github.com/paralus/paralus/pkg/sentry/cryptoutil"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBootstrapService implements service.BootstrapService with only the
// three methods loadAgentTemplates actually calls wired via function
// fields; the rest panic since nothing here exercises them.
type fakeBootstrapService struct {
	getBootstrapInfraFn           func(ctx context.Context, name string) (*sentry.BootstrapInfra, error)
	patchBootstrapInfraFn         func(ctx context.Context, infra *sentry.BootstrapInfra) error
	patchBootstrapAgentTemplateFn func(ctx context.Context, template *sentry.BootstrapAgentTemplate) error
}

func (f *fakeBootstrapService) PatchBootstrapInfra(ctx context.Context, infra *sentry.BootstrapInfra) error {
	return f.patchBootstrapInfraFn(ctx, infra)
}
func (f *fakeBootstrapService) GetBootstrapInfra(ctx context.Context, name string) (*sentry.BootstrapInfra, error) {
	return f.getBootstrapInfraFn(ctx, name)
}
func (f *fakeBootstrapService) PatchBootstrapAgentTemplate(ctx context.Context, template *sentry.BootstrapAgentTemplate) error {
	return f.patchBootstrapAgentTemplateFn(ctx, template)
}
func (f *fakeBootstrapService) GetBootstrapAgentTemplate(ctx context.Context, name string) (*sentry.BootstrapAgentTemplate, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentTemplateForToken(ctx context.Context, token string) (*sentry.BootstrapAgentTemplate, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentTemplateForHost(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) SelectBootstrapAgentTemplates(ctx context.Context, opts ...query.Option) (*sentry.BootstrapAgentTemplateList, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) CreateBootstrapAgent(ctx context.Context, agent *sentry.BootstrapAgent) error {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgent(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgents(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgentList, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentForToken(ctx context.Context, token string) (*sentry.BootstrapAgent, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentCountForClusterID(ctx context.Context, clusterID string, orgID string) (int, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentForClusterID(ctx context.Context, clusterID string, orgID string) (*sentry.BootstrapAgent, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) SelectBootstrapAgents(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgentList, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) RegisterBootstrapAgent(ctx context.Context, token, ip, fingerprint string) error {
	panic("not implemented")
}
func (f *fakeBootstrapService) DeleteBootstrapAgent(ctx context.Context, templateRef string, opts ...query.Option) error {
	panic("not implemented")
}
func (f *fakeBootstrapService) PatchBootstrapAgent(ctx context.Context, ba *sentry.BootstrapAgent, templateRef string, opts ...query.Option) error {
	panic("not implemented")
}

func TestLoadRelayTemplate(t *testing.T) {
	// No service dependency: reads the real embedded fixture files.
	RelayTemplate = nil
	RelayAgentTemplate = nil

	err := loadRelayTemplate()
	require.NoError(t, err)
	assert.NotNil(t, RelayTemplate)
	assert.NotNil(t, RelayAgentTemplate)
}

func TestLoadAgentTemplates(t *testing.T) {
	t.Run("creates a CA and patches every template when no bootstrap infra exists yet", func(t *testing.T) {
		var patchInfraCalls, patchTemplateCalls int
		bs := &fakeBootstrapService{
			getBootstrapInfraFn: func(ctx context.Context, name string) (*sentry.BootstrapInfra, error) {
				return nil, constants.ErrNotFound
			},
			patchBootstrapInfraFn: func(ctx context.Context, infra *sentry.BootstrapInfra) error {
				patchInfraCalls++
				assert.NotEmpty(t, infra.Spec.CaCert)
				assert.NotEmpty(t, infra.Spec.CaKey)
				return nil
			},
			patchBootstrapAgentTemplateFn: func(ctx context.Context, template *sentry.BootstrapAgentTemplate) error {
				patchTemplateCalls++
				assert.NotEmpty(t, template.Spec.Token, "loadAgentTemplates should assign a fresh token to each item")
				return nil
			},
		}

		err := loadAgentTemplates(context.Background(), bs, map[string]interface{}{"sentryPeeringHost": "peering.example.com"}, cryptoutil.NoPassword)
		require.NoError(t, err)
		assert.Positive(t, patchInfraCalls)
		assert.Positive(t, patchTemplateCalls)
	})

	t.Run("skips CA creation when a bootstrap infra with a cert already exists", func(t *testing.T) {
		var patchInfraCalls int
		bs := &fakeBootstrapService{
			getBootstrapInfraFn: func(ctx context.Context, name string) (*sentry.BootstrapInfra, error) {
				return &sentry.BootstrapInfra{Spec: &sentry.BootstrapInfraSpec{CaCert: "existing-cert"}}, nil
			},
			patchBootstrapInfraFn: func(ctx context.Context, infra *sentry.BootstrapInfra) error {
				patchInfraCalls++
				return nil
			},
			patchBootstrapAgentTemplateFn: func(ctx context.Context, template *sentry.BootstrapAgentTemplate) error {
				return nil
			},
		}

		err := loadAgentTemplates(context.Background(), bs, map[string]interface{}{"sentryPeeringHost": "peering.example.com"}, cryptoutil.NoPassword)
		require.NoError(t, err)
		assert.Equal(t, 0, patchInfraCalls, "should not patch infra when a cert already exists")
	})

	t.Run("propagates a template-patch error", func(t *testing.T) {
		bs := &fakeBootstrapService{
			getBootstrapInfraFn: func(ctx context.Context, name string) (*sentry.BootstrapInfra, error) {
				return &sentry.BootstrapInfra{Spec: &sentry.BootstrapInfraSpec{CaCert: "existing-cert"}}, nil
			},
			patchBootstrapInfraFn: func(ctx context.Context, infra *sentry.BootstrapInfra) error {
				return nil
			},
			patchBootstrapAgentTemplateFn: func(ctx context.Context, template *sentry.BootstrapAgentTemplate) error {
				return assertErr("boom")
			},
		}

		err := loadAgentTemplates(context.Background(), bs, map[string]interface{}{"sentryPeeringHost": "peering.example.com"}, cryptoutil.NoPassword)
		assert.Error(t, err)
	})
}

type assertErr string

func (e assertErr) Error() string { return string(e) }

func TestLoad(t *testing.T) {
	bs := &fakeBootstrapService{
		getBootstrapInfraFn: func(ctx context.Context, name string) (*sentry.BootstrapInfra, error) {
			return &sentry.BootstrapInfra{Spec: &sentry.BootstrapInfraSpec{CaCert: "existing-cert"}}, nil
		},
		patchBootstrapInfraFn: func(ctx context.Context, infra *sentry.BootstrapInfra) error {
			return nil
		},
		patchBootstrapAgentTemplateFn: func(ctx context.Context, template *sentry.BootstrapAgentTemplate) error {
			return nil
		},
	}

	err := Load(context.Background(), bs, map[string]interface{}{"sentryPeeringHost": "peering.example.com"}, cryptoutil.NoPassword)
	require.NoError(t, err)
	assert.NotNil(t, RelayTemplate)
	assert.NotNil(t, RelayAgentTemplate)
}
