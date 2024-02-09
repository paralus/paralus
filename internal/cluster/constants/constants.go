package constants

import (
	"regexp"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
)

type ClusterGeneration int

const (
	Cluster_Verr ClusterGeneration = 0
	Cluster_V1   ClusterGeneration = 1
	Cluster_V2   ClusterGeneration = 2
)

var KubernetesLabelNameRegex = regexp.MustCompile(`^[A-Za-z0-9]([\.A-Za-z0-9_-]*[A-Za-z0-9])?$`)

const (
	ParalusDomainLabel       = "paralus.dev"
	ClusterLabelKey          = "paralus.dev/clusterName"
	ClusterLocationKey       = "paralus.dev/clusterLocation"
	ClusterTypeKey           = "paralus.dev/clusterType"
	KubernetesVersionKey     = "paralus.dev/k8sVersion"
	ClusterGPU               = "paralus.dev/clusterGPU"
	ClusterGPUVendor         = "paralus.dev/clusterGPUVendor"
	ClusterUpgradeProtection = "paralus.dev/clusterUpgradeProtection"
	EdgeSuffix               = "EDGE_SUFFIX"
	EdgeCnameSuffix          = "EDGE_CNAME_SUFFIX"
	DefaultBlueprint         = "default"
	OverrideCluster          = "paralus.dev/overrideCluster"
	ClusterID                = "paralus.dev/clusterID"
	Public                   = "paralus.dev/public"
	ClusterName              = "paralus.dev/clusterName"
)

const (
	ApiVersion      = "infra.k8smgmt.io/v3"
	ClusterKind     = "Cluster"
	ClusterListKind = "ClusterList"
)

const (
	NotSet       = commonv3.ParalusConditionStatus_NotSet
	Pending      = commonv3.ParalusConditionStatus_Pending
	InProgress   = commonv3.ParalusConditionStatus_InProgress
	Success      = commonv3.ParalusConditionStatus_Success
	Failed       = commonv3.ParalusConditionStatus_Failed
	Retry        = commonv3.ParalusConditionStatus_Retry
	Skipped      = commonv3.ParalusConditionStatus_Skipped
	Stopped      = commonv3.ParalusConditionStatus_Stopped
	Expired      = commonv3.ParalusConditionStatus_Expired
	Stopping     = commonv3.ParalusConditionStatus_Stopping
	Submitted    = commonv3.ParalusConditionStatus_Submitted
	Healthy      = commonv3.ParalusConditionStatus_Healthy
	Unhealthy    = commonv3.ParalusConditionStatus_Unhealthy
	Disconnected = commonv3.ParalusConditionStatus_Disconnected
)
