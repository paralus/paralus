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
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/server"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/cluster-scheduler/proto/rpc/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/auth/interceptors"
	authv3 "github.com/RafaySystems/rcloud-base/components/common/pkg/auth/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/gateway"
	grpcutil "github.com/RafaySystems/rcloud-base/components/common/pkg/grpc"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	configrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/config"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
	peerPortENV            = "PEER_PORT"
	connectorPortENV       = "CONNECTOR_PORT"
	configAddrENV          = "CONFIG_ADDR"
	sentryAddrENV          = "SENTRY_ADDR"
	dbAddrEnv              = "DB_ADDR"
	dbNameEnv              = "DB_NAME"
	dbUserEnv              = "DB_USER"
	dbPasswordEnv          = "DB_PASSWORD"
	devEnv                 = "DEV"
	secretPathEnv          = "SECRET_PATH"
	controlAddrEnv         = "CONTROL_ADDR"
	apiAddrEnv             = "API_ADDR"
	authAddrEnv            = "AUTH_ADDR"
	connectorCert          = "connector.pem"
	connectorKey           = "connector-key.pem"
	caCert                 = "ca.pem"
	caKey                  = "ca-key.pem"
	controllerImageEnv     = "CONTROLLER_IMAGE"
	connectorImageEnv      = "CONNECTOR_IMAGE"
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
	peerPort           int
	connectorPort      int
	configAddr         string
	controlAddr        string
	sentryAddr         string
	apiAddr            string
	dbAddr             string
	dbName             string
	dbUser             string
	dbPassword         string
	dev                bool
	secretPath         string
	podName            string
	_log               = log.GetLogger()
	db                 *bun.DB
	cs                 service.ClusterService
	signer             credentials.Signer
	allowReuseToken    string // featureflag
	controllerImage    string
	connectorImage     string
	relayAgentImage    string
	downloadData       *bootstrapper.DownloadData
	authAddr           string
	schedulerNamespace string
	authPool           authv3.AuthPool
	configPool         configrpc.ConfigPool
)

