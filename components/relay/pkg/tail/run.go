package tail

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/spf13/viper"
)

const (
	auditPathEnv         = "AUDIT_PATH"
	podNameEnv           = "POD_NAME"
	podNamespaceEnv      = "POD_NAMESPACE"
	peerServiceURIEnv    = "RAFAY_RELAY_PEERSERVICE"
	sentryAddrEnv        = "SENTRY_ADDR" // used by relay running inside core
	bootstrapAddrEnv     = "BOOTSTRAP_ADDR"
	relayPeeringTokenEnv = "RELAY_PEERING_TOKEN"
)

var (
	podName         string
	podNamespace    string
	relayPeeringURI string
	sentryAddr      string // used by relay running inside core
	auditPath       string

	// for relays running outside of core
	relayPeeringToken string
	bootstrapAddr     string
)

func setupTail() error {
	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetDefault(podNameEnv, "relay-pod")
	viper.SetDefault(podNamespace, "rafay-system")
	viper.SetDefault(peerServiceURIEnv, "https://peering.sentry.rafay.local:10001")
	viper.SetDefault(sentryAddrEnv, "localhost:10000")

	// for relays running outside of core
	viper.SetDefault(bootstrapAddrEnv, "")
	viper.SetDefault(relayPeeringTokenEnv, "-")
	viper.SetDefault(auditPathEnv, "/tmp")

	viper.BindEnv(podNameEnv)
	viper.BindEnv(podNamespaceEnv)
	viper.BindEnv(relayPeeringTokenEnv)
	viper.BindEnv(sentryAddrEnv)

	// outside core configs
	viper.BindEnv(bootstrapAddrEnv)

	podName = viper.GetString(podNameEnv)
	podNamespace = viper.GetString(podNamespaceEnv)
	sentryAddr = viper.GetString(sentryAddrEnv)
	auditPath = viper.GetString(auditPathEnv)

	relayPeeringURI = viper.GetString(peerServiceURIEnv)

	// outside core relay configurations
	bootstrapAddr = viper.GetString(bootstrapAddrEnv)

	if sentryAddr == "" && bootstrapAddr == "" {
		_log.Infow("missing sentry & bootstrap uri, please set one of the environment variable RAFAY_SENTRY[relay deployed inside core] (or) BOOTSTRAP_ADDR[outside core]")
		return fmt.Errorf("relay server failed in setupserver")
	}

	if bootstrapAddr != "" {
		// outside config found, unset inside boostrap endpoint.
		// both cannot be set, only one mode is possible
		sentryAddr = ""
	}

	if relayPeeringURI == "" {
		_log.Infow("missing relay peer service uri, please set RAFAY_RELAY_PEERSERVICE environment variable")
		return fmt.Errorf("relay server failed in setupserver")
	}

	relayPeeringToken = viper.GetString(relayPeeringTokenEnv)

	_log.Infow(
		"relay server setup values",
		"podName", podName,
		"podNamespace", podNamespace,
		"relayPeeringURI", relayPeeringURI,
		"auditPath", auditPath,
	)

	if sentryAddr != "" {
		_log.Infow("relay inside core", "sentryAddr", sentryAddr)
	} else {
		_log.Infow("outside core relay", "bootStrapURI", bootstrapAddr,
			"relayPeeringToken", relayPeeringToken,
		)
		if relayPeeringToken == "-" {
			_log.Infow("missing env variable", "RELAY_PEERING_TOKEN", relayPeeringToken)
			return fmt.Errorf("relay server failed in setupserver")
		}

	}
	return nil
}

func runTail(ctx context.Context) {
	sap, err := newSentryAuthorizationPool(ctx)
	if err != nil {
		_log.Panicw("unable to create sentry authorization pool", "error", err)
	}

	//_log.Infow("created sentry authorization pool")

	transformer, err := NewTransformer(sap)
	if err != nil {
		_log.Panicw("unable to create transformer", "error", err)
	}

	//_log.Infow("created audit transformer")

	logChan := make(chan LogMsg)

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()

		watchedDir := make(map[string]context.CancelFunc)
		var m sync.Mutex

		// find new dirs in audit path
		for ; true; <-ticker.C {
			//_log.Infow("ticker started")
			// check if context is done
			select {
			case <-ctx.Done():
				_log.Infow("context done")
				for _, c := range watchedDir {
					c()
				}
				break
			default:
			}

			staleDirs, err := findStaleDirs(auditPath)
			if err != nil {
				_log.Infow("unable to find state directories", "error", err)
				continue
			}

		remStaleLoop:
			for _, staleDir := range staleDirs {
				err := os.RemoveAll(staleDir)
				if err != nil {
					_log.Infow("unable to remove stale dir", "error", err)
					continue remStaleLoop
				}
			}

			dirs, err := findLogDirs(auditPath)
			if err != nil {
				_log.Infow("unable to find log dirs", "error", err)
				continue
			}

			//_log.Infow("found log dirs", "dirs", dirs)

			for _, dir := range dirs {
				m.Lock()

				if _, ok := watchedDir[dir]; !ok {
					wCtx, cancel := context.WithCancel(ctx)
					watchedDir[dir] = cancel

					go tailDir(wCtx, dir, logChan)
				}
				m.Unlock()
			}
		}

	}()

transformLoop:
	for {
		select {
		case lm := <-logChan:
			var am AuditMsg
			err := transformer.Transform(&lm, &am)
			if err != nil {
				_log.Infow("unable to transform message", "error", err)
				continue
			}

			err = json.NewEncoder(os.Stdout).Encode(&am)
			if err != nil {
				_log.Infow("unable to encode audit msg", "error", err)
				continue
			}

		case <-ctx.Done():
			break transformLoop
		}
	}

}

// RunRelayTail runs relay tail
func RunRelayTail(ctx context.Context) {
	setupTail()
	runTail(ctx)
}
