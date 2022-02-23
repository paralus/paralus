package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	goruntime "runtime"
	"sync"
	"time"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/internal/fixtures"
	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/sentry/util"
	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/service"
	adminrpc "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/rpc"
	"github.com/RafaySystems/rcloud-base/components/adminsrv/server"
	authv3 "github.com/RafaySystems/rcloud-base/components/common/pkg/auth/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/gateway"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/grpc"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	configrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/config"
	schedulerrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/scheduler"
	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	_grpc "google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	rpcPortEnv                = "RPC_PORT"
	apiPortEnv                = "API_PORT"
	debugPortEnv              = "DEBUG_PORT"
	dbAddrEnv                 = "DB_ADDR"
	dbNameEnv                 = "DB_NAME"
	dbUserEnv                 = "DB_USER"
	dbPasswordEnv             = "DB_PASSWORD"
	devEnv                    = "DEV"
	configAddrENV             = "CONFIG_ADDR"
	schedulerAddrENV          = "SCHEDULER_ADDR"
	sentryPeeringHostEnv      = "SENTRY_PEERING_HOST"
	coreRelayConnectorHostEnv = "CORE_RELAY_CONNECTOR_HOST"
	coreRelayUserHostEnv      = "CORE_RELAY_USER_HOST"
	bootstrapKEKEnv           = "BOOTSTRAP_KEK"
	authAddrEnv               = "AUTH_ADDR"
	sentryBootstrapEnv        = "SENTRY_BOOTSTRAP_ADDR"
	relayImageEnv             = "RELAY_IMAGE"

	// cd relay
	coreCDRelayUserHostEnv      = "CORE_CD_RELAY_USER_HOST"
	coreCDRelayConnectorHostEnv = "CORE_CD_RELAY_CONNECTOR_HOST"
)

var (
	rpcPort                int
	apiPort                int
	debugPort              int
	rpcRelayPeeringPort    int
	dbAddr                 string
	dbName                 string
	dbUser                 string
	dbPassword             string
	dev                    bool
	db                     *bun.DB
	ps                     service.PartnerService
	os                     service.OrganizationService
	pps                    service.ProjectService
	bs                     service.BootstrapService
	aps                    service.AccountPermissionService
	gps                    service.GroupPermissionService
	krs                    service.KubeconfigRevocationService
	kss                    service.KubeconfigSettingService
	kcs                    service.KubectlClusterSettingsService
	_log                   = log.GetLogger()
	authPool               authv3.AuthPool
	configPool             configrpc.ConfigPool
	schedulerPool          schedulerrpc.SchedulerPool
	configAddr             string
	authAddr               string
	schedulerAddr          string
	bootstrapKEK           string
	sentryPeeringHost      string
	coreRelayConnectorHost string
	coreRelayUserHost      string

	// cd relay
	coreCDRelayUserHost      string
	coreCDRelayConnectorHost string

	kekFunc = func() ([]byte, error) {
		if len(bootstrapKEK) == 0 {
			return nil, errors.New("empty KEK")
		}
		return []byte(bootstrapKEK), nil
	}
)

