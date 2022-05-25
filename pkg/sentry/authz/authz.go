package authz

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RafayLabs/rcloud-base/internal/constants"
	"github.com/RafayLabs/rcloud-base/pkg/controller/runtime"
	"github.com/RafayLabs/rcloud-base/pkg/log"
	"github.com/RafayLabs/rcloud-base/pkg/query"
	"github.com/RafayLabs/rcloud-base/pkg/sentry/kubeconfig"
	"github.com/RafayLabs/rcloud-base/pkg/service"
	sentryrpc "github.com/RafayLabs/rcloud-base/proto/rpc/sentry"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	"github.com/RafayLabs/rcloud-base/proto/types/controller"
	"github.com/RafayLabs/rcloud-base/proto/types/sentry"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

var _log = log.GetLogger()

var permissions = []string{
	sentry.KubectlFullAccessPermission,
	sentry.KubectlNamespaceReadPermission,
	sentry.KubectlNamespaceWritePermission,
	sentry.KubectlClusterReadPermission,
	sentry.KubectlClusterWritePermission,
}

var clusterScopePermissions = []string{
	sentry.KubectlClusterReadPermission,
	sentry.KubectlClusterWritePermission,
	sentry.KubectlFullAccessPermission,
}

var namespaceScopePermissions = []string{
	sentry.KubectlNamespaceReadPermission,
	sentry.KubectlNamespaceWritePermission,
}

const (
	rafayRelayLabel     = "rafay-relay"
	relayUserLabel      = "relay-user"
	authzRefreshedLabel = "authz-refreshed"
	systemUsername      = "admin@rafay.co"
)

type roleBindExclusionList struct {
	exclude   bool
	namespace string
}

func getCurrentEpoch() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func getAuthzLabels(userName string) map[string]string {
	return map[string]string{
		rafayRelayLabel:     "true",
		relayUserLabel:      userName,
		authzRefreshedLabel: getCurrentEpoch(),
	}
}

/*TODO: pending along with namespaces
func getAccountProjectNamespace(ctx context.Context, projectID, accountID, orgID int64, apn service.AccountProjectNamespaceService) ([]string, error) {
	var ns []string

	apns, err := apn.GetAccountProjectNamesapce(ctx, orgID, accountID, projectID)
	if err != nil {
		return nil, err
	}

	for _, apn := range apns {
		ns = append(ns, apn.NamespaceName)
	}

	return ns, nil
}

func getSSOAccountProjectNamespace(ctx context.Context, projectID, accountID, orgID int64, apn service.AccountProjectNamespaceService) ([]string, error) {
	var ns []string

	apns, err := apn.GetSSOAccountProjectNamesapce(ctx, orgID, accountID, projectID)
	if err != nil {
		return nil, err
	}

	for _, apn := range apns {
		ns = append(ns, apn.NamespaceName)
	}

	return ns, nil
}

func getGroupAccountProjectNamespace(ctx context.Context, projectID, accountID, orgID int64, apn service.AccountProjectNamespaceService) ([]string, error) {
	var ns []string

	apns, err := apn.GetGroupProjectNamesapce(ctx, orgID, accountID, projectID)
	if err != nil {
		return nil, err
	}

	for _, apn := range apns {
		ns = append(ns, apn.NamespaceName)
	}

	return ns, nil
}
*/

func getProjectPermissions(ctx context.Context, projects []string, accountID, orgID, partnerID string, aps service.AccountPermissionService) (map[string][]string, string, error) {
	projects = append(projects, "")
	accountPermissions, err := aps.GetAccountPermissionsByProjectIDPermissions(ctx, accountID, orgID, partnerID, projects, permissions)
	if err != nil {
		return nil, "", err
	}

	accountData, err := aps.GetAccount(ctx, accountID)
	if err != nil {
		return nil, "", err
	}

	projectPermissions := make(map[string][]string)
OUTER:
	for _, accountPermission := range accountPermissions {
		p := accountPermission.ProjectID
		if projectPermissions[p] == nil {
			projectPermissions[p] = []string{}
		}
		for _, permission := range projectPermissions[p] {
			if permission == accountPermission.PermissionName {
				continue OUTER
			}
		}
		projectPermissions[p] = append(projectPermissions[p], accountPermission.PermissionName)
	}
	return projectPermissions, accountData.Username, nil
}

