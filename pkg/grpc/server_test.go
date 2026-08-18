package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

func TestNewServer(t *testing.T) {
	srv, err := NewServer()
	require.NoError(t, err)
	assert.NotNil(t, srv)
}

func TestNewSecureServerWithPEM(t *testing.T) {
	caCert, _, _ := genCertKeyPair(t, "test-ca", nil, true)
	cert, key, _ := genCertKeyPair(t, "server", nil, false)

	t.Run("succeeds with a valid cert/key/ca", func(t *testing.T) {
		srv, err := NewSecureServerWithPEM(cert, key, caCert)
		require.NoError(t, err)
		assert.NotNil(t, srv)
	})

	t.Run("errors on mismatched cert/key pair", func(t *testing.T) {
		_, otherKey, _ := genCertKeyPair(t, "other", nil, false)
		_, err := NewSecureServerWithPEM(cert, otherKey, caCert)
		assert.Error(t, err)
	})

	t.Run("errors on invalid ca PEM", func(t *testing.T) {
		_, err := NewSecureServerWithPEM(cert, key, []byte("not-a-cert"))
		assert.Error(t, err)
	})
}

func TestNewSecureServer(t *testing.T) {
	caCert, _, _ := genCertKeyPair(t, "test-ca", nil, true)
	cert, key, _ := genCertKeyPair(t, "server", nil, false)

	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	caPath := filepath.Join(dir, "ca.pem")
	require.NoError(t, os.WriteFile(certPath, cert, 0o600))
	require.NoError(t, os.WriteFile(keyPath, key, 0o600))
	require.NoError(t, os.WriteFile(caPath, caCert, 0o600))

	t.Run("succeeds with valid file paths", func(t *testing.T) {
		srv, err := NewSecureServer(certPath, keyPath, caPath)
		require.NoError(t, err)
		assert.NotNil(t, srv)
	})

	t.Run("errors when cert file is missing", func(t *testing.T) {
		_, err := NewSecureServer(filepath.Join(dir, "missing.pem"), keyPath, caPath)
		assert.Error(t, err)
	})

	t.Run("errors when ca file is missing", func(t *testing.T) {
		_, err := NewSecureServer(certPath, keyPath, filepath.Join(dir, "missing-ca.pem"))
		assert.Error(t, err)
	})

	t.Run("errors on invalid ca PEM content", func(t *testing.T) {
		badCaPath := filepath.Join(dir, "bad-ca.pem")
		require.NoError(t, os.WriteFile(badCaPath, []byte("not-a-cert"), 0o600))
		_, err := NewSecureServer(certPath, keyPath, badCaPath)
		assert.Error(t, err)
	})
}

func peerContextWithCert(t *testing.T, cn string, ou []string) context.Context {
	t.Helper()
	_, _, cert := genCertKeyPair(t, cn, ou, false)
	p := &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}},
		},
	}
	return peer.NewContext(context.Background(), p)
}

func TestGetClientName(t *testing.T) {
	t.Run("returns the CommonName from the peer cert", func(t *testing.T) {
		ctx := peerContextWithCert(t, "client-1", nil)
		name, err := GetClientName(ctx)
		require.NoError(t, err)
		assert.Equal(t, "client-1", name)
	})

	t.Run("errors when no peer is present", func(t *testing.T) {
		_, err := GetClientName(context.Background())
		assert.ErrorIs(t, err, ErrInvalidClient)
	})
}

func TestGetClientOU(t *testing.T) {
	t.Run("returns the joined OrganizationalUnit from the peer cert", func(t *testing.T) {
		ctx := peerContextWithCert(t, "client-1", []string{"eng", "platform"})
		ou, err := GetClientOU(ctx)
		require.NoError(t, err)
		assert.Equal(t, "eng-platform", ou)
	})

	t.Run("errors when the cert has no OrganizationalUnit", func(t *testing.T) {
		ctx := peerContextWithCert(t, "client-1", nil)
		_, err := GetClientOU(ctx)
		assert.ErrorIs(t, err, ErrInvalidClient)
	})

	t.Run("errors when no peer is present", func(t *testing.T) {
		_, err := GetClientOU(context.Background())
		assert.ErrorIs(t, err, ErrInvalidClient)
	})
}
