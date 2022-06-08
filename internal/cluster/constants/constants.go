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
	RafayDomainLabel         = "rafay.dev"
	ClusterLabelKey          = "rafay.dev/clusterName"
	ClusterLocationKey       = "rafay.dev/clusterLocation"
	ClusterTypeKey           = "rafay.dev/clusterType"
	KubernetesVersionKey     = "rafay.dev/k8sVersion"
	ClusterGPU               = "rafay.dev/clusterGPU"
	ClusterGPUVendor         = "rafay.dev/clusterGPUVendor"
	ClusterUpgradeProtection = "rafay.dev/clusterUpgradeProtection"
	EdgeSuffix               = "EDGE_SUFFIX"
	EdgeCnameSuffix          = "EDGE_CNAME_SUFFIX"
	DefaultBlueprint         = "default"
	OverrideCluster          = "rafay.dev/overrideCluster"
	ClusterID                = "rafay.dev/clusterID"
	Public                   = "rafay.dev/public"
	ClusterName              = "rafay.dev/clusterName"
)

const (
	ApiVersion      = "infra.k8smgmt.io/v3"
	ClusterKind     = "Cluster"
	ClusterListKind = "ClusterList"
)

const (
	NotSet     = commonv3.RafayConditionStatus_NotSet
	Pending    = commonv3.RafayConditionStatus_Pending
	InProgress = commonv3.RafayConditionStatus_InProgress
	Success    = commonv3.RafayConditionStatus_Success
	Failed     = commonv3.RafayConditionStatus_Failed
	Retry      = commonv3.RafayConditionStatus_Retry
	Skipped    = commonv3.RafayConditionStatus_Skipped
	Stopped    = commonv3.RafayConditionStatus_Stopped
	Expired    = commonv3.RafayConditionStatus_Expired
	Stopping   = commonv3.RafayConditionStatus_Stopping
	Submitted  = commonv3.RafayConditionStatus_Submitted
)