func getSSOProjectPermissions(ctx context.Context, projects []string, orgID, partnerID, accountID string, aps service.AccountPermissionService, gps service.GroupPermissionService) (map[string][]string, string, []string, error) {
	acc, err := aps.GetAccount(ctx, accountID)
	if err != nil {
		return nil, "", nil, err
	}
	groups, err := aps.GetAccountGroups(ctx, accountID)
	if err != nil {
		return nil, "", nil, err
	}
	projects = append(projects, "")
	groupPermissions, err := gps.GetGroupPermissionsByProjectIDPermissions(ctx, groups, orgID, partnerID, projects, permissions)
	if err != nil {
		return nil, "", nil, err
	}

	projectPermissions := make(map[string][]string)
OUTER:
	for _, groupPermission := range groupPermissions {
		p := groupPermission.ProjectID
		if projectPermissions[p] == nil {
			projectPermissions[p] = []string{}
		}
		for _, permission := range projectPermissions[p] {
			if permission == groupPermission.PermissionName {
				continue OUTER
			}
		}
		projectPermissions[p] = append(projectPermissions[p], groupPermission.PermissionName)
	}
	return projectPermissions, acc.Username, groups, nil
}

func getClusterRole(permission string) (cr *rbacv1.ClusterRole, err error) {
	switch permission {
	case sentry.KubectlFullAccessPermission:
		cr, err = GetFullAccessClusterRole()
	case sentry.KubectlNamespaceWritePermission:
		cr, err = GetWriteNamespaceClusterRole()
	case sentry.KubectlNamespaceReadPermission:
		cr, err = GetReadNamespaceClusterRole()
	case sentry.KubectlClusterWritePermission:
		cr, err = GetWriteClusterScopeClusterRole()
	case sentry.KubectlClusterReadPermission:
		cr, err = GetReadClusterScopeClusterRole()
	default:
		err = fmt.Errorf("permission not valid - %s", permission)
	}

	if cr != nil {

	}

	return
}

func getRole(permission string) (r *rbacv1.Role, err error) {
	switch permission {
	case sentry.KubectlNamespaceWritePermission:
		r, err = GetWriteNamespaceRole()
	case sentry.KubectlNamespaceReadPermission:
		r, err = GetReadNamespaceRole()
	default:
		err = fmt.Errorf("permission not valid - %s", permission)
	}

	return
}

func getRoleName(nsName, permission string) string {
	switch permission {
	case sentry.KubectlNamespaceWritePermission:
		return "rafay-ns-role-write-" + nsName
	case sentry.KubectlNamespaceReadPermission:
		return "rafay-ns-role-read-" + nsName
	default:
		_log.Infow("getRoleName invalid namespace", "permission", permission)
	}
	return ""
}

func setRoleValues(r *rbacv1.Role, nsName, permission string) {
	r.Name = getRoleName(nsName, permission)
	r.Namespace = nsName
}

func getClusterRoleBindingName(saName, clusterRole string) string {
	return clusterRole + "-" + saName + "-cr-binding"
}

func getClusterRoleBinding(sa *corev1.ServiceAccount, clusterRole string) *rbacv1.ClusterRoleBinding {
	crb := &rbacv1.ClusterRoleBinding{}
	crb.APIVersion = "rbac.authorization.k8s.io/v1"
	crb.Kind = "ClusterRoleBinding"
	crb.Name = getClusterRoleBindingName(sa.Name, clusterRole)
	// crb.Labels = map[string]string{
	// 	"rafay-relay": "true",
	// 	"relay-user":  sa.Name,
	// }
	subject := rbacv1.Subject{}
	subject.Kind = "ServiceAccount"
	subject.Name = sa.Name
	subject.Namespace = sa.Namespace
	crb.Subjects = append(crb.Subjects, subject)

	crb.RoleRef.Kind = "ClusterRole"
	crb.RoleRef.APIGroup = "rbac.authorization.k8s.io"
	crb.RoleRef.Name = clusterRole

	return crb
}

func getDeleteClusterRoleBinding(name string) *rbacv1.ClusterRoleBinding {
	crb := &rbacv1.ClusterRoleBinding{}
	crb.APIVersion = "rbac.authorization.k8s.io/v1"
	crb.Kind = "ClusterRoleBinding"
	crb.Name = name
	return crb
}

func getRoleBindingName(saName, roleName string) string {
	return roleName + "-" + saName + "-r-binding"
}