func setup() {
	viper.SetDefault(rpcPortEnv, 10000)
	viper.SetDefault(apiPortEnv, 11000)
	viper.SetDefault(debugPortEnv, 12000)
	viper.SetDefault(dbAddrEnv, "localhost:5432")
	viper.SetDefault(dbNameEnv, "admindb")
	viper.SetDefault(dbUserEnv, "admindbuser")
	viper.SetDefault(dbPasswordEnv, "admindbpassword")
	viper.SetDefault(devEnv, true)
	viper.SetDefault(configAddrENV, "localhost:7000")
	viper.SetDefault(schedulerAddrENV, "localhost:5000")
	viper.SetDefault(sentryPeeringHostEnv, "peering.sentry.rafay.local:10001")
	viper.SetDefault(coreRelayConnectorHostEnv, "*.core-connector.relay.rafay.local:10002")
	viper.SetDefault(coreRelayUserHostEnv, "*.user.relay.rafay.local:10002")
	viper.SetDefault(bootstrapKEKEnv, "rafay")
	viper.SetDefault(authAddrEnv, "authsrv.rcloud-admin.svc.cluster.local:50011")
	viper.SetDefault(coreCDRelayUserHostEnv, "*.user.cdrelay.rafay.local:10012")
	viper.SetDefault(coreCDRelayConnectorHostEnv, "*.core-connector.cdrelay.rafay.local:10012")
	viper.SetDefault(sentryBootstrapEnv, "console.rafay.dev:443")
	viper.SetDefault(relayImageEnv, "registry.rafay-edge.net/rafay/rafay-relay-agent:r1.10.0-24")

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
	viper.BindEnv(schedulerAddrENV)
	viper.BindEnv(bootstrapKEKEnv)
	viper.BindEnv(authAddrEnv)
	viper.BindEnv(sentryPeeringHostEnv)
	viper.BindEnv(coreRelayConnectorHostEnv)
	viper.BindEnv(coreRelayUserHostEnv)
	viper.BindEnv(coreCDRelayConnectorHostEnv)
	viper.BindEnv(coreCDRelayUserHostEnv)
	viper.BindEnv(sentryBootstrapEnv)
	viper.BindEnv(relayImageEnv)

	rpcPort = viper.GetInt(rpcPortEnv)
	apiPort = viper.GetInt(apiPortEnv)
	debugPort = viper.GetInt(debugPortEnv)
	dbAddr = viper.GetString(dbAddrEnv)
	dbName = viper.GetString(dbNameEnv)
	dbUser = viper.GetString(dbUserEnv)
	dbPassword = viper.GetString(dbPasswordEnv)
	dev = viper.GetBool(devEnv)
	configAddr = viper.GetString(configAddrENV)
	schedulerAddr = viper.GetString(schedulerAddrENV)
	authAddr = viper.GetString(authAddrEnv)
	bootstrapKEK = viper.GetString(bootstrapKEKEnv)
	sentryPeeringHost = viper.GetString(sentryPeeringHostEnv)
	coreRelayConnectorHost = viper.GetString(coreRelayConnectorHostEnv)
	coreRelayUserHost = viper.GetString(coreRelayUserHostEnv)
	coreCDRelayConnectorHost = viper.GetString(coreCDRelayConnectorHostEnv)
	coreCDRelayUserHost = viper.GetString(coreCDRelayUserHostEnv)

	rpcRelayPeeringPort = rpcPort + 1

	// DB setup
	dsn := "postgres://" + dbUser + ":" + dbPassword + "@" + dbAddr + "/" + dbName + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db = bun.NewDB(sqldb, pgdialect.New())

	if dev {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	_log.Infow("printing db", "db", db)

	configPool = configrpc.NewConfigPool(configAddr, 5*goruntime.NumCPU())
	schedulerPool = schedulerrpc.NewSchedulerPool(schedulerAddr, 5*goruntime.NumCPU())

	ps = service.NewPartnerService(db)
	os = service.NewOrganizationService(db)
	pps = service.NewProjectService(db)

	//sentry related services
	bs = service.NewBootstrapService(db, schedulerPool)
	krs = service.NewKubeconfigRevocationService(db)
	kss = service.NewKubeconfigSettingService(db)
	kcs = service.NewkubectlClusterSettingsService(db)
	aps = service.NewAccountPermissionService(db)
	gps = service.NewGroupPermissionService(db)

	/* TODO: to be revisited if required
	apn = modelsv2.NewaccountProjectNamespaceService(db)
	*/
}

func run() {

	ctx := signals.SetupSignalHandler()

	replace := map[string]interface{}{
		"sentryPeeringHost":   sentryPeeringHost,
		"coreRelayServerHost": coreRelayConnectorHost,
		"coreRelayUserHost":   coreRelayUserHost,

		// cd relay
		"coreCDRelayUserHost":      coreCDRelayUserHost,
		"coreCDRelayConnectorHost": coreCDRelayConnectorHost,
	}

	_log.Infow("loading fixtures", "data", replace)

	fixtures.Load(ctx, bs, replace, kekFunc)

	var wg sync.WaitGroup
	wg.Add(4)

	go runAPI(&wg, ctx)
	go runRPC(&wg, ctx)
	go runRelayPeerRPC(&wg, ctx)
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
		adminrpc.RegisterPartnerHandlerFromEndpoint,
		adminrpc.RegisterOrganizationHandlerFromEndpoint,
		adminrpc.RegisterProjectHandlerFromEndpoint,
		sentryrpc.RegisterBootstrapHandlerFromEndpoint,
		sentryrpc.RegisterKubeConfigHandlerFromEndpoint,
		sentryrpc.RegisterKubectlClusterSettingsHandlerFromEndpoint,
		sentryrpc.RegisterClusterAuthorizationHandlerFromEndpoint,
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

func runRelayPeerRPC(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	_log.Infow("waiting to fetch peering server creds")
	time.Sleep(time.Second * 25)
	cert, key, ca, err := util.GetPeeringServerCreds(context.Background(), bs, rpcPort, sentryPeeringHost)
	if err != nil {
		_log.Fatalw("unable to get peering server cerds", "error", err)
	}

	relayPeerService, err := server.NewRelayPeerService()
	if err != nil {
		_log.Fatalw("unable to get create relay peer service")
	}
	clusterAuthzServer := server.NewClusterAuthzServer(bs, aps, gps, krs, kcs, kss, configPool)

	/*
		auditInfoServer := server.NewAuditInfoServer(bs, aps)
	*/

	s, err := grpc.NewSecureServerWithPEM(cert, key, ca)
	if err != nil {
		_log.Fatalw("cannot grpc secure server failed", "error", err)

	}

	go func() {
		defer s.GracefulStop()

		<-ctx.Done()
		_log.Infow("peer service stoped due to context done")
	}()

	sentryrpc.RegisterRelayPeerServiceServer(s, relayPeerService)
	sentryrpc.RegisterClusterAuthorizationServer(s, clusterAuthzServer)
	/*sentryrpc.RegisterAuditInformationServer(s, auditInfoServer)*/

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcRelayPeeringPort))
	if err != nil {
		_log.Fatalw("failed to listen relay peer service port", "port", rpcRelayPeeringPort, "error", err)
		return
	}

	go server.RunRelaySurveyHandler(ctx.Done(), relayPeerService)

	_log.Infow("started relay rpc service ", "port", rpcRelayPeeringPort)
	if err = s.Serve(l); err != nil {
		_log.Fatalw("failed to serve relay peer service", "error", err)
	}

}

