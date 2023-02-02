package peering

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509/pkix"
	"database/sql"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/paralus/paralus/pkg/log"
	"github.com/paralus/paralus/pkg/query"
	"github.com/paralus/paralus/pkg/sentry/cryptoutil"
	"github.com/paralus/paralus/pkg/sentry/register"
	"github.com/paralus/paralus/pkg/service"
	sentryrpc "github.com/paralus/paralus/proto/rpc/sentry"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/rs/xid"
)

var (
	_log = log.GetLogger()
)

// RelayClusterConnectionInfo relay conn info
type RelayClusterConnectionInfo struct {
	Relayuuid string
	Relayip   string
}

type OnEvict = func(item *ristretto.Item)

// InitPeerCache initialize the cache to store dialin cluster-connection
// information of peers. When a dialin miss happens look into this cache
// to find the peer IP address to forward the user connection.
func InitPeerCache(evict OnEvict) (*ristretto.Cache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // Num keys to track frequency of (10M).
		MaxCost:     1 << 30, // Maximum cost of cache (1GB).
		BufferItems: 64,      // Number of keys per Get buffer.
		OnEvict:     evict,
	})
}

// InsertPeerCache inserts the value to cache
func InsertPeerCache(cache *ristretto.Cache, expiry time.Duration, key, value interface{}) bool {
	return cache.SetWithTTL(key, value, 100, expiry)
}

// GetPeerCache get value from cache and if more than 1
// rnadomly select the peer
func GetPeerCache(cache *ristretto.Cache, key interface{}) (string, bool) {
	value, found := cache.Get(key)
	if found {
		if value == nil {
			return "", false
		}
		cacheItems := value.([]RelayClusterConnectionInfo)
		cnt := len(cacheItems)
		if cnt > 1 {
			rand.Seed(time.Now().UnixNano())
			min := 0
			max := cnt - 1
			indx := rand.Intn(max-min+1) + min
			return cacheItems[indx].Relayip, found
		}
		return cacheItems[0].Relayip, found
	}
	return "", found
}

