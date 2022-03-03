package authz

import (
	"encoding/json"

	"github.com/shurcooL/httpfs/vfsutil"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"sigs.k8s.io/yaml"
)

const (
	defaultClusterRolePath           = "relay_default_cluster_role.yaml"
	defaultRolePath                  = "relay_default_role.yaml"
	fullAccessClusterRolePath        = "cluster_role_full_access.yaml"
	readNamespaceClusterRolePath     = "cluster_role_namespace_read.yaml"
	writeNamespaceClusterRolePath    = "cluster_role_namespace_write.yaml"
	readClusterScopeClusterRolePath  = "cluster_role_cluster_read.yaml"
	writeClusterScopeClusterRolePath = "cluster_role_cluster_write.yaml"
	readNamespaceRolePath            = "role_read_access.yaml"
	writeNamespaceRolePath           = "role_write_access.yaml"
	nameSpacePath                    = "namespace.yaml"

	fullAccessClusterRoleName  = "full-access-cluster-role"
	readAccessClusterRoleName  = "read-access-cluster-role"
	writeAccessClusterRoleName = "write-access-cluster-role"
)

// GetDefaultClusterRole returns default cluster role for relay user
func GetDefaultClusterRole() (*rbacv1.ClusterRole, error) {
	return getClusterRoleFromFile(defaultClusterRolePath)
}

// GetDefaultRole return default role for relay user
func GetDefaultRole() (*rbacv1.Role, error) {
	return getRoleFromFile(defaultRolePath)
}

// GetFullAccessClusterRole gets cluster role with full access
func GetFullAccessClusterRole() (*rbacv1.ClusterRole, error) {
	return getClusterRoleFromFile(fullAccessClusterRolePath)
}

// GetReadNamespaceClusterRole gets cluster role with read access
func GetReadNamespaceClusterRole() (*rbacv1.ClusterRole, error) {
	return getClusterRoleFromFile(readNamespaceClusterRolePath)
}

// GetWriteNamespaceClusterRole gets cluster role with write access
func GetWriteNamespaceClusterRole() (*rbacv1.ClusterRole, error) {
	return getClusterRoleFromFile(writeNamespaceClusterRolePath)
}

// GetReadClusterScopeClusterRole gets cluster role with read access
func GetReadClusterScopeClusterRole() (*rbacv1.ClusterRole, error) {
	return getClusterRoleFromFile(readClusterScopeClusterRolePath)
}

// GetWriteClusterScopeClusterRole gets cluster role with write access
func GetWriteClusterScopeClusterRole() (*rbacv1.ClusterRole, error) {
	return getClusterRoleFromFile(writeClusterScopeClusterRolePath)
}

// GetReadNamespaceRole gets cluster role with read access
func GetReadNamespaceRole() (*rbacv1.Role, error) {
	return getRoleFromFile(readNamespaceRolePath)
}

// GetWriteNamespaceRole gets cluster role with write access
func GetWriteNamespaceRole() (*rbacv1.Role, error) {
	return getRoleFromFile(writeNamespaceRolePath)
}

// GetNamespace gets namespace
func GetNamespace() (*corev1.Namespace, error) {
	return getNameSpaceFromFile(nameSpacePath)
}

func getClusterRoleFromFile(path string) (*rbacv1.ClusterRole, error) {
	yb, err := vfsutil.ReadFile(defaults, path)
	if err != nil {
		return nil, err
	}
	jb, err := yaml.YAMLToJSON(yb)
	if err != nil {
		return nil, err
	}

	var cr rbacv1.ClusterRole
	err = json.Unmarshal(jb, &cr)
	if err != nil {
		return nil, err
	}

	return &cr, nil
}

func getRoleFromFile(path string) (*rbacv1.Role, error) {
	yb, err := vfsutil.ReadFile(defaults, path)
	if err != nil {
		return nil, err
	}
	jb, err := yaml.YAMLToJSON(yb)
	if err != nil {
		return nil, err
	}

	var r rbacv1.Role
	err = json.Unmarshal(jb, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func getNameSpaceFromFile(path string) (*corev1.Namespace, error) {
	yb, err := vfsutil.ReadFile(defaults, path)
	if err != nil {
		return nil, err
	}
	jb, err := yaml.YAMLToJSON(yb)
	if err != nil {
		return nil, err
	}

	var n corev1.Namespace
	err = json.Unmarshal(jb, &n)
	if err != nil {
		return nil, err
	}

	return &n, nil
}
