package authv3

import (
	"net/http"
	"regexp"

	commonpbv3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	"github.com/urfave/negroni"
)

type authMiddleware struct {
	ac  authContext
	opt Option
}

func NewAuthMiddleware(opt Option) negroni.Handler {
	return &authMiddleware{
		ac:  NewAuthContext(),
		opt: opt,
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
	req := &commonpbv3.IsRequestAllowedRequest{
		Url:           r.URL.String(),
		Method:        r.Method,
		XSessionToken: r.Header.Get("X-Session-Token"),
		Cookie:        r.Header.Get("Cookie"),
	}
	res, err := am.ac.IsRequestAllowed(r.Context(), req)
	if err != nil {
		_log.Errorf("Failed to authenticate a request: %s", err)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	s := res.GetStatus()
	switch s {
	case commonpbv3.RequestStatus_RequestAllowed:
		ctx := NewSessionContext(r.Context(), res.SessionData)
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
	return
}
