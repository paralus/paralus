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
			"peering.sentry.rafay.local",
			"peering.sentry.rafay.local",
			"rafay-sentry",
			"rafay-sentry.rafay-system",
			"rafay-sentry.rafay-system.cluster.local",
		},
	}
	err := cert.VerifyHostname("peering.sentry.rafay.local")
	if err != nil {
		t.Error(err)
		return
	}
	err = cert.VerifyHostname("rafay-sentry.rafay-system")
	if err != nil {
		t.Error(err)
		return
	}
	err = cert.VerifyHostname("rafay-sentry")
	if err != nil {
		t.Error(err)
		return
	}
	err = cert.VerifyHostname("rafay-sentry.rafay-system.cluster.local")
	if err != nil {
		t.Error(err)
		return
	}
}
