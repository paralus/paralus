package authv3

import (
	"os"

	logv2 "github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	kclient "github.com/ory/kratos-client-go"
)

var _log *logv2.Logger

func init() {
	_log = logv2.GetLogger()
}

type Option struct {
	// ExcludeRPCMethods is a list of full RPC method string in
	// format /package.service/method (for example,
	// /rafay.dev.rpc.v3.Idp/ListIdps). These RPC methods are to
	// be excluded from the auth interceptor.
	ExcludeRPCMethods []string
}

type authContext struct {
	kc *kclient.APIClient
}

// NewAuthContext setup authentication and authorization dependencies.
func NewAuthContext() authContext {
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

	return authContext{kc: kc}
}
