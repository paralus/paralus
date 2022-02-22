package relay

import (
	"context"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/sentry/cryptoutil"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/sentry/register"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/tunnel"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
	"github.com/spf13/viper"
)

const (
	podNameEnv        = "POD_NAME"
	podNamespaceEnv   = "POD_NAMESPACE"
	peerServiceURIEnv = "RAFAY_RELAY_PEERSERVICE"
	relayListenIPEnv  = "RELAY_LISTENIP"
	sentryAddrEnv     = "SENTRY_ADDR" // used by relay running inside core
	auditPathEnv      = "AUDIT_PATH"
	metricListenEnv   = "METRIC_LISTEN"

	// for relays running outside of core
	// supply bootstrapAddr and token as environment variables
	bootstrapAddrEnv          = "BOOTSTRAP_ADDR"
	relayPeeringTokenEnv      = "RELAY_PEERING_TOKEN"
	relayUserTokenEnv         = "RELAY_USER_TOKEN"
	relayConnectorTokenEnv    = "RELAY_CONNECTOR_TOKEN"
	relayUserHostPortEnv      = "RELAY_USER_HOST_PORT"
	relayConnectorHostPortEnv = "RELAY_CONNECTOR_HOST_PORT"
	relayNetworkIDEnv         = "RELAY_NETWORK_ID"

	// cd relay
	cdRelayUserTokenEnv      = "CD_RELAY_USER_TOKEN"
	cdRelayConnectorTokenEnv = "CD_RELAY_CONNECTOR_TOKEN"
)

var (
	podName         string
	podNamespace    string
	relayPeeringURI string
	relayListenIP   string
	log             *relaylogger.RelayLog
	sentryAddr      string // used by relay running inside core
	auditPath       string
	metricListen    string

	// for relays running outside of core
	relayPeeringToken      string
	relayConnectorToken    string
	relayUserToken         string
	bootstrapAddr          string
	relayUserHostPort      string
	relayConnectorHostPort string
	relayNetworkID         string

	// cd relay
	cdRelayConnectorToken string
	cdRelayUserToken      string
)

