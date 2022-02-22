package tunnel

import (
	"context"
	"crypto/tls"
	"net/url"

	"time"

	comgrpc "github.com/RafaySystems/rcloud-base/components/common/pkg/grpc"
	peerclient "github.com/RafaySystems/rcloud-base/components/common/pkg/sentry/peering"
	relayrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
	"google.golang.org/grpc/credentials"
)

var peerlog *relaylogger.RelayLog

//StartPeeringMgr will start the peering RPCs
func StartPeeringMgr(ctx context.Context, log *relaylogger.RelayLog, exitChan chan<- bool, config *ServerConfig) {
	var tlsConfig *tls.Config
	var err error
	var peerSeviceHost string

	peerlog = log.WithName("PeeringMgr")

	if utils.PeerCache == nil {
		peerlog.Error(
			nil,
			"peer client cache is nil",
		)
		exitChan <- true
		return
	}

	//Peer Service Clint certs is popultaed from boot strap use it
	if len(utils.PeerCertificate) > 0 && len(utils.PeerPrivateKey) > 0 && len(utils.PeerCACertificate) > 0 {
		u, err := url.Parse(utils.PeerServiceURI)
		if err != nil {
			peerlog.Error(
				nil,
				"peer servic uri parse failed",
			)
			exitChan <- true
			return
		}
		//Load certificates
		tlsConfig, err = ClientTLSConfigFromBytes(utils.PeerCertificate, utils.PeerPrivateKey, utils.PeerCACertificate, u.Host)
		peerSeviceHost = u.Host
	} else {
		//Load certificates
		tlsConfig, err = ClientTLSConfig(config.Controller.ClientCRT, config.Controller.ClientKEY, config.Controller.RootCA, config.Controller.PeerProbeSNI)
		peerSeviceHost = config.Controller.PeerProbeSNI
	}

	if err != nil {
		peerlog.Error(
			err,
			"Error loading peer TLC config",
		)
		exitChan <- true
		return
	}

	//enforce TLS mutual authN
	transportCreds := credentials.NewTLS(tlsConfig)

connectRetry:
	log.Debug("Start Peer grpc dial")
	conn, err := comgrpc.NewSecureClientConn(ctx, peerSeviceHost, transportCreds)
	if err != nil {
		peerlog.Error(
			err,
			"failed to connect to peer server, retry in 10 seconds",
			peerSeviceHost,
		)
		time.Sleep(10 * time.Second)
		goto connectRetry
	}

	log.Info("grpc connected to peer service", peerSeviceHost)
	client := relayrpc.NewRelayPeerServiceClient(conn)

	// create RPC streams
	helloStream, err := client.RelayPeerHelloRPC(context.Background())
	if err != nil {
		peerlog.Error(
			err,
			"failed to create HelloRPC stream with peer server, retry in 10 seconds",
			peerSeviceHost,
		)
		conn.Close()
		time.Sleep(10 * time.Second)
		goto connectRetry
	}

	probeStream, err := client.RelayPeerProbeRPC(context.Background())
	if err != nil {
		peerlog.Error(
			err,
			"failed to create ProbeRPC stream with peer server, retry in 10 seconds",
			peerSeviceHost,
		)
		conn.Close()
		time.Sleep(10 * time.Second)
		goto connectRetry
	}

	surveyStream, err := client.RelayPeerSurveyRPC(context.Background())
	if err != nil {
		peerlog.Error(
			err,
			"failed to create SurveyRPC stream with peer server, retry in 10 seconds",
			peerSeviceHost,
		)
		conn.Close()
		time.Sleep(10 * time.Second)
		goto connectRetry
	}

	peerlog.Debug(
		"created RPC streams with peer server",
		peerSeviceHost,
	)

	rpcctx, rpccancel := context.WithCancel(context.Background())
	go peerclient.ClientHelloRPC(rpcctx, helloStream, utils.PeerHelloInterval, utils.RelayUUID, utils.GetRelayIP)
	//Add a init time wait for hello stream to finish
	time.Sleep(2 * time.Second)

	go peerclient.ClientProbeRPC(rpcctx, probeStream, utils.PeerCache, utils.RelayUUID, utils.PeerCacheDefaultExpiry, PeerProbeChanel, utils.GetRelayIPPort)

	go peerclient.ClientSurveyRPC(rpcctx, surveyStream, utils.RelayUUID, utils.GetRelayIPPort, dialinCountLookup)

	for {
		//Watch for errors in rpc streams.
		//On error cancel the streams and reconnect.
		//HelloRPC send heartbeats in every 60Sec.
		//If there is underlying connectivity issues
		//HelloRPC will detect it with in 60Sec,
		select {
		case <-helloStream.Context().Done():
			rpccancel()
			conn.Close()
			time.Sleep(5 * time.Second)
			goto connectRetry
		case <-probeStream.Context().Done():
			rpccancel()
			conn.Close()
			time.Sleep(5 * time.Second)
			goto connectRetry
		case <-surveyStream.Context().Done():
			rpccancel()
			conn.Close()
			time.Sleep(5 * time.Second)
			goto connectRetry
		case <-ctx.Done():
			rpccancel()
			slog.Error(
				ctx.Err(),
				"Stopping  PeeringMgr",
			)
			return
		}
	}

}
