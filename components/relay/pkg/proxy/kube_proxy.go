package proxy

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"hash"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
	"github.com/dgraph-io/ristretto"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	k8proxy "k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

// TODO convert to JSON
type responder struct{}

func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

type unixCachedRoundTrip struct {
	proxy http.Handler
	t     *http.Transport
}

type relayCachedPeerRoundTrip struct {
	proxy http.Handler
	t     *http.Transport
}

var (
	cacheTTL = time.Minute * 15
	// cache map of unix round tripper
	unixCachedRoundTripper *ristretto.Cache
	// cache map for peer upstream round tripper
	relayPeerRoundTripper *ristretto.Cache
)

// InitUnixCacheRoundTripper initialize the cache
func InitUnixCacheRoundTripper() error {
	var err error
	unixCachedRoundTripper, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // Num keys to track frequency of (10k).
		MaxCost:     1 << 25, // Maximum cost of cache (32MB).
		BufferItems: 64,      // Number of keys per Get buffer.
	})

	return err
}

// InitPeerCacheRoundTripper initialize the cache
func InitPeerCacheRoundTripper() error {
	var err error
	relayPeerRoundTripper, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // Num keys to track frequency of (10k).
		MaxCost:     1 << 25, // Maximum cost of cache (32MB).
		BufferItems: 64,      // Number of keys per Get buffer.
	})

	return err
}

// makeUpgradeTransport creates a transport that explicitly bypasses HTTP2 support
// for proxy connections that must upgrade.
func makeUpgradeTransport(config *rest.Config, keepalive time.Duration) (k8proxy.UpgradeRequestRoundTripper, error) {

	transportConfig, err := config.TransportConfig()
	if err != nil {
		return nil, err
	}
	tlsConfig, err := transport.TLSConfigFor(transportConfig)
	if err != nil {
		return nil, err
	}
	tlsConfig.NextProtos = []string{"http/1.1"}
	rt := utilnet.SetOldTransportDefaults(&http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: keepalive,
		}).DialContext,
		TLSNextProto:    make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		MaxIdleConns:    idleConns,
		MaxConnsPerHost: idleConnsPerHost,
	})

	upgrader, err := transport.HTTPWrappersForConfig(transportConfig, k8proxy.MirrorRequest)
	if err != nil {
		return nil, err
	}

	return k8proxy.NewUpgradeRequestRoundTripper(rt, upgrader), nil
}

// NewCachedKubeHandler is cached kube handler
func NewCachedKubeHandler(cfg *rest.Config, keepalive time.Duration) (*k8proxy.UpgradeAwareHandler, error) {
	host := cfg.Host
	if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}
	target, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	responder := &responder{}

	transport, err := rest.TransportFor(cfg)
	if err != nil {
		return nil, err
	}

	rt, err := NewCachedRoundTripper(transport)
	if err != nil {
		return nil, err
	}

	upgradeTransport, err := makeUpgradeTransport(cfg, keepalive)
	if err != nil {
		return nil, err
	}

	kproxy := k8proxy.NewUpgradeAwareHandler(target, rt, false, false, responder)
	kproxy.UpgradeTransport = upgradeTransport
	kproxy.UseRequestLocation = true

	return kproxy, nil
}

