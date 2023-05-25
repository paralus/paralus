package kubeconfig

import (
	"context"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/log"
	"github.com/paralus/paralus/pkg/query"
	sentryrpc "github.com/paralus/paralus/proto/rpc/sentry"
	rpcv3 "github.com/paralus/paralus/proto/rpc/user"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	sentry "github.com/paralus/paralus/proto/types/sentry"
	"go.uber.org/zap"

	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"

	"github.com/paralus/paralus/internal/constants"
	"github.com/paralus/paralus/pkg/sentry/cryptoutil"
	"github.com/paralus/paralus/pkg/sentry/util"
	"github.com/paralus/paralus/pkg/service"
)

const (
	kubeconfigPermission = sentry.KubeconfigReadPermission
	systemUsername       = "admin@paralus.local"
)

var _log = log.GetLogger()

// GetUserCN returns user cn from attrs
func GetUserCN(attrs map[string]string) string {
	var keys []string
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	sb := new(strings.Builder)

	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteRune('=')
		sb.WriteString(attrs[key])
		sb.WriteRune('/')
	}
	return strings.TrimRight(sb.String(), "/")
}

// GetUserAttrs returns attrs from cn
func GetUserAttrs(cn string) map[string]string {
	attrs := strings.Split(cn, "/")
	ret := make(map[string]string)
	for _, attr := range attrs {
		vals := strings.Split(attr, "=")
		if len(vals) != 2 {
			continue
		}
		ret[vals[0]] = vals[1]
	}
	return ret
}

func getProjectsForSSOAccount(ctx context.Context, groups []string, orgID, partnerID string, gps service.GroupPermissionService) ([]string, bool, error) {
	isOrgScope := false
	groupProjectPermissions, err := gps.GetGroupProjectsByPermission(ctx, groups, orgID, partnerID, kubeconfigPermission)
	if err != nil {
		_log.Errorw("error getting group project permissions", "permission", kubeconfigPermission, "error", err.Error())
		return nil, isOrgScope, err
	}
	projectsMap := make(map[string]string)
	projects := make([]string, 0)
	for _, gp := range groupProjectPermissions {
		if _, ok := projectsMap[gp.ProjectID]; ok {
			continue
		}
		projects = append(projects, gp.ProjectID)
		projectsMap[gp.ProjectID] = gp.Scope
		if gp.Scope == "ORGANIZATION" {
			isOrgScope = true
		}
	}
	return projects, isOrgScope, nil
}

func getProjectsForAccount(ctx context.Context, accountID, orgID, partnerID string, aps service.AccountPermissionService) ([]string, bool, error) {
	isOrgScope := false
	accountProjectPermissions, err := aps.GetAccountProjectsByPermission(ctx, accountID, orgID, partnerID, kubeconfigPermission)
	if err != nil {
		_log.Errorw("error getting account project permissions", "permission", kubeconfigPermission, "error", err.Error())
		return nil, isOrgScope, err
	}
	projectsMap := make(map[string]string)
	projects := make([]string, 0)
	for _, ap := range accountProjectPermissions {
		if _, ok := projectsMap[ap.ProjectID]; ok {
			continue
		}
		projects = append(projects, ap.ProjectID)
		projectsMap[ap.ProjectID] = ap.Scope
		if ap.Scope == "organization" {
			isOrgScope = true
		}
	}
	return projects, isOrgScope, nil
}

