package notify

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/common"
	"github.com/paralus/paralus/pkg/event"
	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyFromMeta(t *testing.T) {
	key := KeyFromMeta(&commonv3.Metadata{Partner: "p1", Organization: "o1", Project: "proj1", Name: "n1"})
	assert.Equal(t, "p1/o1/proj1/n1", key)
}

func TestMetaFromKey(t *testing.T) {
	t.Run("parses a well-formed 4-segment key", func(t *testing.T) {
		meta := MetaFromKey("p1/o1/proj1/n1")
		assert.Equal(t, "p1", meta.Partner)
		assert.Equal(t, "o1", meta.Organization)
		assert.Equal(t, "proj1", meta.Project)
		assert.Equal(t, "n1", meta.Name)
	})

	t.Run("returns zero-value Metadata for a malformed key", func(t *testing.T) {
		meta := MetaFromKey("too/few/segments")
		assert.Equal(t, commonv3.Metadata{}, meta)
	})
}

// fakeClusterService implements service.ClusterService with only Get and
// ListenClusters wired; the rest panic since nothing here exercises them.
type fakeClusterService struct {
	getFn            func(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error)
	listenClustersFn func(ctx context.Context, mChan chan<- commonv3.Metadata)
}

func (f *fakeClusterService) Create(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {
	panic("not implemented")
}
func (f *fakeClusterService) Select(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error) {
	panic("not implemented")
}
func (f *fakeClusterService) Get(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error) {
	return f.getFn(ctx, opts...)
}
func (f *fakeClusterService) Update(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {
	panic("not implemented")
}
func (f *fakeClusterService) Delete(ctx context.Context, cluster *infrav3.Cluster) error {
	panic("not implemented")
}
func (f *fakeClusterService) List(ctx context.Context, opts ...query.Option) (*infrav3.ClusterList, error) {
	panic("not implemented")
}
func (f *fakeClusterService) UpdateClusterConditionStatus(ctx context.Context, current *infrav3.Cluster) error {
	panic("not implemented")
}
func (f *fakeClusterService) UpdateClusterAnnotations(ctx context.Context, cluster *infrav3.Cluster) error {
	panic("not implemented")
}
func (f *fakeClusterService) ListenClusters(ctx context.Context, mChan chan<- commonv3.Metadata) {
	f.listenClustersFn(ctx, mChan)
}
func (f *fakeClusterService) GetClusterProjects(ctx context.Context, cluster *infrav3.Cluster) ([]models.ProjectCluster, error) {
	panic("not implemented")
}
func (f *fakeClusterService) UpdateStatus(ctx context.Context, current *infrav3.Cluster, opts ...query.Option) error {
	panic("not implemented")
}
func (f *fakeClusterService) CreateBootstrapAgentForCluster(ctx context.Context, cluster *infrav3.Cluster) error {
	panic("not implemented")
}
func (f *fakeClusterService) GetRelaysConfigForCluster(ctx context.Context, cluster *infrav3.Cluster) ([]common.Relay, error) {
	panic("not implemented")
}
func (f *fakeClusterService) UpdateProjectsForBootstrapAgentForCluster(ctx context.Context, cluster *infrav3.Cluster) error {
	panic("not implemented")
}
func (f *fakeClusterService) AddEventHandler(evh event.Handler) {}

func TestNotifier_AddRemoveListener(t *testing.T) {
	n := New(&fakeClusterService{}).(*notifier)

	t.Run("invalid selector returns an error", func(t *testing.T) {
		ch := make(chan infrav3.Cluster, 1)
		err := n.AddListener(ch, query.WithSelector("==="))
		assert.Error(t, err)
	})

	t.Run("adds and removes a listener", func(t *testing.T) {
		ch := make(chan infrav3.Cluster, 1)
		require.NoError(t, n.AddListener(ch))
		assert.Len(t, n.listeners, 1)

		n.RemoveListener(ch)
		assert.Len(t, n.listeners, 0)
	})
}

func TestNotifier_NotifyListeners(t *testing.T) {
	n := New(&fakeClusterService{}).(*notifier)

	cluster := infrav3.Cluster{Metadata: &commonv3.Metadata{Name: "cluster-1"}}

	t.Run("delivers to a matching listener", func(t *testing.T) {
		ch := make(chan infrav3.Cluster, 1)
		require.NoError(t, n.AddListener(ch, query.WithName("cluster-1")))
		defer n.RemoveListener(ch)

		n.notifyListeners(cluster)

		select {
		case got := <-ch:
			assert.Equal(t, "cluster-1", got.Metadata.Name)
		case <-time.After(time.Second):
			t.Fatal("expected a delivered notification")
		}
	})

	t.Run("does not deliver to a non-matching listener", func(t *testing.T) {
		ch := make(chan infrav3.Cluster, 1)
		require.NoError(t, n.AddListener(ch, query.WithName("other-cluster")))
		defer n.RemoveListener(ch)

		n.notifyListeners(cluster)

		select {
		case <-ch:
			t.Fatal("did not expect a delivered notification")
		case <-time.After(50 * time.Millisecond):
		}
	})

	t.Run("drops the notification instead of blocking when the listener channel is full", func(t *testing.T) {
		ch := make(chan infrav3.Cluster) // unbuffered, nothing reading
		require.NoError(t, n.AddListener(ch))
		defer n.RemoveListener(ch)

		done := make(chan struct{})
		go func() {
			n.notifyListeners(cluster)
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("notifyListeners should not block on a full/unread channel")
		}
	})
}

func TestStart_EndToEnd(t *testing.T) {
	meta := commonv3.Metadata{Name: "cluster-1"}
	cluster := &infrav3.Cluster{Metadata: &meta}

	cs := &fakeClusterService{
		getFn: func(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error) {
			return cluster, nil
		},
		listenClustersFn: func(ctx context.Context, mChan chan<- commonv3.Metadata) {
			select {
			case mChan <- meta:
			case <-ctx.Done():
			}
			<-ctx.Done()
		},
	}
	n := New(cs)

	ch := make(chan infrav3.Cluster, 1)
	require.NoError(t, n.AddListener(ch, query.WithName("cluster-1")))

	stop := make(chan struct{})
	go n.Start(stop)
	defer close(stop)

	select {
	case got := <-ch:
		assert.Equal(t, "cluster-1", got.Metadata.Name)
	case <-time.After(2 * time.Second):
		t.Fatal("expected Start to deliver a notification end-to-end")
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	// Reset package-level state so this test is self-contained regardless
	// of what other tests in this file have already done to it.
	_notifier = nil
	once = sync.Once{}

	t.Run("returns ErrNotInitialized before Init is called", func(t *testing.T) {
		assert.ErrorIs(t, Start(make(chan struct{})), ErrNotInitialized)
		assert.ErrorIs(t, AddListener(make(chan infrav3.Cluster, 1)), ErrNotInitialized)
		assert.ErrorIs(t, RemoveListener(make(chan infrav3.Cluster, 1)), ErrNotInitialized)
	})

	t.Run("works after Init", func(t *testing.T) {
		Init(&fakeClusterService{
			listenClustersFn: func(ctx context.Context, mChan chan<- commonv3.Metadata) {
				<-ctx.Done()
			},
		})

		ch := make(chan infrav3.Cluster, 1)
		require.NoError(t, AddListener(ch))
		require.NoError(t, RemoveListener(ch))

		stop := make(chan struct{})
		require.NoError(t, Start(stop))
		close(stop)
	})
}
