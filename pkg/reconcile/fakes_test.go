package reconcile

import (
	"context"

	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/common"
	"github.com/paralus/paralus/pkg/event"
	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
)

// fakeClusterService implements service.ClusterService with only the
// methods exercised in this package's tests wired via function fields; the
// rest panic since nothing here calls them.
type fakeClusterService struct {
	createFn                       func(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error)
	selectFn                       func(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error)
	getFn                          func(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error)
	updateFn                       func(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error)
	updateClusterConditionStatusFn func(ctx context.Context, current *infrav3.Cluster) error
	updateStatusFn                 func(ctx context.Context, current *infrav3.Cluster, opts ...query.Option) error
	updateClusterAnnotationsFn     func(ctx context.Context, cluster *infrav3.Cluster) error
	createBootstrapAgentFn         func(ctx context.Context, cluster *infrav3.Cluster) error
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
	panic("not implemented")
}
func (f *fakeClusterService) List(ctx context.Context, opts ...query.Option) (*infrav3.ClusterList, error) {
	panic("not implemented")
}
func (f *fakeClusterService) UpdateClusterConditionStatus(ctx context.Context, current *infrav3.Cluster) error {
	return f.updateClusterConditionStatusFn(ctx, current)
}
func (f *fakeClusterService) UpdateClusterAnnotations(ctx context.Context, cluster *infrav3.Cluster) error {
	return f.updateClusterAnnotationsFn(ctx, cluster)
}
func (f *fakeClusterService) ListenClusters(ctx context.Context, mChan chan<- commonv3.Metadata) {
	panic("not implemented")
}
func (f *fakeClusterService) GetClusterProjects(ctx context.Context, cluster *infrav3.Cluster) ([]models.ProjectCluster, error) {
	panic("not implemented")
}
func (f *fakeClusterService) UpdateStatus(ctx context.Context, current *infrav3.Cluster, opts ...query.Option) error {
	return f.updateStatusFn(ctx, current, opts...)
}
func (f *fakeClusterService) CreateBootstrapAgentForCluster(ctx context.Context, cluster *infrav3.Cluster) error {
	return f.createBootstrapAgentFn(ctx, cluster)
}
func (f *fakeClusterService) GetRelaysConfigForCluster(ctx context.Context, cluster *infrav3.Cluster) ([]common.Relay, error) {
	panic("not implemented")
}
func (f *fakeClusterService) UpdateProjectsForBootstrapAgentForCluster(ctx context.Context, cluster *infrav3.Cluster) error {
	panic("not implemented")
}
func (f *fakeClusterService) AddEventHandler(evh event.Handler) {}