// GetConfigForUser returns YAML encoding of kubeconfig
func GetConfigForUser(ctx context.Context, bs service.BootstrapService, aps service.AccountPermissionService, gps service.GroupPermissionService, req *sentryrpc.GetForUserRequest, pf cryptoutil.PasswordFunc, kss service.KubeconfigSettingService, ksvc service.ApiKeyService, os service.OrganizationService, ps service.PartnerService, al *zap.Logger) ([]byte, error) {
	opts := req.Opts
	if opts.Selector != "" {
		opts.Selector = fmt.Sprintf("%s,!paralus.dev/cdRelayAgent", opts.Selector)
	} else {
		opts.Selector = "!paralus.dev/cdRelayAgent"
	}
	batl, err := bs.SelectBootstrapAgentTemplates(ctx, query.WithSelector("paralus.dev/defaultUser=true"), query.WithGlobalScope())
	if err != nil {
		_log.Errorw("error getting default user bootstrap agent templates", "error", err.Error())
		return nil, err
	}

	if len(batl.Items) < 1 {
		_log.Errorw("no user bootstrap agent template found")
		return nil, fmt.Errorf("no user bootstrap agent template found")
	}

	_log.Infow("get config for user ", "opts", opts)

	bi, err := bs.GetBootstrapInfra(ctx, batl.Items[0].Spec.InfraRef)
	if err != nil {
		_log.Errorw("error getting bootstrap infra", "infraRef", batl.Items[0].Spec.InfraRef, "error", err.Error())
		return nil, err
	}
	isSSOAcc := opts.GetIsSSOUser()

	username := opts.Username
	sessionUserName := opts.Username
	groups := opts.Groups
	enforceSession := false

	if sessionUserName == "" && opts.Account != "" {
		accountData, err := aps.GetAccount(ctx, opts.Account)
		if err != nil {
			_log.Errorw("error getting account data", "error", err.Error())
			return nil, err
		}
		sessionUserName = accountData.Username
		username = accountData.Username
	} else if sessionUserName == "" && opts.ID != "" {
		apiKey, err := ksvc.GetByKey(ctx, &rpcv3.ApiKeyRequest{Id: opts.ID})
		if err != nil {
			_log.Errorw("error getting account data", "error", err.Error())
			return nil, err
		}
		sessionUserName = apiKey.Name
		username = apiKey.Name
		opts.Account = apiKey.AccountID.String()
	} else if sessionUserName == "" && opts.Account == "" {
		_log.Errorw("error getting account data", "error", err.Error())
		return nil, fmt.Errorf("account information not present in request")
	}

	//validate if organization id or name is given, should support both
	if opts.Organization == "" {
		_log.Errorw("error getting organization data", "error", err.Error())
		return nil, fmt.Errorf("organization information is missing in request")
	}
	oid, err := uuid.Parse(opts.Organization)
	if err != nil {
		//looks like name is provided, fetch org id
		org, err := os.GetByName(ctx, opts.Organization)
		if err != nil {
			_log.Errorw("error getting organization data", "error", err.Error())
			return nil, fmt.Errorf("failed to retrieve organization %s", err.Error())
		}
		oid = uuid.MustParse(org.Metadata.Id)
		opts.Organization = oid.String()
		opts.Partner = org.Metadata.Partner
	}

	if opts.Partner == "" {
		_log.Errorw("error getting partner data", "error", err.Error())
		return nil, fmt.Errorf("partner information is missing in request")
	}
	_, err = uuid.Parse(opts.Partner)
	if err != nil {
		part, err := ps.GetByName(ctx, opts.Partner)
		if err != nil {
			_log.Errorw("error getting partner data", "error", err.Error())
			return nil, fmt.Errorf("failed to retrieve partner %s", err.Error())
		}
		opts.Partner = part.Metadata.Id
	}

	// get user level settings if exist
	ks, err := kss.Get(ctx, opts.Organization, opts.Account, isSSOAcc)
	if err == constants.ErrNotFound {
		// get user org settings if exist
		ks, err = kss.Get(ctx, opts.Organization, "", isSSOAcc)
		if err != nil && err != constants.ErrNotFound {
			return nil, err
		}
	} else if err != nil && err != constants.ErrNotFound {
		return nil, err
	}
	if ks != nil && ks.EnableSessionCheck {
		enforceSession = true
	}
	// {"account": "", "username": "", "partner": "", "org": "", "project":, "sso":,  "enforceSession"}
	// TODO: figure out how SSO works
	// CN=account=<aid>/partner=<pid>/orgid=<id>/username=<un>
	cnAttr := CNAttributes{
		AccountID:      opts.Account,
		PartnerID:      opts.Partner,
		OrganizationID: opts.Organization,
		IsSSO:          isSSOAcc,
		EnforceSession: enforceSession,
		Username:       util.SanitizeUsername(username),
		SessionType:    TerminalShell,
		RelayNetwork:   false,
	}
	cn := cnAttr.GetCN()

	// get account projects with kubeconfig.read permission
	projects := make([]string, 0)
	isOrgScope := false
	if !isSSOAcc {
		projects, isOrgScope, err = getProjectsForAccount(ctx, opts.Account, opts.Organization, opts.Partner, aps)
		if err != nil {
			_log.Errorw("error getting project for paralus ", "account", opts.Account, "error", err.Error())
			return nil, err
		}
	} else {
		projects, isOrgScope, err = getProjectsForSSOAccount(ctx, groups, opts.Organization, opts.Partner, gps)
		if err != nil {
			_log.Errorw("error getting project for sso ", "account", opts.Account, "error", err.Error())
			return nil, err
		}
	}

	// get list of bootstrap agents
	bas := []*sentry.BootstrapAgent{}
	if isOrgScope {
		bal, err := bs.SelectBootstrapAgents(ctx, "-",
			query.WithOptions(opts),
			// ignore project id, because kubeconfig is not project scoped
			query.WithIgnoreScopeDefault(),
		)
		if err != nil {
			_log.Errorw("error getting bootstrap agents", "error", err.Error())
			return nil, err
		}
		bas = bal.Items
	} else {
		set := make(map[string]interface{})
		for _, ap := range projects {
			selector := ""
			if opts.Selector != "" {
				selector = fmt.Sprintf("%s,project/%s", opts.Selector, ap)
			} else {
				selector = fmt.Sprintf("project/%s", ap)
			}
			bal, err := bs.SelectBootstrapAgents(ctx, "-",
				query.WithOptions(opts),
				query.WithSelector(selector),
				// ignore project id, because kubeconfig is not project scoped
				query.WithIgnoreScopeDefault(),
			)
			if err != nil {
				_log.Errorw("error getting bootstrap agents", "error", err.Error())
				return nil, err
			}
			for _, ba := range bal.Items {
				if ba.Spec.TemplateRef != "paralus-core-relay-agent" && ba.Spec.TemplateRef != "paralus-core-cd-relay-agent" {
					if vi, ok := set[ba.Metadata.Name]; ok {
						v := vi.(sentry.BootstrapAgent)
						if v.Spec.TemplateRef == ba.Spec.TemplateRef {
							continue
						}
					}

					set[ba.Metadata.Name] = ba
					bas = append(bas, ba)
				} else {
					if _, ok := set[ba.Metadata.Name]; ok {
						continue
					}
					set[ba.Metadata.Name] = ba
					bas = append(bas, ba)
				}
			}
		}
	}

	serverHost := ""
	for _, host := range batl.Items[0].Spec.Hosts {
		if host.Type == sentry.BootstrapTemplateHostType_HostTypeExternal {
			serverHost = host.Host
		}
	}

	if serverHost == "" {
		return nil, fmt.Errorf("no externals hosts found")
	}

	// get cert validity setting
	certValidity, err := getCertValidity(ctx, opts.Organization, opts.Account, isSSOAcc, kss)
	if err != nil {
		_log.Errorw("error getting cert validity settings", "error", err.Error())
		return nil, err
	}

	if certValidity == 0 {
		// Set 1 second to avoid default value from cert Sign
		certValidity = 1 * time.Second
	}

	config, err := getUserConfig(ctx, *opts, username, req.Namespace, cn, serverHost, bi, bas, pf, certValidity, bs)
	if err != nil {
		_log.Errorw("error generating kubeconfig", "error", err.Error())
		return nil, err
	}

	jb, err := json.Marshal(&config)
	if err != nil {
		return nil, err
	}

	service.DownloadKubeconfigAuditEvent(ctx, al, username)

	return yaml.JSONToYAML(jb)
}

