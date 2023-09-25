package server

import (
	"context"
	"net"
	"os"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/grpc"
	sentryrpc "github.com/paralus/paralus/proto/rpc/sentry"
)

// The peering service operate a s follows.
// Performs all grpc send/recv with connected relays
// Recvs probe request from a relay and perform survey across other relays.
// It maintains 3 streams
// 1) HelloRPC heartbeat stream
// 2) Probe stream to recv probe requests and send probe responses
// 3) Survey stream to send survey requests and recv survey responses.
// Maintains the list of relays connected to the service.
// - Relay list is keyed with the UUID of the relay
// - Each relay object has a survey send chnl. Use this chnl to send survey requests.
// - Each relay obect has a probe chnl. Use this chnl to send probe response

// used for survey broadcasting
type surveyBroadCastRequest struct {
	clustersni string
	relayuuid  string // relay requsting the survey
	ou         string
}

// used to maintain list of connected relays
type relayObject struct {
	timeStamp         int64
	refCnt            uint8
	relayip           string
	ou                string
	probeReplyChnl    chan sentryrpc.PeerProbeResponse
	surveyRequestChnl chan sentryrpc.PeerSurveyRequest
}

// relayPeerService relay peer service
type relayPeerService struct {
	cert   []byte // rpc server certifciate
	key    []byte // rpc server key
	rootCA []byte // rpc rootCA to verify client certificates.
	port   int

	//ServiceUUID ...
	ServiceUUID string

	//relayMutex to syncrinise operations
	relayMutex sync.RWMutex

	//RelayMap list of active
	RelayMap map[string]map[string]*relayObject

	//SurveyBroadCast send survey request to all the relays
	surveyBroadCast chan surveyBroadCastRequest

	//SurveyCacheExpiry default expiry
	surveyCacheExpiry time.Duration

	//peerServiceCache stores peer dialin info
	peerServiceCache *ristretto.Cache
}

var maxRelayIdle = 300 //5 min

var _ sentryrpc.RelayPeerServiceServer = (*relayPeerService)(nil)

//var _log = logv2.GetLogger()

// initPeerServiceCache initialize the cache to store dialin cluster-connection
// information of peers. When a dialin miss happens look into this cache
// to find the peer IP address to forward the user connection.
func initPeerServiceCache() (*ristretto.Cache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // Num keys to track frequency of (10M).
		MaxCost:     1 << 30, // Maximum cost of cache (1GB).
		BufferItems: 64,      // Number of keys per Get buffer.
	})
}

// insertPeerServiceCache inserts the value to cache
func (s *relayPeerService) insertPeerServiceCache(key, value interface{}) bool {
	return s.peerServiceCache.SetWithTTL(key, value, 100, s.surveyCacheExpiry)
}

// getPeerServiceCache get value from cache and if more than 1
// rnadomly select the peer
func (s *relayPeerService) getPeerServiceCache(key interface{}) (string, bool) {
	value, found := s.peerServiceCache.Get(key)
	if found {
		if value == nil {
			return "", false
		}
		relayip := value.(string)

		return relayip, true
	}
	return "", found
}

// NewRelayPeerService returns new placement server implementation
func NewRelayPeerService() (sentryrpc.RelayPeerServiceServer, error) {
	cache, err := initPeerServiceCache()
	if err != nil {
		_log.Errorw("failed to init cache", "error", err)
		return nil, err
	}

	return &relayPeerService{
		ServiceUUID:       uuid.New().String(),
		RelayMap:          make(map[string]map[string]*relayObject),
		surveyBroadCast:   make(chan surveyBroadCastRequest, 256),
		surveyCacheExpiry: 60 * time.Second,
		peerServiceCache:  cache,
	}, nil
}

// RunRelaySurveyHandler is the cotrol loop that maintains the peer suvey
// messages.
func RunRelaySurveyHandler(stop <-chan struct{}, svc interface{}) {
	s := svc.(*relayPeerService)
	_log.Infow("started survey request handler")
	for {
		select {
		case <-stop:
			_log.Errorw("stopping relay servey handler")
			return
		case surveyReq := <-s.surveyBroadCast:
			go s.handleSurveyReq(surveyReq)
		}
	}
}

