package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/RafayLabs/rcloud-base/pkg/grpc"
	"github.com/RafayLabs/rcloud-base/pkg/sentry/peering"
	relayrpc "github.com/RafayLabs/rcloud-base/proto/rpc/sentry"
	"github.com/google/uuid"
	"google.golang.org/grpc/credentials"
)

var relayUUID1 string
var relayUUID2 string

// steps to test peering protocols
// * start a server
// * runs two client instances, send hello rpc to build active relay list
// * client1 send probe for a dummy cluster id
// * peer service broadcast survey request to clients
// * client2 sends survey reply back
// * peer service sends probe reply
// * client1 get probe reply back

func exit(cancel context.CancelFunc) {
	cancel()
	os.Exit(0)
}

//GenUUID generates a google UUID
func genUUID() string {
	id := uuid.New()
	return id.String()
}

//GetRelayIP get relay IP address
func getRelayIP1() string {
	return "1.1.1.1"
}

//GetRelayIP get relay IP address
func getRelayIP2() string {
	return "2.2.2.2"
}

func dummyDialinLookup(clustersni string) int {
	return 1
}

func readPEM(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return b
}

//os signal handler
func signalHandler(sig os.Signal, cancel context.CancelFunc) {
	if sig == syscall.SIGINT || sig == syscall.SIGKILL || sig == syscall.SIGTERM || sig == syscall.SIGQUIT {
		fmt.Println("Received", "signal", sig)
		exit(cancel)
		return
	}

	fmt.Println("Received", "signal", sig)
}

func startClient(ctx context.Context, t *testing.T, name, relayUUID string, exitChan chan<- bool, getRelayIP func() string, peerProbeChanel chan string) {
	tlsConfig, err := ClientTLSConfig("./testdata/peersvc.crt", "./testdata/peersvc.key", "./testdata/ca.crt", "star.probe.relay.rafay.dev:7001")
	if err != nil {
		t.Error("Error loading peer TLC config", err)
		exitChan <- true
		return
	}

	//enforce TLS mutual authN
	transportCreds := credentials.NewTLS(tlsConfig)
	conn, err := grpc.NewSecureClientConn(ctx, "127.0.0.1:7001", transportCreds)

	if err != nil {
		t.Error("Error connecting to peer service", err)
		exitChan <- true
		return
	}

	fmt.Println("connected to grpc server")
	client := relayrpc.NewRelayPeerServiceClient(conn)

	// create RPC streams
	helloStream, err := client.RelayPeerHelloRPC(context.Background())
	if err != nil {
		t.Error(
			"failed to create HelloRPC stream with peer server", name, err,
		)
		conn.Close()
		exitChan <- true
	}

	probeStream, err := client.RelayPeerProbeRPC(context.Background())
	if err != nil {
		t.Error(
			err,
			"failed to create ProbeRPC stream with peer server", name, err,
		)
		conn.Close()
		exitChan <- true
	}

	surveyStream, err := client.RelayPeerSurveyRPC(context.Background())
	if err != nil {
		t.Error(
			err,
			"failed to create SurveyRPC stream with peer server", name, err,
		)
		conn.Close()
		exitChan <- true
	}

	fmt.Println("created RPC streams with peer server")

	rpcctx, rpccancel := context.WithCancel(context.Background())

	go ClientHelloRPC(rpcctx, helloStream, 60*time.Second, relayUUID, getRelayIP)
	//Add a init time wait for hello stream to finish
	time.Sleep(2 * time.Second)

	pcache, err := InitPeerCache(nil)
	if err != nil {
		t.Error(
			err,
			"failed to init peer client cache", name, err,
		)
		conn.Close()
		exitChan <- true
	}

	go peering.ClientProbeRPC(rpcctx, probeStream, pcache, relayUUID, 600*time.Second, peerProbeChanel, getRelayIP)

	go peering.ClientSurveyRPC(rpcctx, surveyStream, relayUUID, getRelayIP, dummyDialinLookup)

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
			t.Error("stopping  client", name, ctx.Err())
			exitChan <- true
			return
		case <-probeStream.Context().Done():
			rpccancel()
			conn.Close()
			t.Error("stopping  client", name, ctx.Err())
			exitChan <- true
			return
		case <-surveyStream.Context().Done():
			rpccancel()
			conn.Close()
			t.Error("stopping  client", name, ctx.Err())
			exitChan <- true
			return
		case <-ctx.Done():
			rpccancel()
			conn.Close()
			t.Error("stopping  client", name, ctx.Err())
			exitChan <- true
			return
		}
	}

}

func TestRelayPeerRPC(t *testing.T) {
	var ExitChan = make(chan bool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		relayPeerService, err := NewRelayPeerService()
		if err != nil {
			_log.Fatalw("unable to get create relay peer service")
			panic(err)
		}

		grpcServer, err := grpc.NewSecureServerWithPEM(readPEM("testdata/peersvc.crt"), readPEM("testdata/peersvc.key"), readPEM("testdata/ca.crt"))
		if err != nil {
			_log.Fatalw("cannot grpc secure server failed", "error", err)

		}

		go func() {
			defer grpcServer.GracefulStop()

			<-ctx.Done()
			_log.Infow("peer service stoped due to context done")
		}()

		relayrpc.RegisterRelayPeerServiceServer(grpcServer, relayPeerService)

		l, err := net.Listen("tcp", fmt.Sprintf(":%d", 7001))
		if err != nil {
			_log.Fatalw("failed to listen relay peer service port", "port", 7001, "error", err)
			return
		}

		_log.Infow("started relay rpc service ", "port", 7001)
		if err = grpcServer.Serve(l); err != nil {
			_log.Fatalw("failed to server relay peer service", "error", err)
		}
	}()

	t.Log("started RunRelayPeerRPC")

	relayUUID1 = genUUID()
	relayUUID2 = genUUID()

	peerProbeChanel1 := make(chan string, 256)
	peerProbeChanel2 := make(chan string, 256)

	go startClient(ctx, t, "client1", relayUUID1, ExitChan, getRelayIP1, peerProbeChanel1)
	go startClient(ctx, t, "client2", relayUUID2, ExitChan, getRelayIP2, peerProbeChanel2)

	time.Sleep(5 * time.Second)

	//send a dummy probe from client1
	peerProbeChanel1 <- "dummycluster.relay.rafay.dev"
	fmt.Println("send probe from client 1")

	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ExitChan:
			t.Errorf("got exit chanl")
			exit(cancel)
		case sig := <-signalChan:
			signalHandler(sig, cancel)
		case <-tick.C:
			fmt.Println("success: test time reached. no errors from peerservice or clients. see logs showing \"cache probeRPC response\"")
			exit(cancel)
		}
	}
}
