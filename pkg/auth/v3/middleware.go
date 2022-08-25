package authv3

import (
	context "context"
	"database/sql"
	"net/http"
	"regexp"
	"strings"

	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/pkg/common"
	rpcv3 "github.com/paralus/paralus/proto/rpc/v3"
	commonpbv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

type authMiddleware struct {
	db  *bun.DB
	ac  authContext
	opt Option
}

// NewAuthMiddleware creates as a middleware for the HTTP server which
// does the auth and authz by talking to kratos server and casbin
func NewAuthMiddleware(al *zap.Logger, opt Option) negroni.Handler {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(getDSN())))
	return &authMiddleware{
		ac:  SetupAuthContext(al),
		opt: opt,
		db:  bun.NewDB(sqldb, pgdialect.New()),
	}
}

type remoteAuthMiddleware struct {
	as  rpcv3.AuthServiceClient
	db  *bun.DB
	opt Option
}

// NewRemoteAuthMiddleware creates a middleware for the HTTP server
// which does auth and authz by talking to the auth service exposed by
// paralus via grpc.
func NewRemoteAuthMiddleware(al *zap.Logger, as string, opt Option) negroni.Handler {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(getDSN())))
	conn, err := grpc.Dial(as, grpc.WithInsecure())
	if err != nil {
		_log.Fatal("Unable to connect to server", err)
	}
	client := rpcv3.NewAuthServiceClient(conn)

	return &remoteAuthMiddleware{
		as:  client,
		opt: opt,
		db:  bun.NewDB(sqldb, pgdialect.New()),
	}
}

func serveHTTP(opt Option,
	db *bun.DB,
	isRequestAllowed func(context.Context, *commonpbv3.IsRequestAllowedRequest) (*commonpbv3.IsRequestAllowedResponse, error),
	rw http.ResponseWriter,
	r *http.Request,
	next http.HandlerFunc,
) {
	for _, ex := range opt.ExcludeURLs {
		match, err := regexp.MatchString(ex, r.URL.Path)
		if err != nil {
			_log.Errorf("failed to match URL expression", err)
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if match {
			next(rw, r)
			return
		}
	}
	// Auth is primarily done via grpc endpoints, this is only used
	// for endoints which do not go through grpc As of now, it is just
	// prompt.
	var poResp dao.ProjectOrg

	if strings.HasPrefix(r.URL.String(), "/v2/debug/prompt/project/") {
		// /v2/debug/prompt/project/:project/cluster/:cluster_name
		splits := strings.Split(r.URL.String(), "/")
		if len(splits) > 5 {
			// we have to fetch the org info for casbin
			res, err := dao.GetProjectOrganization(r.Context(), db, splits[5])
			if err != nil {
				_log.Errorf("Failed to authenticate: unable to find project")
				http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			_log.Info("found project with organization ", res.Organization)
			poResp = res
		}
	} else {
		// The middleware to only used with routes which does not have
		// a grpc and so fail for any other requests.
		_log.Errorf("Failed to authenticate: not a prompt request")
		http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	req := &commonpbv3.IsRequestAllowedRequest{
		Url:           r.URL.String(),
		Method:        r.Method,
		XSessionToken: r.Header.Get("X-Session-Token"),
		XApiKey:       r.Header.Get("X-API-KEYID"),
		XApiToken:     r.Header.Get("X-API-TOKEN"),
		Cookie:        r.Header.Get("Cookie"),
		Project:       poResp.Project,
		Org:           poResp.Organization,
	}
	res, err := isRequestAllowed(r.Context(), req)
	if err != nil {
		_log.Errorf("Failed to authenticate a request: %s", err)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	s := res.GetStatus()
	switch s {
	case commonpbv3.RequestStatus_RequestAllowed:
		// update the session data response to be used within prompt
		res.SessionData.Organization = poResp.OrganizationId
		res.SessionData.Partner = poResp.PartnerId
		res.SessionData.Project = &commonpbv3.ProjectData{
			List: []*commonpbv3.ProjectRole{
				{
					Project:   poResp.Project,
					ProjectId: poResp.ProjectId,
				},
			},
		}
		ctx := context.WithValue(r.Context(), common.SessionDataKey, res.SessionData)
		next(rw, r.WithContext(ctx))
		return
	case commonpbv3.RequestStatus_RequestMethodOrURLNotAllowed:
		http.Error(rw, res.GetReason(), http.StatusForbidden)
		return
	case commonpbv3.RequestStatus_RequestNotAuthenticated:
		http.Error(rw, res.GetReason(), http.StatusUnauthorized)
		return
	}

	// status is unknown
	http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// ServeHTTP function is called by the HTTP server to invoke the
// middleware
func (am *authMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	serveHTTP(am.opt, am.db, am.ac.IsRequestAllowed, rw, r, next)
}

// ServeHTTP function is called by the HTTP server to invoke the
// middleware.  Same as previous ServeHTTP, but uses
// remoteAuthMiddleware instead of authMiddleware
func (am *remoteAuthMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	serveHTTP(
		am.opt,
		am.db,
		func(ctx context.Context, isr *commonpbv3.IsRequestAllowedRequest) (*commonpbv3.IsRequestAllowedResponse, error) {
			return am.as.IsRequestAllowed(ctx, isr)
		},
		rw,
		r,
		next)
}
