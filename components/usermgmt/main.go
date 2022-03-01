package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	kclient "github.com/ory/kratos-client-go"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	authzrpcv1 "github.com/RafaySystems/rcloud-base/components/authz/proto/rpc/v1"
	authv3 "github.com/RafaySystems/rcloud-base/components/common/pkg/auth/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/gateway"
	grpc "github.com/RafaySystems/rcloud-base/components/common/pkg/grpc"
	logv2 "github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	configrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/config"
	rolerpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/role"
	systemrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/system"
	userrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/user"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/providers"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/service"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/server"
	_grpc "google.golang.org/grpc"
)

const (
	rpcPortEnv      = "RPC_PORT"
	apiPortEnv      = "API_PORT"
	authzPortEnv    = "AUTHZ_SERVER_PORT"
	debugPortEnv    = "DEBUG_PORT"
	kratosSchemeEnv = "KRATOS_SCHEME"
	kratosAddrEnv   = "KRATOS_ADDR"
	dbAddrEnv       = "DB_ADDR"
	dbNameEnv       = "DB_NAME"
	dbUserEnv       = "DB_USER"
	dbPasswordEnv   = "DB_PASSWORD"
	devEnv          = "DEV"
	configAddrENV   = "CONFIG_ADDR"
	appHostHTTPEnv  = "APP_HOST_HTTP"
)

var (
	rpcPort             int
	apiPort             int
	debugPort           int
	rpcRelayPeeringPort int
	kratosScheme        string
	kratosAddr          string
	authzPort           int
	kc                  *kclient.APIClient
	azc                 authzrpcv1.AuthzClient
	dbAddr              string
	dbName              string
	dbUser              string
	dbPassword          string
	db                  *bun.DB
	us                  service.UserService
	gs                  service.GroupService
	rs                  service.RoleService
	rrs                 service.RolepermissionService
	is                  service.IdpService
	ps                  service.OIDCProviderService
	dev                 bool
	_log                = logv2.GetLogger()
	authPool            authv3.AuthPool
	configPool          configrpc.ConfigPool
	configAddr          string
	appHostHTTP         string
)

