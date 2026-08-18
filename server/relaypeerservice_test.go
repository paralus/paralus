package server

// Only the plain data-structure and cache logic in relaypeerservice.go is
// unit tested here: the relay registry (RelayMap) bookkeeping
// (updateRelayIfExist/getRelayObject/putRelayObject/insertRelayObject/
// handleHelloRequest), the peer service cache helpers, peerServiceCacheKey,
// and the NewRelayPeerService constructor.
//
// Deliberately NOT unit tested (left to integration/e2e coverage per the
// Test Pyramid): RelayPeerHelloRPC, RelayPeerProbeRPC, RelayPeerSurveyRPC,
// relayPeerProbeSender, relayPeerSurveySender, tryResponseFromCache,
// RunRelaySurveyHandler, and handleSurveyReq. These drive live
// bidirectional gRPC streams (sentryrpc.RelayPeerService_*RPCServer),
// extract client identity from TLS certificates via
// grpc.GetClientName/GetClientOU, and (in handleSurveyReq's case) contain
// real multi-second time.Sleep/time.Ticker retry loops. Faking the stream
// interfaces well enough to exercise them would mostly test the fakes'
// Send/Recv sequencing rather than this file's logic, and the sleep-based
// retry loop would make the suite slow and flaky, so they are out of scope
// for this pass.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeerServiceCacheKey(t *testing.T) {
	assert.Equal(t, "sni1uuid1ou1", peerServiceCacheKey("sni1", "uuid1", "ou1"))
	assert.Equal(t, "", peerServiceCacheKey("", "", ""))
}

func TestNewRelayPeerService(t *testing.T) {
	svc, err := NewRelayPeerService()
	require.NoError(t, err)
	require.NotNil(t, svc)

	s, ok := svc.(*relayPeerService)
	require.True(t, ok)
	assert.NotEmpty(t, s.ServiceUUID)
	assert.NotNil(t, s.RelayMap)
	assert.NotNil(t, s.surveyBroadCast)
	assert.Equal(t, 60*time.Second, s.surveyCacheExpiry)
	assert.NotNil(t, s.peerServiceCache)
}

func newTestRelayPeerService(t *testing.T) *relayPeerService {
	t.Helper()
	svc, err := NewRelayPeerService()
	require.NoError(t, err)
	return svc.(*relayPeerService)
}

func TestRelayPeerService_PeerServiceCache(t *testing.T) {
	s := newTestRelayPeerService(t)

	t.Run("getPeerServiceCache reports not found for a missing key", func(t *testing.T) {
		ip, found := s.getPeerServiceCache("missing")
		assert.False(t, found)
		assert.Equal(t, "", ip)
	})

	t.Run("insertPeerServiceCache followed by getPeerServiceCache returns the stored ip", func(t *testing.T) {
		ok := s.insertPeerServiceCache("key1", "10.0.0.5")
		require.True(t, ok)
		s.peerServiceCache.Wait()

		ip, found := s.getPeerServiceCache("key1")
		require.True(t, found)
		assert.Equal(t, "10.0.0.5", ip)
	})
}

