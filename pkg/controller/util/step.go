package util

import (
	clusterv2 "github.com/paralus/paralus/proto/types/controller"
)

const (
	crdKind                = "CustomResourceDefinition"
	clusterRoleKind        = "ClusterRole"
	clusterRoleBindingKind = "ClusterRoleBinding"
	roleBindingKind        = "RoleBinding"
	roleKind               = "Role"
)

// SetNamespace is a utility function setting namespace for a step objects
// it preserves namespace if already set for selective resources
func SetNamespace(so *clusterv2.StepObject, namespace string) error {
	accessor, err := so.Accessor()
	if err != nil {
		return err
	}

	kind, err := accessor.Kind()
	if err != nil {
		return err
	}

	switch kind {
	case crdKind, clusterRoleKind, clusterRoleBindingKind:
	case roleKind, roleBindingKind:
		ens, err := accessor.Namespace()
		if err != nil {
			return err
		}
		if ens == "" {
			accessor.SetNamespace(namespace)
		}
	default:
		accessor.SetNamespace(namespace)
	}

	accessor.ResetAutoFields()
	so.Raw = accessor.Bytes()

	return nil
}
