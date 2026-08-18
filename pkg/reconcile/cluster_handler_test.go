package reconcile

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paralus/paralus/pkg/event"
	"github.com/paralus/paralus/pkg/query"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHandler(cs *fakeClusterService) *clusterEventHandler {
	return NewClusterEventHandler(cs, nil, nil, nil).(*clusterEventHandler)
}

func TestClusterHooks_EnqueueEvents(t *testing.T) {
	h := newTestHandler(&fakeClusterService{})
	r := event.Resource{Name: "r1"}

	t.Run("ClusterHook adds to the cluster queue", func(t *testing.T) {
		h.ClusterHook().OnChange(r)
		item, shutdown := h.cwq.Get()
		require.False(t, shutdown)
		defer h.cwq.Done(item)
		assert.Equal(t, resourceToKey(r), item)
	})

	t.Run("AntiEntropyHook adds to the cluster queue", func(t *testing.T) {
		h.AntiEntropyHook().OnChange(r)
		item, shutdown := h.cwq.Get()
		require.False(t, shutdown)
		defer h.cwq.Done(item)
		assert.Equal(t, resourceToKey(r), item)
	})

	t.Run("ClusterWorkloadHook adds to the workload queue", func(t *testing.T) {
		h.ClusterWorkloadHook().OnChange(r)
		item, shutdown := h.wwq.Get()
		require.False(t, shutdown)
		defer h.wwq.Done(item)
		assert.Equal(t, resourceToKey(r), item)
	})
}

func TestProcessNextCluster(t *testing.T) {
	t.Run("processes an item and logs+returns on a lookup error, without panicking", func(t *testing.T) {
		cs := &fakeClusterService{
			getFn: func(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error) {
				return nil, errors.New("not found")
			},
		}
		h := newTestHandler(cs)
		h.cwq.Add(resourceToKey(event.Resource{Name: "r1"}))

		done := make(chan bool, 1)
		go func() { done <- h.processNextCluster() }()

		select {
		case ok := <-done:
			assert.True(t, ok)
		case <-time.After(2 * time.Second):
			t.Fatal("processNextCluster did not return in time")
		}
	})

	t.Run("returns false when the queue is shut down", func(t *testing.T) {
		h := newTestHandler(&fakeClusterService{})
		h.cwq.ShutDown()
		assert.False(t, h.processNextCluster())
	})
}

func TestProcessNextClusterWorkload(t *testing.T) {
	t.Run("processes an item and logs+returns on a lookup error, without panicking", func(t *testing.T) {
		cs := &fakeClusterService{
			getFn: func(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error) {
				return nil, errors.New("not found")
			},
		}
		h := newTestHandler(cs)
		h.wwq.Add(resourceToKey(event.Resource{Name: "r1"}))

		done := make(chan bool, 1)
		go func() { done <- h.processNextClusterWorkload() }()

		select {
		case ok := <-done:
			assert.True(t, ok)
		case <-time.After(2 * time.Second):
			t.Fatal("processNextClusterWorkload did not return in time")
		}
	})

	t.Run("returns false when the queue is shut down", func(t *testing.T) {
		h := newTestHandler(&fakeClusterService{})
		h.wwq.ShutDown()
		assert.False(t, h.processNextClusterWorkload())
	})
}

func TestHandle_StopsCleanly(t *testing.T) {
	h := newTestHandler(&fakeClusterService{})

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		h.Handle(stop)
		close(done)
	}()

	close(stop)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Handle did not return after stop was closed")
	}
}
