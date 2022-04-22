package authv3

import (
	context "context"
	"database/sql"
	"net/http"
	"regexp"
	"strings"

	"github.com/RafayLabs/rcloud-base/internal/dao"
	"github.com/RafayLabs/rcloud-base/pkg/common"
	commonpbv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
)

type authMiddleware struct {
	db  *bun.DB
	ac  authContext
	opt Option
}

func NewAuthMiddleware(al *zap.Logger, opt Option) negroni.Handler {
	// Initialize database
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(getDSN())))
	return &authMiddleware{
		ac:  SetupAuthContext(al),
		opt: opt,
		db:  bun.NewDB(sqldb, pgdialect.New()),
	}
}

func (am *authMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	for _, ex := range am.opt.ExcludeURLs {
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
	var proj string
	var org string

	if strings.HasPrefix(r.URL.String(), "/v2/debug/prompt/project/") {
		// /v2/debug/prompt/project/:project/cluster/:cluster_name
		splits := strings.Split(r.URL.String(), "/")
		if len(splits) > 5 {
			// we have to fetch the org info for casbin
			proj, org, err := dao.GetProjectOrganization(r.Context(), am.db, splits[5])
			if err != nil {
				_log.Errorf("Failed to authenticate: unable to find project")
				http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			_log.Info("found project with organization %s %s", proj, org)
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
		XApiKey:       r.Header.Get("X-RAFAY-API-KEYID"),
		Cookie:        r.Header.Get("Cookie"),
		Project:       proj,
		Org:           org,
	}
	res, err := am.ac.IsRequestAllowed(r.Context(), r, req)
	if err != nil {
		_log.Errorf("Failed to authenticate a request: %s", err)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	s := res.GetStatus()
	switch s {
	case commonpbv3.RequestStatus_RequestAllowed:
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
