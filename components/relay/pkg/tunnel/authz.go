package tunnel

import (
	"context"
	"fmt"
	"hash"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/controller/apply"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/grpc"
	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/proxy"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
	"github.com/dgraph-io/ristretto"
	"github.com/twmb/murmur3"
	"google.golang.org/grpc/credentials"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	_log = log.GetLogger()
)

const (
	bypassNamespace = "rafay-system"
)

var bypassUserNames = []string{"system-sa", "default"}

func getClusterID(sni string) (string, error) {
	idx := strings.Index(sni, ".")
	if idx < 0 {
		return "", fmt.Errorf("invalid user sni format %s", sni)
	}
	return sni[0:idx], nil
}

type bypassWrapper struct {
	sni, key, bypassUserName string
	rt                       http.RoundTripper
}

func (w *bypassWrapper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("X-Rafay-User", w.bypassUserName)
	r.Header.Set("X-Rafay-Key", w.key)
	r.Header.Set("X-Rafay-Namespace", bypassNamespace)

	if !strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path = r.URL.Path + "/"
	}

	return w.rt.RoundTrip(r)
}

func newBypassWrapper(sni, key, bypassUserName string) transport.WrapperFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &bypassWrapper{sni: sni, key: key, rt: rt, bypassUserName: bypassUserName}
	}
}

func getKubeRESTConfig(socketPath, sni, dialinLookupKey, username string) *rest.Config {

	cfg := &rest.Config{
		Host: fmt.Sprintf("%s_%s", sni, dialinLookupKey),
		// Set transport expicitly to prevent client for caching transport
		Transport: &http.Transport{
			DialContext: proxy.UnixDialContext(socketPath, dialinLookupKey, username, sni),
		},
		Timeout: time.Second * 5,
	}

	cfg.Wrap(newBypassWrapper(sni, dialinLookupKey, username))

	return cfg

}

var _hashPool = sync.Pool{
	New: func() interface{} {
		// The Pool's New function should generally only return pointer
		// types, since a pointer can be put into the return interface
		// value without an allocation:
		return murmur3.New64()
	},
}

type authzProvisioner struct {
	clientCache *ristretto.Cache
	authzCache  *ristretto.Cache
}

func newAuthzProvisioner() (*authzProvisioner, error) {
	clientCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // Num keys to track frequency of (10k).
		MaxCost:     1 << 27, // Maximum cost of cache (256MB).
		BufferItems: 64,      // Number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}

	authzCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // Num keys to track frequency of (10k).
		MaxCost:     1 << 25, // Maximum cost of cache (32MB).
		BufferItems: 64,      // Number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}

	return &authzProvisioner{clientCache: clientCache, authzCache: authzCache}, nil
}

