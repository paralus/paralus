package interceptors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	authv3 "github.com/RafaySystems/rcloud-base/components/common/pkg/auth/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/gateway"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/hasher"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _log = log.GetLogger()

var _dummyHandler = func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {}

const (
	cookieHeader = "cookie"
	apiKeyHeader = "x-rafay-api-keyid"
	rafaySession = "rsid"
)

func getAPIKey(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	apiKeys, ok := md[apiKeyHeader]
	if !ok {
		return ""
	}
	return apiKeys[0]
}

func getRsid(ctx context.Context) string {
	//If rsid is present in the header, use it directly
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	_log.Debugw("incoming metadta", "md", md)
	//rsid is part of Cookies
	cookies, ok := md[cookieHeader]
	if !ok {
		return ""
	}
	return readCookies(cookies[0], rafaySession)
}

func readCookies(cookies, filter string) string {
	parts := strings.Split(strings.TrimSpace(cookies), ";")
	// Per-line attributes
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.TrimSpace(parts[i])
		if len(parts[i]) == 0 {
			continue
		}
		name, val := parts[i], ""
		if j := strings.Index(name, "="); j >= 0 {
			name, val = name[:j], name[j+1:]
		}
		if filter != "" && filter != name {
			continue
		}
		return val
	}
	return ""
}

func allowRequest(ctx context.Context, opts *options, req interface{}) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		isGateway := func() bool {
			if vals := md.Get(gateway.GatewayRequest); vals != nil {
				return true
			}
			return false
		}

		url := func() string {
			if urls, ok := md[gateway.GatewayURL]; ok {
				return urls[0]
			}
			return ""
		}

		rsid := func() string {
			if rsids, ok := md[gateway.GatewayRSID]; ok {
				return rsids[0]
			}
			return ""
		}

		apiKey := func() string {
			if apiKeys, ok := md[gateway.GatewayAPIKey]; ok {
				return apiKeys[0]
			}
			return ""
		}

		method := func() string {
			if methods, ok := md[gateway.GatewayMethod]; ok {
				return methods[0]
			}
			return ""
		}

		excluded := func() bool {
			if opts.exclude == nil {
				return false
			}

			if h, _, _ := opts.exclude.Lookup(method(), url()); h != nil {
				return true
			}
			return false
		}

		if opts.logRequest {
			_log.Infow("allowRequest", "url", url(), "isGateway", isGateway(), "rsid", rsid(), "apiKey", apiKey(), "method", method(), "excluded", excluded())
		}

		// if request originates from gateway and method and path are not in excluded
		if isGateway() && !excluded() {
			allowed := func() error {
				aCtx, cancel := context.WithTimeout(ctx, time.Second*10)
				defer cancel()

				client, err := opts.pool.NewClient(aCtx)
				if err != nil {
					_log.Infow("unable to get new auth client", "error", err.Error())
					return err
				}
				defer client.Close()

				resp, err := client.IsRequestAllowed(aCtx, func() *authv3.IsRequestAllowedRequest {
					return &authv3.IsRequestAllowedRequest{
						Url:    url(),
						Method: method(),
						Rsid:   rsid(),
						ApiKey: apiKey(),
					}
				}())
				if err != nil {
					return err
				}

				if resp.GetStatus() != authv3.RequestStatus_RequestAllowed {
					err = fmt.Errorf("%s", resp.Reason)
					return err
				}

				if resp.SessionData != nil {

					var partnerID int64
					var organizationID int64

					if _, ok := req.(commonv3.Metadata); ok {
						_log.Debugw("adding request meta")
						_, err := hasher.IDFromString(resp.SessionData.Account)
						if err != nil {
							_log.Infow("unable to convert account id", "accountID", resp.SessionData.Account, "error", err.Error())
							return err
						}
						partnerID, err = hasher.IDFromString(resp.SessionData.Partner)
						if err != nil {
							_log.Infow("unable to convert partner id", "partnerID", resp.SessionData.Partner, "error", err.Error())
							return err
						}
						organizationID, err = hasher.IDFromString(resp.SessionData.Organization)
						if err != nil {
							_log.Infow("unable to convert organization id", "organizationID", resp.SessionData.Organization, "error", err.Error())
							return err
						}
						fmt.Println(partnerID)
						fmt.Println(organizationID)
						/*rmo.SetAccountID(accountID)
						rmo.SetPartnerID(partnerID)
						rmo.SetOrganizationID(organizationID)
						rmo.SetIsSSOUser(resp.SessionData.IsSsoUser)
						rmo.SetGroups(resp.SessionData.Groups)
						rmo.SetUsername(resp.SessionData.Username)
						rmo.SetAuthType(resp.SessionData.AuthType.String())
						rmo.SetClientType(resp.SessionData.ClientType.String())
						rmo.SetIdp(resp.SessionData.Idp)
						rmo.SetIsAllNsAccess(resp.SessionData.IsAllNsAccess)
						rmo.SetIsOrgAdmin(resp.SessionData.IsOrgAdmin)
						rmo.SetIsPartnerAdmin(resp.SessionData.IsPartnerAdmin)
						rmo.SetIsSuperAdmin(resp.SessionData.IsSuperAdmin)
						rmo.SetNamespaces(authv3.ConvertFromAuthNamespaces(resp.SessionData.Namespaces))
						rmo.SetProject(authv3.ConvertFormAuthProject(resp.SessionData.Project))
						rmo.SetIsReadonlyOrgAdmin(resp.SessionData.IsReadonlyOrgAdmin)*/
					}
				}

				return nil
			}

			if err := allowed(); err != nil {
				return err
			}
			return nil
		}

	}
	return nil

}

