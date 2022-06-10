package authv3

import (
	"database/sql"
	"fmt"
	"os"

	kclient "github.com/ory/kratos-client-go"
	"github.com/paralus/paralus/pkg/enforcer"
	logv2 "github.com/paralus/paralus/pkg/log"
	"github.com/paralus/paralus/pkg/service"
	"github.com/uptrace/bun"
	"go.uber.org/zap"

	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var _log = logv2.GetLogger()

type Option struct {
	// ExcludeRPCMethods is a list of full RPC method string in
	// format /package.service/method (for example,
	// /paralus.dev.rpc.v3.Idp/ListIdps). These RPC methods are to
	// be excluded from the auth interceptor.
	ExcludeRPCMethods []string

	// ExcludeURLs is a list of URL regular expressions that are
	// excluded from the auth middleware.
	ExcludeURLs []string

	// ExcludeAuthzMethods is a list of RPC method strings which only
	// do authentication and not authorization.
	ExcludeAuthzMethods []string
}

type authContext struct {
	db *bun.DB
	kc *kclient.APIClient
	ks service.ApiKeyService
	as service.AuthzService
}

// SetupAuthContext sets up new authContext along with its
// dependencies. If the caller already has instances of authContext
// fields created then use NewAuthContext instead.
func SetupAuthContext(auditLogger *zap.Logger) authContext {
	var (
		kc         *kclient.APIClient
		kratosAddr string
		db         *bun.DB
	)

	// Initialize database
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(getDSN())))
	db = bun.NewDB(sqldb, pgdialect.New())

	if v, ok := os.LookupEnv("KRATOS_PUB_ADDR"); ok {
		kratosAddr = v
	} else {
		kratosAddr = "http://localhost:4433"
	}
	kratosConfig := kclient.NewConfiguration()
	kratosConfig.Servers[0].URL = kratosAddr
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

	return authContext{db: db, kc: kc, as: as, ks: service.NewApiKeyService(db, auditLogger)}
}

func getDSN() string {
	dsn := getEnvWithDefault("DSN", "")
	if dsn == "" {
		dbUser := getEnvWithDefault("DB_USER", "admindbuser")
		dbPassword := getEnvWithDefault("DB_PASSWORD", "admindbpassword")
		dbAddr := getEnvWithDefault("DB_ADDR", "localhost:5432")
		dbName := getEnvWithDefault("DB_NAME", "admindb")
		dsn = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbAddr, dbName)
	}
	return dsn
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
	db *bun.DB,
	kc *kclient.APIClient,
	apiKeySvc service.ApiKeyService,
	authzSvc service.AuthzService,
) authContext {
	return authContext{
		db: db,
		kc: kc,
		ks: apiKeySvc,
		as: authzSvc,
	}
}
