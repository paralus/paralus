package controller

const (
	// PreDeleteFinalizer is the finalizer for all cluster CRD pre delete
	PreDeleteFinalizer = "cluster.paralus.dev.v2.predelete"
	// OwnerRef is set if a kubernetes resource is owned by paralus cluster controllers
	// this is used in place of k8s owner ref to enable managing namespaced/non namespaced
	// resources across namespaces
	OwnerRef = "paralus.dev/ownerRef"

	// PrunedSteps is the annotation describing steps pruned from the object
	PrunedSteps = "paralus.dev/pruned"

	// OrignalConfig is the annotation which stores the last applied config (this is equivalent to kubectl last applied config)
	// on an k8s resource
	// this is used to caliculating 3 way merge patches
	// Note: Paralus CRDs are not patched, they are updated and all the resouces created through them are patched
	OrignalConfig = "paralus.dev/original"
)

// StepOnFailed determines what should be done when execution of a step fails
// +kubebuilder:validation:Enum=StepBreak;StepContinue
type StepOnFailed string

// StepObjectState is the state of the object described in the step
// +kubebuilder:validation:Enum=StepObjectNotCreated;StepObjectCreated;StepObjectFailed;StepObjectComplete
type StepObjectState string

// StepJobState is the state of the job described in the step
// +kubebuilder:validation:Enum=StepJobNotCreated;StepJobCreated;StepJobFailed;StepJobComplete
type StepJobState string

// StepState is the aggregate state of the step
// +kubebuilder:validation:Enum=StepNotExecuted;StepExecuted;StepFailed;StepComplete
type StepState string

// ConditionStatus is the status of condition
// +kubebuilder:validation:Enum=Pending;InProgress;Configured;Complete;Failed
type ConditionStatus string

// NamespaceConditionType is the condition type of namespace
// +kubebuilder:validation:Enum=NamespaceInit;NamespaceCreate;NamespacePostCreate;NamespacePreDelete;NamespaceReady
type NamespaceConditionType string

// TaskConditionType is the condition type of task
// +kubebuilder:validation:Enum=TaskInit;TaskletCreate;TaskPreDelete;TaskReady
type TaskConditionType string

// TaskletConditionType is the condition type of task
// +kubebuilder:validation:Enum=TaskletInit;TaskletInstall;TaskletPostInstall;TaskletPreDelete;TaskletReady
type TaskletConditionType string

// NodeConditionType is the condition type of node
// +kubebuilder:validation:Enum=TaskletInit;TaskletInstall;TaskletPostInstall;TaskletPreDelete;TaskletReady
type NodeConditionType string

// enum values types
const (
	StepBreak    StepOnFailed = "StepBreak"
	StepContinue StepOnFailed = "StepContinue"

	StepObjectNotCreated StepObjectState = "StepObjectNotCreated"
	StepObjectCreated    StepObjectState = "StepObjectCreated"
	StepObjectFailed     StepObjectState = "StepObjectFailed"
	StepObjectComplete   StepObjectState = "StepObjectComplete"
	StepObjectRetry      StepObjectState = "StepObjectRetry"

	StepJobNotCreated StepJobState = "StepJobNotCreated"
	StepJobCreated    StepJobState = "StepJobCreated"
	StepJobFailed     StepJobState = "StepJobFailed"
	StepJobComplete   StepJobState = "StepJobComplete"

	StepNotExecuted StepState = "StepNotExecuted"
	StepExecuted    StepState = "StepExecuted"
	StepFailed      StepState = "StepFailed"
	StepComplete    StepState = "StepComplete"

	Pending    ConditionStatus = "Pending"
	InProgress ConditionStatus = "InProgress"
	Configured ConditionStatus = "Configured"
	Complete   ConditionStatus = "Complete"
	Failed     ConditionStatus = "Failed"

	NamespaceUpsert     NamespaceConditionType = "NamespaceUpsert"
	NamespaceInit       NamespaceConditionType = "NamespaceInit"
	NamespaceCreate     NamespaceConditionType = "NamespaceCreate"
	NamespacePostCreate NamespaceConditionType = "NamespacePostCreate"
	NamespacePreDelete  NamespaceConditionType = "NamespacePreDelete"
	NamespaceReady      NamespaceConditionType = "NamespaceReady"

	TaskUpsert    TaskConditionType = "TaskUpsert"
	TaskInit      TaskConditionType = "TaskInit"
	TaskletCreate TaskConditionType = "TaskletCreate"
	TaskPreDelete TaskConditionType = "TaskPreDelete"
	TaskReady     TaskConditionType = "TaskReady"

	TaskletUpsert      TaskletConditionType = "TaskletUpsert"
	TaskletInit        TaskletConditionType = "TaskletInit"
	TaskletInstall     TaskletConditionType = "TaskletInstall"
	TaskletPostInstall TaskletConditionType = "TaskletPostInstall"
	TaskletPreDelete   TaskletConditionType = "TaskletPreDelete"
	TaskletReady       TaskletConditionType = "TaskletReady"

	// NodeReady means kubelet is healthy and ready to accept pods.
	NodeReady NodeConditionType = "Ready"
	// NodeOutOfDisk means the kubelet will not accept new pods due to insufficient free disk
	// space on the node.
	NodeOutOfDisk NodeConditionType = "OutOfDisk"
	// NodeMemoryPressure means the kubelet is under pressure due to insufficient available memory.
	NodeMemoryPressure NodeConditionType = "MemoryPressure"
	// NodeDiskPressure means the kubelet is under pressure due to insufficient available disk.
	NodeDiskPressure NodeConditionType = "DiskPressure"
	// NodeNetworkUnavailable means that network for the node is not correctly configured.
	NodeNetworkUnavailable NodeConditionType = "NetworkUnavailable"
)
