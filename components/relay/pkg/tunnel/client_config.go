package tunnel

import (
	"time"
)

// Default backoff configuration.
const (
	DefaultBackoffInterval    = 500 * time.Millisecond
	DefaultBackoffMultiplier  = 1.5
	DefaultBackoffMaxInterval = 20 * time.Second
	DefaultBackoffMaxTime     = 2 * time.Minute
)

// BackoffConfig defines behavior of staggering reconnection retries.
type BackoffConfig struct {
	Interval    time.Duration
	Multiplier  float64
	MaxInterval time.Duration
	MaxTime     time.Duration
}

// Dialout defines the dialout.
type Dialout struct {
	Protocol           string
	Addr               string
	ServiceSNI         string
	RootCA             string
	ClientCRT          string
	ClientKEY          string
	Upstream           string
	UpstreamClientCRT  string
	UpstreamClientKEY  string
	UpstreamRootCA     string
	UpstreamSkipVerify bool
	UpstreamKubeConfig string
	Version            string
}