func setupserver(log *relaylogger.RelayLog) error {
	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetDefault(podNameEnv, "relay-pod")
	viper.SetDefault(podNamespace, "rafay-system")
	viper.SetDefault(peerServiceURIEnv, "https://peering.sentry.rafay.local:10001")
	viper.SetDefault(sentryAddrEnv, "localhost:10000")

	// for relays running outside of core
	viper.SetDefault(bootstrapAddrEnv, "")
	viper.SetDefault(relayPeeringTokenEnv, "-")
	viper.SetDefault(relayUserToken, "-")
	viper.SetDefault(relayConnectorToken, "-")
	viper.SetDefault(relayUserHostPortEnv, "")
	viper.SetDefault(relayConnectorHostPortEnv, "")
	viper.SetDefault(auditPathEnv, "/tmp")
	viper.SetDefault(metricListenEnv, ":8003")
	viper.SetDefault(relayNetworkIDEnv, "rafay-core-64a66h0asw")

	viper.BindEnv(podNameEnv)
	viper.BindEnv(podNamespaceEnv)
	viper.BindEnv(relayPeeringTokenEnv)
	viper.BindEnv(sentryAddrEnv)

	// outside core configs
	viper.BindEnv(bootstrapAddrEnv)
	viper.BindEnv(relayUserHostPortEnv)
	viper.BindEnv(relayConnectorHostPortEnv)
	viper.BindEnv(relayUserTokenEnv)
	viper.BindEnv(relayConnectorTokenEnv)
	viper.BindEnv(relayNetworkIDEnv)

	// for testing
	viper.BindEnv(relayListenIPEnv)

	podName = viper.GetString(podNameEnv)
	podNamespace = viper.GetString(podNamespaceEnv)
	sentryAddr = viper.GetString(sentryAddrEnv)
	auditPath = viper.GetString(auditPathEnv)

	utils.PODNAME = podName

	relayPeeringURI = viper.GetString(peerServiceURIEnv)

	// outside core relay configurations
	bootstrapAddr = viper.GetString(bootstrapAddrEnv)
	relayUserHostPort = viper.GetString(relayUserHostPortEnv)
	relayConnectorHostPort = viper.GetString(relayConnectorHostPortEnv)
	relayNetworkID = viper.GetString(relayNetworkIDEnv)

	if sentryAddr == "" && bootstrapAddr == "" {
		log.Error(
			nil,
			"missing sentry & bootstrap uri, please set one of the environment variable RAFAY_SENTRY[relay deployed inside core] (or) BOOTSTRAP_ADDR[outside core]",
		)
		return fmt.Errorf("relay server failed in setupserver")
	}

	if bootstrapAddr != "" {
		// outside config found, unset inside boostrap endpoint.
		// both cannot be set, only one mode is possible
		sentryAddr = ""
	}

	if relayPeeringURI == "" {
		log.Error(
			nil,
			"missing relay peer service uri, please set RAFAY_RELAY_PEERSERVICE environment variable",
		)
		return fmt.Errorf("relay server failed in setupserver")
	}

	relayPeeringToken = viper.GetString(relayPeeringTokenEnv)
	relayUserToken = viper.GetString(relayUserTokenEnv)
	relayConnectorToken = viper.GetString(relayConnectorTokenEnv)
	relayListenIP = viper.GetString(relayListenIPEnv) // used for running multiple relays in local test hosts
	metricListen = viper.GetString(metricListenEnv)

	log.Info(
		"relay server setup values",
		"podName", podName,
		"podNamespace", podNamespace,
		"relayPeeringURI", relayPeeringURI,
		"relayListenIP", relayListenIP,
	)

	if sentryAddr != "" {
		log.Info("relay inside core: sentryAddr", sentryAddr)
	} else {
		log.Info(
			"outside core relay: bootStrapURI", bootstrapAddr,
			"relayPeeringToken", relayPeeringToken,
			"relayUserToken", relayUserToken,
			"relayConnectorToken", relayConnectorToken,
		)
		if relayPeeringToken == "-" || relayUserToken == "-" || relayConnectorToken == "-" || relayUserHostPort == "" || relayConnectorHostPort == "" {
			log.Error(
				fmt.Errorf("missing env variable"),
				"outside core relay: one of the required token is missing, please set missing environment variables",
				"RELAY_PEERING_TOKEN", relayPeeringToken,
				"RELAY_USER_TOKEN", relayUserToken,
				"RELAY_CONNECTOR_TOKEN", relayConnectorToken,
				"RELAY_USER_HOST_PORT", relayUserHostPort,
				"RELAY_CONNECTOR_HOST_PORT", relayConnectorHostPort,
			)
			return fmt.Errorf("relay server failed in setupserver")
		}

		// process the user facing host:port
		host, port, err := net.SplitHostPort(relayUserHostPort)
		if err != nil {
			log.Error(
				err,
				"failed to process RELAY_USER_HOST_PORT env variable", relayUserHostPort,
			)
			return fmt.Errorf("relay server failed in setupserver")
		}
		utils.RelayUserPort, err = strconv.Atoi(port)
		if err != nil {
			log.Error(
				err,
				"invalid port in RELAY_USER_HOST_PORT env variable", relayUserHostPort,
			)
			return fmt.Errorf("relay server failed in setupserver")
		}
		utils.RelayUserHost = host

		// process the connector facing host:port
		host, port, err = net.SplitHostPort(relayConnectorHostPort)
		if err != nil {
			log.Error(
				err,
				"failed to process RELAY_CONNECTOR_HOST_PORT env variable", relayConnectorHostPort,
			)
			return fmt.Errorf("relay server failed in setupserver")
		}
		utils.RelayConnectorPort, err = strconv.Atoi(port)
		if err != nil {
			log.Error(
				err,
				"invalid port in RELAY_CONNECTOR_HOST_PORT env variable", relayConnectorHostPort,
			)
			return fmt.Errorf("relay server failed in setupserver")
		}
		utils.RelayConnectorHost = host
	}

	utils.RelayIPFromConfig = relayListenIP
	return nil
}

