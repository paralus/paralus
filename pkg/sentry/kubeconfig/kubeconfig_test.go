package kubeconfig

import (
	"crypto/x509/pkix"
	"testing"
	"time"

	"github.com/paralus/paralus/pkg/sentry/cryptoutil"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserCN(t *testing.T) {
	t.Run("sorts keys and joins as key=value pairs", func(t *testing.T) {
		cn := GetUserCN(map[string]string{"b": "2", "a": "1"})
		assert.Equal(t, "a=1/b=2", cn)
	})

	t.Run("empty map yields empty string", func(t *testing.T) {
		assert.Equal(t, "", GetUserCN(map[string]string{}))
	})
}

func TestGetUserAttrs(t *testing.T) {
	t.Run("parses key=value pairs", func(t *testing.T) {
		attrs := GetUserAttrs("a=1/b=2")
		assert.Equal(t, map[string]string{"a": "1", "b": "2"}, attrs)
	})

	t.Run("ignores malformed segments", func(t *testing.T) {
		attrs := GetUserAttrs("a=1/malformed/b=2")
		assert.Equal(t, map[string]string{"a": "1", "b": "2"}, attrs)
	})

	t.Run("GetUserCN and GetUserAttrs round trip", func(t *testing.T) {
		original := map[string]string{"a": "1", "b": "2", "c": "3"}
		assert.Equal(t, original, GetUserAttrs(GetUserCN(original)))
	})
}

func testBootstrapInfra(t *testing.T) *sentry.BootstrapInfra {
	t.Helper()
	cert, key, err := cryptoutil.GenerateCA(pkix.Name{CommonName: "test-ca"}, cryptoutil.NoPassword)
	require.NoError(t, err)
	return &sentry.BootstrapInfra{
		Spec: &sentry.BootstrapInfraSpec{CaCert: string(cert), CaKey: string(key)},
	}
}

func testBootstrapAgent(name, displayName, templateRef string) *sentry.BootstrapAgent {
	return &sentry.BootstrapAgent{
		Metadata: &commonv3.Metadata{Name: name, DisplayName: displayName},
		Spec:     &sentry.BootstrapAgentSpec{TemplateRef: templateRef},
	}
}

func TestGetConfig(t *testing.T) {
	infra := testBootstrapInfra(t)

	t.Run("builds a config for a matching relay agent", func(t *testing.T) {
		agents := []*sentry.BootstrapAgent{
			testBootstrapAgent("agent-1", "My Cluster", "paralus-core-relay-agent"),
		}
		cfg, err := getConfig("bob", "", "u=bob", "*.relay.example.com", infra, agents, cryptoutil.NoPassword, time.Hour, "")
		require.NoError(t, err)

		require.Len(t, cfg.Clusters, 1)
		assert.Equal(t, "My Cluster", cfg.Clusters[0].Name)
		assert.Equal(t, "https://agent-1.relay.example.com", cfg.Clusters[0].Cluster.Server)
		require.Len(t, cfg.Contexts, 1)
		assert.Equal(t, "My Cluster", cfg.Contexts[0].Context.Cluster)
		assert.Equal(t, "default", cfg.Contexts[0].Context.Namespace)
		assert.Equal(t, "My Cluster", cfg.CurrentContext) // defaults to the first context
		require.Len(t, cfg.AuthInfos, 1)
		assert.Equal(t, "bob", cfg.AuthInfos[0].Name)
		assert.NotEmpty(t, cfg.AuthInfos[0].AuthInfo.ClientCertificateData)
		assert.NotEmpty(t, cfg.AuthInfos[0].AuthInfo.ClientKeyData)
	})

	t.Run("skips agents whose template is not a default relay", func(t *testing.T) {
		agents := []*sentry.BootstrapAgent{
			testBootstrapAgent("agent-1", "Other Cluster", "some-other-template"),
		}
		cfg, err := getConfig("bob", "", "u=bob", "*.relay.example.com", infra, agents, cryptoutil.NoPassword, time.Hour, "")
		require.NoError(t, err)
		assert.Empty(t, cfg.Clusters)
		assert.Empty(t, cfg.Contexts)
		assert.Empty(t, cfg.CurrentContext)
	})

	t.Run("also accepts the cd-relay-agent template", func(t *testing.T) {
		agents := []*sentry.BootstrapAgent{
			testBootstrapAgent("agent-1", "CD Cluster", "paralus-core-cd-relay-agent"),
		}
		cfg, err := getConfig("bob", "", "u=bob", "*.relay.example.com", infra, agents, cryptoutil.NoPassword, time.Hour, "")
		require.NoError(t, err)
		require.Len(t, cfg.Clusters, 1)
	})

	t.Run("uses the given namespace when set", func(t *testing.T) {
		agents := []*sentry.BootstrapAgent{
			testBootstrapAgent("agent-1", "My Cluster", "paralus-core-relay-agent"),
		}
		cfg, err := getConfig("bob", "custom-ns", "u=bob", "*.relay.example.com", infra, agents, cryptoutil.NoPassword, time.Hour, "")
		require.NoError(t, err)
		require.Len(t, cfg.Contexts, 1)
		assert.Equal(t, "custom-ns", cfg.Contexts[0].Context.Namespace)
	})

	t.Run("sanitizes the username for the AuthInfo name", func(t *testing.T) {
		cfg, err := getConfig("Bob@Example.com", "", "u=bob", "*.relay.example.com", infra, nil, cryptoutil.NoPassword, time.Hour, "")
		require.NoError(t, err)
		require.Len(t, cfg.AuthInfos, 1)
		assert.NotContains(t, cfg.AuthInfos[0].Name, "@")
	})

	t.Run("prefers an explicit non-empty clusterName as CurrentContext", func(t *testing.T) {
		agents := []*sentry.BootstrapAgent{
			testBootstrapAgent("agent-1", "cluster-a", "paralus-core-relay-agent"),
			testBootstrapAgent("agent-2", "cluster-b", "paralus-core-relay-agent"),
		}
		cfg, err := getConfig("bob", "", "u=bob", "*.relay.example.com", infra, agents, cryptoutil.NoPassword, time.Hour, "  cluster-b  ")
		require.NoError(t, err)
		assert.Equal(t, "cluster-b", cfg.CurrentContext)
	})

	t.Run("propagates a signer error from an invalid CA", func(t *testing.T) {
		badInfra := &sentry.BootstrapInfra{Spec: &sentry.BootstrapInfraSpec{CaCert: "not-a-cert", CaKey: "not-a-key"}}
		_, err := getConfig("bob", "", "u=bob", "*.relay.example.com", badInfra, nil, cryptoutil.NoPassword, time.Hour, "")
		assert.Error(t, err)
	})
}
