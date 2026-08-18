package server

// Only the pure, non-streaming logic in relaypeerclient.go is unit tested
// here: the peer cache helpers (InitPeerCache/InsertPeerCache/GetPeerCache)
// and ClientTLSConfig's config-building logic. The RPC loops
// (helloRPCSend, ClientHelloRPC, probeRPCSend, ClientProbeRPC,
// ClientSurveyRPC) drive a live bidirectional gRPC stream with goroutines,
// tickers, and context cancellation; faking relayrpc.RelayPeerService_*RPCClient
// stream interfaces to exercise them meaningfully would mostly test the
// fake's own Send/Recv sequencing rather than this file's logic, so per the
// Test Pyramid they are deliberately left to integration/e2e coverage.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
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

func TestPeerCache(t *testing.T) {
	t.Run("InsertPeerCache followed by GetPeerCache with a single item returns that item's IP", func(t *testing.T) {
		cache, err := InitPeerCache(nil)
		require.NoError(t, err)
		defer cache.Close()

		items := []RelayClusterConnectionInfo{{Relayuuid: "u1", Relayip: "10.0.0.1"}}
		ok := InsertPeerCache(cache, time.Minute, "cluster1", items)
		assert.True(t, ok)
		cache.Wait()

		ip, found := GetPeerCache(cache, "cluster1")
		assert.True(t, found)
		assert.Equal(t, "10.0.0.1", ip)
	})

	t.Run("GetPeerCache picks one of multiple cached items", func(t *testing.T) {
		cache, err := InitPeerCache(nil)
		require.NoError(t, err)
		defer cache.Close()

		items := []RelayClusterConnectionInfo{
			{Relayuuid: "u1", Relayip: "10.0.0.1"},
			{Relayuuid: "u2", Relayip: "10.0.0.2"},
		}
		ok := InsertPeerCache(cache, time.Minute, "cluster1", items)
		require.True(t, ok)
		cache.Wait()

		ip, found := GetPeerCache(cache, "cluster1")
		require.True(t, found)
		assert.Contains(t, []string{"10.0.0.1", "10.0.0.2"}, ip)
	})

	t.Run("GetPeerCache reports not found for a missing key", func(t *testing.T) {
		cache, err := InitPeerCache(nil)
		require.NoError(t, err)
		defer cache.Close()

		ip, found := GetPeerCache(cache, "missing")
		assert.False(t, found)
		assert.Equal(t, "", ip)
	})
}

// writeSelfSignedCert generates a throwaway self-signed EC cert/key pair
// under dir and returns the cert and key file paths.
func writeSelfSignedCert(t *testing.T, dir string) (certPath, keyPath string) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)

	certPath = filepath.Join(dir, "cert.pem")
	keyPath = filepath.Join(dir, "key.pem")

	certOut, err := os.Create(certPath)
	require.NoError(t, err)
	require.NoError(t, pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der}))
	require.NoError(t, certOut.Close())

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	require.NoError(t, err)
	keyOut, err := os.Create(keyPath)
	require.NoError(t, err)
	require.NoError(t, pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}))
	require.NoError(t, keyOut.Close())

	return certPath, keyPath
}

func TestClientTLSConfig(t *testing.T) {
	t.Run("returns an error when the cert/key pair cannot be loaded", func(t *testing.T) {
		_, err := ClientTLSConfig("/does/not/exist.crt", "/does/not/exist.key", "", "example.com:443")
		assert.Error(t, err)
	})

	t.Run("returns an error when rootCA cannot be read", func(t *testing.T) {
		dir := t.TempDir()
		certPath, keyPath := writeSelfSignedCert(t, dir)

		_, err := ClientTLSConfig(certPath, keyPath, "/does/not/exist-ca.pem", "example.com:443")
		assert.Error(t, err)
	})

	t.Run("returns an error when addr has no host:port", func(t *testing.T) {
		dir := t.TempDir()
		certPath, keyPath := writeSelfSignedCert(t, dir)

		_, err := ClientTLSConfig(certPath, keyPath, "", "not-a-host-port")
		assert.Error(t, err)
	})

	t.Run("builds a config with InsecureSkipVerify when no rootCA is given", func(t *testing.T) {
		dir := t.TempDir()
		certPath, keyPath := writeSelfSignedCert(t, dir)

		cfg, err := ClientTLSConfig(certPath, keyPath, "", "example.com:443")
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "example.com", cfg.ServerName)
		assert.True(t, cfg.InsecureSkipVerify)
		assert.Nil(t, cfg.RootCAs)
		assert.Len(t, cfg.Certificates, 1)
		assert.Equal(t, uint16(tls.VersionTLS12), cfg.MinVersion)
	})

	t.Run("builds a config with a root CA pool and verification enabled when rootCA is given", func(t *testing.T) {
		dir := t.TempDir()
		certPath, keyPath := writeSelfSignedCert(t, dir)
		// Reuse the leaf cert itself as the CA bundle; ClientTLSConfig only
		// cares that the PEM parses into a non-empty pool.
		rootCA := certPath

		cfg, err := ClientTLSConfig(certPath, keyPath, rootCA, "example.com:443")
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.False(t, cfg.InsecureSkipVerify)
		require.NotNil(t, cfg.RootCAs)
	})
}