func setupCDServer(log *relaylogger.RelayLog) error {
	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetDefault(podNameEnv, "cdrelay-pod")
	viper.SetDefault(podNamespace, "rafay-system")
	viper.SetDefault(peerServiceURIEnv, "https://peering.sentry.rafay.local:10001")
	viper.SetDefault(sentryAddrEnv, "localhost:10000")

	// for relays running outside of core
	viper.SetDefault(bootstrapAddrEnv, "")
	viper.SetDefault(relayPeeringTokenEnv, "-")
	viper.SetDefault(cdRelayUserTokenEnv, "-")
	viper.SetDefault(cdRelayConnectorTokenEnv, "-")
	viper.SetDefault(auditPathEnv, "/tmp")
	viper.SetDefault(metricListenEnv, ":8003")

	viper.BindEnv(podNameEnv)
	viper.BindEnv(podNamespaceEnv)
	viper.BindEnv(relayPeeringTokenEnv)
	viper.BindEnv(sentryAddrEnv)

	// outside core configs
	viper.BindEnv(bootstrapAddrEnv)
	viper.BindEnv(cdRelayUserTokenEnv)
	viper.BindEnv(cdRelayConnectorTokenEnv)

	// for testing
	viper.BindEnv(relayListenIPEnv)

	podName = viper.GetString(podNameEnv)
	podNamespace = viper.GetString(podNamespaceEnv)
	sentryAddr = viper.GetString(sentryAddrEnv)
	auditPath = viper.GetString(auditPathEnv)

	utils.PODNAME = podName

	relayPeeringURI = viper.GetString(peerServiceURIEnv)

	if sentryAddr == "" {
		log.Error(
			nil,
			"missing sentry uri, please set one of the environment variable RAFAY_SENTRY[relay deployed inside core] (or) BOOTSTRAP_ADDR[outside core]",
		)
		return fmt.Errorf("relay server failed in setupserver")
	}

	if relayPeeringURI == "" {
		log.Error(
			nil,
			"missing relay peer service uri, please set RAFAY_RELAY_PEERSERVICE environment variable",
		)
		return fmt.Errorf("relay server failed in setupserver")
	}

	relayPeeringToken = viper.GetString(relayPeeringTokenEnv)
	cdRelayUserToken = viper.GetString(cdRelayUserTokenEnv)
	cdRelayConnectorToken = viper.GetString(cdRelayConnectorTokenEnv)
	relayListenIP = viper.GetString(relayListenIPEnv) // used for running multiple relays in local test hosts
	metricListen = viper.GetString(metricListenEnv)

	log.Debug(
		"relay server setup values",
		"podName", podName,
		"podNamespace", podNamespace,
		"relayPeeringURI", relayPeeringURI,
		"relayListenIP", relayListenIP,
	)

	utils.RelayIPFromConfig = relayListenIP
	return nil
}

// prepare config for outside relay boot strapping
func prepareConfigCSRForBootStrapOutSideCore(config *register.Config, CN string, log *relaylogger.RelayLog) error {
	privKey, err := cryptoutil.GenerateECDSAPrivateKey()
	if err != nil {
		log.Error(
			err,
			"failed generate ecd private key with CN", CN,
		)
		return err
	}

	key, err := cryptoutil.EncodePrivateKey(privKey, cryptoutil.NoPassword)
	if err != nil {
		log.Error(
			err,
			"failed encode ecd private key with CN", CN,
		)
		return err
	}

	config.PrivateKey = key

	csr, err := cryptoutil.CreateCSR(pkix.Name{
		CommonName:         CN,
		Country:            []string{"USA"},
		Organization:       []string{"Rafay Systems Inc"},
		Province:           []string{"California"},
		Locality:           []string{"Sunnyvale"},
		OrganizationalUnit: []string{relayNetworkID},
	}, privKey)
	if err != nil {
		log.Error(
			err,
			"failed create CSR with CN", CN,
		)
		return err
	}

	config.CSR = csr
	return nil
}

