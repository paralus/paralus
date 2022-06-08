package converter

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// namespace gvk
	namespaceGVK = schema.GroupVersionKind{Version: "v1", Kind: "Namespace"}
	placementGVK = schema.GroupVersionKind{Group: "config.paralus.dev", Version: "v2", Kind: "Placement"}

	// namespace post install gvks
	limitRangeGVK    = schema.GroupVersionKind{Version: "v1", Kind: "LimitRange"}
	resourceQuotaGVK = schema.GroupVersionKind{Version: "v1", Kind: "ResourceQuota"}

	// task init gvks
	serviceAccountGVK = schema.GroupVersionKind{Version: "v1", Kind: "ServiceAccount"}
)

// IsListGVK checks if gvk is of List Kind
func IsListGVK(gvk schema.GroupVersionKind) bool {
	switch gvk.Group {
	case "":
		switch gvk.Kind {
		case "List":
			return true
		}
	}
	return false
}

// IsTaskInitGVK checks if given gvk should go into init phase
// of a Task
func IsTaskInitGVK(gvk schema.GroupVersionKind) bool {
	switch gvk.Group {
	case "rbac.authorization.k8s.io", "scheduling.k8s.io",
		"apiextensions.k8s.io", "policy",
		"admissionregistration.k8s.io", "storage.k8s.io",
		"cert-manager.io":
		return true
	}

	switch gvk {
	case serviceAccountGVK:
		return true
	}

	return false
}

// IsTaskletInitGVK checks if given gvk should go into init phase of a
// Tasklet
func IsTaskletInitGVK(gvk schema.GroupVersionKind) bool {
	switch gvk.Group {
	case "":
		switch gvk.Kind {
		case "Pod":
		default:
			return true
		}
	case "networking.k8s.io":
		return true
	}
	return false
}

// IsTaskletInstallGVK checks if given gvk should go into install phase of
// a Tasklet
func IsTaskletInstallGVK(gvk schema.GroupVersionKind) bool {
	switch gvk.Group {
	case "apps", "batch":
		return true
	case "extensions":
		switch gvk.Kind {
		case "Deployment", "DaemonSet":
			return true
		}
	case "":
		switch gvk.Kind {
		case "Pod":
			return true
		}
	}
	return false
}

// IsTaskletPostInstallGVK checks if given gvk should go into post install phase
// of a TaskSet
func IsTaskletPostInstallGVK(gvk schema.GroupVersionKind) bool {
	switch gvk.Group {
	case "autoscaling":
		return true
	case "extensions":
		switch gvk.Kind {
		case "NetworkPolicy", "Ingress":
			return true
		}
	}
	return false
}

// IsNamespaceGVK checks if given gvk is namespace
func IsNamespaceGVK(gvk schema.GroupVersionKind) bool {
	switch gvk {
	case namespaceGVK:
		return true
	}
	return false
}

// IsNamespacePostCreate checks if given gvk should go into namespace post install
func IsNamespacePostCreate(gvk schema.GroupVersionKind) bool {
	// switch gvk {
	// case limitRangeGVK, resourceQuotaGVK:
	// 	return true
	// }
	return false
}

// IsPlacementGVK checks gvk is placement
func IsPlacementGVK(gvk schema.GroupVersionKind) bool {
	switch gvk {
	case placementGVK:
		return true
	}
	return false
}

// IsPlacementGVK checks gvk is placement
func IsIngressGVK(gvk schema.GroupVersionKind) bool {
	switch gvk.Group {
	case "networking.k8s.io":
		switch gvk.Kind {
		case "Ingress":
			return true
		}
	case "extensions":
		switch gvk.Kind {
		case "Ingress":
			return true
		}
	}
	return false
}
