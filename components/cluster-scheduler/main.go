package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	goruntime "runtime"
	"sync"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/bootstrapper"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/credentials"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/notify"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/reconcile"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/service"
	adminrpc "github.com/RafaySystems/rcloud-base/components/cluster-scheduler/proto/rpc"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/server"
	authv3 "github.com/RafaySystems/rcloud-base/components/common/pkg/auth/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/gateway"
	grpcutil "github.com/RafaySystems/rcloud-base/components/common/pkg/grpc"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/leaderelection"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	configrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/config"
	schedulerrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/scheduler"
	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/xid"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	rpcPortEnv             = "RPC_PORT"
	apiPortEnv             = "API_PORT"
	configAddrENV          = "CONFIG_ADDR"
	sentryAddrENV          = "SENTRY_ADDR"
	dbAddrEnv              = "DB_ADDR"
	dbNameEnv              = "DB_NAME"
	dbUserEnv              = "DB_USER"
	dbPasswordEnv          = "DB_PASSWORD"
	adbNameEnv             = "ADMIN_DB_NAME"
	adbUserEnv             = "ADMIN_DB_USER"
	adbPasswordEnv         = "ADMIN_DB_PASSWORD"
	devEnv                 = "DEV"
	secretPathEnv          = "SECRET_PATH"
	controlAddrEnv         = "CONTROL_ADDR"
	apiAddrEnv             = "API_ADDR"
	authAddrEnv            = "AUTH_ADDR"
	caCert                 = "ca.pem"
	caKey                  = "ca-key.pem"
	relayAgentImageEnv     = "RELAY_AGENT_IMAGE"
	sentryBootstrapAddrENV = "SENTRY_BOOTSTRAP_ADDR"
	podNameEnv             = "POD_NAME"
	schedulerNamespaceEnv  = "SCHEDULER_NAMESPACE"
	maxSendSize            = 16 * 1024 * 1024
	maxRecvSize            = 16 * 1024 * 1024
)

var (
	rpcPort            int
	apiPort            int
	configAddr         string
	controlAddr        string
	sentryAddr         string
	apiAddr            string
	dbAddr             string
	dbName             string
	dbUser             string
	dbPassword         string
	adbName            string
	adbUser            string
	adbPassword        string
	dev                bool
	secretPath         string
	podName            string
	_log               = log.GetLogger()
	db                 *bun.DB
	adb                *bun.DB
	cs                 service.ClusterService
	ms                 service.MetroService
	signer             credentials.Signer
	allowReuseToken    string // featureflag
	relayAgentImage    string
	downloadData       *bootstrapper.DownloadData
	authAddr           string
	schedulerNamespace string
	authPool           authv3.AuthPool
	configPool         configrpc.ConfigPool
	sentryPool         sentryrpc.SentryPool
)