// registerRelayPeerService will register with  rafay-sentry-peering-client template token
// registration fetches client-certificate/root-ca to connect to sentry peer service
func registerRelayPeerService(ctx context.Context, log *relaylogger.RelayLog) error {
	cfg := &register.Config{
		ClientID: "peering-" + podName,
		ClientIP: utils.GetRelayIP(),
		Name:     "peering-" + podName,
		Mode:     "client",
	}

	if bootstrapAddr != "" {
		// outside core bootstrap
		cfg.Addr = bootstrapAddr
		if relayPeeringToken == "" {
			return fmt.Errorf("empty peering token for bootstrap")
		}
		cfg.TemplateToken = relayPeeringToken
		cfg.Scheme = "http"
		// check port is 443 then set scheme as https
		if utils.IsHTTPS(bootstrapAddr) {
			cfg.Scheme = "https"
		}
		// this is a client certificate CN is same as ClientID
		err := prepareConfigCSRForBootStrapOutSideCore(cfg, cfg.ClientID, log)
		if err != nil {
			return fmt.Errorf("failed in config csr for relay peering bootstrap")
		}
	} else {
		cfg.TemplateToken = "template/-"
		cfg.Addr = sentryAddr
		cfg.Scheme = "grpc"
		cfg.TemplateName = "rafay-sentry-peering-client"
	}

	if err := register.Register(ctx, cfg); err != nil {
		log.Error(
			err,
			"failed to register peering relay",
		)
		return err
	}

	log.Info(
		"peer certificate:",
		string(cfg.Certificate),
	)

	utils.PeerCertificate = cfg.Certificate
	utils.PeerPrivateKey = cfg.PrivateKey
	utils.PeerCACertificate = cfg.CACertificate
	utils.PeerServiceURI = relayPeeringURI

	return nil
}

// registerCDRelayPeerService will register with  rafay-sentry-cd-peering-client template token
// registration fetches client-certificate/root-ca to connect to sentry peer service
func registerCDRelayPeerService(ctx context.Context, log *relaylogger.RelayLog) error {
	cfg := &register.Config{
		ClientID: "peering-" + podName,
		ClientIP: utils.GetRelayIP(),
		Name:     "peering-" + podName,
		Mode:     "client",
	}

	if bootstrapAddr != "" {
		// outside core bootstrap
		cfg.Addr = bootstrapAddr
		if relayPeeringToken == "" {
			return fmt.Errorf("empty peering token for bootstrap")
		}
		cfg.TemplateToken = relayPeeringToken
		cfg.Scheme = "http"
		// check port is 443 then set scheme as https
		if utils.IsHTTPS(bootstrapAddr) {
			cfg.Scheme = "https"
		}
		// this is a client certificate CN is same as ClientID
		err := prepareConfigCSRForBootStrapOutSideCore(cfg, cfg.ClientID, log)
		if err != nil {
			return fmt.Errorf("failed in config csr for relay peering bootstrap")
		}
	} else {
		cfg.TemplateToken = "template/cd-relay"
		cfg.Addr = sentryAddr
		cfg.Scheme = "grpc"
		cfg.TemplateName = "rafay-sentry-cd-peering-client"
	}

	if err := register.Register(ctx, cfg); err != nil {
		log.Error(
			err,
			"failed to register cd peering relay",
		)
		return err
	}

	log.Info(
		"peer certificate:",
		string(cfg.Certificate),
	)

	utils.PeerCertificate = cfg.Certificate
	utils.PeerPrivateKey = cfg.PrivateKey
	utils.PeerCACertificate = cfg.CACertificate
	utils.PeerServiceURI = relayPeeringURI

	return nil
}

