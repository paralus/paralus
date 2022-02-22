package leaderelection

import (
	"time"
)

const (
	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack.
	//
	// A client needs to wait a full LeaseDuration without observing a change to
	// the record before it can attempt to take over. When all clients are
	// shutdown and a new set of clients are started with different names against
	// the same leader record, they must wait the full LeaseDuration before
	// attempting to acquire the lease. Thus LeaseDuration should be as short as
	// possible (within your tolerance for clock skew rate) to avoid a possible
	// long waits in the scenario.
	//
	LeaseDuration = 15 * time.Second
	// RenewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up.
	//
	RenewDeadline = 10 * time.Second

	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	//
	RetryPeriod = 2 * time.Second
)
