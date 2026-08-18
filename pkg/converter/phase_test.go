package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestIsListGVK(t *testing.T) {
	assert.True(t, IsListGVK(schema.GroupVersionKind{Kind: "List"}))
	assert.False(t, IsListGVK(schema.GroupVersionKind{Kind: "Pod"}))
	assert.False(t, IsListGVK(schema.GroupVersionKind{Group: "apps", Kind: "List"}))
}

func TestIsTaskInitGVK(t *testing.T) {
	tests := []struct {
		name string
		gvk  schema.GroupVersionKind
		want bool
	}{
		{"rbac group", schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Kind: "Role"}, true},
		{"scheduling group", schema.GroupVersionKind{Group: "scheduling.k8s.io", Kind: "PriorityClass"}, true},
		{"apiextensions group", schema.GroupVersionKind{Group: "apiextensions.k8s.io", Kind: "CustomResourceDefinition"}, true},
		{"policy group", schema.GroupVersionKind{Group: "policy", Kind: "PodDisruptionBudget"}, true},
		{"admissionregistration group", schema.GroupVersionKind{Group: "admissionregistration.k8s.io", Kind: "Webhook"}, true},
		{"storage group", schema.GroupVersionKind{Group: "storage.k8s.io", Kind: "StorageClass"}, true},
		{"cert-manager group", schema.GroupVersionKind{Group: "cert-manager.io", Kind: "Certificate"}, true},
		{"service account", schema.GroupVersionKind{Version: "v1", Kind: "ServiceAccount"}, true},
		{"unrelated gvk", schema.GroupVersionKind{Group: "apps", Kind: "Deployment"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsTaskInitGVK(tt.gvk))
		})
	}
}

func TestIsTaskletInitGVK(t *testing.T) {
	tests := []struct {
		name string
		gvk  schema.GroupVersionKind
		want bool
	}{
		{"core group non-pod", schema.GroupVersionKind{Group: "", Kind: "ConfigMap"}, true},
		{"core group pod excluded", schema.GroupVersionKind{Group: "", Kind: "Pod"}, false},
		{"networking group", schema.GroupVersionKind{Group: "networking.k8s.io", Kind: "Ingress"}, true},
		{"unrelated group", schema.GroupVersionKind{Group: "apps", Kind: "Deployment"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsTaskletInitGVK(tt.gvk))
		})
	}
}

func TestIsTaskletInstallGVK(t *testing.T) {
	tests := []struct {
		name string
		gvk  schema.GroupVersionKind
		want bool
	}{
		{"apps group", schema.GroupVersionKind{Group: "apps", Kind: "Deployment"}, true},
		{"batch group", schema.GroupVersionKind{Group: "batch", Kind: "Job"}, true},
		{"extensions deployment", schema.GroupVersionKind{Group: "extensions", Kind: "Deployment"}, true},
		{"extensions daemonset", schema.GroupVersionKind{Group: "extensions", Kind: "DaemonSet"}, true},
		{"extensions unrelated kind", schema.GroupVersionKind{Group: "extensions", Kind: "Ingress"}, false},
		{"core pod", schema.GroupVersionKind{Group: "", Kind: "Pod"}, true},
		{"core non-pod", schema.GroupVersionKind{Group: "", Kind: "ConfigMap"}, false},
		{"unrelated group", schema.GroupVersionKind{Group: "networking.k8s.io", Kind: "Ingress"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsTaskletInstallGVK(tt.gvk))
		})
	}
}

func TestIsTaskletPostInstallGVK(t *testing.T) {
	tests := []struct {
		name string
		gvk  schema.GroupVersionKind
		want bool
	}{
		{"autoscaling group", schema.GroupVersionKind{Group: "autoscaling", Kind: "HorizontalPodAutoscaler"}, true},
		{"extensions network policy", schema.GroupVersionKind{Group: "extensions", Kind: "NetworkPolicy"}, true},
		{"extensions ingress", schema.GroupVersionKind{Group: "extensions", Kind: "Ingress"}, true},
		{"extensions unrelated kind", schema.GroupVersionKind{Group: "extensions", Kind: "Deployment"}, false},
		{"unrelated group", schema.GroupVersionKind{Group: "apps", Kind: "Deployment"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsTaskletPostInstallGVK(tt.gvk))
		})
	}
}

func TestIsNamespaceGVK(t *testing.T) {
	assert.True(t, IsNamespaceGVK(schema.GroupVersionKind{Version: "v1", Kind: "Namespace"}))
	assert.False(t, IsNamespaceGVK(schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}))
}

func TestIsNamespacePostCreate(t *testing.T) {
	// Currently hard-coded to always return false (implementation is
	// commented out); documents the actual, present-day behavior.
	assert.False(t, IsNamespacePostCreate(schema.GroupVersionKind{Version: "v1", Kind: "Namespace"}))
}

func TestIsPlacementGVK(t *testing.T) {
	assert.True(t, IsPlacementGVK(schema.GroupVersionKind{Group: "config.paralus.dev", Version: "v2", Kind: "Placement"}))
	assert.False(t, IsPlacementGVK(schema.GroupVersionKind{Group: "config.paralus.dev", Version: "v2", Kind: "Other"}))
}

func TestIsIngressGVK(t *testing.T) {
	tests := []struct {
		name string
		gvk  schema.GroupVersionKind
		want bool
	}{
		{"networking.k8s.io ingress", schema.GroupVersionKind{Group: "networking.k8s.io", Kind: "Ingress"}, true},
		{"networking.k8s.io unrelated kind", schema.GroupVersionKind{Group: "networking.k8s.io", Kind: "NetworkPolicy"}, false},
		{"extensions ingress", schema.GroupVersionKind{Group: "extensions", Kind: "Ingress"}, true},
		{"extensions unrelated kind", schema.GroupVersionKind{Group: "extensions", Kind: "Deployment"}, false},
		{"unrelated group", schema.GroupVersionKind{Group: "apps", Kind: "Ingress"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsIngressGVK(tt.gvk))
		})
	}
}
