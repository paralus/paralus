package cryptoutil

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
)

func TestVerifyCertHostname(t *testing.T) {
	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "---",
		},
		DNSNames: []string{
			"peering.sentry.paralus.local",
			"peering.sentry.paralus.local",
			"paralus-sentry",
			"paralus-sentry.paralus-system",
			"paralus-sentry.paralus-system.cluster.local",
		},
	}
	err := cert.VerifyHostname("peering.sentry.paralus.local")
	if err != nil {
		t.Error(err)
		return
	}
	err = cert.VerifyHostname("paralus-sentry.paralus-system")
	if err != nil {
		t.Error(err)
		return
	}
	err = cert.VerifyHostname("paralus-sentry")
	if err != nil {
		t.Error(err)
		return
	}
	err = cert.VerifyHostname("paralus-sentry.paralus-system.cluster.local")
	if err != nil {
		t.Error(err)
		return
	}
}
