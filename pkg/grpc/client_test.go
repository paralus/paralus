package grpc

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	grpclib "google.golang.org/grpc"
)

func TestNewSecureClientConn(t *testing.T) {
	t.Run("fails fast on an already-cancelled context", func(t *testing.T) {
		caCert, _, _ := genCertKeyPair(t, "test-ca", nil, true)
		cert, key, _ := genCertKeyPair(t, "client", nil, false)
		creds, err := NewClientTransportCredentials(cert, key, caCert, "localhost:443")
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = NewSecureClientConn(ctx, "127.0.0.1:1", creds)
		assert.Error(t, err)
	})
}

func TestNewClientTransportCredentials(t *testing.T) {
	caCert, _, _ := genCertKeyPair(t, "test-ca", nil, true)
	cert, key, _ := genCertKeyPair(t, "client", nil, false)

	t.Run("succeeds with a valid cert/key/ca and addr", func(t *testing.T) {
		creds, err := NewClientTransportCredentials(cert, key, caCert, "localhost:443")
		require.NoError(t, err)
		assert.NotNil(t, creds)
	})

	t.Run("succeeds without a CA (falls back to InsecureSkipVerify)", func(t *testing.T) {
		creds, err := NewClientTransportCredentials(cert, key, nil, "localhost:443")
		require.NoError(t, err)
		assert.NotNil(t, creds)
	})

	t.Run("errors on mismatched cert/key pair", func(t *testing.T) {
		_, otherKey, _ := genCertKeyPair(t, "other", nil, false)
		_, err := NewClientTransportCredentials(cert, otherKey, caCert, "localhost:443")
		assert.Error(t, err)
	})

	// BUG: newClientTLSConfig's CA-parse failure branch does `return nil, err`
	// where `err` is still nil from the earlier (successful) tls.X509KeyPair
	// call, instead of constructing a new error. So an invalid CA PEM is
	// silently swallowed: no error is returned, but the resulting config
	// still ends up with an empty (non-nil) RootCAs pool and
	// InsecureSkipVerify: false, which would reject every real server cert
	// with no indication of why.
	t.Run("BUG: invalid ca PEM is silently swallowed, no error returned", func(t *testing.T) {
		creds, err := NewClientTransportCredentials(cert, key, []byte("not-a-cert"), "localhost:443")
		assert.NoError(t, err)
		assert.NotNil(t, creds) // returns a usable-looking but effectively broken credentials object
	})

	t.Run("errors when addr has no port", func(t *testing.T) {
		_, err := NewClientTransportCredentials(cert, key, caCert, "no-port-here")
		assert.Error(t, err)
	})
}

// startLocalGRPCServer starts a minimal insecure gRPC server on an
// OS-assigned local port and returns its address, for use as a fast
// same-process dial target instead of a real network endpoint.
func startLocalGRPCServer(t *testing.T) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpclib.NewServer()
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	return lis.Addr().String()
}

func TestNewClientConn(t *testing.T) {
	t.Run("connects to a real listener", func(t *testing.T) {
		addr := startLocalGRPCServer(t)

		cc, err := NewClientConn(context.Background(), addr)
		require.NoError(t, err)
		defer cc.Close()
		assert.NotNil(t, cc)
	})

	t.Run("fails fast on an already-cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := NewClientConn(ctx, "127.0.0.1:1")
		assert.Error(t, err)
	})
}

func TestNewGrpcClientClientConn(t *testing.T) {
	t.Run("connects to a real listener", func(t *testing.T) {
		addr := startLocalGRPCServer(t)
		host, port := splitHostPortForTest(t, addr)

		cc, err := NewGrpcClientClientConn(context.Background(), host, port)
		require.NoError(t, err)
		defer cc.Close()
		assert.NotNil(t, cc)
	})
}

func TestNewGrpcClientClientConnWithTimeout(t *testing.T) {
	t.Run("connects to a real listener", func(t *testing.T) {
		addr := startLocalGRPCServer(t)
		host, port := splitHostPortForTest(t, addr)

		cc, err := NewGrpcClientClientConnWithTimeout(context.Background(), host, port, time.Minute)
		require.NoError(t, err)
		defer cc.Close()
		assert.NotNil(t, cc)
	})
}

func splitHostPortForTest(t *testing.T, addr string) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)
	return host, port
}
