package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/RafayLabs/rcloud-base/internal/fixtures"
	providers "github.com/RafayLabs/rcloud-base/internal/provider/kratos"
	"github.com/RafayLabs/rcloud-base/pkg/audit"
	authv3 "github.com/RafayLabs/rcloud-base/pkg/auth/v3"
	"github.com/RafayLabs/rcloud-base/pkg/common"
	"github.com/RafayLabs/rcloud-base/pkg/enforcer"
	"github.com/RafayLabs/rcloud-base/pkg/gateway"
	"github.com/RafayLabs/rcloud-base/pkg/grpc"
	"github.com/RafayLabs/rcloud-base/pkg/log"
	"github.com/RafayLabs/rcloud-base/pkg/notify"
	"github.com/RafayLabs/rcloud-base/pkg/reconcile"
	"github.com/RafayLabs/rcloud-base/pkg/sentry/peering"
	"github.com/RafayLabs/rcloud-base/pkg/service"
	auditrpc "github.com/RafayLabs/rcloud-base/proto/rpc/audit"
	rolerpc "github.com/RafayLabs/rcloud-base/proto/rpc/role"
	schedulerrpc "github.com/RafayLabs/rcloud-base/proto/rpc/scheduler"
	sentryrpc "github.com/RafayLabs/rcloud-base/proto/rpc/sentry"
	systemrpc "github.com/RafayLabs/rcloud-base/proto/rpc/system"
	userrpc "github.com/RafayLabs/rcloud-base/proto/rpc/user"
	"github.com/RafayLabs/rcloud-base/server"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	kclient "github.com/ory/kratos-client-go"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	_grpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	// application
	rpcPortEnv   = "RPC_PORT"
	apiPortEnv   = "API_PORT"
	debugPortEnv = "DEBUG_PORT"
	apiAddrEnv   = "API_ADDR"
	devEnv       = "DEV"

	// db
	dbAddrEnv     = "DB_ADDR"
	dbNameEnv     = "DB_NAME"
	dbUserEnv     = "DB_USER"
	dbPasswordEnv = "DB_PASSWORD"

	// relay
	sentryPeeringHostEnv      = "SENTRY_PEERING_HOST"
	coreRelayConnectorHostEnv = "CORE_RELAY_CONNECTOR_HOST"
	coreRelayUserHostEnv      = "CORE_RELAY_USER_HOST"
	sentryBootstrapEnv        = "SENTRY_BOOTSTRAP_ADDR"
	bootstrapKEKEnv           = "BOOTSTRAP_KEK"
	relayImageEnv             = "RELAY_IMAGE"

	// audit
	auditFileEnv               = "AUDIT_LOG_FILE"
	esEndPointEnv              = "ES_END_POINT"
	esIndexPrefixEnv           = "ES_INDEX_PREFIX"
	relayAuditESIndexPrefixEnv = "RELAY_AUDITS_ES_INDEX_PREFIX"
	relayCommandESIndexPrefix  = "RELAY_COMMANDS_ES_INDEX_PREFIX"

	// cd relay
	coreCDRelayUserHostEnv      = "CORE_CD_RELAY_USER_HOST"
	coreCDRelayConnectorHostEnv = "CORE_CD_RELAY_CONNECTOR_HOST"
	schedulerNamespaceEnv       = "SCHEDULER_NAMESPACE"

	// kratos
	kratosAddrEnv       = "KRATOS_ADDR"
	kratosPublicAddrEnv = "KRATOS_PUB_ADDR"
)

var (
	// application
	rpcPort             int
	apiPort             int
	debugPort           int
	apiAddr             string
	dev                 bool
	rpcRelayPeeringPort int
	_log                = log.GetLogger()

	// db
	dbAddr     string
	dbName     string
	dbUser     string
	dbPassword string
	db         *bun.DB
	gormDb     *gorm.DB

	// relay
	sentryPeeringHost      string
	coreRelayConnectorHost string
	coreRelayUserHost      string
	bootstrapKEK           string
	relayImage             string

	// audit
	auditFile                  string
	elasticSearchUrl           string
	esIndexPrefix              string
	relayAuditsESIndexPrefix   string
	relayCommandsESIndexPrefix string

	// cd relay
	coreCDRelayUserHost      string
	coreCDRelayConnectorHost string
	schedulerNamespace       string
	sentryBootstrapAddr      string

	// kratos
	kratosAddr       string
	kratosPublicAddr string
	kc               *kclient.APIClient
	akc              *kclient.APIClient

	// services
	ps    service.PartnerService
	os    service.OrganizationService
	pps   service.ProjectService
	bs    service.BootstrapService
	aps   service.AccountPermissionService
	gps   service.GroupPermissionService
	krs   service.KubeconfigRevocationService
	kss   service.KubeconfigSettingService
	kcs   service.KubectlClusterSettingsService
	as    service.AuthzService
	cs    service.ClusterService
	ms    service.MetroService
	us    service.UserService
	ks    service.ApiKeyService
	gs    service.GroupService
	rs    service.RoleService
	rrs   service.RolepermissionService
	is    service.IdpService
	oidcs service.OIDCProviderService
	aus   *service.AuditLogService
	ras   *service.RelayAuditService
	rcs   *service.AuditLogService

	schedulerPool schedulerrpc.SchedulerPool
	schedulerAddr string
	downloadData  *common.DownloadData

	kekFunc = func() ([]byte, error) {
		if len(bootstrapKEK) == 0 {
			return nil, errors.New("empty KEK")
		}
		return []byte(bootstrapKEK), nil
	}
)