func getRoleBinding(sa *corev1.ServiceAccount, roleName, namespace string) *rbacv1.RoleBinding {
	rb := &rbacv1.RoleBinding{}
	rb.APIVersion = "rbac.authorization.k8s.io/v1"
	rb.Kind = "RoleBinding"
	rb.Name = getRoleBindingName(sa.Name, roleName)
	rb.Namespace = namespace
	// rb.Labels = map[string]string{
	// 	"rafay-relay": "true",
	// 	"relay-user":  sa.Name,
	// }
	subject := rbacv1.Subject{}
	subject.Kind = "ServiceAccount"
	subject.Name = sa.Name
	subject.Namespace = sa.Namespace
	rb.Subjects = append(rb.Subjects, subject)

	rb.RoleRef.Kind = "Role"
	rb.RoleRef.APIGroup = "rbac.authorization.k8s.io"
	rb.RoleRef.Name = roleName

	return rb
}

func getDeleteRoleBinding(name, namespace string) *rbacv1.RoleBinding {
	rb := &rbacv1.RoleBinding{}
	rb.APIVersion = "rbac.authorization.k8s.io/v1"
	rb.Kind = "RoleBinding"
	rb.Name = name
	rb.Namespace = namespace
	return rb
}

func getProjectsFromLabels(labels map[string]string) ([]string, error) {
	projects := make([]string, 0)
	for key := range labels {
		if !strings.HasPrefix(key, "project/") {
			continue
		}
		s := strings.Split(key, "/")
		if len(s) != 2 {
			continue
		}
		projectID := s[1]
		projects = append(projects, projectID)
	}
	return projects, nil
}

