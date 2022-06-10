package gateway

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

// Paralus Gateway annotations
const (
	GatewayRequest       = "x-gateway-request"
	GatewayURL           = "x-gateway-url"
	GatewaySessionCookie = "ory_kratos_session"
	GatewayAPIKey        = "X-Session-Token"
	APIKey               = "X-API-KEYID"
	APIKeyToken          = "X-API-TOKEN"
	GatewayMethod        = "x-gateway-method"
	UserAgent            = "x-gateway-user-agent"
	Host                 = "x-gateway-host"
	RemoteAddr           = "x-gateway-remote-addr"
)

// paralusGatewayAnnotator adds paralus gateway specific annotations
var paralusGatewayAnnotator = func(ctx context.Context, r *http.Request) metadata.MD {
	return metadata.New(map[string]string{
		GatewayRequest: "true",
		GatewayURL:     r.URL.EscapedPath(),
		GatewayAPIKey:  r.Header.Get(GatewayAPIKey),
		APIKey:         r.Header.Get(APIKey),
		APIKeyToken:    r.Header.Get(APIKeyToken),
		GatewayMethod:  r.Method,
		UserAgent:      r.UserAgent(),
		Host:           r.Host,
		RemoteAddr:     r.RemoteAddr,
	})
}
