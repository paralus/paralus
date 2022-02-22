package tunnel

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/RafaySystems/rcloud-base/components/relay/pkg/proxy"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
	"github.com/cenkalti/backoff"
	"golang.org/x/net/http2"
)

//ClientConfig ..
type ClientConfig struct {
	//ServiceName name of the service
	ServiceName string
	// ServerAddr specifies  address of the tunnel server.
	ServerAddr string
	//Upstream upstream address
	Upstream string
	//Protocol ..
	Protocol string
	// TLSClientConfig specifies the tls configuration to use with
	// tls.Client.
	TLSClientConfig *tls.Config
	// Backoff specifies backoff policy on server connection retry. If nil
	// when dial fails it will not be retried.
	Backoff Backoff

	//ServiceProxy is Func responsible for transferring data between server and local services.
	ServiceProxy proxy.Func
	// Logger is optional logger. If nil logging is disabled.
	Logger *relaylogger.RelayLog
}

//Client struct
type Client struct {
	sync.Mutex
	conn               net.Conn
	httpServer         *http2.Server
	serverErr          error
	lastDisconnect     time.Time
	config             *ClientConfig
	logger             *relaylogger.RelayLog
	closeChnl          chan bool
	streams            int64
	lastRequest        time.Time
	lastRequestStreams int64
	isScaledConnection bool
}

var (
	clog *relaylogger.RelayLog
	//Clients map, key by ServerName
	//Clients = make(map[string]*Client)
	ScaleClients = make(chan bool, 5)
)

func expBackoff(c BackoffConfig) *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = c.Interval
	b.Multiplier = c.Multiplier
	b.MaxInterval = c.MaxInterval
	b.MaxElapsedTime = c.MaxTime

	return b
}

//loadNewRelayNetwork start the relay agent for a given network
func loadNewRelayNetwork(ctx context.Context, rnc utils.RelayNetworkConfig) error {
	var spxy proxy.Func
	var tlsconf *tls.Config
	var dialout *Dialout
	var err error

	backoff := BackoffConfig{
		Interval:    DefaultBackoffInterval,
		Multiplier:  DefaultBackoffMultiplier,
		MaxInterval: DefaultBackoffMaxInterval,
		MaxTime:     DefaultBackoffMaxTime,
	}

	//proxy handler
	switch rnc.Network.Name {
	case utils.CDAGENTCORE:
		tlsconf, err = ClientTLSConfigFromBytes(rnc.RelayAgentCert, rnc.RelayAgentKey, rnc.RelayAgentCACert, strings.ReplaceAll(rnc.Network.Domain, "*", utils.AgentID))
		if err != nil {
			clog.Error(
				err,
				"failed to create tlsconfig",
				"service name", rnc.Network.Name,
				"addr", rnc.Network.Domain,
			)
			return fmt.Errorf("failed to load client certs")
		}
		dialout = &Dialout{
			Protocol:   rnc.Network.Name,
			Upstream:   rnc.Network.Upstream,
			Addr:       strings.ReplaceAll(rnc.Network.Domain, "*", utils.AgentID),
			ServiceSNI: utils.AgentID,
		}
		spxy = clientProxy(utils.GENTCP, dialout, clog.WithName(rnc.Network.Name))
	default:
		tlsconf, err = ClientTLSConfigFromBytes(rnc.RelayAgentCert, rnc.RelayAgentKey, rnc.RelayAgentCACert, strings.ReplaceAll(rnc.Network.Domain, "*", utils.ClusterID))
		if err != nil {
			clog.Error(
				err,
				"failed to create tlsconfig",
				"service name", rnc.Network.Name,
				"addr", rnc.Network.Domain,
			)
			return fmt.Errorf("failed to load client certs")
		}
		dialout = &Dialout{
			Protocol:   rnc.Network.Name,
			Upstream:   rnc.Network.Upstream,
			Addr:       strings.ReplaceAll(rnc.Network.Domain, "*", utils.ClusterID),
			ServiceSNI: utils.ClusterID,
		}
		spxy = clientProxy(utils.KUBECTL, dialout, clog.WithName(rnc.Network.Name))
	}

	if spxy == nil {
		return fmt.Errorf("failed to load client proxy")
	}

	ccfg := &ClientConfig{
		ServiceName:     rnc.Network.Name,
		ServerAddr:      dialout.Addr,
		Protocol:        dialout.Protocol,
		TLSClientConfig: tlsconf,
		Upstream:        dialout.Upstream,
		ServiceProxy:    spxy,
		Backoff:         expBackoff(backoff),
		Logger:          clog.WithName(rnc.Network.Name),
	}

	for i := 0; i < utils.MaxDials; i++ {
		c := &Client{
			config:             ccfg,
			httpServer:         &http2.Server{},
			logger:             clog.WithName(rnc.Network.Name),
			isScaledConnection: false,
		}

		c.closeChnl = make(chan bool)
		go c.runClient(ctx)
	}

	// scale clients
	// based on rate of new streams or
	// based on total concurrent streams
	go runClientScaling(ctx, ccfg, rnc)

	return nil
}

