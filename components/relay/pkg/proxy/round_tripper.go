package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"hash"
	"net/http"
	"sync"
	"time"

	"github.com/twmb/murmur3"

	clientutil "github.com/RafaySystems/rcloud-base/components/common/pkg/controller/client"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
	"github.com/dgraph-io/ristretto"
	utilnet "k8s.io/apimachinery/pkg/util/net"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = relaylogger.NewLogger(utils.LogLevel).WithName("Proxy")
	// ErrInvalidUser is returned for invalid user
	ErrInvalidUser = errors.New("invalid user")
)

var _hashPool = sync.Pool{
	New: func() interface{} {
		// The Pool's New function should generally only return pointer
		// types, since a pointer can be put into the return interface
		// value without an allocation:
		return murmur3.New64()
	},
}

const (
	defaulTLSHandshakeTimeout = 10 * time.Second
	idleConnsPerHost          = 5
	idleConns                 = 25
	defaultTimeout            = 30 * time.Second
	defaultKeepAlive          = 30 * time.Second
	disableCompression        = false
)

func getUnixTransport(sockPath, key, username, sni string) *http.Transport {
	return utilnet.SetTransportDefaults(&http.Transport{
		TLSHandshakeTimeout: defaulTLSHandshakeTimeout,
		MaxIdleConnsPerHost: idleConnsPerHost,
		DialContext:         UnixDialContext(sockPath, key, username, sni),
		DisableCompression:  disableCompression,
		MaxIdleConns:        idleConns,
	})
}

func getPeerTransport(tlscfg *tls.Config, relayIP string) *http.Transport {
	if tlscfg != nil {
		tlscfg.NextProtos = []string{"http/1.1"}
	}
	return &http.Transport{
		TLSHandshakeTimeout: defaulTLSHandshakeTimeout,
		TLSClientConfig:     tlscfg,
		MaxIdleConns:        idleConns,
		MaxIdleConnsPerHost: idleConnsPerHost,
		DialContext:         peerDialContext(relayIP),
		DisableCompression:  disableCompression,
		TLSNextProto:        make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		//Keep the peer upstream non http2
		ForceAttemptHTTP2: false,
	}
}

// NewCachedRoundTripper returns new cached round tripper
func NewCachedRoundTripper(rt http.RoundTripper) (http.RoundTripper, error) {
	logger = relaylogger.NewLogger(utils.LogLevel).WithName("CachedRoundTripper")
	c, err := clientutil.New()
	if err != nil {
		logger.Error(
			err,
			"failed in clientutil.New",
		)
		return nil, err
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // Num keys to track frequency of (10k).
		MaxCost:     1 << 25, // Maximum cost of cache (32MB).
		BufferItems: 64,      // Number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}

	crt := &cachedRoundTripper{client: c, cache: cache, rt: rt}

	return crt, nil
}

type cachedRoundTripper struct {
	client client.Client
	rt     http.RoundTripper
	cache  *ristretto.Cache
}

func getCacheKey(userName, namespace string) (key uint64) {
	hasher := _hashPool.Get().(hash.Hash64)
	hasher.Reset()
	hasher.Write([]byte(namespace))
	hasher.Write([]byte(userName))
	key = hasher.Sum64()
	_hashPool.Put(hasher)
	return
}

func (rt *cachedRoundTripper) getToken(req *http.Request) (token string, err error) {
	userName := req.Header.Get(utils.HeaderRafayUserName)
	namespace := req.Header.Get(utils.HeaderRafayNamespace)
	clearCache := req.Header.Get(utils.HeaderClearSecret)
	if userName == "" || namespace == "" {
		logger.Error(
			nil,
			"getToken failed",
			"userName", userName,
			"namespace", namespace,
			"req", req,
		)
		return "", ErrInvalidUser
	}

	logger.Debug(
		"getToken for ",
		"user", userName,
		"namespace", namespace,
		"clearCache", clearCache,
	)

	key := getCacheKey(userName, namespace)

	if clearCache == "" {
		if val, ok := rt.cache.Get(key); ok {
			if strVal, ok := val.(string); ok {
				logger.Debug(
					"token found in cache ",
					"user", userName,
					"namespace", namespace,
				)
				token = strVal
				return
			}
		}
	} else {
		// Clear the header
		req.Header.Del(utils.HeaderClearSecret)
	}

	ctx, cancel := context.WithTimeout(req.Context(), time.Second*10)
	defer cancel()

	secret, err := getServiceAccountSecret(ctx, rt.client, userName, namespace)
	if err != nil {
		return "", err
	}

	token = string(secret.Data[tokenKey])
	rt.cache.SetWithTTL(key, token, 100, time.Minute*5)
	return

}

func (rt *cachedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	logger.Debug("RoundTrip", "request", req)
	// cache here
	token, err := rt.getToken(req)
	if err != nil {
		logger.Error(
			err,
			"failed getToken in RoundTrip",
			"req", req,
		)
		return nil, err
	}

	req = utilnet.CloneRequest(req)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	// remove x-rafay heders before sending to kube-apiserver
	req.Header.Del("X-Rafay-Key")
	req.Header.Del("X-Rafay-Namespace")
	req.Header.Del("X-Rafay-User")

	return rt.rt.RoundTrip(req)
}

func (rt *cachedRoundTripper) CancelRequest(req *http.Request) {
	tryCancelRequest(rt.WrappedRoundTripper(), req)
}

func (rt *cachedRoundTripper) WrappedRoundTripper() http.RoundTripper { return rt.rt }

func tryCancelRequest(rt http.RoundTripper, req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	switch rt := rt.(type) {
	case canceler:
		rt.CancelRequest(req)
	case utilnet.RoundTripperWrapper:
		tryCancelRequest(rt.WrappedRoundTripper(), req)
	default:
		logger.Error(nil, "Unable to cancel request")
	}
}
