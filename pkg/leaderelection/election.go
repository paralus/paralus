package leaderelection

import (
	"context"

	log "github.com/paralus/paralus/pkg/log"
	le "k8s.io/client-go/tools/leaderelection"
	rl "k8s.io/client-go/tools/leaderelection/resourcelock"
)

var (
	_log = log.GetLogger()
)

// Run runs leader election and calls onStarted when runner becomes leader
func Run(lock rl.Interface, onStarted func(stop <-chan struct{}), stop <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_log.Infow("starting leader election", "for", lock.Describe(), "id", lock.Identity())
	elector, err := le.NewLeaderElector(le.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   LeaseDuration,
		RenewDeadline:   RenewDeadline,
		RetryPeriod:     RetryPeriod,
		Callbacks: le.LeaderCallbacks{
			OnStartedLeading: func(_ context.Context) {
				_log.Infow("started leading", "for", lock.Describe(), "id", lock.Identity())
				onStarted(stop)
			},
			OnStoppedLeading: func() {
				_log.Infow("stopped leading", "for", lock.Describe(), "id", lock.Identity())
			},
			OnNewLeader: func(identity string) {
				_log.Infow("new leader", "for", lock.Describe(), "id", identity)
			},
		},
	})

	if err != nil {
		return err
	}

	go elector.Run(ctx)
	_log.Infow("started leader election", "for", lock.Describe(), "id", lock.Identity())

	<-stop

	return nil

}
