package reconcile

import (
	"context"
	"fmt"
	"time"

	clstrutil "github.com/paralus/paralus/internal/cluster"
	"github.com/paralus/paralus/internal/cluster/constants"
	"github.com/paralus/paralus/pkg/service"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"

	"github.com/paralus/paralus/pkg/query"
	"github.com/pkg/errors"
)

const (
	clusterCoolDown = time.Minute
)

// ClusterReconciler reconciles cluster state
type ClusterReconciler interface {
	Reconcile(ctx context.Context, cluster *infrav3.Cluster) error
}

type clusterReconciler struct {
	cs service.ClusterService
}

// NewClusterReconciler returns new cluster reconciler
func NewClusterReconciler(cs service.ClusterService) ClusterReconciler {
	return &clusterReconciler{cs: cs}
}

func (r *clusterReconciler) Reconcile(ctx context.Context, cluster *infrav3.Cluster) error {

	switch {
	case canReconcileClusterDelete(cluster):
		return r.handleClusterDelete(ctx, cluster)
	case canReconcileClusterBootstrapAgent(cluster):
		return r.handleClusterBootstrapAgent(ctx, cluster)
	default:
		return nil
	}
}

func canReconcileClusterDelete(c *infrav3.Cluster) bool {
	switch {
	case clstrutil.IsClusterDeletePending(c):
		return true
	case clstrutil.IsClusterDeleteRetry(c) && clstrutil.IsClusterDeleteCooledDown(c, clusterCoolDown):
		return true
	default:
		return false
	}
}

func canReconcileClusterBootstrapAgent(c *infrav3.Cluster) bool {
	switch {
	case clstrutil.IsClusterBootstrapAgentPending(c):
		return true
	case clstrutil.IsClusterBootstrapAgentRetry(c) && clstrutil.IsClusterBootstrapAgentCooledDown(c, clusterCoolDown):
		return true
	default:
		return false
	}
}

func (r *clusterReconciler) handleClusterDelete(ctx context.Context, cluster *infrav3.Cluster) error {
	_log.Infow("handling cluster delete", "cluster.Name", cluster.Metadata.Name)
	try := func() error {
		_log.Debugw("no relay networks to disassociate", "cluster name", cluster.Metadata.Name)
		return nil
	}

	onSuccess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		err := r.cs.UpdateStatus(ctx, &infrav3.Cluster{
			Metadata: cluster.Metadata,
			Spec: &infrav3.ClusterSpec{
				ClusterData: &infrav3.ClusterData{
					ClusterStatus: &infrav3.ClusterStatus{
						Conditions: []*infrav3.ClusterCondition{
							clstrutil.NewClusterDelete(constants.Success, "cluster deleted")},
					},
				},
			},
		}, query.WithMeta(cluster.Metadata))
		return err
	}

	onFailure := func(reason string) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		err := r.cs.UpdateStatus(ctx, &infrav3.Cluster{
			Metadata: cluster.Metadata,
			Spec: &infrav3.ClusterSpec{
				ClusterData: &infrav3.ClusterData{
					ClusterStatus: &infrav3.ClusterStatus{
						Conditions: []*infrav3.ClusterCondition{
							clstrutil.NewClusterDelete(constants.Retry, reason)},
					},
				},
			},
		}, query.WithMeta(cluster.Metadata))
		return err
	}

	if err := try(); err != nil {
		_log.Infow("unable to delete cluster", "name", cluster.Metadata.Name, "error", err)
		err = errors.Wrap(err, "unable to delete cluster")
		sErr := onFailure(err.Error())
		if sErr != nil {
			_log.Infow("unable to update status", "error", err)
		}
		return err
	}

	if err := onSuccess(); err != nil {
		_log.Infow("unable to update status", "error", err)
	}

	return nil
}

func (r *clusterReconciler) handleClusterBootstrapAgent(ctx context.Context, cluster *infrav3.Cluster) error {
	_log.Infow("handling cluster bootstrap agent", "cluster", cluster.Metadata)
	try := func() error {
		err := r.cs.CreateBootstrapAgentForCluster(ctx, cluster)
		return err
	}

	onSuccess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		fmt.Print("annotations in reconciler :: ", cluster.Metadata.Annotations)

		err := r.cs.UpdateStatus(ctx, &infrav3.Cluster{
			Metadata: cluster.Metadata,
			Spec: &infrav3.ClusterSpec{
				ClusterData: &infrav3.ClusterData{
					ClusterStatus: &infrav3.ClusterStatus{
						Conditions: []*infrav3.ClusterCondition{
							clstrutil.NewClusterBootstrapAgent(constants.Success, "bootstrap agent created")},
					},
				},
			},
		}, query.WithMeta(cluster.Metadata))

		//update relays to annotations
		if err == nil {
			err = r.cs.UpdateClusterAnnotations(ctx, cluster)
		}
		return err
	}

	onFailure := func(reason string) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		err := r.cs.UpdateStatus(ctx, &infrav3.Cluster{
			Metadata: cluster.Metadata,
			Spec: &infrav3.ClusterSpec{
				ClusterData: &infrav3.ClusterData{
					ClusterStatus: &infrav3.ClusterStatus{
						Conditions: []*infrav3.ClusterCondition{
							clstrutil.NewClusterBootstrapAgent(constants.Retry, reason)},
					},
				},
			},
		}, query.WithMeta(cluster.Metadata))
		return err
	}

	if err := try(); err != nil {
		_log.Infow("unable to create cluster bootstrap agent", "error", err)
		err = errors.Wrap(err, "unable to create cluster boostrap agent")
		sErr := onFailure(err.Error())
		if sErr != nil {
			_log.Infow("unable to update status", "error", err)
		}
		return err
	}

	if err := onSuccess(); err != nil {
		_log.Infow("unable to update status", "error", err)
	}

	return nil
}