// registerRelayUserServer will register with rafay-core-relay-user template token
// registration fetches client-certificate/root-ca to terminate user connections
// the same certificate will be used as the client cert for peer-upstreams
func registerRelayUserServer(ctx context.Context, log *relaylogger.RelayLog) error {
	cfg := &register.Config{
		ClientID: "user-" + podName,
		ClientIP: utils.GetRelayIP(),
		Name:     "user-" + podName,
		Mode:     "server",
	}

	if bootstrapAddr != "" {
		// outside core bootstrap
		cfg.Addr = bootstrapAddr
		if relayUserToken == "" {
			return fmt.Errorf("empty user token for bootstrap")
		}
		cfg.TemplateToken = relayUserToken
		cfg.Scheme = "http"
		// check port is 443 then set scheme as https
		if utils.IsHTTPS(bootstrapAddr) {
			cfg.Scheme = "https"
		}
		// this is a server certificate CN is same as ServerHost
		cfg.ServerHost = utils.RelayUserHost
		cfg.ServerPort = utils.RelayUserPort
		err := prepareConfigCSRForBootStrapOutSideCore(cfg, cfg.ServerHost, log)
		if err != nil {
			return fmt.Errorf("failed in config csr for relay user server bootstrap")
		}
	} else {
		cfg.Addr = sentryAddr
		cfg.Scheme = "grpc"
		cfg.TemplateName = "rafay-core-relay-user"
	}

	if err := register.Register(ctx, cfg); err != nil {
		log.Error(
			err,
			"failed to register user relay",
		)
		return err
	}

	log.Info(
		"user certificate:",
		string(cfg.Certificate),
	)

	utils.RelayUserCert = cfg.Certificate
	utils.RelayUserKey = cfg.PrivateKey
	utils.RelayUserCACert = cfg.CACertificate
	utils.RelayUserPort = cfg.ServerPort
	utils.RelayUserHost = cfg.ServerHost

	return nil
}

// registerCDRelayUserServer will register with rafay-core-cd-relay-user template token
// registration fetches client-certificate/root-ca to terminate user connections
// the same certificate will be used as the client cert for peer-upstreams
func registerCDRelayUserServer(ctx context.Context, log *relaylogger.RelayLog) error {
	cfg := &register.Config{
		ClientID: "cd-user-" + podName,
		ClientIP: utils.GetRelayIP(),
		Name:     "cd-user-" + podName,
		Mode:     "server",
	}

	//inside core bootstrap
	cfg.Addr = sentryAddr
	cfg.Scheme = "grpc"
	cfg.TemplateName = "rafay-core-cd-relay-user"

	if err := register.Register(ctx, cfg); err != nil {
		log.Error(
			err,
			"failed to register user cd relay",
		)
		return err
	}

	log.Info(
		"user certificate:",
		string(cfg.Certificate),
	)

	utils.CDRelayUserCert = cfg.Certificate
	utils.CDRelayUserKey = cfg.PrivateKey
	utils.CDRelayUserCACert = cfg.CACertificate
	utils.CDRelayUserPort = cfg.ServerPort
	utils.CDRelayUserHost = cfg.ServerHost

	return nil
}

