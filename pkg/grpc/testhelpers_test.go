package grpc

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// genCertKeyPair generates a minimal self-signed PEM-encoded cert/key pair
// for the given CommonName/OrganizationalUnit, for use as either a CA or a
// leaf certificate in tests -- no real CA infra needed.
func genCertKeyPair(t *testing.T, commonName string, ou []string, isCA bool) (certPEM, keyPEM []byte, cert *x509.Certificate) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:         commonName,
			OrganizationalUnit: ou,
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
		DNSNames:              []string{"localhost"},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	certBuf := new(bytes.Buffer)
	require.NoError(t, pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: der}))

	keyBuf := new(bytes.Buffer)
	require.NoError(t, pem.Encode(keyBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))

	parsed, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return certBuf.Bytes(), keyBuf.Bytes(), parsed
}
