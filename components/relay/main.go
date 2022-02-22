// Copyright (C) 2020 Rafay Systems https://rafay.co/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/RafaySystems/rcloud-base/components/relay/pkg/agent"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relay"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/tail"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
)

var (
	log *relaylogger.RelayLog
)

func exit(cancel context.CancelFunc) {
	cancel()
	os.Exit(0)
}

func terminate(cancel context.CancelFunc) {
	cancel()
	os.Exit(1)
}

//os signal handler
func signalHandler(sig os.Signal, cancel context.CancelFunc) {
	if sig == syscall.SIGINT || sig == syscall.SIGKILL || sig == syscall.SIGTERM || sig == syscall.SIGQUIT {
		log.Error(
			nil,
			"Received",
			"signal", sig,
		)
		exit(cancel)
		return
	}

	log.Info(
		"Received",
		"signal", sig,
	)
}

func main() {

	flag.String("mode", "server", "mode of relay, accepeted values server/client")
	flag.Int("log-level", 1, "log leve, accepted values 0-3")
	flag.Bool("profile", false, "enable profiling")
	flag.Int("profile-port", 13000, "profiler http port")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	mode := viper.GetString("mode")
	logLevel := viper.GetInt("log-level")
	profile := viper.GetBool("profile")
	profilePort := viper.GetInt("profile-port")

	utils.GenUUID()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log = relaylogger.NewLogger(logLevel).WithName("Relay")

	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	log.Info(
		"Relay",
		"mode", mode,
		"loglevel", logLevel,
	)

	utils.LogLevel = logLevel

	if profile {
		go func() {
			log.Info("profiling enabled", "profilePort", profilePort)
			err := http.ListenAndServe(fmt.Sprintf(":%d", profilePort), nil)
			if err != nil {
				panic(err)
			}
		}()
	}

restart:
	switch mode {
	case "client":
		log.Info("Starting relay agent client")
		utils.Mode = utils.RELAYAGENT
		go agent.RunRelayKubeCTLAgent(ctx, logLevel)
	case "cdclient":
		log.Info("Starting relay agent client")
		utils.Mode = utils.CDRELAYAGENT
		go agent.RunRelayCDAgent(ctx, logLevel)
	case "tail":
		log.Info("Starting relay log tail")
		go tail.RunRelayTail(ctx)
	case "cdrelay":
		utils.Mode = utils.CDRELAY
		log.Info("Starting cdrelay server")
		go relay.RunCDRelayServer(ctx, logLevel)
	default:
		// default server
		utils.Mode = utils.RELAY
		log.Info("Starting relay server")
		go relay.RunRelayServer(ctx, logLevel)
	}

	for {
		select {
		case <-utils.ExitChan:
			log.Info(
				"got exit",
			)
			time.Sleep(time.Second * 5)
			goto restart
		case <-utils.TerminateChan:
			terminate(cancel)
		case sig := <-signalChan:
			signalHandler(sig, cancel)
		}
	}
}