func runRPC(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	defer ps.Close()
	defer configPool.Close()
	defer schedulerPool.Close()

	partnerServer := server.NewPartnerServer(ps)
	organizationServer := server.NewOrganizationServer(os)
	projectServer := server.NewProjectServer(pps)

	bootstrapServer := server.NewBootstrapServer(bs, kekFunc, configPool)
	kubeConfigServer := server.NewKubeConfigServer(bs, aps, gps, kss, krs, kekFunc)
	/*auditInfoServer := rpcv2.NewAuditInfoServer(bs, aps)*/
	clusterAuthzServer := server.NewClusterAuthzServer(bs, aps, gps, krs, kcs, kss, configPool)
	kubectlClusterSettingsServer := server.NewKubectlClusterSettingsServer(bs, kcs)

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

	adminrpc.RegisterPartnerServer(s, partnerServer)
	adminrpc.RegisterOrganizationServer(s, organizationServer)
	adminrpc.RegisterProjectServer(s, projectServer)
	sentryrpc.RegisterBootstrapServer(s, bootstrapServer)
	sentryrpc.RegisterKubeConfigServer(s, kubeConfigServer)
	sentryrpc.RegisterClusterAuthorizationServer(s, clusterAuthzServer)
	/*pbrpcv2.RegisterAuditInformationServer(s, auditInfoServer)*/
	sentryrpc.RegisterKubectlClusterSettingsServer(s, kubectlClusterSettingsServer)

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
