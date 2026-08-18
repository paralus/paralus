package pool

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestNewGRPCPool(t *testing.T) {
	t.Run("enforces the minimum pool capacity", func(t *testing.T) {
		p := NewGRPCPool("addr:1234", 1, nil)
		assert.Equal(t, DefaultMaxPoolConn, p.capacity)
	})

	t.Run("keeps a capacity above the minimum", func(t *testing.T) {
		p := NewGRPCPool("addr:1234", 50, nil)
		assert.Equal(t, 50, p.capacity)
	})

	t.Run("keeps the exact minimum", func(t *testing.T) {
		p := NewGRPCPool("addr:1234", DefaultMaxPoolConn, nil)
		assert.Equal(t, DefaultMaxPoolConn, p.capacity)
	})
}

// startLocalGRPCServer starts a minimal insecure gRPC server on an
// OS-assigned local port so GetConnection has a fast, real dial target
// instead of a real network endpoint (which could take up to the library's
// 30s dial timeout to fail against an unreachable address).
func startLocalGRPCServer(t *testing.T) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	return lis.Addr().String()
}

func TestGRPCPool_GetConnection(t *testing.T) {
	t.Run("connects and lazily initializes the underlying pool", func(t *testing.T) {
		addr := startLocalGRPCServer(t)
		p := NewGRPCPool(addr, 1, nil)

		cc, err := p.GetConnection(context.Background())
		require.NoError(t, err)
		require.NotNil(t, cc)
		defer cc.Close()

		assert.NotNil(t, p.Pool, "pool should be lazily created on first GetConnection call")
	})

	t.Run("reuses the pool across calls", func(t *testing.T) {
		addr := startLocalGRPCServer(t)
		p := NewGRPCPool(addr, 1, nil)

		cc1, err := p.GetConnection(context.Background())
		require.NoError(t, err)
		defer cc1.Close()
		firstPool := p.Pool

		cc2, err := p.GetConnection(context.Background())
		require.NoError(t, err)
		defer cc2.Close()

		assert.Same(t, firstPool, p.Pool)
	})
}

// newSecurePool's dial factory (see pool.go) opens its own hardcoded 30s
// context.WithTimeout(context.Background(), ...) independent of whatever
// context callers pass to GetConnection, so a TLS-handshake-mismatch
// failure genuinely blocks for the full 30s no matter what deadline is set
// here. That makes the secure-pool error path impractical to unit test
// without a slow (30s+) test, so it's intentionally left uncovered here
// rather than added as a slow test.
