package reconcile

import (
	"context"

	"github.com/paralus/paralus/pkg/log"
	"github.com/paralus/paralus/pkg/service"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
)

var _log = log.GetLogger()

const (
	_bpInprogress int = iota
	_bpFailed
)

type blueprintError struct {
	errorType int
	reason    string
}

func (e blueprintError) Error() string {
	return e.reason
}

type clusterConditionReconciler struct {
	cs service.ClusterService
	/*ps models.PlacementService*/
}

// ClusterConditionReconciler is the interface for reconciling cluster conditions
type ClusterConditionReconciler interface {
	Reconcile(ctx context.Context, cluster *infrav3.Cluster) error
}

// NewClusterConditionReconciler returns cluster condition reconciler
func NewClusterConditionReconciler(cs service.ClusterService) ClusterConditionReconciler {
	return &clusterConditionReconciler{cs: cs}
}

func (r *clusterConditionReconciler) Reconcile(ctx context.Context, cluster *infrav3.Cluster) error {
	_log.Debugw("reconciling cluster conditions", "cluster", cluster.Metadata)

	var conditions []*infrav3.ClusterCondition
	clusterStatus := &infrav3.Cluster{
		Metadata: cluster.Metadata,
		Spec: &infrav3.ClusterSpec{
			ClusterData: &infrav3.ClusterData{
				ClusterStatus: &infrav3.ClusterStatus{
					Conditions: conditions,
				},
			},
		},
	}

	if shouldUpdateClusterStatus(clusterStatus, cluster) {
		err := r.cs.UpdateClusterConditionStatus(ctx, cluster)
		if err != nil {
			_log.Infow("unable to update cluster status", "error", err)
			return err
		}

		_log.Debugw("successfully reconciled cluster condition", "cluster", cluster.Metadata)
	}

	return nil
}

func mergeClusterConditions(conditions []infrav3.ClusterCondition) []infrav3.ClusterCondition {
	condMap := map[infrav3.ClusterConditionType]infrav3.ClusterCondition{}
	var retConditions []infrav3.ClusterCondition

	for _, cond := range conditions {
		if ec, ok := condMap[cond.Type]; ok {
			ec.Reason = ec.Reason + ", " + cond.Reason
			condMap[cond.Type] = ec
		} else {
			condMap[cond.Type] = cond
		}
	}

	for _, cond := range condMap {
		retConditions = append(retConditions, cond)
	}

	return retConditions
}

func shouldUpdateClusterStatus(current, modified *infrav3.Cluster) bool {

	// check if any of the modified conditions are different from
	// current conditions
	for _, modifiedCondition := range modified.Spec.ClusterData.ClusterStatus.Conditions {
		for _, currentCondition := range current.Spec.ClusterData.ClusterStatus.Conditions {
			if modifiedCondition.Type == currentCondition.Type {
				if (modifiedCondition.Status != currentCondition.Status) ||
					(modifiedCondition.Reason != currentCondition.Reason) {
					return true
				}
			}
		}
	}

	return false
}
