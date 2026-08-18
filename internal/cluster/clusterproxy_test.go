package cluster

import (
	"strings"
	"testing"

	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/stretchr/testify/assert"
)

func TestUpdateProxyData(t *testing.T) {
	cidr := map[string]string{"PodNetworkCidr": "10.0.0.0/16", "ServiceCidr": "10.1.0.0/16"}

	t.Run("includes http/https proxy vars when set", func(t *testing.T) {
		out := UpdateProxyData("existing-cert", infrav3.ProxyConfig{
			HttpProxy:  "http://proxy:8080",
			HttpsProxy: "https://proxy:8443",
			NoProxy:    "example.com",
		}, cidr)

		assert.True(t, strings.HasPrefix(out, "existing-cert"))
		assert.Contains(t, out, BEGIN_PROXY_DATA)
		assert.Contains(t, out, END_PROXY_DATA)
		assert.Contains(t, out, "export http_proxy=http://proxy:8080")
		assert.Contains(t, out, "export HTTP_PROXY=http://proxy:8080")
		assert.Contains(t, out, "export https_proxy=https://proxy:8443")
		assert.Contains(t, out, "export no_proxy="+NO_PROXY_PARALUS_DATA+",10.0.0.0/16,10.1.0.0/16,example.com")
	})

	t.Run("omits http/https proxy vars when unset", func(t *testing.T) {
		out := UpdateProxyData("", infrav3.ProxyConfig{}, cidr)
		assert.NotContains(t, out, "http_proxy")
		assert.NotContains(t, out, "https_proxy")
	})

	t.Run("strips a previous proxy data block from the existing cert", func(t *testing.T) {
		oldCert := "actual-cert-content\n" + BEGIN_PROXY_DATA + "\nstale data\n" + END_PROXY_DATA + "\n"
		out := UpdateProxyData(oldCert, infrav3.ProxyConfig{}, cidr)

		assert.True(t, strings.HasPrefix(out, "actual-cert-content\n"))
		assert.NotContains(t, out, "stale data")
		assert.Equal(t, 1, strings.Count(out, BEGIN_PROXY_DATA))
	})
}

func TestGetNoProxyDataString(t *testing.T) {
	cidr := map[string]string{"PodNetworkCidr": "10.0.0.0/16", "ServiceCidr": "10.1.0.0/16"}

	got := GetNoProxyDataString("custom.example.com", cidr)
	assert.Equal(t, NO_PROXY_PARALUS_DATA+",10.0.0.0/16,10.1.0.0/16,custom.example.com", got)
}