// GetAuthorization returns authorization for user, cluster
// The RBAC model mapped to the existing role
// PROJECT_ADMIN:
//   -   Read/Write access to all cluster scoped resources
//   -   Read/Write access to all namespace scoped resources
// PROJECT_READ:
//   -   Read access to all cluster scoped resources
//   -   Read access to all namespace scoped resources
// INFRA_ADMIN:
//   -   Read/Write access to all cluster scoped resources
//   -   Read/Write access to all namespace scoped resources
// INFRA_READ:
//   -   Read access to all cluster scoped resources
//   -   Read access to all namespace scoped resources
// ENV_ADMIN
//   -   NO Access to cluster scoped resources
//   -   Read/Write Access to namespace scoped resources (only within the environment)
// ENV_READ
//   -   NO Access to cluster scoped resources
//   -   Read Access to namespace scoped resources (only within the environment)
func GetAuthorization(ctx context.Context, req *sentryrpc.GetUserAuthorizationRequest, bs service.BootstrapService, aps service.AccountPermissionService, gps service.GroupPermissionService, krs service.KubeconfigRevocationService, kcs service.KubectlClusterSettingsService, kss service.KubeconfigSettingService) (resp *sentryrpc.GetUserAuthorizationResponse, err error) {
	var userName string
	var groups []string
	var rolePrevilage int
	var highestRole string
	var enforceOrgAdminOnlySecretAccess, isOrgAdmin bool

	resp = new(sentryrpc.GetUserAuthorizationResponse)

	// get attributes from user CN
	cnAttr := kubeconfig.GetCNAttributes(req.UserCN)
	accountID := cnAttr.AccountID
	orgID := cnAttr.OrganizationID
	partnerID := cnAttr.PartnerID

	if cnAttr.SystemUser {
		return getSystemUserAuthz(cnAttr)
	}

	isOrgAdmin, _ = aps.IsOrgAdmin(ctx, accountID, partnerID)

	// Check user is partner / super admin to bypass cluster/user checks.
	// Partner Super admins has full access.
	isPartnerAdmin, isSuperAdmin, err := aps.IsPartnerSuperAdmin(ctx, accountID, partnerID)
	if err != nil {
		_log.Infow("Error getting partner/super admin permission info", "accountID", accountID, "orgID", orgID, "partnerID", partnerID, "error", err)
	}
	_log.Infow("check for partner/super admin", " isPartnerAdmin ", isPartnerAdmin, " isSuperAdmin ", isSuperAdmin)
	if !isSuperAdmin && !isPartnerAdmin {
		existUserLevel := true
		// get org level setting if exist
		ksOrg, errOrg := kss.Get(ctx, orgID, "", cnAttr.IsSSO)
		if errOrg != nil && errOrg != constants.ErrNotFound {
			_log.Errorw("failed to fetch organization level kubectl settings for", "userCN", req.UserCN)
			return nil, errOrg
		}
		if errOrg == nil && ksOrg != nil {
			// check for kubectl org settings
			err = verifyKubectlSettings(cnAttr, ksOrg, "organization")
			if err != nil {
				_log.Errorw("kubectl denied as per org level kubectl settings for", "userCN", req.UserCN)
				return nil, err
			}
			enforceOrgAdminOnlySecretAccess = ksOrg.EnforceOrgAdminSecretAccess
		}

		// check for kubectl cluster settings
		err = verifyClusterKubectlSettings(ctx, bs, kcs, cnAttr, req.ClusterID, orgID)
		if err != nil {
			_log.Errorw("failed to verify kubectl cluster settings for", "userCN", req.UserCN)
			return nil, err
		}

		// get user level settings if exist
		ks, errUser := kss.Get(ctx, orgID, accountID, cnAttr.IsSSO)
		if errUser == constants.ErrNotFound {
			existUserLevel = false
			// set org settings if exist
			errUser = errOrg
			ks = ksOrg
		} else if errUser != nil && errUser != constants.ErrNotFound {
			return nil, errUser
		}

		if existUserLevel && ks != nil {
			// check for kubectl user settings
			errVerify := verifyKubectlSettings(cnAttr, ks, "user")
			if errVerify != nil {
				_log.Errorw("kubectl denied as per user level kubectl settings for", "userCN", req.UserCN)
				return nil, errVerify
			}
		}

		if errUser == nil && ks != nil && ks.EnableSessionCheck {
			// check the last login timestamp
			var lastLogin time.Time
			if cnAttr.IsSSO {
				accountData, err := aps.GetAccount(ctx, accountID)
				if err != nil {
					return nil, err
				}
				lastLogin = accountData.LastLogin
			} else {
				accountData, err := aps.GetAccount(ctx, accountID)
				if err != nil {
					return nil, err
				}
				lastLogin = accountData.LastLogin
			}
			t1 := time.Now()
			if t1.Sub(lastLogin) > time.Hour*12 {
				_log.Infow("get kubectl authorization block access. user did not login to portal in last 12 Hour")
				return nil, fmt.Errorf("enforce session enabled. user did not login to portal in last 12 Hour")
			}
		}

		// is user active
		if !cnAttr.IsSSO {
			active, err := aps.IsAccountActive(ctx, accountID, orgID)
			if err != nil {
				return nil, err
			}
			if !active {
				return nil, fmt.Errorf("kubeconfig user deactivated")
			}
		}

		// get revocation timestamp
		kr, err := krs.Get(ctx, orgID, accountID, cnAttr.IsSSO)
		if err != nil && err != constants.ErrNotFound {
			return nil, err
		} else if err == nil && kr.RevokedAt.AsTime().Unix() >= req.CertIssueSeconds {
			return nil, fmt.Errorf("kubeconfig revoked")
		}
	}

	opts := commonv3.QueryOptions{
		Name:         req.ClusterID,
		Organization: orgID,
		Partner:      partnerID,
	}
	bal, err := bs.GetBootstrapAgents(ctx, "-",
		query.WithOptions(&opts),
		// ignore project id because kubeconfig is not project scoped
		query.WithIgnoreScopeDefault(),
	)
	if err != nil {
		return nil, err
	}
	if bal.Metadata.Count <= 0 {
		return nil, fmt.Errorf("no bootstrap agents found")
	}
	ba := bal.Items[0]
	labels := ba.Metadata.GetLabels()

	// get projects
	projects, err := getProjectsFromLabels(labels)
	if err != nil {
		_log.Errorw("error getting projects from bootstrap agents labels", "labels", labels, "error", err.Error())
		return nil, err
	}

	// get permissions in the cluster's projects
	var projectPermissions map[string][]string
	if !cnAttr.IsSSO {
		projectPermissions, userName, err = getProjectPermissions(ctx, projects, accountID, orgID, partnerID, aps)
	} else {
		projectPermissions, userName, groups, err = getSSOProjectPermissions(ctx, projects, orgID, partnerID, accountID, aps, gps)
	}
	if err != nil {
		_log.Errorw("error getting project permission", "projects", projects, "userCN", req.UserCN, "error", err.Error())
		return nil, err
	}

	// get sa, clusterroles, roles, bindings
	sa := &corev1.ServiceAccount{}
	sa.APIVersion = "v1"
	sa.Kind = "ServiceAccount"
	sa.Name = cnAttr.Username
	sa.Namespace = "rafay-system"

	crMap := make(map[string]*rbacv1.ClusterRole)
	crbMap := make(map[string]*rbacv1.ClusterRoleBinding)
	rMap := make(map[string]*rbacv1.Role)
	rbMap := make(map[string]*rbacv1.RoleBinding)
	nsMap := make(map[string]*corev1.Namespace)
	crbExclusionMap := make(map[string]bool)
	rbExclusionMap := make(map[string]*roleBindExclusionList)

	projectNamespaces := make([]string, 0)
	/*TODO: pending with namespaces
	projectNamespaces, err := func() ([]string, error) {
		var nsl []string
		configClient, err := cPool.NewClient(ctx)
		if err != nil {
			err = errors.Wrap(err, "unable to get config client")
			return nil, err
		}
		defer configClient.Close()

		for _, project := range projects {
			namespaces, err := configClient.GetNamespaces(ctx, &configrpc.GetAllNamespacesRequest{
				QueryOptions: commonv3.QueryOptions{
					Project:      project,
					Organization: orgID,
					Partner:      partnerID,
				},
			})

			if err == nil {
				_log.Debugw("Get namespaces ", "orgID", orgID, "partnerID", partnerID, "project", project, "namespaces", namespaces.Items, "itemslen", len(namespaces.Items))
				for _, namespace := range namespaces.Items {
					nsl = append(nsl, namespace.Name)
				}
			}
		}
		return nsl, nil
	}()

	if err != nil {
		_log.Infow("unable to get project namespaces", "error", err)
		return nil, err
	}

	_log.Debugw("projectNamespaces", "names", projectNamespaces)
	*/

	for _, pm := range sentry.GetKubeConfigClusterPermissions() {
		cr, err := getClusterRole(pm)
		if err != nil {
			continue
		}
		crbName := getClusterRoleBindingName(sa.Name, cr.Name)
		crbExclusionMap[crbName] = true
	}

	for _, pm := range sentry.GetKubeConfigNameSpacePermissions() {
		if len(projectNamespaces) > 0 {
			for _, nsName := range projectNamespaces {
				roleName := getRoleName(nsName, pm)
				rbName := getRoleBindingName(sa.Name, roleName)
				rbExclusionMap[rbName] = &roleBindExclusionList{true, nsName}
			}
		}
	}

	rolePrevilage = -1
	for project, permissions := range projectPermissions {
		var namespaces []string
		_log.Infow("authorization", "project", project, "user", sa.Name, "permissions", permissions)
		groups = append(groups, permissions...)
		/* TODO: pending with namespaces
		// need to get the namesapces assigned to this user.
		if !cnAttr.IsSSO {
			ns1, _ := getAccountProjectNamespace(ctx, project, accountID, orgID, apn)
			ns2, _ := getGroupAccountProjectNamespace(ctx, project, accountID, orgID, apn)
			if len(ns1) > 0 {
				namespaces = append(namespaces, ns1...)
			}
			if len(ns2) > 0 {
				namespaces = append(namespaces, ns2...)
			}
			_log.Infow("namespaces", "project", project, "accountID", accountID, "orgID", orgID, "namespaces", namespaces)
		} else {
			namespaces, _ = getSSOAccountProjectNamespace(ctx, project, accountID, orgID, apn)
			_log.Infow("namespacesSSO", "project", project, "accountID", accountID, "orgID", orgID, "namespaces", namespaces)
		}
		*/
		// org scope
		if project == "" {
			for _, permission := range permissions {
				cr, err := getClusterRole(permission)
				if err != nil {
					return nil, err
				}

				rp := sentry.GetKubeConfigPermissionPrivilage(permission)
				if rp > rolePrevilage {
					rolePrevilage = rp
					highestRole = permission
				}

				crb := getClusterRoleBinding(sa, cr.Name)
				crMap[cr.Name] = cr
				crbMap[crb.Name] = crb
				crbExclusionMap[crb.Name] = false
			}
			break
		}
		for _, permission := range permissions {

			rp := sentry.GetKubeConfigPermissionPrivilage(permission)
			if rp > rolePrevilage {
				rolePrevilage = rp
				highestRole = permission
			}

			if isClusterScopePermission(permission) {
				cr, err := getClusterRole(permission)
				if err != nil {
					return nil, err
				}
				crb := getClusterRoleBinding(sa, cr.Name)
				crMap[cr.Name] = cr
				crbMap[crb.Name] = crb
				crbExclusionMap[crb.Name] = false
			} else if isNamespaceScopePermission(permission) {
				for _, namespace := range namespaces {
					ns, err := GetNamespace()
					if err != nil {
						return nil, err
					}
					ns.Name = namespace
					nsMap[namespace] = ns

					r, err := getRole(permission)
					if err != nil {
						return nil, err
					}
					setRoleValues(r, namespace, permission)
					rb := getRoleBinding(sa, r.Name, namespace)
					rMap[r.Name] = r
					rbMap[rb.Name] = rb
					rbExclusionMap[rb.Name] = &roleBindExclusionList{false, namespace}
				}
			}
		}

	}

	// add authz labels
	authzLabels := getAuthzLabels(cnAttr.Username)

	sa.Labels = authzLabels
	for k := range crMap {
		crMap[k].Labels = authzLabels
	}
	for k := range crbMap {
		crbMap[k].Labels = authzLabels
	}

	for k := range rMap {
		rMap[k].Labels = authzLabels
	}
	for k := range rbMap {
		rbMap[k].Labels = authzLabels
	}

	// convert to step objects
	saObject, err := runtime.FromObject(sa)
	if err != nil {
		return nil, err
	}
	resp.ServiceAccount = saObject

	for _, cr := range crMap {
		crObject, err := runtime.FromObject(cr)
		if err != nil {
			return nil, err
		}
		resp.ClusterRoles = append(resp.ClusterRoles, crObject)
	}

	for _, crb := range crbMap {
		crbObject, err := runtime.FromObject(crb)
		if err != nil {
			return nil, err
		}
		resp.ClusterRoleBindings = append(resp.ClusterRoleBindings, crbObject)
	}

	for _, r := range rMap {
		rObject, err := runtime.FromObject(r)
		if err != nil {
			return nil, err
		}
		resp.Roles = append(resp.Roles, rObject)
	}

	for _, ns := range nsMap {
		nObject, err := runtime.FromObject(ns)
		if err != nil {
			return nil, err
		}
		resp.Namespaces = append(resp.Namespaces, nObject)
	}

	for _, rb := range rbMap {
		rbObject, err := runtime.FromObject(rb)
		if err != nil {
			return nil, err
		}
		resp.RoleBindings = append(resp.RoleBindings, rbObject)
	}

	for dcrbName, val := range crbExclusionMap {
		if val {
			crbObject, err := runtime.FromObject(getDeleteClusterRoleBinding(dcrbName))
			if err == nil {
				resp.DeleteClusterRoleBindings = append(resp.DeleteClusterRoleBindings, crbObject)
			}
		}
	}

	for drbName, val := range rbExclusionMap {
		if val.exclude {
			rbObject, err := runtime.FromObject(getDeleteRoleBinding(drbName, val.namespace))
			if err == nil {
				resp.DeleteRoleBindings = append(resp.DeleteRoleBindings, rbObject)
			}
		}
	}

	resp.UserName = cnAttr.Username
	resp.RoleName = highestRole
	resp.IsRead = sentry.GetKubeConfigPermissionIsRead(highestRole)
	resp.EnforceOrgAdminOnlySecretAccess = enforceOrgAdminOnlySecretAccess
	resp.IsOrgAdmin = isOrgAdmin

	//to be removed along with events
	_log.Infow("username", userName)
	/*TODO: pending with events
	// system audit log event to notify success authz
	clusterName := labels["rafay.dev/clusterName"]
	kubectlAuthzEvent("user.login.success", req.ClusterID, clusterName, cnAttr.OrganizationID, cnAttr.PartnerID, userName, cnAttr.AccountID, groups)
	*/

	return resp, nil
}

