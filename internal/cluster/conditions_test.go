package cluster

import (
	"testing"
	"time"

	"github.com/paralus/paralus/internal/cluster/constants"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/paralus/paralus/proto/types/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func clusterWithConditions(conditions ...*infrav3.ClusterCondition) *infrav3.Cluster {
	return &infrav3.Cluster{
		Spec: &infrav3.ClusterSpec{
			ClusterData: &infrav3.ClusterData{
				ClusterStatus: &infrav3.ClusterStatus{Conditions: conditions},
			},
		},
	}
}

func TestNewClusterConditionFuncs(t *testing.T) {
	cond := NewClusterApprove(constants.Success, "approved by admin")
	assert.Equal(t, infrav3.ClusterConditionType_ClusterApprove, cond.Type)
	assert.Equal(t, constants.Success, cond.Status)
	assert.Equal(t, "approved by admin", cond.Reason)
	assert.NotNil(t, cond.LastUpdated)
}

func TestIsClusterConditionFuncs(t *testing.T) {
	t.Run("IsClusterApproved true when success condition present", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterApprove, Status: constants.Success,
		})
		assert.True(t, IsClusterApproved(c))
	})

	t.Run("IsClusterApproved false when condition is a different status", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterApprove, Status: constants.Pending,
		})
		assert.False(t, IsClusterApproved(c))
	})

	t.Run("IsClusterApproved false when condition type is absent", func(t *testing.T) {
		c := clusterWithConditions()
		assert.False(t, IsClusterApproved(c))
	})

	t.Run("IsClusterBootstrapAgentPending matches an explicit status across multiple types", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type: infrav3.ClusterConditionType_ClusterBootstrapAgent, Status: constants.Pending,
		})
		assert.True(t, IsClusterBootstrapAgentPending(c))
		assert.False(t, IsClusterBootstrapAgentRetry(c))
	})
}

func TestSetClusterCondition(t *testing.T) {
	c := clusterWithConditions(&infrav3.ClusterCondition{
		Type: infrav3.ClusterConditionType_ClusterApprove, Status: constants.Pending, Reason: "old",
	})

	SetClusterCondition(c, &infrav3.ClusterCondition{
		Type: infrav3.ClusterConditionType_ClusterApprove, Status: constants.Success, Reason: "new",
	})

	require.Len(t, c.Spec.ClusterData.ClusterStatus.Conditions, 1)
	assert.Equal(t, constants.Success, c.Spec.ClusterData.ClusterStatus.Conditions[0].Status)
	assert.Equal(t, "new", c.Spec.ClusterData.ClusterStatus.Conditions[0].Reason)
}

func TestDefaultClusterConditions(t *testing.T) {
	// Built once at package-init time by iterating infrav3.ClusterConditionType_name;
	// just assert it produced a non-empty, correctly-shaped list.
	require.NotEmpty(t, DefaultClusterConditions)
	for _, c := range DefaultClusterConditions {
		assert.Equal(t, commonv3.ParalusConditionStatus_NotSet, c.Status)
		assert.Equal(t, "pending", c.Reason)
	}
}

func namespaceWithConditions(conditions ...*scheduler.ClusterNamespaceCondition) *scheduler.ClusterNamespace {
	return &scheduler.ClusterNamespace{
		Status: &scheduler.ClusterNamespaceStatus{Conditions: conditions},
	}
}

func TestNamespaceConditionFuncs(t *testing.T) {
	cond := NewNamespaceReady(constants.Success, "ready")
	assert.Equal(t, scheduler.ClusterNamespaceConditionType_ClusterNamespaceReady, cond.Type)

	t.Run("IsNamespaceReady true on matching success condition", func(t *testing.T) {
		n := namespaceWithConditions(&scheduler.ClusterNamespaceCondition{
			Type: scheduler.ClusterNamespaceConditionType_ClusterNamespaceReady, Status: constants.Success,
		})
		assert.True(t, IsNamespaceReady(n))
		assert.False(t, IsNamespaceReadyFailed(n))
	})

	t.Run("NamespaceConvergeFailedReason returns the reason on a matching failed condition", func(t *testing.T) {
		n := namespaceWithConditions(&scheduler.ClusterNamespaceCondition{
			Type: scheduler.ClusterNamespaceConditionType_ClusterNamespaceConverged, Status: constants.Failed, Reason: "boom",
		})
		assert.Equal(t, "boom", NamespaceConvergeFailedReason(n))
	})

	t.Run("NamespaceConvergeFailedReason returns empty string when no match", func(t *testing.T) {
		n := namespaceWithConditions()
		assert.Equal(t, "", NamespaceConvergeFailedReason(n))
	})
}

func TestIsClusterCooledDownFuncs(t *testing.T) {
	t.Run("true when retry condition was last updated before the cooldown window", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type:        infrav3.ClusterConditionType_ClusterBootstrapAgent,
			Status:      constants.Retry,
			LastUpdated: timestamppb.New(time.Now().Add(-time.Hour)),
		})
		assert.True(t, IsClusterBootstrapAgentCooledDown(c, time.Minute))
	})

	t.Run("false when retry condition is still within the cooldown window", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type:        infrav3.ClusterConditionType_ClusterBootstrapAgent,
			Status:      constants.Retry,
			LastUpdated: timestamppb.New(time.Now()),
		})
		assert.False(t, IsClusterBootstrapAgentCooledDown(c, time.Hour))
	})

	t.Run("false when condition status is not Retry", func(t *testing.T) {
		c := clusterWithConditions(&infrav3.ClusterCondition{
			Type:        infrav3.ClusterConditionType_ClusterBootstrapAgent,
			Status:      constants.Success,
			LastUpdated: timestamppb.New(time.Now().Add(-time.Hour)),
		})
		assert.False(t, IsClusterBootstrapAgentCooledDown(c, time.Minute))
	})
}
