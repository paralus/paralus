package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	authv3 "github.com/RafaySystems/rcloud-base/components/common/pkg/auth/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/gateway"
	logv2 "github.com/RafaySystems/rcloud-base/components/common/pkg/log/v2"
	configrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/config"
	pbrpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
)

const (
	rpcPortEnv    = "RPC_PORT"
	apiPortEnv    = "API_PORT"
	debugPortEnv  = "DEBUG_PORT"
	kratosAddrEnv = "KRATOS_ADDR"
	devEnv        = "DEV"
	configAddrENV = "CONFIG_ADDR"
)

var (
	rpcPort             int
	apiPort             int
	debugPort           int
	rpcRelayPeeringPort int
	kratosAddr          string
	dev                 bool
	// ps                  service.PartnerService
	_log       = logv2.GetLogger()
	authPool   authv3.AuthPool
	configPool configrpc.ConfigPool
	configAddr string
)

func setup() {
	viper.SetDefault(rpcPortEnv, 10000)
	viper.SetDefault(apiPortEnv, 11000)
	viper.SetDefault(debugPortEnv, 12000)
	viper.SetDefault(kratosAddr, "localhost:5443")
	viper.SetDefault(devEnv, true)
	viper.SetDefault(configAddrENV, "localhost:7000")

	viper.BindEnv(rpcPortEnv)
	viper.BindEnv(apiPortEnv)
	viper.BindEnv(debugPortEnv)
	viper.BindEnv(kratosAddrEnv)
	viper.BindEnv(devEnv)
	viper.BindEnv(configAddrENV)

	rpcPort = viper.GetInt(rpcPortEnv)
	apiPort = viper.GetInt(apiPortEnv)
	debugPort = viper.GetInt(debugPortEnv)
	kratosAddr = viper.GetString(kratosAddrEnv)
	dev = viper.GetBool(devEnv)
	configAddr = viper.GetString(configAddrENV)

	rpcRelayPeeringPort = rpcPort + 1

	_log.Infow("usrmgmt setup")

	// ps = service.NewPartnerService(db)
}

func run() {

	ctx := signals.SetupSignalHandler()

	var wg sync.WaitGroup

	wg.Add(1)

	go runAPI(&wg, ctx)
	// go runRPC(&wg, ctx)
	// go runDebug(&wg, ctx)

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
		pbrpcv3.RegisterUserHandlerFromEndpoint,
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

func main() {
	setup()
	run()
}
