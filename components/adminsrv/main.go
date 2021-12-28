package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	goruntime "runtime"
	"sync"
	"time"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/service"
	pbrpcv3 "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/rpc/v3"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/rpc/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/auth/interceptors"
	authv3 "github.com/RafaySystems/rcloud-base/components/common/pkg/auth/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/gateway"
	logv2 "github.com/RafaySystems/rcloud-base/components/common/pkg/log/v2"
	configrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/config"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"google.golang.org/grpc"
	_grpc "google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	rpcPortEnv    = "RPC_PORT"
	apiPortEnv    = "API_PORT"
	debugPortEnv  = "DEBUG_PORT"
	dbAddrEnv     = "DB_ADDR"
	dbNameEnv     = "DB_NAME"
	dbUserEnv     = "DB_USER"
	dbPasswordEnv = "DB_PASSWORD"
	devEnv        = "DEV"
	configAddrENV = "CONFIG_ADDR"
)

var (
	rpcPort             int
	apiPort             int
	debugPort           int
	rpcRelayPeeringPort int
	dbAddr              string
	dbName              string
	dbUser              string
	dbPassword          string
	dev                 bool
	db                  *bun.DB
	ps                  service.PartnerService
	_log                = logv2.GetLogger()
	authPool            authv3.AuthPool
	configPool          configrpc.ConfigPool
	configAddr          string
)

func setup() {
	viper.SetDefault(rpcPortEnv, 10000)
	viper.SetDefault(apiPortEnv, 11000)
	viper.SetDefault(debugPortEnv, 12000)
	viper.SetDefault(dbAddr, ":5432")
	viper.SetDefault(dbNameEnv, "admindb")
	viper.SetDefault(dbUserEnv, "admindbuser")
	viper.SetDefault(dbPasswordEnv, "admindbpassword")
	viper.SetDefault(devEnv, true)
	viper.SetDefault(configAddrENV, "localhost:7000")

	viper.BindEnv(rpcPortEnv)
	viper.BindEnv(apiPortEnv)
	viper.BindEnv(debugPortEnv)
	viper.BindEnv(dbAddrEnv)
	viper.BindEnv(dbNameEnv)
	viper.BindEnv(dbPasswordEnv)
	viper.BindEnv(dbPasswordEnv)
	viper.BindEnv(devEnv)
	viper.BindEnv(dbUserEnv)
	viper.BindEnv(configAddrENV)

	rpcPort = viper.GetInt(rpcPortEnv)
	apiPort = viper.GetInt(apiPortEnv)
	debugPort = viper.GetInt(debugPortEnv)
	dbAddr = viper.GetString(dbAddrEnv)
	dbName = viper.GetString(dbNameEnv)
	dbUser = viper.GetString(dbUserEnv)
	dbPassword = viper.GetString(dbPasswordEnv)
	dev = viper.GetBool(devEnv)
	configAddr = viper.GetString(configAddrENV)

	rpcRelayPeeringPort = rpcPort + 1

	dsn := "postgres://admindbuser:admindbpassword@localhost:5432/admindb?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	_log.Infow("printing db", "db", db)

	configPool = configrpc.NewConfigPool(configAddr, 5*goruntime.NumCPU())

	ps = service.NewPartnerService(db)
}

func run() {

	ctx := signals.SetupSignalHandler()

	var wg sync.WaitGroup

	wg.Add(4)

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
		pbrpcv3.RegisterPartnerHandlerFromEndpoint,
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
	defer ps.Close()
	defer configPool.Close()

	partnerServer := rpcv3.NewPartnerServer(ps)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		_log.Fatalw("unable to start rpc listener", "error", err)
	}

	var opts []_grpc.ServerOption
	if !dev {
		opts = append(opts, _grpc.UnaryInterceptor(
			interceptors.NewAuthInterceptorWithOptions(
				interceptors.WithLogRequest(),
				interceptors.WithAuthPool(authPool),
				interceptors.WithExclude("POST", "/v2/sentry/bootstrap/:templateToken/register"),
			),
		))
		defer authPool.Close()
	} else {
		opts = append(opts, _grpc.UnaryInterceptor(
			interceptors.NewAuthInterceptorWithOptions(interceptors.WithDummy())),
		)
	}
	s := grpc.NewServer(opts...)
	if err != nil {
		_log.Fatalw("unable to create grpc server", "error", err)
	}

	go func() {
		defer s.GracefulStop()

		<-ctx.Done()
		_log.Infow("context done")
	}()

	rpcv3.RegisterPartnerServer(s, partnerServer)

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
