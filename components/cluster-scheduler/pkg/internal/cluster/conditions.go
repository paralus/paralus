package cluster

import (
	"time"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/cluster/constants"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ClusterConditionFunc is the function signature for creating new cluster condition
type ClusterConditionFunc func(status commonv3.RafayConditionStatus, reason string) *infrav3.ClusterCondition

// ClusterConditionReadyFunc checks if condition type is ready
type ClusterConditionReadyFunc func(c *infrav3.Cluster) bool

var (
	NewClusterRegister          ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_ClusterRegister)
	NewClusterApprove           ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_ClusterApprove)
	NewClusterCheckIn           ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_ClusterCheckIn)
	NewClusterNodeSync          ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_ClusterNodeSync)
	NewClusterNamespaceSync     ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_ClusterNamespaceSync)
	NewClusterBlueprintSync     ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_ClusterBlueprintSync)
	NewClusterReady             ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_ClusterReady)
	NewClusterAuxiliaryTaskSync ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_ClusterAuxiliaryTaskSync)
	NewClusterBootstrapAgent    ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_ClusterBootstrapAgent)
	NewClusterDelete            ClusterConditionFunc = newClusterCondition(infrav3.ClusterConditionType_ClusterDelete)

	IsClusterBootstrapAgentPending   ClusterConditionReadyFunc = isClusterCondition(constants.Pending, infrav3.ClusterConditionType_ClusterBootstrapAgent)
	IsClusterBootstrapAgentRetry     ClusterConditionReadyFunc = isClusterCondition(constants.Retry, infrav3.ClusterConditionType_ClusterBootstrapAgent)
	IsClusterBootstrapAgentCreated   ClusterConditionReadyFunc = isClusterCondition(constants.Success, infrav3.ClusterConditionType_ClusterBootstrapAgent)
	IsClusterRegisterd               ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_ClusterRegister)
	IsClusterRegisterPending         ClusterConditionReadyFunc = isClusterCondition(constants.Pending, infrav3.ClusterConditionType_ClusterRegister)
	IsClusterApproved                ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_ClusterApprove)
	IsClusterCheckedIn               ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_ClusterCheckIn)
	IsClusterNodeSynced              ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_ClusterNodeSync)
	IsClusterNamespaceSynced         ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_ClusterNamespaceSync)
	IsClusterBlueprintSynced         ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_ClusterBlueprintSync)
	IsClusterBlueprintSyncPending    ClusterConditionReadyFunc = isClusterCondition(constants.Pending, infrav3.ClusterConditionType_ClusterBlueprintSync)
	IsClusterBlueprintSyncRetry      ClusterConditionReadyFunc = isClusterCondition(constants.Retry, infrav3.ClusterConditionType_ClusterBlueprintSync)
	IsClusterBlueprintSyncSuccess    ClusterConditionReadyFunc = isClusterCondition(constants.Success, infrav3.ClusterConditionType_ClusterBlueprintSync)
	IsClusterBlueprintSyncInprogress ClusterConditionReadyFunc = isClusterCondition(constants.InProgress, infrav3.ClusterConditionType_ClusterBlueprintSync)
	IsClusterBlueprintSyncFailed     ClusterConditionReadyFunc = isClusterCondition(constants.Failed, infrav3.ClusterConditionType_ClusterBlueprintSync)
	IsClusterReady                   ClusterConditionReadyFunc = isClusterConditionSuccess(infrav3.ClusterConditionType_ClusterReady)
	IsClusterDeletePending           ClusterConditionReadyFunc = isClusterCondition(constants.Pending, infrav3.ClusterConditionType_ClusterDelete)
	IsClusterDeleteRetry             ClusterConditionReadyFunc = isClusterCondition(constants.Retry, infrav3.ClusterConditionType_ClusterDelete)
	IsClusterDeleted                 ClusterConditionReadyFunc = isClusterCondition(constants.Success, infrav3.ClusterConditionType_ClusterDelete)
	IsClusterDeleteNotSet            ClusterConditionReadyFunc = isClusterCondition(constants.NotSet, infrav3.ClusterConditionType_ClusterDelete)
)

// DefaultClusterConditions is the default cluster conditions list
var DefaultClusterConditions = func() []*infrav3.ClusterCondition {
	var conditions []*infrav3.ClusterCondition

	var i int32 = 0
	for {
		if _, ok := infrav3.ClusterConditionType_name[i]; !ok {
			break
		}
		conditions = append(conditions, newClusterCondition(infrav3.ClusterConditionType(i))(commonv3.RafayConditionStatus_NotSet, "pending"))
		i++
	}

	return conditions
}()

func newClusterCondition(conditionType infrav3.ClusterConditionType) func(status commonv3.RafayConditionStatus, reason string) *infrav3.ClusterCondition {
	return func(status commonv3.RafayConditionStatus, reason string) *infrav3.ClusterCondition {
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

func isClusterCondition(conditionStatus commonv3.RafayConditionStatus, conditionTypes ...infrav3.ClusterConditionType) func(c *infrav3.Cluster) bool {
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