func setup() {
	viper.SetDefault(apiPortEnv, 8000)
	viper.SetDefault(rpcPortEnv, 5000)
	viper.SetDefault(dbAddrEnv, "localhost:5432")
	viper.SetDefault(dbNameEnv, "clusterdb")
	viper.SetDefault(dbUserEnv, "clusterdbuser")
	viper.SetDefault(dbPasswordEnv, "clusterdbpassword")
	viper.SetDefault(adbNameEnv, "admindb")
	viper.SetDefault(adbUserEnv, "admindbuser")
	viper.SetDefault(adbPasswordEnv, "admindbpassword")
	viper.SetDefault(devEnv, true)
	viper.SetDefault(secretPathEnv, "/home/infracloud/Documents/warehouse/rafay/test data/cluster-scheduler")
	viper.SetDefault(configAddrENV, ":7000")
	viper.SetDefault(allowReuseToken, "true")
	viper.SetDefault(controlAddrEnv, "localhost:5002")
	viper.SetDefault(apiAddrEnv, "localhost:8000")
	viper.SetDefault(relayAgentImageEnv, "rafaysystems/rafay-relay:latest")
	viper.SetDefault(authAddrEnv, "authsrv.rcloud-admin.svc.cluster.local:50011")
	viper.SetDefault(sentryAddrENV, "localhost:10000")
	viper.SetDefault(sentryBootstrapAddrENV, "api.sentry.rafay.local:11000")
	viper.SetDefault(podNameEnv, "local-pod-0")
	viper.SetDefault(schedulerNamespaceEnv, "rafay-system")

	viper.BindEnv(rpcPortEnv)
	viper.BindEnv(apiPortEnv)
	viper.BindEnv(dbAddrEnv)
	viper.BindEnv(dbNameEnv)
	viper.BindEnv(dbPasswordEnv)
	viper.BindEnv(dbPasswordEnv)
	viper.BindEnv(devEnv)
	viper.BindEnv(secretPathEnv)
	viper.BindEnv(configAddrENV)
	viper.BindEnv(allowReuseToken)
	viper.BindEnv(controlAddrEnv)
	viper.BindEnv(apiAddrEnv)
	viper.BindEnv(relayAgentImageEnv)
	viper.BindEnv(authAddrEnv)
	viper.BindEnv(sentryAddrENV)
	viper.BindEnv(sentryBootstrapAddrENV)
	viper.BindEnv(podNameEnv)
	viper.BindEnv(schedulerNamespaceEnv)

	rpcPort = viper.GetInt(rpcPortEnv)
	apiPort = viper.GetInt(apiPortEnv)
	dbAddr = viper.GetString(dbAddrEnv)
	dbName = viper.GetString(dbNameEnv)
	dbUser = viper.GetString(dbUserEnv)
	dbPassword = viper.GetString(dbPasswordEnv)
	adbName = viper.GetString(adbNameEnv)
	adbUser = viper.GetString(adbUserEnv)
	adbPassword = viper.GetString(adbPasswordEnv)
	dev = viper.GetBool(devEnv)
	secretPath = viper.GetString(secretPathEnv)
	configAddr = viper.GetString(configAddrENV)
	allowReuseToken = viper.GetString(allowReuseToken)
	controlAddr = viper.GetString(controlAddrEnv)
	apiAddr = viper.GetString(apiAddrEnv)
	relayAgentImage = viper.GetString(relayAgentImageEnv)
	authAddr = viper.GetString(authAddrEnv)
	sentryAddr = viper.GetString(sentryAddrENV)
	podName = viper.GetString(podNameEnv)
	schedulerNamespace = viper.GetString(schedulerNamespaceEnv)

	dsn := "postgres://" + dbUser + ":" + dbPassword + "@" + dbAddr + "/" + dbName + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db = bun.NewDB(sqldb, pgdialect.New())

	//TODO: this should be microservice call to admin service
	dsn = "postgres://" + adbUser + ":" + adbPassword + "@" + dbAddr + "/" + adbName + "?sslmode=disable"
	sqldb = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	adb = bun.NewDB(sqldb, pgdialect.New())

	if dev {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	_log.Infow("setup finished",
		"rpcPort", rpcPort,
		"apiPort", apiPort,
		"dbAddr", dbAddr,
		"dbName", dbName,
		"dbUser", dbUser,
		"dbPassword", dbPassword,
		"secretPath", secretPath,
		"controlAddr", controlAddr,
		"apiAddr", apiAddr,
		"configAddr", configAddr,
		"dev", dev,
		"authAddr", authAddr,
		"sentryAddr", sentryAddr,
		"podName", podName,
		"schedulerNamespace", schedulerNamespace,
	)

	if dev {
		lc := make(chan string)
		go _log.ChangeLevel(lc)
		lc <- "debug"
		_log.Debugw("Debug mode set in log because this is a dev environment")
	}

	downloadData = &bootstrapper.DownloadData{
		ControlAddr:     controlAddr,
		APIAddr:         apiAddr,
		RelayAgentImage: relayAgentImage,
	}

	//TODO: add auth pool as required
	configPool = configrpc.NewConfigPool(configAddr, 5*goruntime.NumCPU())
	sentryPool = sentryrpc.NewSentryPool(sentryAddr, 5*goruntime.NumCPU())

	cs = service.NewClusterService(db, adb, downloadData, sentryPool)
	ms = service.NewMetroService(db, adb)

	notify.Init(cs)

	var err error
	signer, err = credentials.NewSigner(secretPath)

	if err != nil {
		_log.Fatalw("unable to create signer", "error", err)
	}

	_log.Infow("queried number of cpus", "numCPUs", goruntime.NumCPU())

}

func run() {
	ctx := signals.SetupSignalHandler()

	notify.Start(ctx.Done())

	healthServer := grpc.NewServer()

	// health server
	_log.Infow("registering grpc health server")
	hs := health.NewServer()
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(healthServer, hs)
	_log.Infow("registerd grpc health server")

	var wg sync.WaitGroup
	wg.Add(5)
	go runAPI(&wg, ctx)
	go runRPC(&wg, ctx)
	go runEventHandlers(&wg, ctx)

	<-ctx.Done()
	_log.Infow("shutting down, waiting for children to die")
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
		schedulerrpc.RegisterClusterHandlerFromEndpoint,
		adminrpc.RegisterLocationHandlerFromEndpoint,
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
	defer cs.Close()
	defer configPool.Close()

	crpc := server.NewClusterServer(cs, signer, downloadData)
	mserver := server.NewLocationServer(ms)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		_log.Fatalw("unable to start rpc listener", "error", err)
	}

	var opts []grpc.ServerOption
	if !dev {
		_log.Infow("adding auth interceptor")
		ac := authv3.NewAuthContext()
		o := authv3.Option{ExcludeRPCMethods: []string{"/rafay.dev.scheduler.rpc.Cluster/RegisterCluster"}}
		opts = append(opts, grpc.UnaryInterceptor(
			ac.NewAuthUnaryInterceptor(o),
		))
	}
	s := grpc.NewServer(opts...)
	if err != nil {
		_log.Fatalw("unable to create grpc server", "error", err)
	}

	s, err = grpcutil.NewServer(opts...)
	if err != nil {
		_log.Fatalw("unable to create grpc server", "error", err)
	}

	go func() {
		defer s.GracefulStop()

		<-ctx.Done()
		_log.Infow("context done")
	}()

	// register all the rpc servers
	schedulerrpc.RegisterClusterServer(s, crpc)
	adminrpc.RegisterLocationServer(s, mserver)

	_log.Infow("starting rpc server", "port", rpcPort)
	err = s.Serve(l)
	if err != nil {
		_log.Fatalw("unable to start rpc server", "error", err)
	}
}

func runEventHandlers(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	//TODO: need to add a bunch of other handlers with gitops
	ceh := reconcile.NewClusterEventHandler(cs, sentryPool, configPool)
	_log.Infow("starting cluster event handler")
	go ceh.Handle(ctx.Done())

	// listen to cluster events
	cs.AddEventHandler(ceh.ClusterHook())

	if !dev {
		rl, err := leaderelection.NewConfigMapLock("cluster-scheduler", schedulerNamespace, xid.New().String())
		if err != nil {
			_log.Fatalw("unable to create configmap lock", "error", err)
		}
		go func() {
			err := leaderelection.Run(rl, func(stop <-chan struct{}) {
			}, ctx.Done())

			if err != nil {
				_log.Fatalw("unable to run leader election", "error", err)
			}
		}()
	}

	<-ctx.Done()
}

func main() {
	setup()
	run()
}
