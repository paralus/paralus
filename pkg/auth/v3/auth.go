package authv3

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/RafayLabs/rcloud-base/pkg/enforcer"
	logv2 "github.com/RafayLabs/rcloud-base/pkg/log"
	"github.com/RafayLabs/rcloud-base/pkg/service"
	kclient "github.com/ory/kratos-client-go"
	"github.com/uptrace/bun"

	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
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

// SetupAuthContext sets up new authContext along with its
// dependencies. If the caller already has instances of authContext
// fields created then use NewAuthContext instead.
func SetupAuthContext() authContext {
	var (
		kc           *kclient.APIClient
		kratosScheme string
		kratosAddr   string
		db           *bun.DB
	)

	// Initialize database
	dbUser := getEnvWithDefault("DB_USER", "admindbuser")
	dbPassword := getEnvWithDefault("DB_PASSWORD", "admindbpassword")
	dbAddr := getEnvWithDefault("DB_ADDR", "localhost:5432")
	dbName := getEnvWithDefault("DB_NAME", "admindb")
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbAddr, dbName)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db = bun.NewDB(sqldb, pgdialect.New())

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

func getEnvWithDefault(env, def string) string {
	val := os.Getenv(env)
	if val == "" {
		return def
	}
	return val
}

// NewAuthContext instantiate authContext. NewAuthContext creates
// authContext reusing dependency instances from calling function
// instead of creating new instances. To create authContext along with
// its dependencies, use SetupAuthContext.
func NewAuthContext(
	kc *kclient.APIClient,
	apiKeySvc service.ApiKeyService,
	authzSvc service.AuthzService,
) authContext {
	return authContext{
		kc: kc,
		ks: apiKeySvc,
		as: authzSvc,
	}
}