func setup() {
	viper.SetDefault(apiPortEnv, 8000)
	viper.SetDefault(rpcPortEnv, 5000)
	viper.SetDefault(peerPortENV, 5001)
	viper.SetDefault(connectorPortENV, 5002)
	viper.SetDefault(dbAddrEnv, "localhost:5432")
	viper.SetDefault(dbNameEnv, "clusterdb")
	viper.SetDefault(dbUserEnv, "clusterdbuser")
	viper.SetDefault(dbPasswordEnv, "clusterdbpassword")
	viper.SetDefault(devEnv, true)
	viper.SetDefault(secretPathEnv, "/secrets")
	viper.SetDefault(configAddrENV, ":7000")
	viper.SetDefault(allowReuseToken, "true")
	viper.SetDefault(controlAddrEnv, "localhost:5002")
	viper.SetDefault(apiAddrEnv, "localhost:8000")
	viper.SetDefault(controllerImageEnv, "rafaysystems/cluster-controller:latest")
	viper.SetDefault(connectorImageEnv, "rafaysystems/rafay-connector:latest")
	viper.SetDefault(relayAgentImageEnv, "registry.dev.rafay-edge.net:5000/rafay/rafay-relay-agent:v1.6.x-122")
	viper.SetDefault(authAddrEnv, "authsrv.rcloud-admin.svc.cluster.local:50011")
	viper.SetDefault(sentryAddrENV, "localhost:10000")
	viper.SetDefault(sentryBootstrapAddrENV, "api.sentry.rafay.local:11000")
	viper.SetDefault(podNameEnv, "local-pod-0")
	viper.SetDefault(schedulerNamespaceEnv, "rafay-system")

	viper.BindEnv(rpcPortEnv)
	viper.BindEnv(apiPortEnv)
	viper.BindEnv(peerPortENV)
	viper.BindEnv(connectorPortENV)
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
	viper.BindEnv(connectorImageEnv)
	viper.BindEnv(relayAgentImageEnv)
	viper.BindEnv(controllerImageEnv)
	viper.BindEnv(authAddrEnv)
	viper.BindEnv(sentryAddrENV)
	viper.BindEnv(sentryBootstrapAddrENV)
	viper.BindEnv(podNameEnv)
	viper.BindEnv(schedulerNamespaceEnv)

	rpcPort = viper.GetInt(rpcPortEnv)
	apiPort = viper.GetInt(apiPortEnv)
	peerPort = viper.GetInt(peerPortENV)
	connectorPort = viper.GetInt(connectorPortENV)
	dbAddr = viper.GetString(dbAddrEnv)
	dbName = viper.GetString(dbNameEnv)
	dbUser = viper.GetString(dbUserEnv)
	dbPassword = viper.GetString(dbPasswordEnv)
	dev = viper.GetBool(devEnv)
	secretPath = viper.GetString(secretPathEnv)
	configAddr = viper.GetString(configAddrENV)
	allowReuseToken = viper.GetString(allowReuseToken)
	controlAddr = viper.GetString(controlAddrEnv)
	apiAddr = viper.GetString(apiAddrEnv)
	controllerImage = viper.GetString(controllerImageEnv)
	connectorImage = viper.GetString(connectorImageEnv)
	relayAgentImage = viper.GetString(relayAgentImageEnv)
	authAddr = viper.GetString(authAddrEnv)
	sentryAddr = viper.GetString(sentryAddrENV)
	podName = viper.GetString(podNameEnv)
	schedulerNamespace = viper.GetString(schedulerNamespaceEnv)

	dsn := "postgres://" + dbUser + ":" + dbPassword + "@" + dbAddr + "/" + dbName + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db = bun.NewDB(sqldb, pgdialect.New())

	if dev {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	_log.Infow("setup finished",
		"rpcPort", rpcPort,
		"apiPort", apiPort,
		"peerPort", peerPort,
		"connectorPort", connectorPort,
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
		ControllerImage: controllerImage,
		ConnectorImage:  connectorImage,
		RelayAgentImage: relayAgentImage,
	}

	var err error

	cs = service.NewClusterService(db, downloadData)

	notify.Init(cs)

	signer, err = credentials.NewSigner(secretPath)

	if err != nil {
		_log.Fatalw("unable to create signer", "error", err)
	}

	_log.Infow("queried number of cpus", "numCPUs", goruntime.NumCPU())

	//TODO: add sentry, auth pool as required
	configPool = configrpc.NewConfigPool(configAddr, 5*goruntime.NumCPU())

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
	go runPeer(&wg, ctx)

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
		rpcv3.RegisterClusterHandlerFromEndpoint,
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

func runPeer(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	<-ctx.Done()
}

func runRPC(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	defer cs.Close()
	defer configPool.Close()

	crpc := server.NewClusterServer(cs, signer, downloadData)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		_log.Fatalw("unable to start rpc listener", "error", err)
	}

	var opts []grpc.ServerOption
	if !dev {
		_log.Infow("adding auth interceptor")
		opts = append(opts, grpc.UnaryInterceptor(
			interceptors.NewAuthInterceptorWithOptions(
				interceptors.WithAuthPool(authPool),
				interceptors.WithExclude("POST", "/infra/v3/scheduler/cluster/register"),
			),
		))
		defer authPool.Close()
	} else {
		opts = append(opts, grpc.UnaryInterceptor(
			interceptors.NewAuthInterceptorWithOptions(interceptors.WithDummy()),
		))
	}

	s, err := grpcutil.NewServer(opts...)
	if err != nil {
		_log.Fatalw("unable to create grpc server", "error", err)
	}

	go func() {
		defer s.GracefulStop()

		<-ctx.Done()
		_log.Infow("context done")
	}()

	// register all the rpc servers
	rpcv3.RegisterClusterServer(s, crpc)

	_log.Infow("starting rpc server", "port", rpcPort)
	err = s.Serve(l)
	if err != nil {
		_log.Fatalw("unable to start rpc server", "error", err)
	}
}

func main() {
	setup()
	run()
}