func setup() {
	viper.SetDefault(rpcPortEnv, 14000)
	viper.SetDefault(apiPortEnv, 15000)
	viper.SetDefault(debugPortEnv, 16000)
	viper.SetDefault(kratosSchemeEnv, "http")
	viper.SetDefault(kratosAddrEnv, "localhost:4433")
	viper.SetDefault(authzPortEnv, 50011)
	viper.SetDefault(dbAddrEnv, "localhost:5432")
	viper.SetDefault(dbNameEnv, "admindb")
	viper.SetDefault(dbUserEnv, "admindbuser")
	viper.SetDefault(dbPasswordEnv, "admindbpassword")
	viper.SetDefault(devEnv, true)
	viper.SetDefault(configAddrENV, "localhost:7000")
	viper.SetDefault(appHostHTTPEnv, "http://localhost:11000")

	viper.BindEnv(rpcPortEnv)
	viper.BindEnv(apiPortEnv)
	viper.BindEnv(debugPortEnv)
	viper.BindEnv(kratosSchemeEnv)
	viper.BindEnv(kratosAddrEnv)
	viper.BindEnv(authzPortEnv)
	viper.BindEnv(dbAddrEnv)
	viper.BindEnv(dbNameEnv)
	viper.BindEnv(dbPasswordEnv)
	viper.BindEnv(dbPasswordEnv)
	viper.BindEnv(devEnv)
	viper.BindEnv(configAddrENV)
	viper.BindEnv(appHostHTTPEnv)

	rpcPort = viper.GetInt(rpcPortEnv)
	apiPort = viper.GetInt(apiPortEnv)
	debugPort = viper.GetInt(debugPortEnv)
	kratosScheme = viper.GetString(kratosSchemeEnv)
	kratosAddr = viper.GetString(kratosAddrEnv)
	authzPort = viper.GetInt(authzPortEnv)
	dbAddr = viper.GetString(dbAddrEnv)
	dbName = viper.GetString(dbNameEnv)
	dbUser = viper.GetString(dbUserEnv)
	dbPassword = viper.GetString(dbPasswordEnv)
	dev = viper.GetBool(devEnv)
	configAddr = viper.GetString(configAddrENV)
	appHostHTTP = viper.GetString(appHostHTTPEnv)

	rpcRelayPeeringPort = rpcPort + 1

	// Kratos client setup
	kratosConfig := kclient.NewConfiguration()
	kratosUrl := kratosScheme + "://" + kratosAddr
	kratosConfig.Servers[0].URL = kratosUrl
	kc = kclient.NewAPIClient(kratosConfig)

	dsn := "postgres://" + dbUser + ":" + dbPassword + "@" + dbAddr + "/" + dbName + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	if dev {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	conn, err := _grpc.Dial(":"+fmt.Sprint(authzPort), _grpc.WithInsecure())
	if err != nil {
		log.Fatal("unable to connect to authz")
	}
	azc = authzrpcv1.NewAuthzClient(conn)

	us = service.NewUserService(providers.NewKratosAuthProvider(kc), db, azc)
	gs = service.NewGroupService(db, azc)
	rs = service.NewRoleService(db, azc)
	rrs = service.NewRolepermissionService(db)
	is = service.NewIdpService(db, appHostHTTP)
	ps = service.NewOIDCProviderService(db, kratosUrl)
	_log.Infow("usermgmt setup complete")
}

func run() {

	ctx := signals.SetupSignalHandler()

	var wg sync.WaitGroup

	wg.Add(1)

	go runAPI(&wg, ctx)
	go runRPC(&wg, ctx)
	go runDebug(&wg, ctx)

	<-ctx.Done()
	wg.Wait()
}

func runAPI(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := http.NewServeMux()

	gwHandler, err := gateway.NewGateway(
		ctx,
		fmt.Sprintf(":%d", rpcPort),
		make([]runtime.ServeMuxOption, 0),
		userrpc.RegisterUserHandlerFromEndpoint,
		userrpc.RegisterGroupHandlerFromEndpoint,
		rolerpc.RegisterRoleHandlerFromEndpoint,
		rolerpc.RegisterRolepermissionHandlerFromEndpoint,
		systemrpc.RegisterIdpHandlerFromEndpoint,
		systemrpc.RegisterOIDCProviderHandlerFromEndpoint,
	)
	if err != nil {
		_log.Fatalw("unable to create gateway", "error", err)
	}
	mux.Handle("/", gwHandler)

	s := http.Server{
		Addr:    fmt.Sprintf(":%d", apiPort),
		Handler: mux,
	}
	go func() {
		defer s.Shutdown(context.TODO())
		<-ctx.Done()
	}()

	_log.Infow("starting gateway server", "port", apiPort)
	err = s.ListenAndServe()
	if err != nil {
		_log.Fatalw("unable to start gateway", "error", err)
	}
}

func runRPC(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	defer gs.Close()
	defer rs.Close()
	defer rrs.Close()

	userServer := server.NewUserServer(us)
	groupServer := server.NewGroupServer(gs)
	roleServer := server.NewRoleServer(rs)
	rolepermissionServer := server.NewRolePermissionServer(rrs)
	idpServer := server.NewIdpServer(is)
	oidcProviderServer := server.NewOIDCServer(ps)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		_log.Fatalw("unable to start rpc listener", "error", err)
	}

	var opts []_grpc.ServerOption
	if !dev {
		ac := authv3.NewAuthContext()
		o := authv3.Option{}
		opts = append(opts, _grpc.UnaryInterceptor(
			ac.NewAuthUnaryInterceptor(o),
		))
	}
	s, err := grpc.NewServer(opts...)
	if err != nil {
		_log.Fatalw("unable to create grpc server", "error", err)
	}

	go func() {
		defer s.GracefulStop()

		<-ctx.Done()
		_log.Infow("context done")
	}()

	userrpc.RegisterUserServer(s, userServer)
	userrpc.RegisterGroupServer(s, groupServer)
	rolerpc.RegisterRoleServer(s, roleServer)
	rolerpc.RegisterRolepermissionServer(s, rolepermissionServer)
	systemrpc.RegisterIdpServer(s, idpServer)
	systemrpc.RegisterOIDCProviderServer(s, oidcProviderServer)

	_log.Infow("starting rpc server", "port", rpcPort)
	err = s.Serve(l)
	if err != nil {
		_log.Fatalw("unable to start rpc server", "error", err)
	}
}

func runDebug(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	s := http.Server{
		Addr: fmt.Sprintf(":%d", debugPort),
	}
	_log.Infow("starting debug server", "port", debugPort)
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			_log.Fatalw("unable to start debug server", "error", err)
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	s.Shutdown(ctx)
}

func main() {
	setup()
	run()
}
