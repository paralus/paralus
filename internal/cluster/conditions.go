package cluster

import (
	"time"

	"github.com/paralus/paralus/internal/cluster/constants"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/paralus/paralus/proto/types/scheduler"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ClusterConditionFunc is the function signature for creating new cluster condition
type ClusterConditionFunc func(status commonv3.ParalusConditionStatus, reason string) *infrav3.ClusterCondition

// ClusterConditionReadyFunc checks if condition type is ready
type ClusterConditionReadyFunc func(c *infrav3.Cluster) bool

// NamespaceConditionFunc is the function signature for creating new cluster namespace condition
type NamespaceConditionFunc func(status commonv3.ParalusConditionStatus, reason string) *scheduler.ClusterNamespaceCondition

// NamespaceConditionReadyFunc checks if condition type is ready
type NamespaceConditionReadyFunc func(n *scheduler.ClusterNamespace) bool

// NamespaceConfitionReasonFunc returns condition status reason
type NamespaceConfitionReasonFunc func(n *scheduler.ClusterNamespace) string

// ClusterConditionCooledDownFunc checks if condition type is in retry and last updated has passed
type ClusterConditionCooledDownFunc func(c *infrav3.Cluster, passed time.Duration) bool

var (
	NewClusterRegister          ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_REGISTER)
	NewClusterApprove           ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_APPROVE)
	NewClusterCheckIn           ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_CHECK_IN)
	NewClusterNodeSync          ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_NODE_SYNC)
	NewClusterNamespaceSync     ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_NAMESPACE_SYNC)
	NewClusterReady             ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_READY)
	NewClusterAuxiliaryTaskSync ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_AUXILIARY_TASK_SYNC)
	NewClusterBootstrapAgent    ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_BOOTSTRAP_AGENT)
	NewClusterDelete            ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_DELETE)

	IsClusterBootstrapAgentPending ClusterConditionReadyFunc = isClusterCondition(constants.Pending, infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_BOOTSTRAP_AGENT)
	IsClusterBootstrapAgentRetry   ClusterConditionReadyFunc = isClusterCondition(constants.Retry, infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_BOOTSTRAP_AGENT)
	IsClusterBootstrapAgentCreated ClusterConditionReadyFunc = isClusterCondition(constants.Success, infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_BOOTSTRAP_AGENT)
	IsClusterRegisterd             ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_REGISTER)
	IsClusterRegisterPending       ClusterConditionReadyFunc = isClusterCondition(constants.Pending, infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_REGISTER)
	IsClusterApproved              ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_APPROVE)
	IsClusterCheckedIn             ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_CHECK_IN)
	IsClusterNodeSynced            ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_NODE_SYNC)
	IsClusterNamespaceSynced       ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_NAMESPACE_SYNC)
	IsClusterReady                 ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_READY)
	IsClusterDeletePending         ClusterConditionReadyFunc = isClusterCondition(constants.Pending, infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_DELETE)
	IsClusterDeleteRetry           ClusterConditionReadyFunc = isClusterCondition(constants.Retry, infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_DELETE)
	IsClusterDeleted               ClusterConditionReadyFunc = isClusterCondition(constants.Success, infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_DELETE)
	IsClusterDeleteNotSet          ClusterConditionReadyFunc = isClusterCondition(constants.NotSet, infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_DELETE)

	NewNamespaceAssigned  NamespaceConditionFunc = newNamespaceCondition(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_ASSIGNED_UNSPECIFIED)
	NewNamespaceConverged NamespaceConditionFunc = newNamespaceCondition(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_CONVERGED)
	NewNamespaceReady     NamespaceConditionFunc = newNamespaceCondition(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_READY)
	NewNamespaceDeleted   NamespaceConditionFunc = newNamespaceCondition(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_DELETE)

	IsNamespaceAssigned           NamespaceConditionReadyFunc  = isNamespaceCondition(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_ASSIGNED_UNSPECIFIED, constants.Success)
	IsNamespaceConverged          NamespaceConditionReadyFunc  = isNamespaceCondition(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_CONVERGED, constants.Success)
	IsNamespaceConvergeFailed     NamespaceConditionReadyFunc  = isNamespaceCondition(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_CONVERGED, constants.Failed)
	IsNamespaceReady              NamespaceConditionReadyFunc  = isNamespaceCondition(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_READY, constants.Success)
	IsNamespaceReadyFailed        NamespaceConditionReadyFunc  = isNamespaceCondition(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_READY, constants.Failed)
	IsNamespaceDeleted            NamespaceConditionReadyFunc  = isNamespaceCondition(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_DELETE, constants.Success)
	NamespaceConvergeFailedReason NamespaceConfitionReasonFunc = namespaceConditionReason(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_CONVERGED, constants.Failed)
	NamespaceReadyFailedReason    NamespaceConfitionReasonFunc = namespaceConditionReason(scheduler.ClusterNamespaceConditionType_CLUSTER_NAMESPACE_CONDITION_TYPE_CLUSTER_NAMESPACE_READY, constants.Failed)

	IsClusterBootstrapAgentCooledDown ClusterConditionCooledDownFunc = isClusterCooledDown(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_BOOTSTRAP_AGENT)
	IsClusterDeleteCooledDown         ClusterConditionCooledDownFunc = isClusterCooledDown(infrav3.ClusterConditionType_CLUSTER_CONDITION_TYPE_CLUSTER_DELETE)
)