// process the survey request.
func (s *relayPeerService) handleSurveyReq(req surveyBroadCastRequest) {
	// Get All relay objects
	// Send survey request to all.
	// Featch from cache in intervals of 1 sec for 5 times
	// Send result of each fetch to the requesting replay
	var relayids []string
	var connInfo []*sentryrpc.RelayClusterConnectionInfo
	var retry int
	var foundStale bool
	sreqmsg := sentryrpc.PeerSurveyRequest{
		Clustersni: req.clustersni,
	}

	_log.Debugw("handleSurveyReq", "RelayMap length", len(s.RelayMap))

	foundStale = false

	//broadcast request to all connected relays
	s.relayMutex.RLock()
	now := time.Now().Unix()
	if relayList, ok := s.RelayMap[req.ou]; ok {
		for relayuuid, robj := range relayList {
			if now > robj.timeStamp && (now-robj.timeStamp) > int64(maxRelayIdle) { //5min max toleration
				//skip the relay that did not have heart beat for 5 mins
				foundStale = true
				continue
			}
			if relayuuid != req.relayuuid {
				relayids = append(relayids, relayuuid)
				//wait max of 2 sec to send to chnl
				tick := time.NewTicker(2 * time.Second)
			handleSurveyReqBreak:
				for {
					select {
					case <-tick.C:
						break handleSurveyReqBreak
					case robj.surveyRequestChnl <- sreqmsg:
						break handleSurveyReqBreak
					}
				}
				tick.Stop()
			}
		}
	}
	s.relayMutex.RUnlock()

	_log.Debugw("handleSurveyReq done broadcasting to relays, wait for response")

	//now waiting for reply. Survey responses get cached
	//poll the cache few times.
	retry = 0
	for {
		connInfo = nil
		//fecth response from cache
		for _, rid := range relayids {
			ckey := peerServiceCacheKey(req.clustersni, rid, req.ou)
			ip, ok := s.getPeerServiceCache(ckey)
			if ok {
				//prepare the probe response
				cinfo := &sentryrpc.RelayClusterConnectionInfo{
					Relayuuid: rid,
					Relayip:   ip,
				}
				connInfo = append(connInfo, cinfo)
			}
		}

		if len(connInfo) > 0 {
			msg := sentryrpc.PeerProbeResponse{
				Clustersni: req.clustersni,
				Items:      connInfo,
			}

			robj := s.getRelayObject(req.relayuuid, req.ou)
			if robj != nil {
				//send the probe response
				robj.probeReplyChnl <- msg
				s.putRelayObject(req.relayuuid, req.ou)
			} else {
				//response chnl not found
				break
			}
		}

		//retry 5 times: total 5 sec wait to get reply from all peers
		retry++
		if retry > 5 {
			break
		}

		time.Sleep(1 * time.Second)
	}

	if foundStale {
		//remove inactive relays
		s.relayMutex.Lock()
		now := time.Now().Unix()
		for _, relayList := range s.RelayMap {
			for relayuuid, robj := range relayList {
				if now > robj.timeStamp && (now-robj.timeStamp) > 300 { //5min max toleration
					if robj.refCnt > 0 {
						_log.Errorw("inactive relay has refcnt")
					} else {
						//delete the relay that did not have heart beat for 5 mins
						delete(relayList, relayuuid)
					}
				}
			}
		}
		s.relayMutex.Unlock()
	}

}

// maintains the timestamp of relays heart beat
func (s *relayPeerService) updateRelayIfExist(relayuuid, ou string) bool {
	s.relayMutex.RLock()
	defer s.relayMutex.RUnlock()
	// check relay exist
	if relayList, ok := s.RelayMap[ou]; ok {
		if robj, found := relayList[relayuuid]; found {
			if robj.ou == ou {
				//update the time stamp
				robj.timeStamp = time.Now().Unix()
				return true
			}
		}
	}
	return false
}

func (s *relayPeerService) getRelayObject(relayuuid, ou string) *relayObject {
	s.relayMutex.Lock()
	defer s.relayMutex.Unlock()

	if relayList, ok := s.RelayMap[ou]; ok {
		if robj, found := relayList[relayuuid]; found {
			if robj.ou == ou {
				robj.refCnt++
				return robj
			}
		}
	}
	return nil
}

func (s *relayPeerService) putRelayObject(relayuuid, ou string) {
	s.relayMutex.Lock()
	defer s.relayMutex.Unlock()

	if relayList, ok := s.RelayMap[ou]; ok {
		if robj, found := relayList[relayuuid]; found {
			if robj.ou == ou && robj.refCnt > 0 {
				robj.refCnt--
				return
			}
		}
	}
}