func setup() {
	// application
	viper.SetDefault(rpcPortEnv, 10000)
	viper.SetDefault(apiPortEnv, 11000)
	viper.SetDefault(debugPortEnv, 12000)
	viper.SetDefault(apiAddrEnv, "localhost:11000")
	viper.SetDefault(devEnv, true)

	// db
	viper.SetDefault(dbAddrEnv, "localhost:5432")
	viper.SetDefault(dbNameEnv, "admindb")
	viper.SetDefault(dbUserEnv, "admindbuser")
	viper.SetDefault(dbPasswordEnv, "admindbpassword")

	// relay
	viper.SetDefault(sentryPeeringHostEnv, "peering.sentry.rafay.local:10001")
	viper.SetDefault(coreRelayConnectorHostEnv, "*.core-connector.relay.rafay.local:10002")
	viper.SetDefault(coreRelayUserHostEnv, "*.user.relay.rafay.local:10002")
	viper.SetDefault(sentryBootstrapEnv, "console.rafay.dev:443")
	viper.SetDefault(bootstrapKEKEnv, "rafay")
	viper.SetDefault(relayImageEnv, "registry.rafay-edge.net/rafay/rafay-relay-agent:r1.10.0-24")

	// audit
	viper.SetDefault(esEndPointEnv, "http://127.0.0.1:9200")
	viper.SetDefault(esIndexPrefixEnv, "ralogs-system")
	viper.SetDefault(relayAuditESIndexPrefixEnv, "ralogs-relay")
	viper.SetDefault(relayCommandESIndexPrefix, "ralogs-prompt")
	viper.SetDefault(auditFileEnv, "audit.log")

	// cd relay
	viper.SetDefault(coreCDRelayUserHostEnv, "*.user.cdrelay.rafay.local:10012")
	viper.SetDefault(coreCDRelayConnectorHostEnv, "*.core-connector.cdrelay.rafay.local:10012")
	viper.SetDefault(schedulerNamespaceEnv, "default")

	// kratos
	viper.SetDefault(kratosAddrEnv, "http://localhost:4433")
	viper.SetDefault(kratosPublicAddrEnv, "http://localhost:4434")

	viper.BindEnv(rpcPortEnv)
	viper.BindEnv(apiPortEnv)
	viper.BindEnv(debugPortEnv)
	viper.BindEnv(apiAddrEnv)
	viper.BindEnv(devEnv)

	viper.BindEnv(dbAddrEnv)
	viper.BindEnv(dbNameEnv)
	viper.BindEnv(dbUserEnv)
	viper.BindEnv(dbPasswordEnv)

	viper.BindEnv(kratosAddrEnv)
	viper.BindEnv(kratosPublicAddrEnv)

	viper.BindEnv(sentryPeeringHostEnv)
	viper.BindEnv(coreRelayConnectorHostEnv)
	viper.BindEnv(coreRelayUserHostEnv)
	viper.BindEnv(sentryBootstrapEnv)
	viper.BindEnv(bootstrapKEKEnv)
	viper.BindEnv(coreCDRelayConnectorHostEnv)
	viper.BindEnv(coreCDRelayUserHostEnv)
	viper.BindEnv(relayImageEnv)
	viper.BindEnv(schedulerNamespaceEnv)

	viper.BindEnv(auditFileEnv)
	viper.BindEnv(esEndPointEnv)
	viper.BindEnv(esIndexPrefixEnv)
	viper.BindEnv(relayAuditESIndexPrefixEnv)
	viper.BindEnv(relayCommandESIndexPrefix)

	rpcPort = viper.GetInt(rpcPortEnv)
	apiPort = viper.GetInt(apiPortEnv)
	debugPort = viper.GetInt(debugPortEnv)
	apiAddr = viper.GetString(apiAddrEnv)
	dev = viper.GetBool(devEnv)

	dbAddr = viper.GetString(dbAddrEnv)
	dbName = viper.GetString(dbNameEnv)
	dbUser = viper.GetString(dbUserEnv)
	dbPassword = viper.GetString(dbPasswordEnv)

	kratosAddr = viper.GetString(kratosAddrEnv)
	kratosPublicAddr = viper.GetString(kratosPublicAddrEnv)

	bootstrapKEK = viper.GetString(bootstrapKEKEnv)
	sentryPeeringHost = viper.GetString(sentryPeeringHostEnv)
	coreRelayConnectorHost = viper.GetString(coreRelayConnectorHostEnv)
	coreRelayUserHost = viper.GetString(coreRelayUserHostEnv)
	coreCDRelayConnectorHost = viper.GetString(coreCDRelayConnectorHostEnv)
	coreCDRelayUserHost = viper.GetString(coreCDRelayUserHostEnv)
	relayImage = viper.GetString(relayImageEnv)
	schedulerNamespace = viper.GetString(schedulerNamespaceEnv)
	sentryBootstrapAddr = viper.GetString(sentryBootstrapEnv)

	auditFile = viper.GetString(auditFileEnv)
	elasticSearchUrl = viper.GetString(esEndPointEnv)
	esIndexPrefix = viper.GetString(esIndexPrefixEnv)
	relayAuditsESIndexPrefix = viper.GetString(relayAuditESIndexPrefixEnv)
	relayCommandsESIndexPrefix = viper.GetString(relayCommandESIndexPrefix)

	rpcRelayPeeringPort = rpcPort + 1

	// Kratos client setup for authentication
	kratosConfig := kclient.NewConfiguration()
	kratosConfig.Servers[0].URL = kratosPublicAddr
	kc = kclient.NewAPIClient(kratosConfig)

	// Kratos client setup for admin purpose
	kratosAdminConfig := kclient.NewConfiguration()
	kratosAdminConfig.Servers[0].URL = kratosAddr
	akc = kclient.NewAPIClient(kratosAdminConfig)

	// db setup
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbAddr, dbName)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db = bun.NewDB(sqldb, pgdialect.New())

	if dev {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
		lc := make(chan string)
		go _log.ChangeLevel(lc)
		lc <- "debug"
		_log.Debugw("Debug mode set in log because this is a dev environment")
	}

	_log.Infow("printing db", "db", db)

	ao := audit.AuditOptions{
		LogPath:    auditFile,
		MaxSizeMB:  1,
		MaxBackups: 10, // Should we let sidecar do rotation?
		MaxAgeDays: 10, // Make these configurable via env
	}
	auditLogger := audit.GetAuditLogger(&ao)

	// authz services
	gormDb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqldb,
	}), &gorm.Config{})
	if err != nil {
		_log.Fatalw("unable to create db connection", "error", err)
	}
	enforcer, err := enforcer.NewCasbinEnforcer(gormDb).Init()
	if err != nil {
		_log.Fatalw("unable to init enforcer", "error", err)
	}
	as = service.NewAuthzService(db, enforcer)

	schedulerPool = schedulerrpc.NewSchedulerPool(schedulerAddr, 5*goruntime.NumCPU())

	ps = service.NewPartnerService(db, auditLogger)
	os = service.NewOrganizationService(db, auditLogger)
	pps = service.NewProjectService(db, as, auditLogger, dev)

	// users and role management services
	cc := common.CliConfigDownloadData{
		RestEndpoint: apiAddr,
		OpsEndpoint:  apiAddr,
	}
	if dev {
		cc.Profile = "staging"
	} else {
		cc.Profile = "production"
	}
	ks = service.NewApiKeyService(db, auditLogger)
	us = service.NewUserService(providers.NewKratosAuthProvider(akc), db, as, ks, cc, auditLogger, dev)
	gs = service.NewGroupService(db, as, auditLogger)
	rs = service.NewRoleService(db, as, auditLogger)
	rrs = service.NewRolepermissionService(db)
	is = service.NewIdpService(db, apiAddr, auditLogger)
	oidcs = service.NewOIDCProviderService(db, sentryBootstrapAddr, auditLogger)

	//sentry related services
	bs = service.NewBootstrapService(db)
	krs = service.NewKubeconfigRevocationService(db)
	kss = service.NewKubeconfigSettingService(db)
	kcs = service.NewkubectlClusterSettingsService(db)
	aps = service.NewAccountPermissionService(db)
	gps = service.NewGroupPermissionService(db)

	// audit services
	aus, err = service.NewAuditLogService(elasticSearchUrl, esIndexPrefix+"-*", "AuditLog API: ")
	if err != nil {
		if dev && strings.Contains(err.Error(), "connect: connection refused") {
			// This is primarily from ES not being available. ES being
			// pretty heavy, you might not always wanna have it
			// running in the background. This way, you can continue
			// working on rcloud-base with ES eating up all the cpu.
			_log.Warn("unable to create auditLog service: ", err)
		} else {
			_log.Fatalw("unable to create auditLog service", "error", err)
		}
	}
	ras, err = service.NewRelayAuditService(elasticSearchUrl, relayAuditsESIndexPrefix+"-*", "RelayAudit API: ")
	if err != nil {
		if dev && strings.Contains(err.Error(), "connect: connection refused") {
			_log.Warn("unable to create relayAudit service: ", err)
		} else {
			_log.Fatalw("unable to create relayAudit service", "error", err)
		}
	}
	rcs, err = service.NewAuditLogService(elasticSearchUrl, relayCommandsESIndexPrefix+"-*", "RelayCommand API: ")
	if err != nil {
		if dev && strings.Contains(err.Error(), "connect: connection refused") {
			_log.Warn("unable to create auditLog service:", err)
		} else {
			_log.Fatalw("unable to create auditLog service", "error", err)
		}
	}

	// cluster bootstrap
	downloadData = &common.DownloadData{
		APIAddr:         apiAddr,
		RelayAgentImage: relayImage,
	}

	cs = service.NewClusterService(db, downloadData, bs, auditLogger)
	ms = service.NewMetroService(db)

	notify.Init(cs)

	_log.Infow("queried number of cpus", "numCPUs", goruntime.NumCPU())
}

