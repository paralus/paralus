package sentry

// paralus specific annotations/labels keys
const (
	ConfigV2Group = "config.paralus.dev/v2"
)

// kubectl/kubeconfig permissions
const (
	KubeconfigReadPermission        = "kubeconfig.read"
	KubectlFullAccessPermission     = "kubectl.fullaccess"
	KubectlClusterReadPermission    = "kubectl.cluster.read"
	KubectlClusterWritePermission   = "kubectl.cluster.write"
	KubectlNamespaceReadPermission  = "kubectl.namespace.read"
	KubectlNamespaceWritePermission = "kubectl.namespace.write"
)

// GetKubeConfigClusterPermissions list of kubeconfig permissions
func GetKubeConfigClusterPermissions() []string {
	return []string{
		KubeconfigReadPermission,
		KubectlFullAccessPermission,
		KubectlClusterReadPermission,
		KubectlClusterWritePermission,
	}
}

// GetKubeConfigNameSpacePermissions list of kubeconfig permissions
func GetKubeConfigNameSpacePermissions() []string {
	return []string{
		KubectlNamespaceReadPermission,
		KubectlNamespaceWritePermission,
	}
}

// GetKubeConfigPermissionIsRead is read permission
func GetKubeConfigPermissionIsRead(permission string) bool {
	switch permission {
	case KubeconfigReadPermission:
		return true
	case KubectlNamespaceReadPermission:
		return true
	case KubectlClusterReadPermission:
		return true
	}
	return false
}

// GetKubeConfigPermissionprivilege privilege order
func GetKubeConfigPermissionPrivilege(permission string) int {
	switch permission {
	case KubeconfigReadPermission:
		return 0
	case KubectlNamespaceReadPermission:
		return 1
	case KubectlNamespaceWritePermission:
		return 2
	case KubectlClusterReadPermission:
		return 3
	case KubectlClusterWritePermission:
		return 4
	case KubectlFullAccessPermission:
		return 5
	}
	return -1
}

// kubeconfig setting scope
const (
	KubeconfigSettingOrganizationScope = "ORGANIZATION"
	KubeconfigSettingUserScope         = "USER"
)

// Kind is kind of resource
type Kind = string

// available config kinds
const (
	PartnerKind Kind = "Partner"
)
