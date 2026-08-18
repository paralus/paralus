package peering

import (
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCache(t *testing.T) *ristretto.Cache {
	t.Helper()
	cache, err := InitPeerCache(nil)
	require.NoError(t, err)
	return cache
}

func TestInitPeerCache(t *testing.T) {
	cache, err := InitPeerCache(nil)
	require.NoError(t, err)
	assert.NotNil(t, cache)
}

func TestInsertAndGetPeerCache(t *testing.T) {
	t.Run("round trip with a single entry", func(t *testing.T) {
		cache := newTestCache(t)
		value := []RelayClusterConnectionInfo{{Relayuuid: "u1", Relayip: "1.1.1.1"}}

		ok := InsertPeerCache(cache, time.Minute, "key1", value)
		require.True(t, ok)
		cache.Wait()

		ip, found := GetPeerCache(cache, "key1")
		require.True(t, found)
		assert.Equal(t, "1.1.1.1", ip)
	})

	t.Run("randomly selects among multiple entries", func(t *testing.T) {
		cache := newTestCache(t)
		value := []RelayClusterConnectionInfo{
			{Relayuuid: "u1", Relayip: "1.1.1.1"},
			{Relayuuid: "u2", Relayip: "2.2.2.2"},
		}
		ok := InsertPeerCache(cache, time.Minute, "key2", value)
		require.True(t, ok)
		cache.Wait()

		ip, found := GetPeerCache(cache, "key2")
		require.True(t, found)
		assert.Contains(t, []string{"1.1.1.1", "2.2.2.2"}, ip)
	})

	t.Run("returns not-found for a missing key", func(t *testing.T) {
		cache := newTestCache(t)
		_, found := GetPeerCache(cache, "missing")
		assert.False(t, found)
	})
}