func runClientScaling(ctx context.Context, ccfg *ClientConfig, rnc utils.RelayNetworkConfig) {
	var scaledClients []*Client
	totalScaledClient := 0

	for {
		select {
		case <-ScaleClients:
			// got signal to scale
			if totalScaledClient < utils.MaxDials*utils.MaxScaleMultiplier {
				c := &Client{
					config:             ccfg,
					httpServer:         &http2.Server{},
					logger:             clog.WithName(rnc.Network.Name),
					isScaledConnection: true,
				}
				c.closeChnl = make(chan bool)
				scaledClients = append(scaledClients, c)
				totalScaledClient++
				go c.runClient(ctx)
				clog.Info(
					"scale client signal",
					" totalScaledClient ", totalScaledClient,
				)
			}
		case <-time.After(5 * time.Minute):
			tempClients := scaledClients[:0]
			for _, c := range scaledClients {
				now := time.Now()
				c.Lock()
				curStreams := atomic.LoadInt64(&c.streams)
				d := now.Sub(c.lastRequest)
				s := int64(d / time.Hour)
				if curStreams <= 0 && s >= int64(utils.HealingInterval) {
					// no request for last 24Hr, close client
					c.closeChnl <- true
					totalScaledClient--
				} else {
					tempClients = append(tempClients, c)
				}
				c.Unlock()
			}
			scaledClients = tempClients
			clog.Info(
				"scaled client timer",
				" totalScaledClient ", totalScaledClient,
			)
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) runClient(ctx context.Context) {
	if err := c.Start(ctx); err != nil {
		if c.isScaledConnection {
			c.logger.Info(
				"closed scaled dialouts after idle time",
				"service name", c.config.ServiceName,
				"addr", c.config.ServerAddr,
			)
		} else {
			c.logger.Error(
				err,
				"failed to start client for dialouts",
				"service name", c.config.ServiceName,
				"addr", c.config.ServerAddr,
			)
		}
	}
}

//Start relay client
func (c *Client) Start(ctx context.Context) error {
	cw := make(chan bool)

	for {
		conn, err := c.connect()
		if err != nil {
			c.logger.Info(
				"connect failed will retry after 2 sec",
			)
			time.Sleep(2 * time.Second)
			continue
		}

		go func() {
			//handles the http/2 tunnel
			c.httpServer.ServeConn(conn, &http2.ServeConnOpts{
				Handler: http.HandlerFunc(c.serveHTTP),
			})

			cw <- true
		}()

		select {
		case <-cw:
			// dialout got closed
			time.Sleep(100 * time.Millisecond)
		case <-c.closeChnl:
			c.conn.Close()
			c.logger.Info(
				"close connection due to idle streams",
			)
			return fmt.Errorf("idle streams")
		case <-ctx.Done():
			c.conn.Close()
			return fmt.Errorf("ctx done")
		}

		c.logger.Info(
			"client disconnected",
		)
		c.Lock()
		now := time.Now()
		err = c.serverErr

		// detect disconnect hiccup
		if err == nil && now.Sub(c.lastDisconnect).Seconds() < 5 {
			err = fmt.Errorf("connection is being cut")
		}

		c.conn = nil
		c.serverErr = nil
		c.lastDisconnect = now
		c.Unlock()

		if err != nil {
			c.logger.Error(
				err,
				"disconnected, will retry after 2 sec",
				"name", c.config.ServiceName,
			)
			time.Sleep(2 * time.Second)
			continue
		}
	}
}

func (c *Client) processDialoutProxy(conn net.Conn, network, addr string) error {
	// send CONNECT addr HTTP/1.1 header
	connHeader := "CONNECT " + addr + " HTTP/1.1\r\nHost: " + addr + "\r\n"
	if utils.DialoutProxyAuth != "" {
		connHeader = connHeader + "Proxy-Authorization:" + utils.DialoutProxyAuth + "\r\n"
	}
	connHeader = connHeader + "Connection: Keep-Alive\r\n\r\n"

	conn.Write([]byte(connHeader))
	tmp := make([]byte, 1)
	checkHeader := ""
	startAppend := false
	respData := ""
	for {
		n, err := conn.Read(tmp)
		if n != 1 || err != nil {
			c.logger.Info(
				"dial failed ",
				"proxy", utils.DialoutProxy,
				"proxy header", connHeader,
				"network", network,
				"checkHeader", checkHeader,
				"err", err,
			)
			return fmt.Errorf("proxy read error")
		}
		respData += string(tmp[0])
		if tmp[0] == '\r' {
			if startAppend {
				checkHeader += string(tmp[0])
			} else {
				checkHeader = "\r"
				startAppend = true
			}
		} else if tmp[0] == '\n' {
			if startAppend {
				if checkHeader == "\r\n\r" {
					// prased the proxy response header
					break
				}
				checkHeader += string(tmp)
			}
		} else {
			startAppend = false
			checkHeader = ""
		}
	}
	c.logger.Info(
		"proxy dialout success",
		"proxy resp", respData,
	)
	return nil
}

//dialout connect
func (c *Client) connect() (net.Conn, error) {
	c.Lock()
	defer c.Unlock()

	if c.conn != nil {
		return nil, fmt.Errorf("already connected")
	}

	conn, err := c.dial()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server %s error %s", c.config.ServerAddr, err)
	}
	c.conn = conn

	return conn, nil
}

//dial
func (c *Client) dial() (net.Conn, error) {
	var (
		network   = "tcp"
		addr      = c.config.ServerAddr
		tlsConfig = c.config.TLSClientConfig
	)

	doDial := func() (conn net.Conn, err error) {

		d := &net.Dialer{
			Timeout: 60 * time.Second,
		}
		if utils.DialoutProxy == "" {
			conn, err = d.Dial(network, addr)
		} else {
			conn, err = d.Dial(network, utils.DialoutProxy)
			if err == nil {
				err = c.processDialoutProxy(conn, network, addr)
			}
		}

		if err == nil {
			err = utils.KeepAlive(conn)
		}
		if err == nil {
			conn = tls.Client(conn, tlsConfig)
		}

		if err == nil {
			err = conn.(*tls.Conn).Handshake()
		}

		if err != nil {
			if conn != nil {
				conn.Close()
				conn = nil
			}

			c.logger.Info(
				"dial failed",
				"network", network,
				"addr", addr,
				"err", err,
			)
		}

		c.logger.Info(
			"action dial out",
			"network", network,
			"addr", addr,
		)

		return
	}

	b := c.config.Backoff
	if b == nil {
		// try once and return
		return doDial()
	}

	for {

		conn, err := doDial()

		// success
		if err == nil {
			b.Reset()
			c.logger.Info(
				"dial success",
			)
			return conn, err
		}

		// failure
		d := b.NextBackOff()
		if d < 0 {
			b.Reset()
			return conn, fmt.Errorf("backoff limit exeded: %s", err)
		}

		// backoff
		c.logger.Info(
			"action backoff",
			"sleep", d,
			"address", c.config.ServerAddr,
		)
		time.Sleep(d)
	}
}

// tunnel serveHTTP handler. Requests form relay lands here.
// request handled based on the action in the ctrl message (header)
// Connect method is used as handshake
// PUT method is used for tunneled requests that carry
// stream data as req.Body
func (c *Client) serveHTTP(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug(
		"invoked handler serveHTTP",
		"method", r.Method,
		"currentStreams", atomic.LoadInt64(&c.streams),
	)

	if r.Method == http.MethodConnect {
		if r.Header.Get(utils.HeaderError) != "" {
			c.handleHandshakeError(w, r)
		} else {
			c.handleHandshake(w, r)
		}
		return
	}

	atomic.AddInt64(&c.streams, 1)
	defer atomic.AddInt64(&c.streams, ^int64(0))

	curStreams := atomic.LoadInt64(&c.streams)
	now := time.Now()

	c.Lock()
	d := now.Sub(c.lastRequest)
	s := int64(d / time.Second)
	if s >= 4 { // check in minimum 4 sec
		if !c.isScaledConnection {
			if curStreams > int64(utils.ScalingStreamsThreshold) {
				// concurrent streams count is high
				ScaleClients <- true
			} else if curStreams > c.lastRequestStreams {
				rate := (curStreams - c.lastRequestStreams) / s
				if rate > int64(utils.ScalingStreamsRateThreshold) {
					// rate of new stream is high
					ScaleClients <- true
				}
			}
		}
		c.lastRequest = now
		c.lastRequestStreams = curStreams
	}
	c.Unlock()

	msg, err := utils.ReadControlMessage(r)
	if err != nil {
		c.logger.Error(
			err,
			"Read Control Message failed",
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.logger.Debug(
		"handle proxy action",
		"ctrlMsg", msg,
		"req", r,
		"currentStreams", atomic.LoadInt64(&c.streams),
	)

	switch msg.Action {
	case utils.ActionProxy:
		c.config.ServiceProxy(w, r.Body, msg, r)
	default:
		c.logger.Info(
			"unknown action",
			"ctrlMsg", msg,
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (c *Client) handleHandshakeError(w http.ResponseWriter, r *http.Request) {
	err := fmt.Errorf(r.Header.Get(utils.HeaderError))

	c.logger.Info(
		"action handshake error",
		"addr", r.RemoteAddr,
		"err", err,
	)

	c.Lock()
	c.serverErr = fmt.Errorf("server error: %s", err)
	c.Unlock()
}

func (c *Client) handleHandshake(w http.ResponseWriter, r *http.Request) {

	c.logger.Info(
		"action", "handshake",
		"addr", r.RemoteAddr,
	)

	w.WriteHeader(http.StatusOK)
	host, _, _ := net.SplitHostPort(c.config.ServerAddr)
	msg := handShakeMsg{
		ServiceName: c.config.ServiceName,
		Protocol:    c.config.Protocol,
		Host:        host,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		c.logger.Info(
			"msg", "handshake failed",
			"err", err,
		)
		return
	}
	w.Write(b)
}

//setups the proxy func handler
func clientProxy(svcName string, d *Dialout, logger *relaylogger.RelayLog) proxy.Func {
	proxyCfg := &utils.ProxyConfig{
		Protocol:           d.Protocol,
		Addr:               d.Addr,
		ServiceSNI:         d.ServiceSNI,
		RootCA:             d.RootCA,
		ClientCRT:          d.ClientCRT,
		ClientKEY:          d.ClientKEY,
		Upstream:           d.Upstream,
		UpstreamClientCRT:  d.UpstreamClientCRT,
		UpstreamClientKEY:  d.UpstreamClientKEY,
		UpstreamRootCA:     d.UpstreamRootCA,
		UpstreamSkipVerify: d.UpstreamSkipVerify,
		UpstreamKubeConfig: d.UpstreamKubeConfig,
		Version:            d.Version,
	}

	switch svcName {
	case utils.KUBECTL:
		if p := proxy.NewKubeCtlTCPProxy(logger.WithName("TCPProxy"), proxyCfg); p != nil {
			return p.Proxy
		}
		logger.Error(
			nil,
			"proxy.NewKubeCtlTCPProxy returned nil",
		)
		return nil
	case utils.GENTCP:
		if p := proxy.NewTCPProxy(logger.WithName("TCPProxy"), proxyCfg); p != nil {
			return p.Proxy
		}
		logger.Error(
			nil,
			"proxy.NewTCPProxy returned nil",
		)
		return nil
	default:
		logger.Error(
			nil,
			"unknown service name",
		)
		return nil
	}
}

//StartClient starts relay clients
func StartClient(ctx context.Context, log *relaylogger.RelayLog, file string, rnc utils.RelayNetworkConfig, exitChan chan<- bool) {
	clog = log.WithName("Client")

	if err := loadNewRelayNetwork(ctx, rnc); err != nil {
		clog.Error(err, "failed to load relay network")
		exitChan <- true
		return
	}

	for {
		select {
		case <-ctx.Done():
			clog.Error(
				ctx.Err(),
				"Stopping client",
			)
			return
		}
	}
}