// registerRelayConnectorServer will register with rafay-core-relay-connector template token
// registration fetches client-certificate/root-ca to terminate connector connections
// the same certificate will be used as the client cert for peer-upstreams
func registerRelayConnectorServer(ctx context.Context, log *relaylogger.RelayLog) error {
	cfg := &register.Config{
		ClientID: "relay-server-" + podName,
		ClientIP: utils.GetRelayIP(),
		Name:     "relay-server-" + podName,
		Mode:     "server",
	}

	if bootstrapAddr != "" {
		// outside core bootstrap
		cfg.Addr = bootstrapAddr
		if relayConnectorToken == "" {
			return fmt.Errorf("empty user token for bootstrap")
		}
		cfg.TemplateToken = relayConnectorToken
		cfg.Scheme = "http"
		// check port is 443 then set scheme as https
		if utils.IsHTTPS(bootstrapAddr) {
			cfg.Scheme = "https"
		}
		// this is a server certificate CN is same as ServerHost
		cfg.ServerHost = utils.RelayConnectorHost
		cfg.ServerPort = utils.RelayConnectorPort
		err := prepareConfigCSRForBootStrapOutSideCore(cfg, cfg.ServerHost, log)
		if err != nil {
			return fmt.Errorf("failed in config csr for relay connector server bootstrap")
		}
	} else {
		//inside core bootstrap
		cfg.Addr = sentryAddr
		cfg.Scheme = "grpc"
		cfg.TemplateName = "rafay-core-relay-server"
	}

	if err := register.Register(ctx, cfg); err != nil {
		log.Error(
			err,
			"failed to register connector relay",
		)
		return err
	}

	log.Info(
		"connector certificate:",
		string(cfg.Certificate),
	)

	utils.RelayConnectorCert = cfg.Certificate
	utils.RelayConnectorKey = cfg.PrivateKey
	utils.RelayConnectorCACert = cfg.CACertificate
	utils.RelayConnectorHost = cfg.ServerHost
	utils.RelayConnectorPort = cfg.ServerPort

	return nil
}

// registerCDRelayConnectorServer will register with rafay-core-cd-relay-connector template token
// registration fetches client-certificate/root-ca to terminate connector connections
// the same certificate will be used as the client cert for peer-upstreams
func registerCDRelayConnectorServer(ctx context.Context, log *relaylogger.RelayLog) error {
	cfg := &register.Config{
		ClientID: "cd-relay-server-" + podName,
		ClientIP: utils.GetRelayIP(),
		Name:     "cd-relay-server-" + podName,
		Mode:     "server",
	}

	//inside core bootstrap
	cfg.Addr = sentryAddr
	cfg.Scheme = "grpc"
	cfg.TemplateName = "rafay-core-cd-relay-server"

	if err := register.Register(ctx, cfg); err != nil {
		log.Error(
			err,
			"failed to register connector cd relay",
		)
		return err
	}

	log.Info(
		"connector certificate:",
		string(cfg.Certificate),
	)

	utils.CDRelayConnectorCert = cfg.Certificate
	utils.CDRelayConnectorKey = cfg.PrivateKey
	utils.CDRelayConnectorCACert = cfg.CACertificate
	utils.CDRelayConnectorHost = cfg.ServerHost
	utils.CDRelayConnectorPort = cfg.ServerPort

	return nil
}

