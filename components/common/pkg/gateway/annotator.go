package gateway

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

// Rafay Gateway annotations
const (
	GatewayRequest     = "rafay-gateway-request"
	GatewayURL         = "rafay-gateway-url"
	GatewayRSID        = "rafay-gateway-rsid"
	GatewayAPIKey      = "rafay-gateway-apikey"
	GatewayMethod      = "rafay-gateway-method"
	rafaySessionCookie = "rsid"
	rafayAPIKeyHeader  = "X-RAFAY-API-KEYID"
	UserAgent          = "rafay-gateway-user-agent"
	Host               = "rafay-gateway-host"
	RemoteAddr         = "rafay-gateway-remote-addr"
)

// rafayGatewayAnnotator adds rafay gateway specific annotations
var rafayGatewayAnnotator = func(ctx context.Context, r *http.Request) metadata.MD {
	return metadata.New(map[string]string{
		GatewayRequest: "true",
		GatewayURL:     r.URL.EscapedPath(),
		GatewayRSID: func() string {
			sid, err := r.Cookie(rafaySessionCookie)
			if err != nil {
				return ""
			}
			return sid.Value
		}(),
		GatewayAPIKey: r.Header.Get(rafayAPIKeyHeader),
		GatewayMethod: r.Method,
		UserAgent:     r.UserAgent(),
		Host:          r.Host,
		RemoteAddr:    r.RemoteAddr,
	})
}