func TestRelayPeerService_RelayRegistry(t *testing.T) {
	t.Run("updateRelayIfExist returns false when the relay is not registered", func(t *testing.T) {
		s := newTestRelayPeerService(t)
		assert.False(t, s.updateRelayIfExist("relay1", "ou1"))
	})

	t.Run("insertRelayObject then updateRelayIfExist refreshes the timestamp", func(t *testing.T) {
		s := newTestRelayPeerService(t)
		robj := &relayObject{timeStamp: 1, ou: "ou1", relayip: "10.0.0.1"}
		s.insertRelayObject(robj, "relay1", "ou1")

		assert.True(t, s.updateRelayIfExist("relay1", "ou1"))
		assert.Greater(t, robj.timeStamp, int64(1))
	})

	t.Run("updateRelayIfExist returns false for a mismatched ou", func(t *testing.T) {
		s := newTestRelayPeerService(t)
		robj := &relayObject{timeStamp: 1, ou: "ou1"}
		s.insertRelayObject(robj, "relay1", "ou1")

		assert.False(t, s.updateRelayIfExist("relay1", "different-ou"))
	})

	t.Run("insertRelayObject creates a fresh per-ou map on first insert", func(t *testing.T) {
		s := newTestRelayPeerService(t)
		robj := &relayObject{ou: "ou1"}
		s.insertRelayObject(robj, "relay1", "ou1")

		require.Contains(t, s.RelayMap, "ou1")
		assert.Same(t, robj, s.RelayMap["ou1"]["relay1"])
	})

	t.Run("insertRelayObject adds to an existing per-ou map on subsequent inserts", func(t *testing.T) {
		s := newTestRelayPeerService(t)
		robj1 := &relayObject{ou: "ou1"}
		robj2 := &relayObject{ou: "ou1"}
		s.insertRelayObject(robj1, "relay1", "ou1")
		s.insertRelayObject(robj2, "relay2", "ou1")

		assert.Len(t, s.RelayMap["ou1"], 2)
		assert.Same(t, robj1, s.RelayMap["ou1"]["relay1"])
		assert.Same(t, robj2, s.RelayMap["ou1"]["relay2"])
	})

	t.Run("getRelayObject increments refCnt and putRelayObject decrements it", func(t *testing.T) {
		s := newTestRelayPeerService(t)
		robj := &relayObject{ou: "ou1"}
		s.insertRelayObject(robj, "relay1", "ou1")

		got := s.getRelayObject("relay1", "ou1")
		require.NotNil(t, got)
		assert.Same(t, robj, got)
		assert.EqualValues(t, 1, got.refCnt)

		s.putRelayObject("relay1", "ou1")
		assert.EqualValues(t, 0, robj.refCnt)
	})

	t.Run("getRelayObject returns nil for an unknown relay or mismatched ou", func(t *testing.T) {
		s := newTestRelayPeerService(t)
		robj := &relayObject{ou: "ou1"}
		s.insertRelayObject(robj, "relay1", "ou1")

		assert.Nil(t, s.getRelayObject("unknown-relay", "ou1"))
		assert.Nil(t, s.getRelayObject("relay1", "different-ou"))
	})

	t.Run("putRelayObject is a no-op when refCnt is already zero", func(t *testing.T) {
		s := newTestRelayPeerService(t)
		robj := &relayObject{ou: "ou1", refCnt: 0}
		s.insertRelayObject(robj, "relay1", "ou1")

		s.putRelayObject("relay1", "ou1")
		assert.EqualValues(t, 0, robj.refCnt)
	})
}

func TestRelayPeerService_HandleHelloRequest(t *testing.T) {
	t.Run("registers a brand-new relay with fresh channels", func(t *testing.T) {
		s := newTestRelayPeerService(t)

		s.handleHelloRequest("relay1", "10.0.0.1", "ou1")

		require.Contains(t, s.RelayMap, "ou1")
		robj := s.RelayMap["ou1"]["relay1"]
		require.NotNil(t, robj)
		assert.Equal(t, "10.0.0.1", robj.relayip)
		assert.Equal(t, "ou1", robj.ou)
		assert.NotNil(t, robj.probeReplyChnl)
		assert.NotNil(t, robj.surveyRequestChnl)
	})

	t.Run("just refreshes the timestamp for an already-registered relay instead of replacing it", func(t *testing.T) {
		s := newTestRelayPeerService(t)
		original := &relayObject{timeStamp: 1, ou: "ou1", relayip: "10.0.0.1"}
		s.insertRelayObject(original, "relay1", "ou1")

		s.handleHelloRequest("relay1", "10.0.0.2", "ou1")

		assert.Same(t, original, s.RelayMap["ou1"]["relay1"])
		assert.Greater(t, original.timeStamp, int64(1))
		// The relay object is not replaced, so the (possibly stale) IP
		// from the first hello is retained rather than updated to the new
		// one reported in this hello.
		assert.Equal(t, "10.0.0.1", original.relayip)
	})
}

func TestGetServiceIP(t *testing.T) {
	// getServiceIP swallows any hostname/lookup error and returns "" in
	// that case, so the only thing we can assert unconditionally in a unit
	// test (without controlling the host's DNS/hosts file) is that it
	// never panics and always returns a string.
	_ = getServiceIP()
}
