package peering

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// genCertKeyFiles generates a minimal self-signed cert/key pair and writes
// them (plus, optionally, a separate CA file) to temp files for use with
// ClientTLSConfig, which takes file paths rather than PEM bytes.
func genCertKeyFiles(t *testing.T) (certPath, keyPath, caPath string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	dir := t.TempDir()
	certPath = filepath.Join(dir, "cert.pem")
	keyPath = filepath.Join(dir, "key.pem")
	caPath = filepath.Join(dir, "ca.pem")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	require.NoError(t, os.WriteFile(certPath, certPEM, 0o600))
	require.NoError(t, os.WriteFile(keyPath, keyPEM, 0o600))
	require.NoError(t, os.WriteFile(caPath, certPEM, 0o600))

	return certPath, keyPath, caPath
}

func TestClientTLSConfig(t *testing.T) {
	certPath, keyPath, caPath := genCertKeyFiles(t)

	t.Run("succeeds with a root CA", func(t *testing.T) {
		cfg, err := ClientTLSConfig(certPath, keyPath, caPath, "localhost:443")
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "localhost", cfg.ServerName)
		assert.False(t, cfg.InsecureSkipVerify)
		assert.NotNil(t, cfg.RootCAs)
	})

	t.Run("falls back to InsecureSkipVerify without a root CA", func(t *testing.T) {
		cfg, err := ClientTLSConfig(certPath, keyPath, "", "localhost:443")
		require.NoError(t, err)
		assert.True(t, cfg.InsecureSkipVerify)
		assert.Nil(t, cfg.RootCAs)
	})

	t.Run("errors when the cert file is missing", func(t *testing.T) {
		_, err := ClientTLSConfig(filepath.Join(t.TempDir(), "missing.pem"), keyPath, caPath, "localhost:443")
		assert.Error(t, err)
	})

	t.Run("errors when addr has no port", func(t *testing.T) {
		_, err := ClientTLSConfig(certPath, keyPath, caPath, "no-port-here")
		assert.Error(t, err)
	})

	// BUG: mirrors pkg/grpc's newClientTLSConfig -- when AppendCertsFromPEM
	// fails, the function does `return nil, err` where `err` is still nil
	// from the earlier successful tls.LoadX509KeyPair call, so an invalid CA
	// file's content is silently swallowed instead of surfacing an error.
	t.Run("BUG: invalid ca file content is silently swallowed", func(t *testing.T) {
		dir := t.TempDir()
		badCaPath := filepath.Join(dir, "bad-ca.pem")
		require.NoError(t, os.WriteFile(badCaPath, []byte("not-a-cert"), 0o600))

		cfg, err := ClientTLSConfig(certPath, keyPath, badCaPath, "localhost:443")
		assert.NoError(t, err)
		assert.Nil(t, cfg)
	})
}