// DefaultClusterConditions is the default cluster conditions list
var DefaultClusterConditions = func() []*infrav3.ClusterCondition {
	var conditions []*infrav3.ClusterCondition

	var i int32 = 1
	for {
		if _, ok := infrav3.ClusterConditionType_name[i]; !ok {
			break
		}
		clstrCnd := newClusterCondition(infrav3.ClusterConditionType(i))(commonv3.ParalusConditionStatus_PARALUS_CONDITION_STATUS_NOT_SET_UNSPECIFIED, "pending")
		conditions = append(conditions, clstrCnd)
		i++
	}

	return conditions
}()

func newClusterCondition(conditionType infrav3.ClusterConditionType) func(status commonv3.ParalusConditionStatus, reason string) *infrav3.ClusterCondition {
	return func(status commonv3.ParalusConditionStatus, reason string) *infrav3.ClusterCondition {
		return &infrav3.ClusterCondition{
			Type:        conditionType,
			Status:      status,
			Reason:      reason,
			LastUpdated: timestamppb.New(time.Now()),
		}
	}
}

// SetClusterCondition sets condition in cluster conditions
var SetClusterCondition = func(c *infrav3.Cluster, condtition *infrav3.ClusterCondition) {
	for i, ec := range c.Spec.ClusterData.ClusterStatus.Conditions {
		if ec.Type == condtition.Type {
			c.Spec.ClusterData.ClusterStatus.Conditions[i] = condtition
			break
		}
	}

}

func isClusterCondition(conditionStatus commonv3.ParalusConditionStatus, conditionTypes ...infrav3.ClusterConditionType) func(c *infrav3.Cluster) bool {
	return func(c *infrav3.Cluster) bool {
		for _, condition := range c.Spec.ClusterData.ClusterStatus.Conditions {
			for _, conditionType := range conditionTypes {
				if condition.Type == conditionType {
					if condition.Status == conditionStatus {
						return true
					}
				}
			}
		}
		return false
	}
}

func isClusterConditionSuccess(conditionType infrav3.ClusterConditionType) func(c *infrav3.Cluster) bool {
	return func(c *infrav3.Cluster) bool {
		for _, condition := range c.Spec.ClusterData.ClusterStatus.Conditions {
			if condition.Type == conditionType {
				if condition.Status == constants.Success {
					return true
				}
			}
		}
		return false
	}
}

func newNamespaceCondition(conditionType scheduler.ClusterNamespaceConditionType) func(status commonv3.ParalusConditionStatus, reason string) *scheduler.ClusterNamespaceCondition {
	return func(status commonv3.ParalusConditionStatus, reason string) *scheduler.ClusterNamespaceCondition {
		return &scheduler.ClusterNamespaceCondition{
			Type:        conditionType,
			Status:      status,
			Reason:      reason,
			LastUpdated: timestamppb.Now(),
		}
	}
}

func isNamespaceCondition(conditionType scheduler.ClusterNamespaceConditionType, conditionStatus commonv3.ParalusConditionStatus) func(n *scheduler.ClusterNamespace) bool {
	return func(n *scheduler.ClusterNamespace) bool {
		for _, condition := range n.Status.Conditions {
			if condition.Type == conditionType {
				if condition.Status == conditionStatus {
					return true
				}
			}
		}
		return false
	}
}

func namespaceConditionReason(conditionType scheduler.ClusterNamespaceConditionType, conditionStatus commonv3.ParalusConditionStatus) func(n *scheduler.ClusterNamespace) string {
	return func(n *scheduler.ClusterNamespace) string {
		for _, condition := range n.Status.Conditions {
			if condition.Type == conditionType {
				if condition.Status == conditionStatus {
					return condition.Reason
				}
			}
		}
		return ""
	}
}

func isClusterCooledDown(conditionType infrav3.ClusterConditionType) func(c *infrav3.Cluster, passed time.Duration) bool {
	return func(c *infrav3.Cluster, passed time.Duration) bool {
		for _, condition := range c.Spec.ClusterData.ClusterStatus.Conditions {
			if condition.Type == conditionType && condition.Status == constants.Retry {
				if condition.LastUpdated.AsTime().Before(time.Now().Add(-passed)) {
					return true
				}
			}
		}
		return false
	}
}
