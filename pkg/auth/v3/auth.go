package authv3

import (
	"os"

	"github.com/RafayLabs/rcloud-base/pkg/enforcer"
	logv2 "github.com/RafayLabs/rcloud-base/pkg/log"
	"github.com/RafayLabs/rcloud-base/pkg/service"
	"github.com/RafayLabs/rcloud-base/pkg/enforcer"
	kclient "github.com/ory/kratos-client-go"
	"github.com/uptrace/bun"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	as service.AuthzService
}

// NewAuthContext setup authentication and authorization dependencies.
func NewAuthContext(db *bun.DB) authContext {
	var (
		kc           *kclient.APIClient
		kratosScheme string
		kratosAddr   string
	)
	// TODO: https://github.com/RafayLabs/prompt/pull/3#issuecomment-1073557206
	// Where exactly should we be getting these values from?
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

	gormDb, err := gorm.Open(
		postgres.New(postgres.Config{Conn: db.DB}),
		&gorm.Config{},
	)
	if err != nil {
		_log.Fatalw("unable to create db connection", "error", err)
	}
	enforcer, err := enforcer.NewCasbinEnforcer(gormDb).Init()
	if err != nil {
		_log.Fatalw("unable to init enforcer", "error", err)
	}
	as := service.NewAuthzService(db, enforcer)

	return authContext{kc: kc, as: as, ks: service.NewApiKeyService(db)}
}
