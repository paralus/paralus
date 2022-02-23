package authv3

import (
	"net/http"

	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"github.com/urfave/negroni"
)

type authMiddleware struct {
	ac authContext
}

// Not maintained. Instead use gRPC interceptor for authentication.
func (ac authContext) NewAuthMiddleware() negroni.Handler {
	return &authMiddleware{ac}
}

func (am *authMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	req := &commonv3.IsRequestAllowedRequest{
		Url:           r.URL.String(),
		Method:        r.Method,
		XSessionToken: r.Header.Get("X-Session-Token"),
		Cookie:        r.Header.Get("Cookie"),
	}
	res, err := am.ac.IsRequestAllowed(r.Context(), req)
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if res.GetStatus() == commonv3.RequestStatus_RequestNotAuthenticated {
		http.Error(rw, res.GetReason(), http.StatusUnauthorized)
		return
	} else if res.GetStatus() == commonv3.RequestStatus_RequestMethodOrURLNotAllowed {
		http.Error(rw, res.GetReason(), http.StatusForbidden)
		return
	}

	if res.GetStatus() == commonv3.RequestStatus_RequestAllowed {
		next(rw, r)
	}
}
