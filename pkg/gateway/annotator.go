package gateway

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

// Rafay Gateway annotations
const (
	GatewayRequest       = "x-gateway-request"
	GatewayURL           = "x-gateway-url"
	GatewaySessionCookie = "ory_kratos_session"
	GatewayAPIKey        = "X-Session-Token"
	GatewayMethod        = "x-gateway-method"
	UserAgent            = "x-gateway-user-agent"
	Host                 = "x-gateway-host"
	RemoteAddr           = "x-gateway-remote-addr"
)

// rafayGatewayAnnotator adds rafay gateway specific annotations
var rafayGatewayAnnotator = func(ctx context.Context, r *http.Request) metadata.MD {
	return metadata.New(map[string]string{
		GatewayRequest: "true",
		GatewayURL:     r.URL.EscapedPath(),
		// GatewaySessionCookie: func() string {
		// 	sid, err := r.Cookie(GatewaySessionCookie)
		// 	if err != nil {
		// 		return ""
		// 	}
		// 	return sid.Value
		// }(),
		GatewayAPIKey: r.Header.Get(GatewayAPIKey),
		GatewayMethod: r.Method,
		UserAgent:     r.UserAgent(),
		Host:          r.Host,
		RemoteAddr:    r.RemoteAddr,
	})
}