func (p *authzProvisioner) getClient(socketPath, sni, dialinLookupKey string) (client.Client, error) {

	key := func() uint64 {
		hasher := _hashPool.Get().(hash.Hash64)
		hasher.Reset()
		hasher.Write([]byte(sni))
		hasher.Write([]byte(dialinLookupKey))
		k := hasher.Sum64()
		_hashPool.Put(hasher)
		return k
	}()

	if i, ok := p.clientCache.Get(key); ok {
		if c, ok := i.(client.Client); ok {
			return c, nil
		}
	}

	getK8sClient := func(bypassUserName string) (client.Client, error) {
		cfg := getKubeRESTConfig(socketPath, sni, dialinLookupKey, bypassUserName)
		mapper, err := apiutil.NewDynamicRESTMapper(cfg)
		if err != nil {
			return nil, err
		}

		c, err := client.New(cfg, client.Options{Scheme: scheme.Scheme, Mapper: mapper})
		return c, err
	}

	var c client.Client
	var err error
	for _, username := range bypassUserNames {
		c, err = getK8sClient(username)
		if err != nil {
			_log.Infow("error getting k8s client for username", "username", username)
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	p.clientCache.SetWithTTL(key, c, 100, time.Minute*30)
	return c, nil
}

func (p *authzProvisioner) getAuthz(ctx context.Context, userCN, sni string, certIssue int64) (*sentryrpc.GetUserAuthorizationResponse, error) {
	clusterID, err := getClusterID(sni)
	if err != nil {
		_log.Infow("unable to get clusterID", "error", err)
		return nil, err
	}

	u, err := url.Parse(utils.PeerServiceURI)
	if err != nil {
		_log.Infow("unable to parse peer service url", "error", err)
		return nil, err
	}
	//Load certificates
	tlsConfig, err := ClientTLSConfigFromBytes(utils.PeerCertificate, utils.PeerPrivateKey, utils.PeerCACertificate, u.Host)
	if err != nil {
		_log.Infow("unable to build tls config for peer service", "error", err)
		return nil, err
	}
	transportCreds := credentials.NewTLS(tlsConfig)
	peerSeviceHost := u.Host

	conn, err := grpc.NewSecureClientConn(ctx, peerSeviceHost, transportCreds)
	if err != nil {
		_log.Infow("unable to connect to sentry", "error", err)
		return nil, err
	}
	defer conn.Close()

	client := sentryrpc.NewClusterAuthorizationClient(conn)

	resp, err := client.GetUserAuthorization(ctx, &sentryrpc.GetUserAuthorizationRequest{
		UserCN:           userCN,
		ClusterID:        clusterID,
		CertIssueSeconds: certIssue,
	})

	if err != nil {
		_log.Infow("unable to get authorization for user", "error", err)
		return nil, err
	}
	return resp, nil
}

func (p *authzProvisioner) provisionAuthzHandleRoleChange(auth *sentryrpc.GetUserAuthorizationResponse, socketPath, sni, dialinLookupKey string) {
	k8sDelClient, err := p.getClient(socketPath, sni, dialinLookupKey)
	if err != nil {
		_log.Infow("unable to get del client", "sni", sni, "dialinLookupKey", dialinLookupKey, "error", err)
		return
	}
	nctx, cancel := context.WithTimeout(context.Background(), time.Second*180)
	defer cancel()
	// Delete cluster role binding to clean if exist to handle change in user role
	for _, crb := range auth.DeleteClusterRoleBindings {
		//crbo, _, err := runtime.ToObject(crb)
		if err == nil {
			err := k8sDelClient.Delete(nctx, crb)
			if err != nil {
				_log.Debugw("unable to delete ClusterRoleBindings", "name", crb.TypeMeta.GetObjectKind(), "obj", crb, "err", err)
			} else {
				_log.Infow("successfully to deleted ClusterRoleBindings", "obj", crb)
			}
		}
	}

	// To handle change in namespace changes delete rolebindings to clean if exist.
	for _, rb := range auth.DeleteRoleBindings {
		//rbo, _, err := runtime.ToObject(rb)
		if err == nil {
			err := k8sDelClient.Delete(nctx, rb)
			if err != nil {
				_log.Debugw("unable to delete RoleBindings", "name", rb.TypeMeta.GetObjectKind(), "obj", rb, "err", err)
			} else {
				_log.Infow("successfully to deleted RoleBindings", "obj", rb)
			}
		}
	}
}

func (p *authzProvisioner) provisionAuthz(ctx context.Context, auth *sentryrpc.GetUserAuthorizationResponse, socketPath, sni, dialinLookupKey string) error {

	k8sClient, err := p.getClient(socketPath, sni, dialinLookupKey)
	if err != nil {
		_log.Infow("unable to get client", "sni", sni, "dialinLookupKey", dialinLookupKey, "error", err)
		return err
	}

	// Handle deletion of excluded clusterrolebindings and rolebindings
	// in the background.
	go p.provisionAuthzHandleRoleChange(auth, socketPath, sni, dialinLookupKey)
	if len(auth.ClusterRoles) > 0 {
		_log.Infow("cluster scope authz for", "user", auth.ServiceAccount, "ClusterRoles", auth.ClusterRoles, "ClusterRoleBindings", auth.ClusterRoleBindings)
	}
	if len(auth.Roles) > 0 {
		_log.Infow("namespace scope authz for", "user", auth.ServiceAccount, "Roles", auth.Roles, "RoleBindings", auth.RoleBindings, "Namespaces", auth.Namespaces)
	}

	applier := apply.NewApplier(k8sClient)

	/*sao, _, err := runtime.ToObject(auth.ServiceAccount)
	if err != nil {
		_log.Infow("unable to make service account runtime object", "error", err)
		return err
	}*/

	err = applier.Apply(ctx, auth.ServiceAccount)
	// Ignore already exist error caused due to
	// multiple inflight request trying to provision
	// ZTKA JIT service account
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		_log.Infow("unable to apply service account", "error", err)
		return err
	}

	for _, clusterRole := range auth.ClusterRoles {
		/*cro, _, err := runtime.ToObject(clusterRole)
		if err != nil {
			_log.Infow("unable to make clusterRole runtime object", "error", err)
			return err
		}*/

		err = applier.Apply(ctx, clusterRole)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			_log.Infow("unable to apply clusterRole", "error", err)
			return err
		}
	}

	for _, clusterRoleBinding := range auth.ClusterRoleBindings {
		/*crbo, _, err := runtime.ToObject(clusterRoleBinding)
		if err != nil {
			_log.Infow("unable to make clusterRoleBinding runtime object", "error", err)
			return err
		}*/

		err = applier.Apply(ctx, clusterRoleBinding)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			_log.Infow("unable to apply clusterRoleBinding", "error", err)
			return err
		}
	}

	for _, role := range auth.Roles {
		/*cro, _, err := runtime.ToObject(role)
		if err != nil {
			_log.Infow("unable to make roles runtime object", "error", err)
			return err
		}*/

		err = applier.Apply(ctx, role)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			_log.Infow("unable to apply Role", "error", err)
			//Ignore any error
		}
	}

	for _, roleBinding := range auth.RoleBindings {
		/*crbo, _, err := runtime.ToObject(roleBinding)
		if err != nil {
			_log.Infow("unable to make RoleBinding runtime object", "error", err)
			return err
		}*/

		err = applier.Apply(ctx, roleBinding)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			_log.Infow("unable to apply RoleBinding", "error", err)
			//Ignore any error
		}
	}

	_log.Infow("applied authz", "sni", sni, "user", auth.UserName)

	return nil
}