func getSystemUserAuthz(cnAttrs kubeconfig.CNAttributes) (resp *sentryrpc.GetUserAuthorizationResponse, err error) {
	resp = new(sentryrpc.GetUserAuthorizationResponse)

	authzLabels := getAuthzLabels(cnAttrs.Username)
	sa := &corev1.ServiceAccount{}
	sa.APIVersion = "v1"
	sa.Kind = "ServiceAccount"
	sa.Name = cnAttrs.Username
	sa.Namespace = "rafay-system"
	sa.Labels = authzLabels

	cr, err := getClusterRole(sentry.KubectlFullAccessPermission)
	if err != nil {
		return nil, err
	}
	cr.Labels = authzLabels
	crb := getClusterRoleBinding(sa, cr.Name)
	crb.Labels = authzLabels

	saObject, err := runtime.FromObject(sa)
	if err != nil {
		return nil, err
	}

	crObject, err := runtime.FromObject(cr)
	if err != nil {
		return nil, err
	}

	crbObject, err := runtime.FromObject(crb)
	if err != nil {
		return nil, err
	}

	resp.UserName = cnAttrs.Username
	resp.ServiceAccount = saObject
	resp.ClusterRoles = []*controller.StepObject{crObject}
	resp.ClusterRoleBindings = []*controller.StepObject{crbObject}
	return
}

