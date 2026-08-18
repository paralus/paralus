package server

import (
	"context"
	"errors"
	"testing"

	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/common"
	"github.com/paralus/paralus/pkg/event"
	"github.com/paralus/paralus/pkg/query"
	rpcv3 "github.com/paralus/paralus/proto/rpc/scheduler"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeClusterService implements service.ClusterService with function fields
// so each test wires only the method it needs; everything else panics to
// surface accidental extra calls.
type fakeClusterService struct {
	createFn                     func(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error)
	selectFn                     func(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error)
	getFn                        func(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error)
	updateFn                     func(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error)
	deleteFn                     func(ctx context.Context, cluster *infrav3.Cluster) error
	listFn                       func(ctx context.Context, opts ...query.Option) (*infrav3.ClusterList, error)
	updateClusterConditionStatus func(ctx context.Context, current *infrav3.Cluster) error
}

func (f *fakeClusterService) Create(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {
	return f.createFn(ctx, cluster)
}
func (f *fakeClusterService) Select(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error) {
	return f.selectFn(ctx, cluster, isExtended)
}
func (f *fakeClusterService) Get(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error) {
	return f.getFn(ctx, opts...)
}
func (f *fakeClusterService) Update(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {
	return f.updateFn(ctx, cluster)
}
func (f *fakeClusterService) Delete(ctx context.Context, cluster *infrav3.Cluster) error {
	return f.deleteFn(ctx, cluster)
}
func (f *fakeClusterService) List(ctx context.Context, opts ...query.Option) (*infrav3.ClusterList, error) {
	return f.listFn(ctx, opts...)
}
func (f *fakeClusterService) UpdateClusterConditionStatus(ctx context.Context, current *infrav3.Cluster) error {
	return f.updateClusterConditionStatus(ctx, current)
}
func (f *fakeClusterService) UpdateClusterAnnotations(ctx context.Context, cluster *infrav3.Cluster) error {
	panic("not implemented")
}
func (f *fakeClusterService) ListenClusters(ctx context.Context, mChan chan<- commonv3.Metadata) {
	panic("not implemented")
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
func (f *fakeClusterService) AddEventHandler(evh event.Handler) {
	panic("not implemented")
}

func newTestClusterServer(cs *fakeClusterService) rpcv3.ClusterServiceServer {
	return NewClusterServer(cs, &common.DownloadData{})
}

func TestClusterServer_CreateCluster(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		cs := &fakeClusterService{createFn: func(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {
			return &infrav3.Cluster{Metadata: cluster.Metadata}, nil
		}}
		s := newTestClusterServer(cs)
		req := &infrav3.Cluster{Metadata: &commonv3.Metadata{Name: "c1"}}

		resp, err := s.CreateCluster(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
		assert.Equal(t, commonv3.ConditionStatus_StatusOK, resp.Status.ConditionStatus)
	})

	t.Run("propagates the service error and stamps the request Failed", func(t *testing.T) {
		cs := &fakeClusterService{createFn: func(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {
			return nil, errors.New("boom")
		}}
		s := newTestClusterServer(cs)
		req := &infrav3.Cluster{}

		resp, err := s.CreateCluster(context.Background(), req)
		assert.Error(t, err)
		require.NotNil(t, resp)
		assert.Same(t, req, resp)
		assert.Equal(t, "boom", resp.Status.Reason)
	})
}

func TestClusterServer_GetClusters(t *testing.T) {
	cs := &fakeClusterService{listFn: func(ctx context.Context, opts ...query.Option) (*infrav3.ClusterList, error) {
		return &infrav3.ClusterList{}, nil
	}}
	s := newTestClusterServer(cs)

	_, err := s.GetClusters(context.Background(), &commonv3.QueryOptions{})
	require.NoError(t, err)
}

func TestClusterServer_GetCluster(t *testing.T) {
	t.Run("success stamps OK on the response", func(t *testing.T) {
		cs := &fakeClusterService{selectFn: func(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error) {
			assert.True(t, isExtended)
			return &infrav3.Cluster{}, nil
		}}
		s := newTestClusterServer(cs)

		resp, err := s.GetCluster(context.Background(), &infrav3.Cluster{})
		require.NoError(t, err)
		require.NotNil(t, resp.Status)
	})

	t.Run("propagates the service error", func(t *testing.T) {
		cs := &fakeClusterService{selectFn: func(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error) {
			return nil, errors.New("not found")
		}}
		s := newTestClusterServer(cs)
		req := &infrav3.Cluster{}

		resp, err := s.GetCluster(context.Background(), req)
		assert.Error(t, err)
		assert.Same(t, req, resp)
	})
}

func TestClusterServer_DeleteCluster(t *testing.T) {
	t.Run("success returns an empty response", func(t *testing.T) {
		cs := &fakeClusterService{deleteFn: func(ctx context.Context, cluster *infrav3.Cluster) error {
			return nil
		}}
		s := newTestClusterServer(cs)

		resp, err := s.DeleteCluster(context.Background(), &infrav3.Cluster{})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("propagates the service error and returns a nil response", func(t *testing.T) {
		want := errors.New("boom")
		cs := &fakeClusterService{deleteFn: func(ctx context.Context, cluster *infrav3.Cluster) error {
			return want
		}}
		s := newTestClusterServer(cs)

		resp, err := s.DeleteCluster(context.Background(), &infrav3.Cluster{})
		assert.Same(t, want, err)
		assert.Nil(t, resp)
	})
}

func TestClusterServer_UpdateCluster(t *testing.T) {
	cs := &fakeClusterService{updateFn: func(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {
		return &infrav3.Cluster{}, nil
	}}
	s := newTestClusterServer(cs)

	resp, err := s.UpdateCluster(context.Background(), &infrav3.Cluster{})
	require.NoError(t, err)
	require.NotNil(t, resp.Status)
}

func TestClusterServer_DownloadCluster(t *testing.T) {
	t.Run("renders the download template on success", func(t *testing.T) {
		cs := &fakeClusterService{selectFn: func(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error) {
			return &infrav3.Cluster{
				Metadata: &commonv3.Metadata{Name: "c1"},
				Spec:     &infrav3.ClusterSpec{},
			}, nil
		}}
		s := newTestClusterServer(cs)

		resp, err := s.DownloadCluster(context.Background(), &infrav3.Cluster{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "application/x-paralus-yaml", resp.ContentType)
		assert.NotEmpty(t, resp.Data)
	})

	t.Run("base64-encodes a non-empty BootstrapCA before rendering", func(t *testing.T) {
		cs := &fakeClusterService{selectFn: func(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error) {
			return &infrav3.Cluster{
				Metadata: &commonv3.Metadata{Name: "c1"},
				Spec: &infrav3.ClusterSpec{
					ProxyConfig: &infrav3.ProxyConfig{BootstrapCA: "raw-ca"},
				},
			}, nil
		}}
		s := newTestClusterServer(cs)

		resp, err := s.DownloadCluster(context.Background(), &infrav3.Cluster{})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("propagates a select error", func(t *testing.T) {
		want := errors.New("boom")
		cs := &fakeClusterService{selectFn: func(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error) {
			return nil, want
		}}
		s := newTestClusterServer(cs)

		resp, err := s.DownloadCluster(context.Background(), &infrav3.Cluster{})
		assert.Same(t, want, err)
		assert.Nil(t, resp)
	})
}

func TestClusterServer_UpdateClusterStatus(t *testing.T) {
	t.Run("success returns an empty response", func(t *testing.T) {
		var gotStatus *infrav3.ClusterStatus
		cs := &fakeClusterService{updateClusterConditionStatus: func(ctx context.Context, current *infrav3.Cluster) error {
			gotStatus = current.Spec.ClusterData.ClusterStatus
			return nil
		}}
		s := newTestClusterServer(cs)

		want := &infrav3.ClusterStatus{Conditions: nil}
		resp, err := s.UpdateClusterStatus(context.Background(), &rpcv3.UpdateClusterStatusRequest{
			Metadata:      &commonv3.Metadata{Name: "c1"},
			ClusterStatus: want,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Same(t, want, gotStatus)
	})

	t.Run("propagates the service error", func(t *testing.T) {
		want := errors.New("boom")
		cs := &fakeClusterService{updateClusterConditionStatus: func(ctx context.Context, current *infrav3.Cluster) error {
			return want
		}}
		s := newTestClusterServer(cs)

		resp, err := s.UpdateClusterStatus(context.Background(), &rpcv3.UpdateClusterStatusRequest{})
		assert.Same(t, want, err)
		assert.Nil(t, resp)
	})
}

func TestClusterServer_GetClusterStatus(t *testing.T) {
	t.Run("success maps the cluster status from the fetched cluster", func(t *testing.T) {
		want := &infrav3.ClusterStatus{}
		cs := &fakeClusterService{getFn: func(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error) {
			return &infrav3.Cluster{
				Metadata: &commonv3.Metadata{Name: "c1"},
				Spec: &infrav3.ClusterSpec{
					ClusterData: &infrav3.ClusterData{ClusterStatus: want},
				},
			}, nil
		}}
		s := newTestClusterServer(cs)

		resp, err := s.GetClusterStatus(context.Background(), &rpcv3.GetClusterStatusRequest{Metadata: &commonv3.Metadata{Name: "c1"}})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Same(t, want, resp.ClusterStatus)
	})

	t.Run("propagates the service error", func(t *testing.T) {
		want := errors.New("boom")
		cs := &fakeClusterService{getFn: func(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error) {
			return nil, want
		}}
		s := newTestClusterServer(cs)

		resp, err := s.GetClusterStatus(context.Background(), &rpcv3.GetClusterStatusRequest{})
		assert.Same(t, want, err)
		assert.Nil(t, resp)
	})
}