func addRequestMeta(ctx context.Context, pool authv3.AuthPool, req interface{}, validateSession bool) error {
	_log.Debugw("adding request meta")
	aCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	client, err := pool.NewClient(aCtx)
	if err != nil {
		_log.Infow("unable to get new auth client", "error", err.Error())
		return err
	}
	defer client.Close()

	rsid := getRsid(ctx)
	apiKey := getAPIKey(ctx)

	var data *authv3.SessionData

	if rsid != "" {
		_log.Debugw("found session", "rsid", rsid)
		resp, err := client.GetSession(ctx, &authv3.GetSessionRequest{
			SessionId: rsid,
		})
		if err != nil {
			_log.Infof("unable to get session for id %s", err.Error())
			return err
		}

		if resp.Status == authv3.SessionStatus_SessionExists {
			_log.Debugw("session exists", "rsid", rsid)
			data = resp.Data
		}
	} else if apiKey != "" {
		resp, err := client.GetAPIKey(ctx, &authv3.GetAPIKeyRequest{
			ApiKey: apiKey,
		})

		if err != nil {
			_log.Infow("unable to get api key", "key", apiKey, "error", err)
			return err
		}

		if resp.Status == authv3.APIKeyStatus_APIKeyExists {
			data = resp.Data
		}
	}

	if data != nil {
		_, err := hasher.IDFromString(data.Partner)
		if err != nil {
			_log.Infow("unable to convert partner id", "partnerID", data.Partner, "error", err.Error())
			return err
		}
		_, err = hasher.IDFromString(data.Organization)
		if err != nil {
			_log.Infow("unable to convert organization id", "organizationID", data.Organization, "error", err.Error())
			return err
		}

		if _, ok := req.(commonv3.Metadata); ok {
			_log.Debugw("adding request meta")
			//rmo.SetPartnerID(partnerID)
			//rmo.SetOrganizationID(organizationID)
		}
	}
	if validateSession {
		if data == nil {
			return errors.New("403 Forbidden error")
		}
	}
	return nil
}

type options struct {
	pool       authv3.AuthPool
	exclude    *httprouter.Router
	dummy      bool
	logRequest bool
}

// Option is the functional arg for building auth interceptor
type Option func(*options)

// WithExclude excludes the method and path from enforcing authn/authz
func WithExclude(method, path string) Option {
	return func(opts *options) {
		if opts.exclude == nil {
			opts.exclude = httprouter.New()
		}

		opts.exclude.Handle(method, path, _dummyHandler)
	}
}

// WithAuthPool adds auth pool
func WithAuthPool(pool authv3.AuthPool) Option {
	return func(opts *options) {
		opts.pool = pool
	}
}

// WithDummy creates a dummy auth interceptor
func WithDummy() Option {
	return func(opts *options) {
		opts.dummy = true
	}
}

// WithLogRequest logs request info
func WithLogRequest() Option {
	return func(opts *options) {
		opts.logRequest = true
	}
}

// authInterceptor implements the Authentication interceptor
type authInterceptor struct {
	opts *options
}

// NewAuthInterceptorWithOptions creates auth interceptor with options
func NewAuthInterceptorWithOptions(opts ...Option) grpc.UnaryServerInterceptor {
	aOpts := &options{}
	for _, opt := range opts {
		opt(aOpts)
	}

	if aOpts.dummy {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			//addDummyMeta(req)
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if err := allowRequest(ctx, aOpts, req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}