// sends periodic heartbeats to core service
func helloRPCSend(ctx context.Context, stream sentryrpc.RelayPeerService_RelayPeerHelloRPCClient, interval time.Duration, relayUUID string, ip func() string) {
	_log.Infow("send first hello")
	msg := &sentryrpc.PeerHelloRequest{
		Relayuuid: relayUUID,
		Relayip:   ip(),
	}
	// send first hello message
	err := stream.Send(msg)
	if err != nil {
		_log.Errorw("failed to send hello message", "error", err)
		return
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

helloRPCSendLoop:
	for {
		select {
		case <-ctx.Done():
			stream.CloseSend()
			break helloRPCSendLoop
		case <-tick.C:
			err := stream.Send(msg)
			if err != nil {
				_log.Errorw("failed to send hello message", err)
				break helloRPCSendLoop
			}
		}
	}
	_log.Debugw("Exit: helloRPCSendLoop")
}

// ClientHelloRPC will handle periodic heartbeat messages between relay and the core service.
func ClientHelloRPC(ctx context.Context, stream sentryrpc.RelayPeerService_RelayPeerHelloRPCClient, interval time.Duration, relayUUID string, ip func() string) {

	go helloRPCSend(ctx, stream, interval, relayUUID, ip)

	for {
		in, err := stream.Recv()
		if err != nil {
			_log.Errorw("helloRPC stream recv error", err)
			stream.CloseSend()
			break
		}
		_log.Debugw("recvd hello resp from service",
			"serviceip", in.GetServiceip(),
			"serviceuuid", in.GetServiceuuid(),
		)
	}

	_log.Debugw("stopping helloRPC routine")
}

// send loop of probe rpc. recvs clustersni from PeerProbeChanel and sends to core service
func probeRPCSend(ctx context.Context, stream sentryrpc.RelayPeerService_RelayPeerProbeRPCClient, relayUUID string, peerProbeChanel chan string) {
probeRPCSendLoop:
	for {
		select {
		case clustersni := <-peerProbeChanel:
			_log.Debugw("probeRPCSend", "clustersni", clustersni)
			msg := &sentryrpc.PeerProbeRequest{
				Relayuuid:  relayUUID,
				Clustersni: clustersni,
			}
			err := stream.Send(msg)
			if err != nil {
				_log.Errorw(
					"failed to send probe message for ",
					"clustersni", clustersni,
					"error", err,
				)
				stream.CloseSend()
				break probeRPCSendLoop
			}
		case <-ctx.Done():
			stream.CloseSend()
			break probeRPCSendLoop
		}

	}
	_log.Debugw("exit: probeRPCSendLoop")
}

// ClientProbeRPC will manage the probes.
// When a relay neeed to probe for connections for a cluster
// it will add a probe message to probe channel. The probe rpc
// will send that message to probe core service. When a probe response
// get inserted to peerCache.
func ClientProbeRPC(ctx context.Context, stream sentryrpc.RelayPeerService_RelayPeerProbeRPCClient, pcache *ristretto.Cache, relayUUID string, expiry time.Duration, peerProbeChanel chan string, ip func() string) {
	//send message with empty cluster sni to init the chanl
	msg := &sentryrpc.PeerProbeRequest{
		Relayuuid:  relayUUID,
		Clustersni: "",
	}

	err := stream.Send(msg)
	if err != nil {
		stream.CloseSend()
		return
	}
	_log.Debugw("probeRPC send first", "msg", msg)

	go probeRPCSend(ctx, stream, relayUUID, peerProbeChanel)

	//probe response recv loop will process the response
	//and push it into peer cache
	for {
		resp, err := stream.Recv()
		if err != nil {
			stream.CloseSend()
			break
		}

		clustersni := resp.GetClustersni()
		items := resp.GetItems()
		if clustersni != "" && items != nil && len(items) > 0 {
			cachevalue := []RelayClusterConnectionInfo{}
			//prepare cachevalue
			for _, item := range items {
				matched, err := regexp.Match(relayUUID, []byte(item.Relayuuid))
				if err == nil && matched {
					_log.Errorw("skip duplicate probe resp",
						"relayuuid", relayUUID,
						"recvd-relayuuid", item.Relayuuid,
					)
					//uuid is same as this relay skip this entry
					continue
				}

				ipAddr := ip()
				if ipAddr != "" && ipAddr == item.Relayip {
					_log.Errorw("skip duplicate probe resp", "ip address", item.Relayip)
					//ip is same as this relay skip this entry
					continue
				}

				v := RelayClusterConnectionInfo{item.Relayuuid, item.Relayip}
				cachevalue = append(cachevalue, v)
			}

			_log.Infow(
				"cache probeRPC response",
				"key", clustersni,
				"value", cachevalue,
			)
			//insert to peer cache
			if !InsertPeerCache(pcache, expiry, clustersni, cachevalue) {
				_log.Errorw(
					"failed cache probeRPC response",
					"key", clustersni,
					"value", cachevalue,
				)
			}
		} else {
			_log.Errorw(
				"prob response with empty items for ", clustersni,
			)
		}
	}

	_log.Debug(
		"stopping probeRPC routine",
	)
}

// ClientSurveyRPC will handle the survey RPC.
// When a relay neeed to probe for connections for a cluster
// it will message to probe core service. The service then sends
// survey to all the connected relays (survey-request).
// On survey request, the relay will lookup its local-dialin-map
// and reply to core if connection from the given cluster is available.
func ClientSurveyRPC(ctx context.Context, stream sentryrpc.RelayPeerService_RelayPeerSurveyRPCClient, relayUUID string, ip func() string, dialinlookup func(string) int) {
	var relayIP string

	relayIP = ip()

	//Send a empty clustersni meesage to init the channel
	msg := &sentryrpc.PeerSurveyResponse{
		Relayuuid:  relayUUID,
		Relayip:    relayIP,
		Clustersni: "",
	}

	err := stream.Send(msg)
	if err != nil {
		stream.CloseSend()
		return
	}

	_log.Debugw("surveyRPC send first", "msg", msg)
	go func() {
	surveyRPCStremWatch:
		for {
			select {
			case <-ctx.Done():
				stream.CloseSend()
				break surveyRPCStremWatch
			}
		}
	}()

	//survey request recv loop will process the survey requests
	for {

		surveyReq, err := stream.Recv()
		if err != nil {
			stream.CloseSend()
			break
		}

		clustersni := surveyReq.GetClustersni()
		if clustersni == "" {
			_log.Errorw(
				"prob response with empty items for ",
				"clustersni", clustersni,
			)
			continue
		}

		//lookup the local dialin table for connections\
		cnt := dialinlookup(clustersni)
		_log.Infow(
			"survey lookup",
			"key", clustersni,
			"count", cnt,
		)

		if relayIP == "" {
			_log.Errorw(
				"survey failed to get relay ip",
				"key", clustersni,
			)
			continue
		}

		if cnt > 0 {
			msg := &sentryrpc.PeerSurveyResponse{
				Relayuuid:  relayUUID,
				Relayip:    relayIP,
				Clustersni: clustersni,
			}
			err = stream.Send(msg)
			if err != nil {
				_log.Errorw(
					"survey send response failed",
					"key", clustersni,
					"error", err,
				)
				stream.CloseSend()
				break
			}
		}

	}

	_log.Debug(
		"stopping probeRPC routine",
	)
}

// GetPeeringServerCreds returns peering cert, key, ca
func GetPeeringServerCreds(ctx context.Context, bs service.BootstrapService, rpcPort int, host string) (cert, key, ca []byte, err error) {
	nctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	var template *sentry.BootstrapAgentTemplate

	template, err = bs.GetBootstrapAgentTemplate(nctx, "paralus-sentry-peering-server")
	if err != nil {
		return
	}

	config := &register.Config{
		TemplateName: "paralus-sentry-peering-server",
		Addr:         fmt.Sprintf("localhost:%d", rpcPort),
		Name:         "paralus-sentry-peering-server",
		Scheme:       "grpc",
		Mode:         "server",
	}

	var privKey *ecdsa.PrivateKey

	privKey, err = cryptoutil.GenerateECDSAPrivateKey()
	if err != nil {
		return
	}

	config.PrivateKey, err = cryptoutil.EncodePrivateKey(privKey, cryptoutil.NoPassword)
	if err != nil {
		return
	}

	var csr []byte

	csr, err = cryptoutil.CreateCSR(pkix.Name{
		CommonName:         host,
		Country:            []string{"USA"},
		Organization:       []string{"Paralus"},
		OrganizationalUnit: []string{"Paralus Sentry Peering Server"},
		Province:           []string{"California"},
		Locality:           []string{"Sunnyvale"},
	}, privKey)
	if err != nil {
		return
	}

	config.CSR = csr

	var agent *sentry.BootstrapAgent

	agent, err = bs.GetBootstrapAgent(nctx, template.Metadata.Name, query.WithName("paralus-sentry-peering-server"), query.WithGlobalScope())

	if err != nil {
		if err != sql.ErrNoRows {
			return
		}
	}

	if agent != nil {
		config.ClientID = agent.Spec.Token
	} else {
		config.ClientID = xid.New().String()
	}

	err = register.Register(nctx, config)
	if err != nil {
		return
	}

	cert = config.Certificate
	key = config.PrivateKey
	ca = config.CACertificate

	return
}
