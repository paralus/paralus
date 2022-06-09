package cluster

import (
	"bytes"
	"fmt"
	"strings"

	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
)

const (
	BEGIN_PROXY_DATA      = "-----BEGIN PROXY DATA-----"
	END_PROXY_DATA        = "-----END PROXY DATA-----"
	NO_PROXY_PARALUS_DATA = "localhost,127.0.0.1,127.0.0.2,k8master.service.consul,ingress-nginx-controller-admission.paralus-system.svc,paralus-drift.paralus-system.svc,secretstore-webhook.paralus-system.svc"
)

func UpdateProxyData(cert string, proxyConfig infrav3.ProxyConfig, clusterCidr map[string]string) string {
	b := new(bytes.Buffer)
	fmt.Fprintf(b, "\n")
	fmt.Fprintf(b, BEGIN_PROXY_DATA+"\n")
	fmt.Fprintf(b, "export %s=%s\n", "no_proxy", NO_PROXY_PARALUS_DATA+","+clusterCidr["PodNetworkCidr"]+","+clusterCidr["ServiceCidr"]+","+proxyConfig.NoProxy)
	fmt.Fprintf(b, "export %s=%s\n", "NO_PROXY", NO_PROXY_PARALUS_DATA+","+clusterCidr["PodNetworkCidr"]+","+clusterCidr["ServiceCidr"]+","+proxyConfig.NoProxy)
	if proxyConfig.HttpProxy != "" {
		fmt.Fprintf(b, "export %s=%s\n", "http_proxy", proxyConfig.HttpProxy)
		fmt.Fprintf(b, "export %s=%s\n", "HTTP_PROXY", proxyConfig.HttpProxy)
	}
	if proxyConfig.HttpsProxy != "" {
		fmt.Fprintf(b, "export %s=%s\n", "https_proxy", proxyConfig.HttpsProxy)
		fmt.Fprintf(b, "export %s=%s\n", "HTTPS_PROXY", proxyConfig.HttpsProxy)
	}
	fmt.Fprintf(b, END_PROXY_DATA+"\n")

	if strings.Contains(cert, BEGIN_PROXY_DATA) {
		cert = strings.Split(cert, BEGIN_PROXY_DATA)[0]
	}
	return cert + b.String()
}

func GetNoProxyDataString(noProxyConfig string, clusterCidr map[string]string) string {
	return NO_PROXY_PARALUS_DATA + "," + clusterCidr["PodNetworkCidr"] + "," + clusterCidr["ServiceCidr"] + "," + noProxyConfig
}
