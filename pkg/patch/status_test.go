package patch

import (
	"testing"
	"time"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/paralus/paralus/proto/types/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
)

func TestClusterStatus(t *testing.T) {
	t.Run("merges a condition of the same type by status/reason", func(t *testing.T) {
		existing := &infrav3.ClusterStatus{
			Conditions: []*infrav3.ClusterCondition{
				{Type: infrav3.ClusterConditionType_ClusterReady, Status: commonv3.ParalusConditionStatus_Pending, Reason: "waiting"},
				{Type: infrav3.ClusterConditionType_ClusterApprove, Status: commonv3.ParalusConditionStatus_Success, Reason: "approved"},
			},
		}
		current := &infrav3.ClusterStatus{
			Conditions: []*infrav3.ClusterCondition{
				{Type: infrav3.ClusterConditionType_ClusterReady, Status: commonv3.ParalusConditionStatus_Success, Reason: "ready now"},
			},
		}

		err := ClusterStatus(existing, current)
		require.NoError(t, err)

		require.Len(t, existing.Conditions, 2)
		var readyCond, approveCond *infrav3.ClusterCondition
		for _, c := range existing.Conditions {
			switch c.Type {
			case infrav3.ClusterConditionType_ClusterReady:
				readyCond = c
			case infrav3.ClusterConditionType_ClusterApprove:
				approveCond = c
			}
		}
		require.NotNil(t, readyCond)
		require.NotNil(t, approveCond)
		assert.Equal(t, commonv3.ParalusConditionStatus_Success, readyCond.Status)
		assert.Equal(t, "ready now", readyCond.Reason)
		// the untouched condition (ClusterApprove) survives the merge
		assert.Equal(t, "approved", approveCond.Reason)
	})

	t.Run("copies PublishedBlueprint from current when set", func(t *testing.T) {
		existing := &infrav3.ClusterStatus{PublishedBlueprint: "old-bp"}
		current := &infrav3.ClusterStatus{PublishedBlueprint: "new-bp"}

		err := ClusterStatus(existing, current)
		require.NoError(t, err)
		assert.Equal(t, "new-bp", existing.PublishedBlueprint)
	})

	t.Run("leaves PublishedBlueprint untouched when current's is empty", func(t *testing.T) {
		existing := &infrav3.ClusterStatus{PublishedBlueprint: "old-bp"}
		current := &infrav3.ClusterStatus{}

		err := ClusterStatus(existing, current)
		require.NoError(t, err)
		assert.Equal(t, "old-bp", existing.PublishedBlueprint)
	})
}

func TestClusterNodeStatus(t *testing.T) {
	t.Run("patches state from current onto existing", func(t *testing.T) {
		existing := &infrav3.ClusterNodeStatus{
			State: infrav3.ClusterNodeState_ClusterNodeReady,
			NodeInfo: &corev1.NodeSystemInfo{
				OSImage: "old-os",
			},
		}
		current := &infrav3.ClusterNodeStatus{
			State: infrav3.ClusterNodeState_ClusterNodeNotReady,
			NodeInfo: &corev1.NodeSystemInfo{
				OSImage: "new-os",
			},
		}

		err := ClusterNodeStatus(existing, current)
		require.NoError(t, err)
		assert.Equal(t, infrav3.ClusterNodeState_ClusterNodeNotReady, existing.State)
		assert.Equal(t, "new-os", existing.NodeInfo.OSImage)
	})

	t.Run("no-op when current is empty", func(t *testing.T) {
		existing := &infrav3.ClusterNodeStatus{State: infrav3.ClusterNodeState_ClusterNodeReady}
		current := &infrav3.ClusterNodeStatus{}

		err := ClusterNodeStatus(existing, current)
		require.NoError(t, err)
		assert.Equal(t, infrav3.ClusterNodeState_ClusterNodeReady, existing.State)
	})
}

func TestNamespaceStatus(t *testing.T) {
	t.Run("patches conditions from current onto existing", func(t *testing.T) {
		existing := &scheduler.ClusterNamespaceStatus{
			Conditions: []*scheduler.ClusterNamespaceCondition{
				{
					Type:        scheduler.ClusterNamespaceConditionType_ClusterNamespaceAssigned,
					Status:      commonv3.ParalusConditionStatus_Pending,
					LastUpdated: timestamppb.New(time.Now()),
				},
			},
		}
		current := &scheduler.ClusterNamespaceStatus{
			Conditions: []*scheduler.ClusterNamespaceCondition{
				{
					Type:        scheduler.ClusterNamespaceConditionType_ClusterNamespaceReady,
					Status:      commonv3.ParalusConditionStatus_Success,
					LastUpdated: timestamppb.New(time.Now()),
				},
			},
		}

		err := NamespaceStatus(existing, current)
		require.NoError(t, err)
		require.NotEmpty(t, existing.Conditions)
	})

	t.Run("no-op when both existing and current are empty", func(t *testing.T) {
		existing := &scheduler.ClusterNamespaceStatus{}
		current := &scheduler.ClusterNamespaceStatus{}

		err := NamespaceStatus(existing, current)
		require.NoError(t, err)
		assert.Empty(t, existing.Conditions)
	})
}