func (s *relayPeerService) insertRelayObject(robj *relayObject, relayuuid, ou string) {
	s.relayMutex.Lock()
	defer s.relayMutex.Unlock()
	relayList := s.RelayMap[ou]
	if relayList == nil {
		relayList := map[string]*relayObject{
			relayuuid: robj,
		}
		s.RelayMap[ou] = relayList
	} else {
		relayList[relayuuid] = robj
	}
}

func (s *relayPeerService) handleHelloRequest(relayuuid, relayip, ou string) {
	res := s.updateRelayIfExist(relayuuid, ou)
	if res {
		return
	}

	robj := &relayObject{
		timeStamp: time.Now().Unix(),
		relayip:   relayip,
		refCnt:    0,
		ou:        ou,
	}

	robj.probeReplyChnl = make(chan sentryrpc.PeerProbeResponse, 128)
	robj.surveyRequestChnl = make(chan sentryrpc.PeerSurveyRequest, 128)

	s.insertRelayObject(robj, relayuuid, ou)
}

// getServiceIP ..
func getServiceIP() string {
	name, err := os.Hostname()
	if err == nil {
		addr, err := net.LookupIP(name)
		if err == nil {
			return addr[0].String()
		}
	}
	return ""
}

// RelayPeerHelloRPC handles PeerHelloMsg
func (s *relayPeerService) RelayPeerHelloRPC(stream sentryrpc.RelayPeerService_RelayPeerHelloRPCServer) error {
	_log.Infow("RelayPeerHelloRPC stream")
	name, err := grpc.GetClientName(stream.Context())
	if err != nil {
		_log.Errorw("error in getting CN from certificate in relay peer hello rpc", "error", err)
		return err
	}

	ou, err := grpc.GetClientOU(stream.Context())
	if err != nil {
		_log.Errorw("error in getting OU from certificate in relay peer hello rpc", "error", err)
		return err
	}

	_log.Infow("RelayPeerHelloRPC client ", "name", name)

	for {
		in, err := stream.Recv()
		if err != nil {
			_log.Errorw("RelayPeerHelloRPC recv", "name", name, "error", err)
			return err
		}

		relayuuid := in.GetRelayuuid()
		relayip := in.GetRelayip()
		_log.Debugw("RelayPeerHelloRPC:Received value", "relayuuid", relayuuid, "relayip", relayip)

		msg := &sentryrpc.PeerHelloResponse{
			Serviceuuid: s.ServiceUUID,
			Serviceip:   getServiceIP(),
		}

		stream.Send(msg)
		go s.handleHelloRequest(relayuuid, relayip, ou)
	}
}

// relayPeerProbeSender send routine to handle sending probe messges
func (s *relayPeerService) relayPeerProbeSender(ctx context.Context, stream sentryrpc.RelayPeerService_RelayPeerProbeRPCServer, relayuuid string, robj *relayObject) {
	for {
		select {
		case <-ctx.Done():
			s.putRelayObject(relayuuid, robj.ou)
			return
		case probeReply := <-robj.probeReplyChnl:
			err := stream.Send(&probeReply)
			if err != nil {
				s.putRelayObject(relayuuid, robj.ou)
				return
			}
		}
	}
}

// try to fill the response form cache
func (s *relayPeerService) tryResponseFromCache(relayuuid, clustersni, ou string) bool {
	var relayids []string
	var connInfo []*sentryrpc.RelayClusterConnectionInfo

	s.relayMutex.RLock()
	// get all other relays in the peer list
	if relayList, ok := s.RelayMap[ou]; ok {
		for key := range relayList {
			if key != relayuuid {
				relayids = append(relayids, key)
			}
		}
	}
	s.relayMutex.RUnlock()

	connInfo = nil
	//fecth response from cache
	for _, rid := range relayids {
		ckey := peerServiceCacheKey(clustersni, rid, ou)
		ip, ok := s.getPeerServiceCache(ckey)
		if ok {
			cinfo := &sentryrpc.RelayClusterConnectionInfo{
				Relayuuid: rid,
				Relayip:   ip,
			}
			connInfo = append(connInfo, cinfo)
		}
	}

	if len(connInfo) > 0 {
		robj := s.getRelayObject(relayuuid, ou)
		if robj != nil {
			msg := sentryrpc.PeerProbeResponse{
				Clustersni: clustersni,
				Items:      connInfo,
			}
			robj.probeReplyChnl <- msg
			s.putRelayObject(relayuuid, ou)
		}
		return true
	}

	return false
}

