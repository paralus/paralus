package authv3

import (
	"os"

	logv2 "github.com/RafaySystems/rcloud-base/pkg/log"
	"github.com/RafaySystems/rcloud-base/pkg/service"
	kclient "github.com/ory/kratos-client-go"
	"github.com/uptrace/bun"
)

var _log = logv2.GetLogger()

type Option struct {
	// ExcludeRPCMethods is a list of full RPC method string in
	// format /package.service/method (for example,
	// /rafay.dev.rpc.v3.Idp/ListIdps). These RPC methods are to
	// be excluded from the auth interceptor.
	ExcludeRPCMethods []string

	// ExcludeURLs is a list of URL regular expressions that are
	// excluded from the auth middleware.
	ExcludeURLs []string
}

type authContext struct {
	kc *kclient.APIClient
	ks service.ApiKeyService
}

// NewAuthContext setup authentication and authorization dependencies.
func NewAuthContext(db *bun.DB) authContext {
	var (
		kc           *kclient.APIClient
		kratosScheme string
		kratosAddr   string
	)
	if v, ok := os.LookupEnv("KRATOS_SCHEME"); ok {
		kratosScheme = v
	} else {
		kratosScheme = "http"
	}

	if v, ok := os.LookupEnv("KRATOS_ADDR"); ok {
		kratosAddr = v
	} else {
		kratosAddr = "localhost:4433"
	}
	kratosConfig := kclient.NewConfiguration()
	kratosConfig.Servers[0].URL = kratosScheme + "://" + kratosAddr
	kc = kclient.NewAPIClient(kratosConfig)

	return authContext{kc: kc, ks: service.NewApiKeyService(db)}
}