func isClusterScopePermission(permission string) bool {
	for _, p := range clusterScopePermissions {
		if permission == p {
			return true
		}
	}
	return false
}

func isNamespaceScopePermission(permission string) bool {
	for _, p := range namespaceScopePermissions {
		if permission == p {
			return true
		}
	}
	return false
}

func verifyClusterKubectlSettings(ctx context.Context, bs service.BootstrapService, kcs service.KubectlClusterSettingsService, cnAttr kubeconfig.CNAttributes, clusterID string, orgID string) error {
	if cnAttr.SessionType == kubeconfig.RafaySystem {
		// internal system sessions are always allowed
		return nil
	}

	_, err := bs.GetBootstrapAgentCountForClusterID(ctx, clusterID, orgID)
	if err != nil {
		_log.Infow("verify cluster kubectl settings invalid clusterid or orgid", "cluster", clusterID, "orgID", orgID)
		return err
	}

	/*
		if cnAttr.RelayNetwork {
			_log.Debugw("skip verify cluster kubectl settings for relaynetwork sessions")
			return nil // allow
		}
	*/

	kc, err := kcs.Get(ctx, orgID, clusterID)
	if err == constants.ErrNotFound {
		// no settings found, hence there is no restriction.
		return nil //allow
	} else if err != nil {
		return err
	}

	if cnAttr.SessionType == "" || cnAttr.SessionType == kubeconfig.TerminalShell {
		// backward combatibility treat "" as terminal session for old kubeconfigs
		if kc.DisableCLIKubectl {
			_log.Infow("kubectl cli is not authorized for ", "cnAttr", cnAttr)
			return fmt.Errorf("kubectl cli is not authorized") //deny
		}
		return nil // allow
	}

	if cnAttr.SessionType == kubeconfig.WebShell {
		if kc.DisableWebKubectl {
			_log.Infow("browser based kubectl is not authorized for ", "cnAttr", cnAttr)
			return fmt.Errorf("browser based kubectl is not authorized") //deny
		}
		return nil // allow
	}

	_log.Infow("unknown kubectl ", "SessionType", cnAttr.SessionType)

	return fmt.Errorf("unknown kubectl session type is not authorized")
}

func verifyKubectlSettings(cnAttr kubeconfig.CNAttributes, ks *sentry.KubeconfigSetting, level string) error {
	if cnAttr.SessionType == kubeconfig.RafaySystem {
		// internal system sessions are always allowed
		return nil
	}

	if cnAttr.SessionType == "" || cnAttr.SessionType == kubeconfig.TerminalShell {
		// backward combatibility treat "" as terminal session for old kubeconfigs
		if ks.DisableCLIKubectl {
			_log.Infow("kubectl cli is not authorized for ", "cnAttr", cnAttr, " by ", level, "config")
			return fmt.Errorf("kubectl cli is not authorized" + " by " + level + "config") //deny
		}
		return nil // allow
	}

	if cnAttr.SessionType == kubeconfig.WebShell {
		if ks.DisableWebKubectl {
			_log.Infow("browser based kubectl is not authorized for ", "cnAttr", cnAttr, " by ", level, "config")
			return fmt.Errorf("browser based kubectl is not authorized" + " by " + level + "config") //deny
		}
		return nil // allow
	}

	_log.Infow("unknown kubectl ", "SessionType", cnAttr.SessionType)

	return fmt.Errorf("unknown kubectl session type is not authorized")
}