// RelayPeerProbeRPC handles PeerHelloMsg
func (s *relayPeerService) RelayPeerProbeRPC(stream sentryrpc.RelayPeerService_RelayPeerProbeRPCServer) error {
	var initSend bool

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	name, err := grpc.GetClientName(stream.Context())
	if err != nil {
		_log.Errorw("RelayPeerProbeRPC", "error", err)
		return err
	}

	ou, err := grpc.GetClientOU(stream.Context())
	if err != nil {
		_log.Errorw("error in getting OU from certificate in relay peer hello rpc", "error", err)
		return err
	}

	_log.Infow("started probe server stream RelayPeerProbeRPC for client", "name", name, "OU", ou)

	initSend = false
	for {
		in, err := stream.Recv()
		if err != nil {
			_log.Errorw("RelayPeerProbeRPC rcv", "name", name, "error", err)
			return err
		}

		//got a probe request
		clustersni := in.GetClustersni()
		relayuuid := in.GetRelayuuid()

		_log.Debugw("RelayPeerProbeRPC recvd values", "relayuuid", relayuuid, "clustersni", clustersni)

		if clustersni == "" && relayuuid != "" {
			//init dummy probe
			if !initSend {
				//Get the relay object for relayuuid
				robj := s.getRelayObject(relayuuid, ou)
				if robj != nil {
					//start the send handler for this relay
					go s.relayPeerProbeSender(ctx, stream, relayuuid, robj)
					initSend = true
					_log.Debug("RelayPeerProbeRPC init done")
				}
			}
			continue
		}

		// find response either from cache or via survey
		go func() {
			if clustersni != "" && relayuuid != "" {
				if !s.tryResponseFromCache(relayuuid, clustersni, ou) {
					//did not find in cache, trigger survey to all relays
					surveyreq := surveyBroadCastRequest{
						clustersni: clustersni,
						relayuuid:  relayuuid,
						ou:         ou,
					}
					s.surveyBroadCast <- surveyreq
				}
			}
		}()

	}

}

// relayPeerSurveySender send routine to handle sending probe messges
func (s *relayPeerService) relayPeerSurveySender(ctx context.Context, stream sentryrpc.RelayPeerService_RelayPeerSurveyRPCServer, relayuuid string, robj *relayObject) {
	_log.Debugw("started relayPeerSurveySender")
	for {
		select {
		case <-ctx.Done():
			s.putRelayObject(relayuuid, robj.ou)
			return
		case surveyRequest := <-robj.surveyRequestChnl:
			_log.Debugw("msg recvd from survey chnl sending to stream")
			err := stream.Send(&surveyRequest)
			if err != nil {
				s.putRelayObject(relayuuid, robj.ou)
				return
			}
		}
	}
}

// RelayPeerSurveyRPC handles relay survey rpc
func (s *relayPeerService) RelayPeerSurveyRPC(stream sentryrpc.RelayPeerService_RelayPeerSurveyRPCServer) error {
	var initSend bool

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	name, err := grpc.GetClientName(stream.Context())
	if err != nil {
		_log.Errorw("relayPeerSurveyRPC", "error", err)
	}

	ou, err := grpc.GetClientOU(stream.Context())
	if err != nil {
		_log.Errorw("error in getting OU from certificate in relay peer hello rpc", "error", err)
		return err
	}

	_log.Infow("started RelayPeerSurveyRPC stream from client ", "name", name, "OU", ou)

	initSend = false
	for {
		in, err := stream.Recv()
		if err != nil {
			_log.Error("RelayPeerSurveyRPC rcv ", "name", name, "error", err)
			return err
		}

		clustersni := in.GetClustersni()
		relayuuid := in.GetRelayuuid()
		relayip := in.GetRelayip()
		_log.Debugw("RelayPeerSurveyRPC:Received values", "clustersni", clustersni, "relayuuid", relayuuid, "relayip", relayip)

		if clustersni == "" && relayuuid != "" {
			if !initSend {
				//Get the relay object for relayuuid
				robj := s.getRelayObject(relayuuid, ou)
				if robj != nil {
					go s.relayPeerSurveySender(ctx, stream, relayuuid, robj)
					initSend = true
					_log.Debugw("RelayPeerSurveyRPC init done")
				}
			}
			continue
		}

		//insert response to cache
		if clustersni != "" && relayuuid != "" && relayip != "" {
			ckey := peerServiceCacheKey(clustersni, relayuuid, ou)
			if !s.insertPeerServiceCache(ckey, relayip) {
				_log.Errorw("failed to insert into cache")
			}
		}

	}

}

func peerServiceCacheKey(clustersni, relayuuid, ou string) string {
	return clustersni + relayuuid + ou
}
