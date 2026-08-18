package reconcile

import (
	"context"
	"testing"
	"time"

	"github.com/paralus/paralus/internal/cluster/constants"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCanReconcileClusterDelete(t *testing.T) {
	t.Run("true when delete is pending", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterDelete, Status: constants.Pending,
		})
		assert.True(t, canReconcileClusterDelete(c))
	})

	t.Run("true when delete is retry and cooled down", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterDelete, Status: constants.Retry,
			LastUpdated: timestamppb.New(time.Now().Add(-time.Hour)),
		})
		assert.True(t, canReconcileClusterDelete(c))
	})

	t.Run("false when delete is retry but still within cooldown", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterDelete, Status: constants.Retry,
			LastUpdated: timestamppb.New(time.Now()),
		})
		assert.False(t, canReconcileClusterDelete(c))
	})

	t.Run("false when there is no delete condition", func(t *testing.T) {
		c := clusterWithConditions()
		assert.False(t, canReconcileClusterDelete(c))
	})
}

func TestCanReconcileClusterBootstrapAgent(t *testing.T) {
	t.Run("true when bootstrap agent is pending", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterBootstrapAgent, Status: constants.Pending,
		})
		assert.True(t, canReconcileClusterBootstrapAgent(c))
	})

	t.Run("true when bootstrap agent is retry and cooled down", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterBootstrapAgent, Status: constants.Retry,
			LastUpdated: timestamppb.New(time.Now().Add(-time.Hour)),
		})
		assert.True(t, canReconcileClusterBootstrapAgent(c))
	})

	t.Run("false when bootstrap agent is retry but still within cooldown", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterBootstrapAgent, Status: constants.Retry,
			LastUpdated: timestamppb.New(time.Now()),
		})
		assert.False(t, canReconcileClusterBootstrapAgent(c))
	})

	t.Run("false when there is no bootstrap agent condition", func(t *testing.T) {
		c := clusterWithConditions()
		assert.False(t, canReconcileClusterBootstrapAgent(c))
	})
}

func TestClusterReconciler_Reconcile_NoOpWhenNothingToDo(t *testing.T) {
	// When neither canReconcileClusterDelete nor canReconcileClusterBootstrapAgent
	// match, Reconcile returns nil without touching any dependency -- the
	// only branch reachable without a BootstrapService/kubeconfig/session
	// fake, which is why this is the only Reconcile path covered here.
	r := NewClusterReconciler(&fakeClusterService{}, nil, nil, nil)
	cluster := clusterWithConditions(&infrav3.ClusterCondition{
		Type: infrav3.ClusterConditionType_ClusterReady, Status: commonv3.ParalusConditionStatus_Success,
	})
	err := r.Reconcile(context.Background(), cluster)
	assert.NoError(t, err)
}
