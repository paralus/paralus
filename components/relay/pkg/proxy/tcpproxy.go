package proxy

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
	k8proxy "k8s.io/apimachinery/pkg/util/proxy"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// TCPProxy forwards TCP streams.
type TCPProxy struct {
	// logger is the proxy logger.
	logger  *relaylogger.RelayLog
	config  *utils.ProxyConfig
	handler *k8proxy.UpgradeAwareHandler
	//kube-apiserver address
	apiHost string
}

func (tp *TCPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(tp.apiHost)
	if err == nil {
		r.URL.Host = u.Host
		r.URL.Scheme = u.Scheme
		r.Host = u.Host
	}
	r.Header.Set("Host", tp.apiHost)
	utils.UnSetXForwardedRafay(r.Header)
	tp.handler.ServeHTTP(w, r)
}

func (tp *TCPProxy) backendProxy() error {
	socketPath := utils.UNIXAGENTSOCKET + tp.config.ServiceSNI
	syscall.Unlink(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		tp.logger.Error(
			err,
			"failed to listen on unix",
			"socketPath", socketPath,
		)
		return err
	}

	tp.logger.Info(
		"started listening on unix",
		"socketPath", socketPath,
	)

	cfg, err := config.GetConfig()
	if err != nil {
		tp.logger.Error(
			err,
			"failed in NewKubeHandler",
		)
		return nil
	}

	hndlr, err := NewCachedKubeHandler(cfg, utils.DefaultKeepAliveInterval)
	if err != nil {
		tp.logger.Error(
			err,
			"unable to create kube handler for config",
		)
		return err
	}

	tp.handler = hndlr

	go func() {

		s := &http.Server{
			Handler:      http.Handler(tp),
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		}

		if err := s.Serve(l); err != nil {
			tp.logger.Error(
				err,
				"unable to start backend server",
			)
		}
	}()

	tp.handler = hndlr

	return nil
}

// NewKubeCtlTCPProxy creates new direct TCPProxy, everything will be proxied to
// unix socket.
func NewKubeCtlTCPProxy(lg *relaylogger.RelayLog, cfg *utils.ProxyConfig) *TCPProxy {
	tp := &TCPProxy{
		logger: lg,
		config: cfg,
	}

	err := tp.backendProxy()
	if err != nil {
		return nil
	}

	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		ec, err := config.GetConfig()
		if err != nil {
			tp.logger.Error(
				err,
				"failed to get kubeapi host:port",
			)
			return nil
		}
		tp.apiHost = ec.Host
	} else {
		tp.apiHost = "https://" + net.JoinHostPort(host, port)
	}

	return tp
}

// NewTCPProxy creates new direct TCPProxy, everything will be proxied to
// unix socket.
func NewTCPProxy(lg *relaylogger.RelayLog, cfg *utils.ProxyConfig) *TCPProxy {
	tp := &TCPProxy{
		logger: lg,
		config: cfg,
	}

	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		ec, err := config.GetConfig()
		if err != nil {
			tp.logger.Error(
				err,
				"failed to get kubeapi host:port",
			)
			return nil
		}
		tp.apiHost = ec.Host
	} else {
		tp.apiHost = "https://" + net.JoinHostPort(host, port)
	}

	return tp
}

// Proxy is a ProxyFunc.
func (tp *TCPProxy) Proxy(w io.Writer, r io.ReadCloser, msg *utils.ControlMessage, req *http.Request) {
	var target string
	var network string
	var directionTO, directionFrom string

	switch tp.config.Protocol {
	case utils.CDAGENTCORE:
		if tp.config.Upstream != "" {
			target = tp.config.Upstream
		} else {
			target = utils.DefaultTCPUpstream
		}
		network = "tcp"
		directionTO = "tunnel to TCP"
		directionFrom = "TCP to tunnel"
	default:
		target = utils.UNIXAGENTSOCKET + tp.config.ServiceSNI
		network = "unix"
		directionTO = "tunnel to unix"
		directionFrom = "unix to tunnel"
	}

	if target == "" {
		tp.logger.Info(
			"no unix target",
			"ctrlMsg", msg,
		)
		return
	}

	local, err := net.DialTimeout(network, target, utils.DefaultMuxTimeout)
	if err != nil {
		tp.logger.Error(
			err,
			"msg dial failed",
			"target", target,
			"ctrlMsg", msg,
			"err", err,
		)
		return
	}
	defer local.Close()

	done := make(chan struct{})
	go func() {
		utils.Transfer(flushWriter{w, tp.logger, local}, local, tp.logger, directionFrom)
		close(done)
	}()

	utils.Transfer(local, r, tp.logger, directionTO)

proxyDone:
	for {
		select {
		case <-req.Context().Done():
			break proxyDone
		case <-done:
			break proxyDone
		}
	}
	// wait fr 2 sec before closing
	time.Sleep(2 * time.Second)
}

type flushWriter struct {
	w  io.Writer
	lg *relaylogger.RelayLog
	c  net.Conn
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	fw.lg.Debug(
		"data wrote",
		"data", string(p),
	)
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	} else {
		fw.lg.Info(
			"flushWriter http.Flusher not ok",
		)
	}
	// tolerate up to utils.IdleTimeout idletime
	fw.c.SetDeadline(time.Now().Add(utils.IdleTimeout))
	return
}