func getCertValidity(ctx context.Context, orgID, accountID string, isSSO bool, kss service.KubeconfigSettingService) (time.Duration, error) {
	ksUser, err := kss.Get(ctx, orgID, accountID, isSSO)
	if err == nil && ksUser.ValiditySeconds >= 0 {
		return time.Second * time.Duration(ksUser.ValiditySeconds), nil
	} else if err != nil && err != constants.ErrNotFound {
		return 0, err
	}

	ksOrg, err := kss.Get(ctx, orgID, "", false)
	if err == nil && ksOrg.ValiditySeconds >= 0 {
		return time.Second * time.Duration(ksOrg.ValiditySeconds), nil
	} else if err != nil && err != constants.ErrNotFound {
		return 0, err
	}

	// by default 1 year validity
	return (360 * (time.Hour * 24)), nil
}

func getConfig(username, namespace, certCN, serverHost string, bootstrapInfra *sentry.BootstrapInfra, bootstrapAgents []*sentry.BootstrapAgent, pf cryptoutil.PasswordFunc, certValidity time.Duration, clusterName string) (*clientcmdapiv1.Config, error) {

	if namespace == "" {
		namespace = "default"
	}
	name := util.SanitizeUsername(username)

	signer, err := cryptoutil.NewSigner([]byte(bootstrapInfra.Spec.CaCert), []byte(bootstrapInfra.Spec.CaKey),
		cryptoutil.WithCAKeyDecrypt(pf),
		cryptoutil.WithCertValidity(certValidity),
		cryptoutil.WithClient(),
	)
	if err != nil {
		return nil, err
	}

	privKey, err := cryptoutil.GenerateECDSAPrivateKey()
	if err != nil {
		return nil, err
	}

	key, err := cryptoutil.EncodePrivateKey(privKey, cryptoutil.NoPassword)
	if err != nil {
		return nil, err
	}

	csr, err := cryptoutil.CreateCSR(pkix.Name{
		CommonName: certCN,
	}, privKey)
	if err != nil {
		return nil, err
	}

	// sign csr and get the cert
	cert, err := signer.Sign(csr)
	if err != nil {
		return nil, err
	}

	users := []clientcmdapiv1.NamedAuthInfo{
		{
			Name: name,
			AuthInfo: clientcmdapiv1.AuthInfo{
				ClientCertificateData: cert,
				ClientKeyData:         key,
			},
		},
	}

	var clusters []clientcmdapiv1.NamedCluster

	var contexts []clientcmdapiv1.NamedContext

	for _, ba := range bootstrapAgents {
		if ba.Spec.TemplateRef != "paralus-core-relay-agent" && ba.Spec.TemplateRef != "paralus-core-cd-relay-agent" {
			// skip non default agents from system kubeconfiog
			continue
		}

		host := strings.ReplaceAll(serverHost, "*", ba.Metadata.Name)

		clusters = append(clusters, clientcmdapiv1.NamedCluster{
			Name: ba.Metadata.DisplayName,
			Cluster: clientcmdapiv1.Cluster{
				Server:                   fmt.Sprintf("https://%s", host),
				CertificateAuthorityData: []byte(bootstrapInfra.Spec.CaCert),
			},
		})

		contexts = append(contexts, clientcmdapiv1.NamedContext{
			Name: ba.Metadata.DisplayName,
			Context: clientcmdapiv1.Context{
				Cluster:   ba.Metadata.DisplayName,
				AuthInfo:  name,
				Namespace: namespace,
			},
		})
	}

	config := &clientcmdapiv1.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters:   clusters,
		AuthInfos:  users,
		Contexts:   contexts,
	}

	if clusterName = strings.Trim(clusterName, " "); clusterName != "" {
		config.CurrentContext = clusterName
	} else if len(contexts) > 0 {
		config.CurrentContext = contexts[0].Name
	}

	return config, nil
}