func run() {

	ctx := signals.SetupSignalHandler()

	notify.Start(ctx.Done())

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

	healthServer, err := grpc.NewServer()
	if err != nil {
		_log.Infow("failed to initialize grpc for health server")
	}
	// health server
	_log.Infow("registering grpc health server")
	hs := health.NewServer()
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(healthServer, hs)
	_log.Infow("registered grpc health server")

	var wg sync.WaitGroup
	wg.Add(5)

	go runAPI(&wg, ctx)
	go runRPC(&wg, ctx)
	go runRelayPeerRPC(&wg, ctx)
	go runDebug(&wg, ctx)
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
		systemrpc.RegisterPartnerHandlerFromEndpoint,
		systemrpc.RegisterOrganizationHandlerFromEndpoint,
		systemrpc.RegisterProjectHandlerFromEndpoint,
		sentryrpc.RegisterBootstrapHandlerFromEndpoint,
		sentryrpc.RegisterKubeConfigHandlerFromEndpoint,
		sentryrpc.RegisterKubectlClusterSettingsHandlerFromEndpoint,
		sentryrpc.RegisterClusterAuthorizationHandlerFromEndpoint,
		schedulerrpc.RegisterClusterHandlerFromEndpoint,
		systemrpc.RegisterLocationHandlerFromEndpoint,
		userrpc.RegisterUserHandlerFromEndpoint,
		userrpc.RegisterGroupHandlerFromEndpoint,
		rolerpc.RegisterRoleHandlerFromEndpoint,
		rolerpc.RegisterRolepermissionHandlerFromEndpoint,
		systemrpc.RegisterIdpHandlerFromEndpoint,
		systemrpc.RegisterOIDCProviderHandlerFromEndpoint,
		auditrpc.RegisterAuditLogHandlerFromEndpoint,
		auditrpc.RegisterRelayAuditHandlerFromEndpoint,
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
	cert, key, ca, err := peering.GetPeeringServerCreds(context.Background(), bs, rpcPort, sentryPeeringHost)
	if err != nil {
		_log.Fatalw("unable to get peering server cerds", "error", err)
	}

	relayPeerService, err := server.NewRelayPeerService()
	if err != nil {
		_log.Fatalw("unable to get create relay peer service")
	}
	clusterAuthzServer := server.NewClusterAuthzServer(bs, aps, gps, krs, kcs, kss)
	auditInfoServer := server.NewAuditInfoServer(bs, aps)

	s, err := grpc.NewSecureServerWithPEM(cert, key, ca)
	if err != nil {
		_log.Fatalw("cannot grpc secure server failed", "error", err)

	}

	go func() {
		defer s.GracefulStop()

		<-ctx.Done()
		_log.Infow("peer service stopped due to context done")
	}()

	sentryrpc.RegisterRelayPeerServiceServer(s, relayPeerService)
	sentryrpc.RegisterClusterAuthorizationServer(s, clusterAuthzServer)
	sentryrpc.RegisterAuditInformationServer(s, auditInfoServer)

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
	defer schedulerPool.Close()
	defer db.Close()

	partnerServer := server.NewPartnerServer(ps)
	organizationServer := server.NewOrganizationServer(os)
	projectServer := server.NewProjectServer(pps)

	bootstrapServer := server.NewBootstrapServer(bs, kekFunc, cs)
	kubeConfigServer := server.NewKubeConfigServer(bs, aps, gps, kss, krs, kekFunc, ks, os, ps)
	auditInfoServer := server.NewAuditInfoServer(bs, aps)
	clusterAuthzServer := server.NewClusterAuthzServer(bs, aps, gps, krs, kcs, kss)
	kubectlClusterSettingsServer := server.NewKubectlClusterSettingsServer(bs, kcs)
	crpc := server.NewClusterServer(cs, downloadData)
	mserver := server.NewLocationServer(ms)

	userServer := server.NewUserServer(us, ks)
	groupServer := server.NewGroupServer(gs)
	roleServer := server.NewRoleServer(rs)
	rolepermissionServer := server.NewRolePermissionServer(rrs)
	idpServer := server.NewIdpServer(is)
	oidcProviderServer := server.NewOIDCServer(oidcs)

	// audit
	auditLogServer, err := server.NewAuditLogServer(aus)
	if err != nil {
		_log.Fatalw("unable to create auditLog server", "error", err)
	}
	relayAuditServer, err := server.NewRelayAuditServer(ras, rcs)
	if err != nil {
		_log.Fatalw("unable to create relayAudit server", "error", err)
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		_log.Fatalw("unable to start rpc listener", "error", err)
	}

	var opts []_grpc.ServerOption
	if !dev {
		_log.Infow("adding auth interceptor")
		ac := authv3.NewAuthContext(kc, ks, as)
		o := authv3.Option{
			ExcludeRPCMethods: []string{
				"/rafay.dev.sentry.rpc.Bootstrap/GetBootstrapAgentTemplate",
				"/rafay.dev.sentry.rpc.Bootstrap/RegisterBootstrapAgent",
				"/rafay.dev.sentry.rpc.KubeConfig/GetForClusterWebSession", //TODO: enable auth from prompt
			},
			ExcludeAuthzMethods: []string{
				"/rafay.dev.rpc.v3.User/GetUserInfo",
			},
		}
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

	systemrpc.RegisterPartnerServer(s, partnerServer)
	systemrpc.RegisterOrganizationServer(s, organizationServer)
	systemrpc.RegisterProjectServer(s, projectServer)
	sentryrpc.RegisterBootstrapServer(s, bootstrapServer)
	sentryrpc.RegisterKubeConfigServer(s, kubeConfigServer)
	sentryrpc.RegisterClusterAuthorizationServer(s, clusterAuthzServer)
	sentryrpc.RegisterAuditInformationServer(s, auditInfoServer)
	sentryrpc.RegisterKubectlClusterSettingsServer(s, kubectlClusterSettingsServer)
	schedulerrpc.RegisterClusterServer(s, crpc)
	systemrpc.RegisterLocationServer(s, mserver)
	userrpc.RegisterUserServer(s, userServer)
	userrpc.RegisterGroupServer(s, groupServer)
	rolerpc.RegisterRoleServer(s, roleServer)
	rolerpc.RegisterRolepermissionServer(s, rolepermissionServer)
	systemrpc.RegisterIdpServer(s, idpServer)
	systemrpc.RegisterOIDCProviderServer(s, oidcProviderServer)
	auditrpc.RegisterAuditLogServer(s, auditLogServer)
	auditrpc.RegisterRelayAuditServer(s, relayAuditServer)

	_log.Infow("starting rpc server", "port", rpcPort)
	err = s.Serve(l)
	if err != nil {
		_log.Fatalw("unable to start rpc server", "error", err)
	}

}

func runEventHandlers(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	//TODO: need to add a bunch of other handlers with gitops
	ceh := reconcile.NewClusterEventHandler(cs)
	_log.Infow("starting cluster event handler")
	go ceh.Handle(ctx.Done())

	// listen to cluster events
	cs.AddEventHandler(ceh.ClusterHook())

	<-ctx.Done()
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