func relayServerBootStrap(ctx context.Context, log *relaylogger.RelayLog) {
	//peer service bootstrap
	ticker := time.NewTicker(5 * time.Second)
	for {
		err := registerRelayPeerService(ctx, log)
		if err != nil {
			log.Error(
				err,
				"failed to register relay with peer-service-bootstrap service, will retry",
			)
			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
		break
	}
	ticker.Stop()

	ticker = time.NewTicker(5 * time.Second)
	// relay user server bootstrap
	for {
		err := registerRelayUserServer(ctx, log)
		if err != nil {
			log.Error(
				err,
				"failed to register relay with relay-user-bootstrap service, will retry",
			)
			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
		break
	}
	ticker.Stop()

	ticker = time.NewTicker(5 * time.Second)
	// relay connector server bootstrap
	for {
		err := registerRelayConnectorServer(ctx, log)
		if err != nil {
			log.Error(
				err,
				"failed to register relay with relay-connector-bootstrap service, will retry",
			)
			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
		break
	}
	ticker.Stop()
}

func cdRelayServerBootStrap(ctx context.Context, log *relaylogger.RelayLog) {
	//peer service bootstrap
	ticker := time.NewTicker(5 * time.Second)
	for {
		err := registerCDRelayPeerService(ctx, log)
		if err != nil {
			log.Error(
				err,
				"failed to register cd relay with peer-service-bootstrap service, will retry",
			)
			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
		break
	}
	ticker.Stop()

	ticker = time.NewTicker(5 * time.Second)
	// relay user server bootstrap
	for {
		err := registerCDRelayUserServer(ctx, log)
		if err != nil {
			log.Error(
				err,
				"failed to register relay with cd-relay-user-bootstrap service, will retry",
			)
			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
		break
	}
	ticker.Stop()

	ticker = time.NewTicker(5 * time.Second)
	// relay connector server bootstrap
	for {
		err := registerCDRelayConnectorServer(ctx, log)
		if err != nil {
			log.Error(
				err,
				"failed to register relay with cd-relay-connector-bootstrap service, will retry",
			)
			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
		break
	}
	ticker.Stop()
}

//RunRelayServer entry to the relay server
func RunRelayServer(ctx context.Context, logLvl int) {
	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log = relaylogger.NewLogger(logLvl).WithName("Relay Server")

	if err := setupserver(log); err != nil {
		//config is not available
		log.Error(
			err,
			"relay server exiting",
		)
		utils.ExitChan <- true
		return
	}

restartServer:
	relayServerBootStrap(rctx, log)
	go tunnel.StartServer(rctx, log, fmt.Sprintf("%s/%s", auditPath, podName), utils.ExitChan)
	go runMetricServer(rctx, log)

	for {
		select {
		case <-utils.ExitChan:
			log.Info(
				"Relay server stopped, restart in 2 secs",
			)
			time.Sleep(2 * time.Second)
			goto restartServer
		case <-ctx.Done():
			log.Info(
				"Relay server exiting",
			)
			return
		}
	}
}

//RunCDRelayServer entry to the relay server
func RunCDRelayServer(ctx context.Context, logLvl int) {
	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log = relaylogger.NewLogger(logLvl).WithName("Relay Server")

	if err := setupCDServer(log); err != nil {
		//config is not available
		log.Error(
			err,
			"relay server exiting",
		)
		utils.ExitChan <- true
		return
	}

restartServer:
	cdRelayServerBootStrap(rctx, log)
	go tunnel.StartCDServer(rctx, log, fmt.Sprintf("%s/%s", auditPath, podName), utils.ExitChan)
	go runMetricServer(rctx, log)

	for {
		select {
		case <-utils.ExitChan:
			log.Info(
				"Relay server stopped, restart in 2 secs",
			)
			time.Sleep(2 * time.Second)
			goto restartServer
		case <-ctx.Done():
			log.Info(
				"Relay server exiting",
			)
			return
		}
	}
}

func runMetricServer(ctx context.Context, log *relaylogger.RelayLog) {

	m := http.NewServeMux()
	s := http.Server{Addr: metricListen, Handler: m}
	m.HandleFunc("/dialins", dialinMetrics)
	m.HandleFunc("/loglevel", setLoglevel)
	m.HandleFunc("/health", healthCheck)

	go func() {
		<-ctx.Done()
		s.Shutdown(context.Background())
	}()
	log.Info("Start Relay Metric Server at", metricListen)
	err := s.ListenAndServe()
	log.Info("Stopping Relay Metric Server at", metricListen, err)
}

func dialinMetrics(w http.ResponseWriter, r *http.Request) {
	tunnel.DialinMetric(w)
}

func setLoglevel(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["level"]
	if !ok || len(keys[0]) < 1 {
		fmt.Fprintf(w, "missing level query string")
		return
	}

	level := keys[0]
	lvl, err := strconv.Atoi(level)
	if err != nil {
		fmt.Fprintf(w, "invalid level query string")
		return
	}

	relaylogger.SetRunTimeLogLevel(lvl)

	fmt.Fprintf(w, "Success: set loglevel to "+level)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "up and feels good")
}