// GetConfigForCluster returns YAML encoded kubeconfig
func GetConfigForCluster(ctx context.Context, bs service.BootstrapService, req *sentryrpc.GetForClusterRequest, pf cryptoutil.PasswordFunc, kss service.KubeconfigSettingService, sessionType string) ([]byte, error) {
	opts := req.Opts
	if opts.Selector != "" {
		opts.Selector = fmt.Sprintf("%s,!paralus.dev/cdRelayAgent", opts.Selector)
	} else {
		opts.Selector = fmt.Sprintf("!paralus.dev/cdRelayAgent")
	}
	_log.Infow("get config for cluster ", "opts", opts, "namespace", req.Namespace, "systemUser", req.SystemUser)
	batl, err := bs.SelectBootstrapAgentTemplates(ctx, query.WithSelector("paralus.dev/defaultUser=true"), query.WithGlobalScope())
	if err != nil {
		return nil, err
	}

	if len(batl.Items) < 1 {
		return nil, fmt.Errorf("no user bootstrap agent template found")
	}

	bi, err := bs.GetBootstrapInfra(ctx, batl.Items[0].Spec.InfraRef)
	if err != nil {
		return nil, err
	}

	serverHost := ""
	for _, host := range batl.Items[0].Spec.Hosts {
		if host.Type == sentry.BootstrapTemplateHostType_HostTypeExternal {
			serverHost = host.Host
		}
	}

	bal, err := bs.SelectBootstrapAgents(ctx, "-",
		query.WithOptions(opts),
		// ignore project id because kubeconfig is not project scoped
		query.WithIgnoreScopeDefault(),
	)
	if err != nil {
		return nil, err
	}

	if bal.Metadata.Count <= 0 {
		return nil, fmt.Errorf("no bootstrap agents found")
	}

	isSSOAcc := opts.GetIsSSOUser()
	username := opts.Username
	if req.SystemUser {
		username = systemUsername
	}
	enforceSession := false

	// get user level settings if exist
	ks, err := kss.Get(ctx, opts.Organization, opts.Account, isSSOAcc)
	if err == constants.ErrNotFound {
		// get user org settings if exist
		ks, err = kss.Get(ctx, opts.Organization, "", isSSOAcc)
		if err != nil && err != constants.ErrNotFound {
			return nil, err
		}
	} else if err != nil && err != constants.ErrNotFound {
		return nil, err
	}
	if ks != nil && ks.EnableSessionCheck {
		enforceSession = true
	}
	// {"account": "", "username": "", "partner": "", "org": "", "project":, "sso":,  "enforceSession"}
	// CN=account=<aid>/partner=<pid>/orgid=<id>/username=<un>
	cnAttr := CNAttributes{
		AccountID:      opts.Account,
		PartnerID:      opts.Partner,
		OrganizationID: opts.Organization,
		IsSSO:          isSSOAcc,
		EnforceSession: enforceSession,
		Username:       util.SanitizeUsername(username),
		SessionType:    sessionType,
	}

	if req.SystemUser {
		cnAttr.AccountID = ""
		cnAttr.PartnerID = ""
		cnAttr.OrganizationID = ""
		cnAttr.IsSSO = false
		cnAttr.EnforceSession = false
		cnAttr.Username = util.SanitizeUsername(username)
		cnAttr.SystemUser = true
	}

	cn := cnAttr.GetCN()

	certValidity, err := getCertValidity(ctx, opts.Organization, opts.Account, isSSOAcc, kss)
	if err != nil {
		_log.Errorw("error getting cert validity settings", "error", err.Error())
		return nil, err
	}

	if certValidity == 0 {
		// Browser based session expire in
		// next 8 hours to avoid default value cert Sign
		certValidity = 8 * time.Hour
	}

	config, err := getConfig(username, req.Namespace, cn, serverHost, bi, bal.Items, pf, certValidity, opts.Name)
	if err != nil {
		_log.Errorw("error generating kubeconfig", "error", err.Error())
		return nil, err
	}

	jb, err := json.Marshal(&config)
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML(jb)

}

