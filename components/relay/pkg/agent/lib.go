package agent

import (
	"context"
	"time"

	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
)

// RunCDAgentRoutine starts cd agent. run as a go routine
func RunCDAgentRoutine(ctx context.Context, logLevel int) {

	_log := relaylogger.NewLogger(logLevel).WithName("Relay")
	_log.Info("Starting relay agent client")

	utils.GenUUID()
	utils.LogLevel = logLevel
	utils.Mode = utils.CDRELAYAGENT

restart:
	go RunRelayCDAgent(ctx, logLevel)

	for {
		select {
		case <-utils.ExitChan:
			log.Info(
				"got exit",
			)
			time.Sleep(time.Second * 5)
			goto restart
		case <-ctx.Done():
			return
		}
	}
}
