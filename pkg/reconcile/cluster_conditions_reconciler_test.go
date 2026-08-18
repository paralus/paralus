package reconcile

import (
	"context"
	"errors"
	"testing"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clusterWithConditions(conditions ...*infrav3.ClusterCondition) *infrav3.Cluster {
	return &infrav3.Cluster{
		Metadata: &commonv3.Metadata{Name: "cluster-1"},
		Spec: &infrav3.ClusterSpec{
			ClusterData: &infrav3.ClusterData{
				ClusterStatus: &infrav3.ClusterStatus{Conditions: conditions},
			},
		},
	}
}

func TestShouldUpdateClusterStatus(t *testing.T) {
	t.Run("true when a shared condition type has a different status", func(t *testing.T) {
		current := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterReady, Status: commonv3.ParalusConditionStatus_Pending,
		})
		modified := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterReady, Status: commonv3.ParalusConditionStatus_Success,
		})
		assert.True(t, shouldUpdateClusterStatus(current, modified))
	})

	t.Run("true when a shared condition type has a different reason", func(t *testing.T) {
		current := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterReady, Status: commonv3.ParalusConditionStatus_Success, Reason: "old",
		})
		modified := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterReady, Status: commonv3.ParalusConditionStatus_Success, Reason: "new",
		})
		assert.True(t, shouldUpdateClusterStatus(current, modified))
	})

	t.Run("false when conditions are identical", func(t *testing.T) {
		current := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterReady, Status: commonv3.ParalusConditionStatus_Success, Reason: "same",
		})
		modified := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterReady, Status: commonv3.ParalusConditionStatus_Success, Reason: "same",
		})
		assert.False(t, shouldUpdateClusterStatus(current, modified))
	})

	t.Run("false when there are no shared condition types", func(t *testing.T) {
		current := clusterWithConditions(&infrav3.ClusterCondition{Type: infrav3.ClusterConditionType_ClusterReady})
		modified := clusterWithConditions(&infrav3.ClusterCondition{Type: infrav3.ClusterConditionType_ClusterApprove})
		assert.False(t, shouldUpdateClusterStatus(current, modified))
	})

	t.Run("false when either side has no conditions", func(t *testing.T) {
		current := clusterWithConditions()
		modified := clusterWithConditions(&infrav3.ClusterCondition{Type: infrav3.ClusterConditionType_ClusterReady})
		assert.False(t, shouldUpdateClusterStatus(current, modified))
	})
}

// BUG: Reconcile builds its "current" comparison baseline as a fresh
// *infrav3.Cluster with an empty (nil) Conditions slice -- it never fetches
// or fills in the cluster's actual prior status. shouldUpdateClusterStatus's
// outer loop iterates modified's conditions but the inner loop iterates
// current's (always empty) conditions, so the inner loop body -- and thus
// any call to UpdateClusterConditionStatus -- can never run, for any input.
// clusterConditionReconciler.Reconcile is therefore a permanent no-op today.
func TestClusterConditionReconciler_Reconcile(t *testing.T) {
	t.Run("never calls UpdateClusterConditionStatus regardless of the cluster's conditions", func(t *testing.T) {
		called := false
		cs := &fakeClusterService{
			updateClusterConditionStatusFn: func(ctx context.Context, current *infrav3.Cluster) error {
				called = true
				return errors.New("should never be reached")
			},
		}
		r := NewClusterConditionReconciler(cs)

		cluster := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterReady, Status: commonv3.ParalusConditionStatus_Success,
		})
		err := r.Reconcile(context.Background(), cluster)
		require.NoError(t, err)
		assert.False(t, called)
	})

	t.Run("also a no-op for a cluster with no conditions at all", func(t *testing.T) {
		cs := &fakeClusterService{}
		r := NewClusterConditionReconciler(cs)
		err := r.Reconcile(context.Background(), clusterWithConditions())
		require.NoError(t, err)
	})
}