type provisionedUser struct {
	UserName                        string
	Role                            string
	IsReadRole                      bool
	IsOrgAdmin                      bool
	EnforceOrgAdminOnlySecretAccess bool
}

func (p *authzProvisioner) ProvisionAuthzForUser(socketPath, userCN, sni, dialinLookupKey string, forceProvision, refreshSession bool, certIssue int64) (string, string, bool, bool, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	key := func() uint64 {
		hasher := _hashPool.Get().(hash.Hash64)
		hasher.Reset()
		hasher.Write([]byte(userCN))
		hasher.Write([]byte(sni))
		k := hasher.Sum64()
		_hashPool.Put(hasher)
		return k
	}()

	if !forceProvision {
		if i, ok := p.authzCache.Get(key); ok {
			if pu, ok := i.(*provisionedUser); ok {
				return pu.UserName, pu.Role, pu.IsReadRole, pu.IsOrgAdmin, pu.EnforceOrgAdminOnlySecretAccess, nil
			}
		}
	}

	resp, err := p.getAuthz(ctx, userCN, sni, certIssue)
	if err != nil {
		return "", "", false, false, false, err
	}

	applyctx, applycancel := context.WithTimeout(context.Background(), time.Second*60)
	defer applycancel()
	err = p.provisionAuthz(applyctx, resp, socketPath, sni, dialinLookupKey)
	// ignore cluster provision failures during refresh.
	// continue with existing RBAC.
	if !refreshSession && err != nil {
		return "", "", false, false, false, err
	}

	p.authzCache.SetWithTTL(key, &provisionedUser{
		UserName:                        resp.UserName,
		Role:                            resp.RoleName,
		IsReadRole:                      resp.IsRead,
		IsOrgAdmin:                      resp.IsOrgAdmin,
		EnforceOrgAdminOnlySecretAccess: resp.EnforceOrgAdminOnlySecretAccess,
	}, 100, time.Minute*5)

	return resp.UserName, resp.RoleName, resp.IsRead, resp.IsOrgAdmin, resp.EnforceOrgAdminOnlySecretAccess, nil
}