// UnixDialContext unix dial
func UnixDialContext(sockPath, key, username, sni string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {

		conn, err := net.DialTimeout("unix", sockPath, 60*time.Second)
		if err != nil {
			return nil, err
		}

		hdr := &utils.ProxyProtocolMessage{
			DialinKey: key,
			UserName:  username,
			SNI:       sni,
		}

		b, err := json.Marshal(hdr)
		if err != nil {
			return nil, err
		}

		buf := make([]byte, utils.ProxyProtocolSize)
		copy(buf, b)

		_, err = conn.Write(buf)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func makeUnixUpgradeTransport(sockPath, key, username, sni string) (k8proxy.UpgradeRequestRoundTripper, error) {

	rt := utilnet.SetOldTransportDefaults(&http.Transport{
		DialContext: UnixDialContext(sockPath, key, username, sni),
	})

	upgrader, err := transport.HTTPWrappersForConfig(new(transport.Config), k8proxy.MirrorRequest)
	if err != nil {
		return nil, err
	}
	return k8proxy.NewUpgradeRequestRoundTripper(rt, upgrader), nil
}

func getRTCacheKey(skey string) (key uint64) {
	hasher := _hashPool.Get().(hash.Hash64)
	hasher.Reset()
	hasher.Write([]byte(skey))
	key = hasher.Sum64()
	_hashPool.Put(hasher)
	return
}

type unixHandlerOptions struct {
	sockPath string
	key      string
	username string
	sni      string
}

//UnixKubeHandler unix handler
func UnixKubeHandler(sockPath, key, username, sni string) (http.Handler, error) {

	hkey := getRTCacheKey(username + key)
	if val, ok := unixCachedRoundTripper.Get(hkey); ok {
		rt := val.(unixCachedRoundTrip)
		return rt.proxy, nil
	}

	target, err := url.Parse("http://unixpath-relay/")
	if err != nil {
		return nil, err
	}

	responder := &responder{}

	upgradeTransport, err := makeUnixUpgradeTransport(sockPath, key, username, sni)
	if err != nil {
		return nil, err
	}

	t := getUnixTransport(sockPath, key, username, sni)
	proxy := k8proxy.NewUpgradeAwareHandler(target, t, false, false, responder)
	proxy.UpgradeTransport = upgradeTransport
	proxy.UseRequestLocation = true

	rt := unixCachedRoundTrip{
		proxy: proxy,
		t:     t,
	}

	unixCachedRoundTripper.SetWithTTL(hkey, rt, 100, cacheTTL)

	return proxy, nil
}

// Peer Relay Proxy Helpers
func peerDialContext(relayIP string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout("tcp", relayIP, 60*time.Second)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

//makePeerUpgradeTransport ...
func makePeerUpgradeTransport(relayIP string, tlscfg *tls.Config) (k8proxy.UpgradeRequestRoundTripper, error) {
	rt := utilnet.SetOldTransportDefaults(&http.Transport{
		DialContext:         peerDialContext(relayIP),
		TLSClientConfig:     tlscfg,
		MaxIdleConns:        10,
		IdleConnTimeout:     utils.IdleTimeout,
		TLSHandshakeTimeout: 30 * time.Second,
		// Keep the peer upstream non http2
		// Found issues in upgradeaware http2
		// missing http2 preface when doing proxy.
		ForceAttemptHTTP2: false,
	})

	upgrader, err := transport.HTTPWrappersForConfig(new(transport.Config), k8proxy.MirrorRequest)
	if err != nil {
		return nil, err
	}
	return k8proxy.NewUpgradeRequestRoundTripper(rt, upgrader), nil
}

//PeerKubeHandler peer proxying handler
func PeerKubeHandler(tlscfg *tls.Config, relayIP string) (http.Handler, error) {
	hkey := getRTCacheKey(tlscfg.ServerName)
	if val, ok := relayPeerRoundTripper.Get(hkey); ok {
		rt := val.(relayCachedPeerRoundTrip)
		return rt.proxy, nil
	}

	target, err := url.Parse("https://" + tlscfg.ServerName + "/")
	if err != nil {
		return nil, err
	}

	responder := &responder{}

	upgradeTransport, err := makePeerUpgradeTransport(relayIP, tlscfg)
	if err != nil {
		return nil, err
	}

	t := getPeerTransport(tlscfg, relayIP)
	proxy := k8proxy.NewUpgradeAwareHandler(target, t, false, false, responder)
	proxy.UpgradeTransport = upgradeTransport
	proxy.UseRequestLocation = true

	rt := relayCachedPeerRoundTrip{
		proxy: proxy,
		t:     t,
	}

	relayPeerRoundTripper.SetWithTTL(hkey, rt, 100, cacheTTL)

	return proxy, nil
}