func getUserConfig(ctx context.Context, opts commonv3.QueryOptions, username, namespace, certCN, serverHost string, bootstrapInfra *sentry.BootstrapInfra, bootstrapAgents []*sentry.BootstrapAgent, pf cryptoutil.PasswordFunc, certValidity time.Duration, bs service.BootstrapService) (*clientcmdapiv1.Config, error) {

	if namespace == "" {
		namespace = "default"
	}
	name := util.SanitizeUsername(username)

	signer, err := cryptoutil.NewSigner([]byte(bootstrapInfra.Spec.CaCert), []byte(bootstrapInfra.Spec.CaKey),
		cryptoutil.WithCAKeyDecrypt(pf),
		cryptoutil.WithCertValidity(certValidity),
		cryptoutil.WithClient(),
	)
	if err != nil {
		return nil, err
	}

	privKey, err := cryptoutil.GenerateECDSAPrivateKey()
	if err != nil {
		return nil, err
	}

	key, err := cryptoutil.EncodePrivateKey(privKey, cryptoutil.NoPassword)
	if err != nil {
		return nil, err
	}

	csr, err := cryptoutil.CreateCSR(pkix.Name{
		CommonName: certCN,
	}, privKey)
	if err != nil {
		return nil, err
	}

	// sign csr and get the cert
	cert, err := signer.Sign(csr)
	if err != nil {
		return nil, err
	}

	users := []clientcmdapiv1.NamedAuthInfo{
		{
			Name: name,
			AuthInfo: clientcmdapiv1.AuthInfo{
				ClientCertificateData: cert,
				ClientKeyData:         key,
			},
		},
	}

	var clusters []clientcmdapiv1.NamedCluster

	var contexts []clientcmdapiv1.NamedContext

	baMaps := make(map[string]sentry.BootstrapAgent)

	// prune agent list
	// if a cluster is added to custom relay then exlude it from default
	for _, ba := range bootstrapAgents {
		if ba.Spec.TemplateRef != "paralus-core-relay-agent" && ba.Spec.TemplateRef != "paralus-core-cd-relay-agent" {
			baMaps[ba.Metadata.Name] = *ba
		} else {
			if _, ok := baMaps[ba.Metadata.Name]; !ok {
				baMaps[ba.Metadata.Name] = *ba
			}
		}
	}

	for _, ba := range baMaps {
		if ba.Spec.TemplateRef != "paralus-core-relay-agent" && ba.Spec.TemplateRef != "paralus-core-cd-relay-agent" {
			// handle custome relay network
		} else {

			host := strings.ReplaceAll(serverHost, "*", ba.Metadata.Name)

			clusters = append(clusters, clientcmdapiv1.NamedCluster{
				Name: ba.Metadata.DisplayName,
				Cluster: clientcmdapiv1.Cluster{
					Server:                   fmt.Sprintf("https://%s", host),
					CertificateAuthorityData: []byte(bootstrapInfra.Spec.CaCert),
				},
			})

			contexts = append(contexts, clientcmdapiv1.NamedContext{
				Name: ba.Metadata.DisplayName,
				Context: clientcmdapiv1.Context{
					Cluster:   ba.Metadata.DisplayName,
					AuthInfo:  name,
					Namespace: namespace,
				},
			})
		}
	}

	config := &clientcmdapiv1.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters:   clusters,
		AuthInfos:  users,
		Contexts:   contexts,
	}

	if len(contexts) > 0 {
		config.CurrentContext = contexts[0].Name
	}

	return config, nil
}
