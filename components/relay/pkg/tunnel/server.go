// Copyright (C) 2020 Rafay Systems https://rafay.co/

package tunnel

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/RafaySystems/rcloud-base/components/relay/pkg/proxy"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/sessions"

	peerclient "github.com/RafaySystems/rcloud-base/components/common/pkg/sentry/peering"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/audit"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
	"github.com/inconshreveable/go-vhost"
	"golang.org/x/net/http2"
)

// Server server definition
type Server struct {
	// Type of the server. Relay means user-facing
	// Dialin means cluster-facing
	Type string

	// Name specifies the service names example kubectl, kubeweb, etc.
	Name string

	// ServerName of the listening server.
	ServerName string

	// Protocol specifies protocol used http(s)
	Protocol string

	// RootCA used to verify TLS client connections
	RootCA []byte

	// ServerCRT used for the server
	ServerCRT []byte

	// ServerKEY used for the server
	ServerKEY []byte

	// DialinServerName specify the dialin server name
	// valid only for relay server types.
	DialinServerName string

	// ProbeSNI specifies sni used to find peers
	//ProbeSNI string

	// DialinPool where dialin connections are parked
	// valid only for dialin server types
	DialinPool *dialinPool

	// httpClient used to forward users connections to dialin
	httpClient *http.Client

	// httpServer used to server user connections
	httpServer *http.Server

	Provisioner *authzProvisioner

	auditPath string
}

//RelayConn connection info
type RelayConn struct {
	// Conn is the network connection
	Conn net.Conn

	// Type of the server. Relay means user-facing
	// Dialin means cluster-facing
	Type string

	// ServerName of the server which accepted the connection
	ServerName string

	// CertSNI derived from client certificate
	CertSNI string

	// PeerID derived from client certificate
	PeerID string

	// server block of this connection
	server *Server

	// dialinCachedKey already stiched to a dialin
	dialinCachedKey string
}

//ServerListen defines a listen object
type ServerListen struct {
	// Addr specifies the listen address
	Addr string
	// Protocol of all servers listening in above address
	Protocol string
	// RootCAs used to verify TLS client connections
	RootCAs []string
	// List of certs used to terminate listening *.format supported
	Certs []utils.SNICertificate
	// ServerList are the servers using above listen address
	ServerList []*Server

	//Mux tls sni muxer
	Mux *vhost.TLSMuxer
}

type kubeError struct {
	Kind       string   `json:"kind"`
	APIVersion string   `json:"apiVersion"`
	Metadata   []string `json:"metadata"`
	Status     string   `json:"status"`
	Message    string   `json:"message"`
	Reason     string   `json:"reason"`
	Code       int      `json:"code"`
}

var (
	//Config server config
	Config *ServerConfig
	//Servers map, key by ServerName
	Servers = make(map[string]*Server)
	//ServerAddrs Servers grouped with listen address
	ServerAddrs = make(map[string][]*Server)
	//ServerListeners list of listen objects
	ServerListeners []*ServerListen

	// connections list of relay net.Conn/tls.Conn ojects
	connections = make(map[string]*RelayConn)
	// connectionsRWMu read write lock for the Connections map
	connectionsRWMu sync.RWMutex

	slog *relaylogger.RelayLog
)

func (srv *Server) disconnected(addr string) {
	s := strings.Split(addr, utils.JoinString)
	if len(s) < 3 {
		slog.Error(
			fmt.Errorf("error from disconnected callback invalid key"),
			addr,
		)
		return
	}
	remoteAddr := s[2]
	if remoteAddr != "" {
		deleteConnections(remoteAddr)
	}
}

// loadRelayServers from configuration
// This creates a server objects and add to
// Servers map with key as server name (*.DNS)
func loadRelayServers(config *ServerConfig) {

	slog.Info("Loading Servers form config")

	// Process config to prepare kubectl relay servers
	for name, relay := range config.Relays {
		provisioner, err := newAuthzProvisioner()
		if err != nil {
			slog.Error(err, "Unable to create provisioned")
			return
		}
		Servers[relay.ServerName] = &Server{
			Type:             utils.RELAY,
			Name:             name,
			ServerName:       relay.ServerName,
			RootCA:           relay.RootCA,
			ServerCRT:        relay.ServerCRT,
			ServerKEY:        relay.ServerKEY,
			Protocol:         relay.Protocol,
			DialinServerName: relay.DialinSfx,
			httpServer:       &http.Server{},
			Provisioner:      provisioner,
			auditPath:        config.AuditPath,
		}

		if len(ServerAddrs[relay.Addr]) > 0 {
			// check protocol
			if ServerAddrs[relay.Addr][0].Protocol == relay.Protocol {
				ServerAddrs[relay.Addr] = append(ServerAddrs[relay.Addr], Servers[relay.ServerName])
			} else {
				// all servers listening on same address should have same
				// protocol
				slog.Info(
					"Protocol missmatch for same addresses in relay",
					"name", name,
					"address", relay.Addr,
					"protocol", relay.Protocol,
					"expected", ServerAddrs[relay.Addr][0].Protocol,
				)
			}
		} else {
			ServerAddrs[relay.Addr] = append(ServerAddrs[relay.Addr], Servers[relay.ServerName])
		}
	}

	// Process config to prepare cd relay servers
	for name, relay := range config.CDRelays {
		Servers[relay.ServerName] = &Server{
			Type:             utils.CDRELAY,
			Name:             name,
			ServerName:       relay.ServerName,
			RootCA:           relay.RootCA,
			ServerCRT:        relay.ServerCRT,
			ServerKEY:        relay.ServerKEY,
			Protocol:         relay.Protocol,
			DialinServerName: relay.DialinSfx,
			auditPath:        config.AuditPath,
		}

		if len(ServerAddrs[relay.Addr]) > 0 {
			// check protocol
			if ServerAddrs[relay.Addr][0].Protocol == relay.Protocol {
				ServerAddrs[relay.Addr] = append(ServerAddrs[relay.Addr], Servers[relay.ServerName])
			} else {
				// all servers listening on same address should have same
				// protocol
				slog.Info(
					"Protocol missmatch for same addresses in relay",
					"name", name,
					"address", relay.Addr,
					"protocol", relay.Protocol,
					"expected", ServerAddrs[relay.Addr][0].Protocol,
				)
			}
		} else {
			ServerAddrs[relay.Addr] = append(ServerAddrs[relay.Addr], Servers[relay.ServerName])
		}
	}

	// Process config to prepare servers
	for name, dialin := range config.Dialins {
		s := &Server{
			Type:       utils.DIALIN,
			Name:       name,
			ServerName: dialin.ServerName,
			RootCA:     dialin.RootCA,
			ServerCRT:  dialin.ServerCRT,
			ServerKEY:  dialin.ServerKEY,
			Protocol:   dialin.Protocol,
		}

		t := &http2.Transport{}
		pool := newDialinPool(t, s.disconnected, slog)
		t.ConnPool = pool
		s.DialinPool = pool
		s.httpClient = &http.Client{
			Transport: t,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		Servers[dialin.ServerName] = s

		if len(ServerAddrs[dialin.Addr]) > 0 {
			// check protocol
			if ServerAddrs[dialin.Addr][0].Protocol == dialin.Protocol {
				ServerAddrs[dialin.Addr] = append(ServerAddrs[dialin.Addr], Servers[dialin.ServerName])
			} else {
				slog.Info(
					"Protocol missmatch for same addresses in dilain",
					"name", name,
					"address", dialin.Addr,
					"protocol", dialin.Protocol,
					"expected", ServerAddrs[dialin.Addr][0].Protocol,
				)
			}
		} else {
			ServerAddrs[dialin.Addr] = append(ServerAddrs[dialin.Addr], Servers[dialin.ServerName])
		}
	}

	slog.Debug(
		"Loaded Servers",
		"Servers", Servers,
	)

	slog.Debug(
		"Loaded ServerAddrs",
		"ServerAddrs", ServerAddrs,
	)

	// Prepare listener list
	for name, srvlst := range ServerAddrs {
		var listenAddr string

		// override server listern address with relayIP if present.
		// This is for running mutliple instance of relay in the same
		// development host.
		if utils.RelayIPFromConfig == "" {
			listenAddr = name
		} else {
			listenAddr = utils.RelayIPFromConfig
		}

		srvlistener := &ServerListen{
			Addr:       listenAddr,
			Protocol:   srvlst[0].Protocol,
			ServerList: srvlst,
		}

		// set the list cert/key paires for this listener
		for _, srv := range srvlst {
			snicert := utils.SNICertificate{
				CertFile: srv.ServerCRT,
				KeyFile:  srv.ServerKEY,
			}
			srvlistener.Certs = append(srvlistener.Certs, snicert)
			// set the rootCA list for this listener
			srvlistener.RootCAs = append(srvlistener.RootCAs, string(srv.RootCA))
		}

		ServerListeners = append(ServerListeners, srvlistener)
	}

	slog.Info(
		"Loaded server ",
		"num-listeners", len(ServerListeners),
	)
}

func (srv *Server) connectRequest(key string, msg *utils.ControlMessage, r io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPut, srv.DialinPool.URL(key), r)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %s", err)
	}
	// write message to header
	utils.WriteToHeader(req.Header, msg)

	return req, nil
}

//dialinLookup check for cluster connection
func dialinCountLookup(sni string) int {
	for _, dsrv := range Servers {
		if dsrv.Type == utils.DIALIN {
			// look dialin pool to find dilain connection key
			cnt, err := dsrv.DialinPool.GetDialinConnectorCount(sni)
			if err == nil {
				return cnt
			}
		}
	}
	return 0
}

//ProcessPeerForwards ...
func (srv *Server) ProcessPeerForwards(w http.ResponseWriter, r *http.Request, lg *relaylogger.RelayLog, relayIP string, certIssue int64) {

	lg.Debug("ProcessPeerForwards:", relayIP, r.TLS.ServerName)
	_, port, err := net.SplitHostPort(relayIP)
	if err != nil {
		port = "443"
	}

	tlscfg, err := ClientTLSConfigFromBytes(srv.ServerCRT, srv.ServerKEY, srv.RootCA, r.TLS.ServerName+":"+port)
	if err != nil {
		lg.Error(
			err,
			"Failed to process peer configs",
		)
		return
	}
	//setsup the upgradeaware peer handler
	peerHandler, err := proxy.PeerKubeHandler(tlscfg, relayIP)
	if err != nil {
		lg.Error(
			err,
			"unable to create peer handler",
		)
		return
	}

	utils.SetXForwardedFor(r.Header, r.RemoteAddr)
	//Update the UUID in the header to detect loops
	utils.SetXRAYUUID(r.Header)

	r.Header.Set("X-Rafay-User-Cert-Issued", strconv.FormatInt(certIssue, 10))

	// set peer security upstream headers
	err = utils.PeerSetHeaderNonce(r.Header)
	if err != nil {
		lg.Error(
			err,
			"unable to set security headers in peer upstream",
		)
		return
	}

	//upgradeaware upstreaming
	peerHandler.ServeHTTP(w, r)
	lg.Debug(
		"done Serving peerHandler.ServeHTTP ",
	)
}

// ProcessCDPeerForwards ...
func (srv *Server) ProcessCDPeerForwards(ctx context.Context, conn net.Conn, lg *relaylogger.RelayLog, relayIP string, state tls.ConnectionState) {
	var peerAddr string
	lg.Debug("ProcessCDPeerForwards:", relayIP, state.ServerName)
	_, port, err := net.SplitHostPort(relayIP)
	if err != nil {
		port = "443"
		peerAddr = relayIP + ":443"
	}

	tlscfg, err := ClientTLSConfigFromBytes(srv.ServerCRT, srv.ServerKEY, srv.RootCA, state.ServerName+":"+port)
	if err != nil {
		lg.Error(
			err,
			"Failed to process peer configs",
		)
		return
	}

	pconn, err := tls.Dial("tcp", peerAddr, tlscfg)
	if err != nil {
		lg.Error(
			err,
			"failed to connect to peer",
			"addr", peerAddr,
		)
		return
	}
	defer pconn.Close()

	utils.Transfer(flushWriter{pconn, lg, pconn}, conn, lg, "client to cdpeer")

processCDPeerForwardDone:
	for {
		select {
		case <-ctx.Done():
			break processCDPeerForwardDone

		}
	}

	// wait 2 sec before closing
	time.Sleep(2 * time.Second)
	return
}

//ProcessRelayRequest process user-facing request
func (srv *Server) ProcessRelayRequest(w http.ResponseWriter, r *http.Request, lg *relaylogger.RelayLog) {
	var (
		rafayUserName     string
		sessionKey        string
		clusterServerName string
		clusterID         string
		dialinSNI         string
		certIssue         int64
		dialinAttempt     int
		ok                bool
		session           *sessions.UserSession
	)

	srvlog := lg.WithName("RelayRequest")
	srvlog.Debug(
		"processing relay request",
	)

	// prepare the dilain lookup key
	if r.TLS.ServerName == srv.ServerName {
		srvlog.Error(
			nil,
			"Wildcard ServerName is expected",
			"ServerName", srv.ServerName,
			"expected SNI", "*."+srv.ServerName,
		)
		errStirng := "ERROR: Unauthenticated access not allowed. Please log in to the portal and download new kubeconfig"
		jsonError(w, errStirng, "invalid cert in kubeconfig", http.StatusUnauthorized)
		return
	}
	srvlog.Debug("ProcessRelayRequest", "r.TLS.ServerName", r.TLS.ServerName, "srv.ServerName", srv.ServerName)

	strs := strings.Split(r.TLS.ServerName, strings.ReplaceAll(srv.ServerName, "*", ""))
	if len(strs) <= 0 {
		srvlog.Error(
			nil,
			"ServeName is not the suffix of  SNI",
			"ServerName", srv.ServerName,
			"SNI", r.TLS.ServerName,
		)
		errStirng := "ERROR: Unauthenticated access not allowed. Please log in to the portal and download new kubeconfig"
		jsonError(w, errStirng, "invalid cert in kubeconfig", http.StatusUnauthorized)
		return
	}

	if len(r.TLS.PeerCertificates) <= 0 {
		srvlog.Error(
			nil,
			"did not find any peer certificates in the request",
		)
		errStirng := "ERROR: Unauthenticated access not allowed. Please log in to the portal and download new kubeconfig"
		jsonError(w, errStirng, "invalid cert in kubeconfig", http.StatusUnauthorized)
		return
	}

	if r.TLS.PeerCertificates[0].Subject.CommonName == "" {
		srvlog.Error(
			nil,
			"cerficate CN is empty",
		)
		errStirng := "ERROR: Unauthenticated access not allowed. Please log in to the portal and download new kubeconfig"
		jsonError(w, errStirng, "invalid cert in kubeconfig", http.StatusUnauthorized)
		return
	}

	issueDate := r.TLS.PeerCertificates[0].NotBefore
	certIssue = issueDate.Unix()
	srvlog.Debug("certificate usse epoch in secs", certIssue)

	sessionKey = ""
	session = nil
	//check the request is from a peer relay
	if r.Header.Get("X-Rafay-XRAY-RELAYUUID") == "" {
		srvlog.Debug(
			"procesing user request common name from cert",
			"CN", r.TLS.PeerCertificates[0].Subject.CommonName,
		)

		//User name is extracted from client certificate CN
		rafayUserName = r.TLS.PeerCertificates[0].Subject.CommonName
		sessionKey = r.TLS.ServerName + ":" + rafayUserName
		session, ok = sessions.GetUserSession(sessionKey)
		if !ok {
			//Create session for the user
			session = &sessions.UserSession{
				Type:            srv.Type,
				ServerName:      srv.ServerName,
				CertSNI:         r.TLS.ServerName,
				DialinCachedKey: "",
				ErrorFlag:       false,
			}
			sessions.AddUserSession(session, sessionKey)
			srvlog.Info(
				"Created new session",
				"key", sessionKey,
			)
		} else {
			srvlog.Debug(
				"Found existing session",
				"key", sessionKey,
				"dialin sticky key", session.DialinCachedKey,
			)
		}

		clusterID = strs[0]
		dialinSNI = strings.ReplaceAll(srv.DialinServerName, "*", strs[0])
		dialinAttempt = 0
	} else {
		var err error
		// to protect from a bad actor injecting rogue headers we need
		// to validate the request tls parametes to make sure the client cert
		// has expcted values.

		// 1. using a shared 256 key aes encryption blob added with a nonce
		// validate the decrypted text using shared key produces expected result
		// to make sure no tampering in headers
		/*
			if utils.CheckRelayLoops(r.Header) {
				var allIds string

				if uuidHdr, ok := r.Header["X-Rafay-XRAY-RELAYUUID"]; ok {
					allIds = strings.Join(uuidHdr, ", ")
				}
				srvlog.Error(
					fmt.Errorf("LOOP detected in peerforwards"),
					"failed peer loop detection check RelayUUID", utils.RelayUUID,
					"header X-Rafay-XRAY-RELAYUUID allIds", allIds,
				)
			}
		*/
		if !utils.CheckPeerHeaders(r.Header) {
			srvlog.Error(
				fmt.Errorf("failed to validate the peer upstream security header"),
				"failed peer upstream due to secuity reasons",
			)
			errStirng := "ERROR: Unauthenticated access not allowed . Please log in to the portal and download new kubeconfig"
			jsonError(w, errStirng, "failed to validate request in peer proxy", http.StatusUnauthorized)
			return
		}

		rafayUserName = r.Header.Get("X-Rafay-User")
		clusterServerName = r.Header.Get("X-Rafay-Cluster-ServerName")
		clusterID = r.Header.Get("X-Rafay-Cluster-ID")
		issDateSecStr := r.Header.Get("X-Rafay-User-Cert-Issued")
		if issDateSecStr == "" {
			srvlog.Error(
				fmt.Errorf("peer did not send a valid user cert issue date header"),
				"failed peer upstream due to missing cert issuedate header",
			)
		}
		certIssue, err = strconv.ParseInt(issDateSecStr, 10, 64)
		if err != nil {
			srvlog.Error(
				fmt.Errorf("peer did not send a valid user cert issue date header"),
				"failed peer upstream due to missing cert issuedate header",
			)
		}

		srvlog.Debug(
			"forwarded request from peer",
			"rafayUserName", rafayUserName,
			"clusterServerName", clusterServerName,
			"clusterID", clusterID,
			"certIssue:", certIssue,
		)

		if rafayUserName == "" || clusterServerName == "" || clusterID == "" {
			srvlog.Error(
				nil,
				"Did not find required headers",
				"X-Rafay-User", rafayUserName,
				"X-Rafay-Cluster-Name", clusterServerName,
				"X-Rafay-Cluster-ID", clusterID,
			)
			errStirng := "ERROR: Unauthenticated access not allowed (kubeconfig invalid cert). Please log in to the portal and download new kubeconfig"
			jsonError(w, errStirng, "failed to find user/cluster details", http.StatusUnauthorized)
			return
		}

		sessionKey = clusterServerName + ":" + rafayUserName
		session, ok = sessions.GetUserSession(sessionKey)
		if !ok {
			//Create session for the user
			session = &sessions.UserSession{
				Type:            srv.Type,
				ServerName:      srv.ServerName,
				CertSNI:         clusterServerName,
				DialinCachedKey: "",
				ErrorFlag:       false,
			}
			sessions.AddUserSession(session, sessionKey)
			srvlog.Info(
				"Created new session",
				"key", sessionKey,
			)
		} else {
			srvlog.Debug(
				"Found existing session",
				"key", sessionKey,
				"dialin sticky key", session.DialinCachedKey,
			)
		}

		//prifixName is the clusterid/uuid for the endpoint
		dialinSNI = strings.ReplaceAll(srv.DialinServerName, "*", clusterID)
		dialinAttempt = 0
	}
	r.Header.Set("X-Rafay-Audit", "yes")

retryDialin:
	// get the dialin server instance
	if dsrv, ok := Servers[srv.DialinServerName]; ok {
		if session.DialinCachedKey == "" {
			// look dialin pool to find dilain connection key
			key, err := dsrv.DialinPool.GetDialinConnectorKey(dialinSNI)
			if err != nil {
				srvlog.Info(
					"No dialins",
					"error", err,
					"dialinAttempt", dialinAttempt,
				)
				// Lookup peerCache to fetch cluster connection
				// On cache-miss send the probe, wait and retry 5 times
				relayIP, found := peerclient.GetPeerCache(utils.PeerCache, dialinSNI)
				if found {
					// forward the client request to peer upstream
					r.Header.Set("X-Rafay-User", rafayUserName)
					r.Header.Set("X-Rafay-Cluster-ServerName", r.TLS.ServerName)
					r.Header.Set("X-Rafay-Cluster-ID", clusterID)
					r.Header.Del("X-Rafay-Audit")
					srv.ProcessPeerForwards(w, r, lg, relayIP, certIssue)
					return
				} else {
					if dialinAttempt == 0 {
						go SendPeerProbe(PeerProbeChanel, dialinSNI)
					}

					dialinAttempt++
					if dialinAttempt > 7 {
						errStirng := "ERROR: failed to forward request to cluster. Please retry"
						jsonError(w, errStirng, "failed to find connection", http.StatusInternalServerError)
						return
					}
					// wait for total ~20 sec to learn from core
					// else drop the connection
					time.Sleep(3 * time.Second)
					goto retryDialin
				}
			} else {
				//cache it
				session.DialinCachedKey = key
			}
		} else {
			//verify cache is valid
			if !dsrv.DialinPool.CheckDialinKeyExist(session.DialinCachedKey) {
				session.DialinCachedKey = ""
				goto retryDialin
			}
		}

		srvlog.Debug(
			"dialin connection lookup",
			"key", session.DialinCachedKey,
			"sessionKey", sessionKey,
		)

		//unix socket for stiching user request to dialin
		socketPath := utils.UNIXSOCKET + srv.DialinServerName

		if session.UserName == "" {
			userName, roleName, isRead, isOrgAdmin, enforceOrgAdminSecret, err := srv.Provisioner.ProvisionAuthzForUser(socketPath, rafayUserName, r.TLS.ServerName, session.DialinCachedKey, session.ErrorFlag, false, certIssue)
			if err != nil {
				var errStirng string
				srvlog.Error(
					err,
					"unable to provision authz for user",
				)
				if userName == "" {
					errStirng = "ERROR: Unauthenticated access not allowed. Please log in to the portal via browser, or set up API key, for access via the secure kubectl proxy. Error:" + err.Error()
				} else {
					errStirng = "ERROR: Connection timed-out. Unable to provision cluster RBAC. Please retry."
				}
				jsonError(w, errStirng, "unable to proxy kubectl service", http.StatusUnauthorized)
				session.DialinCachedKey = ""
				sessions.DeleteUserSession(sessionKey)
				return
			}
			session.UserName = userName
			session.IsReadrole = isRead
			session.RoleName = roleName
			session.IsOrgAdmin = isOrgAdmin
			session.EnforceOrgAdminOnlySecret = enforceOrgAdminSecret
		} else {
			go func() {
				_, roleName, isRead, isOrgAdmin, enforceOrgAdminSecret, err := srv.Provisioner.ProvisionAuthzForUser(socketPath, rafayUserName, r.TLS.ServerName, session.DialinCachedKey, session.ErrorFlag, true, certIssue)
				if err != nil {
					srvlog.Error(
						err,
						"unable to provision authz for user",
					)
					session.UserName = ""
				}
				session.IsReadrole = isRead
				session.RoleName = roleName
				session.IsOrgAdmin = isOrgAdmin
				session.EnforceOrgAdminOnlySecret = enforceOrgAdminSecret
			}()
		}

		//check readrole vs pod exec
		if session.IsReadrole {
			if sessions.GetRoleCheck(r.Method, r.URL.Path) {
				errStirng := fmt.Sprintf("ERROR: request forbidden for the role %s.", session.RoleName)
				jsonError(w, errStirng, "unable to authorize user request", http.StatusUnauthorized)
				return
			}
		}

		if session.EnforceOrgAdminOnlySecret && !session.IsOrgAdmin {
			if sessions.GetSecretRoleCheck(r.Method, r.URL.Path) {
				errStirng := "ERROR: request for secret resource is forbidden. Contact your organization admin"
				jsonError(w, errStirng, "unable to authorize user request to secret", http.StatusUnauthorized)
				return
			}
		}

		//setsup the upgradeaware unix handler
		unixHandler, err := proxy.UnixKubeHandler(socketPath, session.DialinCachedKey, rafayUserName, r.TLS.ServerName)
		if err != nil {
			srvlog.Error(
				err,
				"unable to create unix handler",
			)
			errStirng := "ERROR: failed to forward request to cluster. Please retry"
			jsonError(w, errStirng, "unable to create forward handler", http.StatusInternalServerError)
			return
		}

		//These are the headers used by relay-agent to fetch
		//user information.
		r.Header.Set("X-Rafay-User", session.UserName)
		r.Header.Set("X-Rafay-Key", session.DialinCachedKey)
		r.Header.Set("X-Rafay-Namespace", "rafay-system")
		r.Header.Set("X-Rafay-Sessionkey", sessionKey)
		utils.SetXForwardedFor(r.Header, r.RemoteAddr)
		utils.SetXRAYUUID(r.Header)
		if session.ErrorFlag {
			r.Header.Set(utils.HeaderClearSecret, "true")
		}

		//upgradeaware uinix upstreaming
		unixHandler.ServeHTTP(w, r)
		srvlog.Debug(
			"done Serving unixHandler.ServeHTTP ",
		)
		if session.ErrorFlag {
			session.DialinCachedKey = ""
			sessions.DeleteUserSession(sessionKey)
		}
	}
}

func jsonError(w http.ResponseWriter, message, reason string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	resp := &kubeError{Kind: "Status", APIVersion: "v1", Metadata: nil, Status: "Failure", Message: message, Reason: reason, Code: code}
	e, err := json.Marshal(resp)
	if err == nil {
		fmt.Fprintln(w, string(e))
		return
	}
	fmt.Fprintln(w, message+" "+reason)
}

//ServeHTTP requests from userfacing connection
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	servelog := slog.WithName("RelayServeHTTP")
	servelog.Debug(
		"start serving user http",
	)

	if len(r.TLS.PeerCertificates) <= 0 {
		servelog.Error(
			nil,
			"no client certificate in request",
		)
		errStirng := "ERROR: Unauthenticated access not allowed. Please log in to the portal and download new kubeconfig"
		jsonError(w, errStirng, "invalid cert in kubeconfig", http.StatusUnauthorized)
		return
	}

	//client certificate
	servelog.Debug(
		"describe client certificate",
		"ServerName", r.TLS.ServerName,
		"Peer CN", r.TLS.PeerCertificates[0].Subject.CommonName,
		"Peer Serial", r.TLS.PeerCertificates[0].Subject.SerialNumber,
		"NegotiatedProtocol", r.TLS.NegotiatedProtocol,
		"NegotiatedProtocolIsMutual", r.TLS.NegotiatedProtocolIsMutual,
	)

	switch srv.Type {
	case utils.RELAY:
		srv.ProcessRelayRequest(w, r, servelog)
	default:
		servelog.Error(
			nil,
			"unexpected server type",
			"ServerName", srv.ServerName,
			"type", srv.Type,
		)
	}

}

//AddToDialinPool add connection to dialin pool of the server
func (srv *Server) AddToDialinPool(rconn *RelayConn, remoteAddr string) (string, error) {
	slog.Info(
		"AddToDialinPool",
		"identifier", rconn.PeerID,
		"ServerName", rconn.ServerName,
		"CertSNI", rconn.CertSNI,
	)

	key, err := srv.DialinPool.AddConn(rconn.Conn, rconn.PeerID, rconn.CertSNI, remoteAddr)
	if err != nil {
		slog.Error(
			err,
			"adding connection failed",
			"server", srv.ServerName,
		)
		return "", err
	}

	slog.Info(
		"Added conn to pool of",
		"server", srv.ServerName,
	)

	return key, nil
}

func getSNIMuxDebugString(tmp []byte) string {
	dstr := ""
	for i := range tmp {
		if i >= len(tmp) {
			break
		}
		b := tmp[i]
		if b < 32 || b > 126 {
			dstr = dstr + " . "
		} else {
			dstr = dstr + " " + string(b)
		}
	}
	return dstr
}

//StartHTTPSListen start TLS listen on address
//Both user & dialin endpoint listen on 443
//Based on SNI traffic is routed/muxed to appropriate handler
func (sl *ServerListen) StartHTTPSListen(ctx context.Context) {
	slistenlog := slog.WithName("Listener")
	slistenlog.Info(
		"starting listener",
		"addr", sl.Addr,
		"protocol", sl.Protocol,
		"num-servers", len(sl.ServerList),
	)

	l, err := net.Listen("tcp", sl.Addr)
	if err != nil {
		slistenlog.Error(
			err,
			"failed to listen",
			"addr", sl.Addr,
		)
		return
	}

	//close the listener when ctx is done
	go func() {
		defer l.Close()

		<-ctx.Done()
		slistenlog.Info(
			"stopping listener",
			"address", sl.Addr,
		)
	}()

	sl.Mux, err = vhost.NewTLSMuxer(l, utils.DefaultMuxTimeout)
	if err != nil {
		slistenlog.Error(
			err,
			"failed in NewTLSMuxer",
		)
		return
	}

	for name, srv := range Servers {
		slistenlog.Info(
			"start mux listen on ",
			"sni", name,
		)
		snil, err := sl.Mux.Listen(name)
		if err != nil {
			slistenlog.Error(
				err,
				"failed in sl.Mux.Listen",
				"sni", name,
			)
			return
		}
		//listen routine for sni (name)
		go srv.sniListen(ctx, name, snil, sl)

		if srv.Type == utils.DIALIN && utils.Mode != utils.CDRELAY {
			//start the unix socket listen for each
			//dialin server block
			go srv.startUnixListen(ctx, slistenlog)
		}
	}

	// custom error handler for the mux
	go func() {
		for {
			conn, err := sl.Mux.NextError()
			vhostName := ""
			tlsConn, ok := conn.(*vhost.TLSConn)
			dbgStr := ""
			if conn != nil {
				tmp := make([]byte, 256)
				n, _ := conn.Read(tmp)
				if n > 0 {
					dbgStr = getSNIMuxDebugString(tmp)
				}
			}

			if ok {
				vhostName = tlsConn.Host()
				slistenlog.Error(
					nil,
					"vhostName", vhostName,
					"raddr", conn.RemoteAddr(),
					"laddr", conn.LocalAddr(),
					"peekbuf", dbgStr,
				)
			} else {
				slistenlog.Error(
					nil,
					"error in getting tlsConn",
					"raddr", conn.RemoteAddr(),
					"laddr", conn.LocalAddr(),
					"peekbuf", dbgStr,
				)
			}

			switch err.(type) {
			case vhost.BadRequest:
				slistenlog.Info(
					"got a bad request!",
					"addr", conn.RemoteAddr(),
					"tlsConn", tlsConn,
					"err", err,
				)
			case vhost.NotFound:
				slistenlog.Error(
					err,
					"got a connection for an unknown vhost",
					"addr", vhostName,
					"tlsConn", tlsConn,
				)
			case vhost.Closed:
				slistenlog.Error(
					err,
					"vhost closed conn",
					"addr", vhostName,
					"tlsConn", tlsConn,
				)
			}

			if conn != nil {
				slistenlog.Debug(
					"closing connection",
					"raddr", conn.RemoteAddr(),
					"laddr", conn.LocalAddr(),
				)
				conn.Close()
			}
		}
	}()
}

func (srv *Server) proxyUnixConnection(ctx context.Context, conn net.Conn, lg *relaylogger.RelayLog) {
	var pmsg utils.ProxyProtocolMessage

	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(utils.DefaultTimeout))
	//first utils.ProxyProtocolSize bytes reserved for proxy protocol like buffer
	buf := make([]byte, utils.ProxyProtocolSize)
	n, err := io.ReadAtLeast(conn, buf, utils.ProxyProtocolSize)
	if err != nil {
		lg.Error(
			err,
			"failed in unix listen to read",
			"size", utils.ProxyProtocolSize,
			"srv", srv.ServerName,
		)
	}

	lg.Debug(
		"unix listen read proxy protocol info",
		"msglen", n,
		"message", string(buf),
	)

	if n == utils.ProxyProtocolSize {
		n = bytes.IndexByte(buf, 0)
		err = json.Unmarshal(buf[:n], &pmsg)
		if err != nil {
			lg.Error(
				err,
				"failed to parse json",
				"msg", string(buf[:n]),
			)
			return
		}
	} else {
		return
	}

	if !srv.DialinPool.CheckDialinKeyExist(pmsg.DialinKey) {
		lg.Error(
			nil,
			"did not find key in the dilain pool",
			"key", pmsg.DialinKey,
		)
		return
	}

	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()

	msg := &utils.ControlMessage{
		Action:           utils.ActionProxy,
		ForwardedHost:    pmsg.SNI,
		ForwardedService: srv.Name,
		RafayUserName:    pmsg.UserName, // for testing only
		RafayNamespace:   "default",     // for testing only
		RafayScope:       "default",     // for testing only
		RafayAllow:       "true",        // for testing only
	}

	req, err := srv.connectRequest(pmsg.DialinKey, msg, pr)
	if err != nil {
		lg.Error(
			err,
			"failed in srv.connectRequest",
		)
		return
	}

	nctx, cancel := context.WithCancel(ctx)
	req = req.WithContext(nctx)

	done := make(chan struct{})
	go func() {
		utils.Transfer(flushWriter{pw, lg, conn}, conn, lg, "unix to tunnel")
		cancel()
		close(done)
	}()

	resp, err := srv.httpClient.Do(req)
	if err != nil {
		lg.Error(
			err,
			"io error:",
		)
		return
	}
	defer resp.Body.Close()

	utils.Transfer(flushWriter{conn, lg, conn}, resp.Body, lg, "tunnel to unix")

proxyUnixConnectionDone:
	for {
		select {
		case <-nctx.Done():
			break proxyUnixConnectionDone
		case <-ctx.Done():
			break proxyUnixConnectionDone
		case <-done:
			break proxyUnixConnectionDone
		case <-time.After(300 * time.Second):
			lg.Error(
				fmt.Errorf("proxyUnixConnection waited 5 min for idle client to close"),
				"break the loop to avoid memory exhaustion",
			)
			break proxyUnixConnectionDone
		}
	}

	// wait 2 sec before closing
	time.Sleep(2 * time.Second)
}

func (srv *Server) proxyCDDilainConnection(ctx context.Context, conn net.Conn, dialinKey, rafayUserName, sni string, lg *relaylogger.RelayLog) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(utils.DefaultTimeout))

	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()

	msg := &utils.ControlMessage{
		Action:           utils.ActionProxy,
		ForwardedHost:    sni,
		ForwardedService: srv.ServerName,
		RafayUserName:    rafayUserName, // for testing only
		RafayNamespace:   "default",     // for testing only
		RafayScope:       "default",     // for testing only
		RafayAllow:       "true",        // for testing only
	}

	req, err := srv.connectRequest(dialinKey, msg, pr)
	if err != nil {
		lg.Error(
			err,
			"failed in srv.connectRequest",
		)
		return
	}

	nctx, cancel := context.WithCancel(ctx)
	req = req.WithContext(nctx)

	done := make(chan struct{})
	go func() {
		utils.Transfer(flushWriter{pw, lg, conn}, conn, lg, "client to tunnel")
		cancel()
		close(done)
	}()

	resp, err := srv.httpClient.Do(req)
	if err != nil {
		lg.Error(
			err,
			"io error:",
		)
		return
	}
	defer resp.Body.Close()

	utils.Transfer(flushWriter{conn, lg, conn}, resp.Body, lg, "tunnel to client")

proxyCDDilainConnectionDone:
	for {
		select {
		case <-nctx.Done():
			break proxyCDDilainConnectionDone
		case <-ctx.Done():
			break proxyCDDilainConnectionDone
		case <-done:
			break proxyCDDilainConnectionDone
		case <-time.After(300 * time.Second):
			lg.Error(
				fmt.Errorf("proxyCDDilainConnection waited 5 min for idle client to close"),
				"break the loop to avoid memory exhaustion",
			)
			break proxyCDDilainConnectionDone
		}
	}

	// wait 2 sec before closing
	time.Sleep(2 * time.Second)
}

//start the unix socket that listens for connections to
//stich to a dialin. User request are handled here
func (srv *Server) startUnixListen(ctx context.Context, lg *relaylogger.RelayLog) {
	//start listen the dialin unix pipe
	socketpath := utils.UNIXSOCKET + srv.ServerName
	syscall.Unlink(socketpath)
	ul, err := net.Listen("unix", socketpath)
	if err != nil {
		lg.Error(
			err,
			"couldn't listen to",
			"socketpath", socketpath,
		)
	}

	//close the listener when ctx is done
	go func() {
		defer ul.Close()

		<-ctx.Done()
		lg.Info(
			"stopping listener",
			"address", socketpath,
		)
	}()

	lg.Info(
		"started unix listen",
		"socketpath", socketpath,
	)

	uctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Wait for a connection and accept it
	for {
		conn, err := ul.Accept()

		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") ||
				strings.Contains(err.Error(), "Listener closed") {
				lg.Error(
					err,
					"unix listener closed",
					"srv", srv.ServerName,
				)
				return
			}
			lg.Error(
				err,
				"Accept failed ",
				"srv", srv.ServerName,
			)
			continue
		}
		go srv.proxyUnixConnection(uctx, conn, lg)
	}
}

//sni based muxing (virtual hosting) handler that get called based on SNI
func (srv *Server) sniListen(ctx context.Context, sni string, l net.Listener, sl *ServerListen) {
	var tlsconfig *tls.Config
	var err error

	snilistenlog := slog.WithName("SNIListen")

	// prepare TLS configuration for the server
	if srv.Type == utils.RELAY {
		tlsconfig, err = ServerTLSConfigFromBytes(sl.Certs, sl.RootCAs, "http/1.1")
	} else if srv.Type == utils.CDRELAY {
		tlsconfig, err = ServerTLSConfigFromBytes(sl.Certs, sl.RootCAs)
	} else if srv.Type == utils.DIALIN {
		tlsconfig, err = ServerTLSConfigFromBytes(sl.Certs, sl.RootCAs)
	} else {
		snilistenlog.Error(nil, "unknown server type ", srv.Type, "addr", sl.Addr)
		return
	}

	if err != nil {
		snilistenlog.Error(
			err,
			"failed to create tlsconfig",
			"addr", sl.Addr,
		)
		return
	}

	if srv.Type == utils.RELAY {
		//User facing server
		srv.httpServer.TLSConfig = tlsconfig
		srv.httpServer.Handler = http.HandlerFunc(srv.ServeHTTP)

		snilistenlog.Debug(
			"starting srv.httpServer.ServeTLS",
			"sni", sni,
			"name", srv.ServerName,
			"addr", sl.Addr,
		)

		srv.httpServer.Handler = audit.WrapWithAudit(
			srv.httpServer.Handler,
			audit.WithBasePath(srv.auditPath),
			audit.WithMaxSizeMB(1),
			audit.WithMaxBackups(20),
			audit.WithMaxAgeDays(10),
		)

		//start TLS handshake
		srv.httpServer.ServeTLS(l, "", "")
	}

	if srv.Type == utils.CDRELAY {
		for {
			// accept connections
			conn, err := l.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					snilistenlog.Error(
						err,
						"control connection listener closed",
						"addr", sl.Addr,
					)
					return
				}

				snilistenlog.Error(
					err,
					"accept of connection failed",
					"addr", sl.Addr,
				)
				continue
			}

			// handler for the accepted CD connections
			go srv.handleCDRelayConnection(ctx, conn, tlsconfig)
		}
	}

	if srv.Type == utils.DIALIN {
		for {
			// accept connections
			conn, err := l.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					snilistenlog.Error(
						err,
						"control connection listener closed",
						"addr", sl.Addr,
					)
					return
				}

				snilistenlog.Error(
					err,
					"accept of connection failed",
					"addr", sl.Addr,
				)
				continue
			}

			// handler for the accepted connection
			go srv.handleDialinConnection(conn, tlsconfig)
		}
	}

	snilistenlog.Error(
		nil,
		"unknown service type ",
		"type", srv.Type,
	)
}

func (srv *Server) sendDialinHandshake(key string, lg *relaylogger.RelayLog) error {
	var hMsg handShakeMsg

	sndlog := lg.WithName("sendDialinHandshake")

	sndlog.Debug(
		"send dialin handhshake",
		"key", key,
		"httpClient.Transport", srv.httpClient.Transport,
		"pool", srv.DialinPool,
	)

	req, err := http.NewRequest(http.MethodConnect, srv.DialinPool.URL(key), nil)
	if err != nil {
		sndlog.Error(
			err,
			"handshake request creation failed",
		)
		return err
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), utils.DefaultTimeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	resp, err := srv.httpClient.Do(req)
	if err != nil {
		sndlog.Error(
			err,
			"handshake failed",
		)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Status %s", resp.Status)
		sndlog.Error(
			err,
			"dialin handshake response is not 200 OK",
			"resp.StatusCode", resp.StatusCode,
		)
		return err
	}

	if resp.ContentLength == 0 {
		err = fmt.Errorf("Tunnels Content-Legth: 0")
		sndlog.Error(
			err,
			"dialin handshake response has 0 content length",
		)
		return err
	}

	if err = json.NewDecoder(&io.LimitedReader{R: resp.Body, N: 126976}).Decode(&hMsg); err != nil {
		sndlog.Error(
			err,
			"dialin handshake failed to parse json",
		)
		return err
	}

	sndlog.Info(
		"recieved handshake message",
		"hMsg", hMsg,
	)

	return nil
}

func (srv *Server) processCDRelayTLSState(ctx context.Context, conn net.Conn, lg *relaylogger.RelayLog, tlsconfig *tls.Config) (*Server, error) {
	var (
		rafayUserName string
		sessionKey    string
		dialinSNI     string
		dialinAttempt int
		ok            bool
		session       *sessions.UserSession
	)

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		slog.Error(
			nil,
			"invalid connection type",
			"err", fmt.Errorf("expected TLS conn, got %T", conn),
		)
		return nil, fmt.Errorf("expected TLS conn, got %T", conn)
	}

	// perform TLS handshake
	if err := tlsConn.Handshake(); err != nil {
		slog.Error(
			err,
			"failed to compete TLS handshake",
		)
		return nil, fmt.Errorf("ailed to compete TLS handshake")
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		slog.Error(
			err,
			"failed to SetDeadline",
		)
		return nil, fmt.Errorf("failed to SetDeadline")
	}

	state := tlsConn.ConnectionState()

	_, err := GetRemoteCertID(tlsConn)
	if err != nil {
		slog.Error(
			err,
			"error in getting client cert id",
		)
		return nil, fmt.Errorf("error in getting client cert id")
	}

	// Handshake completed
	// prepare the dilain lookup key
	if state.ServerName == srv.ServerName {
		slog.Error(
			nil,
			"Wildcard ServerName is expected",
			"ServerName", srv.ServerName,
			"expected SNI", "*."+srv.ServerName,
		)
		return nil, fmt.Errorf("ERROR: Unauthenticated access not allowed")
	}

	slog.Debug("processCDRelayTLSState", "ServerName", state.ServerName, "srv.ServerName", srv.ServerName)

	strs := strings.Split(state.ServerName, strings.ReplaceAll(srv.ServerName, "*", ""))
	if len(strs) <= 0 {
		slog.Error(
			nil,
			"ServeName is not the suffix of  SNI",
			"ServerName", srv.ServerName,
			"SNI", state.ServerName,
		)
		return nil, fmt.Errorf("ERROR: Unauthenticated access not allowed")
	}

	if len(state.PeerCertificates) <= 0 {
		slog.Error(
			nil,
			"did not find any peer certificates in the request",
		)
		return nil, fmt.Errorf("ERROR: Unauthenticated access not allowed")
	}

	if state.PeerCertificates[0].Subject.CommonName == "" {
		slog.Error(
			nil,
			"cerficate CN is empty",
		)
		return nil, fmt.Errorf("ERROR: Unauthenticated access not allowed")
	}

	sessionKey = ""
	session = nil
	rafayUserName = state.PeerCertificates[0].Subject.CommonName
	sessionKey = state.ServerName + ":" + rafayUserName
	session, ok = sessions.GetUserSession(sessionKey)
	if !ok {
		//Create session for the user
		session = &sessions.UserSession{
			Type:            srv.Type,
			ServerName:      srv.ServerName,
			CertSNI:         state.ServerName,
			DialinCachedKey: "",
			ErrorFlag:       false,
		}
		sessions.AddUserSession(session, sessionKey)
		slog.Info(
			"Created new session",
			"key", sessionKey,
		)
	} else {
		slog.Debug(
			"Found existing session",
			"key", sessionKey,
			"dialin sticky key", session.DialinCachedKey,
		)
	}

	clusterID := strs[0]
	dialinSNI = strings.ReplaceAll(srv.DialinServerName, "*", clusterID)
	dialinAttempt = 0

retryCDDialin:
	// get the dialin server instance
	if dsrv, ok := Servers[srv.DialinServerName]; ok {
		if session.DialinCachedKey == "" {
			// look dialin pool to find dilain connection key
			key, err := dsrv.DialinPool.GetDialinConnectorKey(dialinSNI)
			if err != nil {
				slog.Info(
					"No dialins",
					"error", err,
					"dialinAttempt", dialinAttempt,
				)
				// Lookup peerCache to fetch cluster connection
				// On cache-miss send the probe, wait and retry 5 times
				relayIP, found := peerclient.GetPeerCache(utils.PeerCache, dialinSNI)
				if found {
					// forward the client request to peer upstream
					srv.ProcessCDPeerForwards(ctx, conn, lg, relayIP, state)
					return srv, nil
				}
				if dialinAttempt == 0 {
					go SendPeerProbe(PeerProbeChanel, dialinSNI)
				}

				dialinAttempt++
				if dialinAttempt > 7 {
					return nil, fmt.Errorf("ERROR: no dialns in both local and peers")
				}
				// wait for total ~20 sec to learn from core
				// else drop the connection
				time.Sleep(3 * time.Second)
				goto retryCDDialin

			} else {
				//cache it
				session.DialinCachedKey = key
			}
		} else {
			//verify cache is valid
			if !dsrv.DialinPool.CheckDialinKeyExist(session.DialinCachedKey) {
				session.DialinCachedKey = ""
				goto retryCDDialin
			}
		}

		slog.Info(
			"dialin connection lookup",
			"key", session.DialinCachedKey,
			"sessionKey", sessionKey,
		)

		session.UserName = rafayUserName

		//dilain proxy
		dsrv.proxyCDDilainConnection(ctx, conn, session.DialinCachedKey, rafayUserName, state.ServerName, slog)
	} else {
		slog.Info(
			"did not find dialin server name ",
			"srv.DialinServerName", srv.DialinServerName,
		)
	}

	return srv, nil
}

func (srv *Server) processDialinTLSState(conn net.Conn, lg *relaylogger.RelayLog, tlsconfig *tls.Config) (*Server, error) {
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		slog.Error(
			nil,
			"invalid connection type",
			"err", fmt.Errorf("expected TLS conn, got %T", conn),
		)
		return nil, fmt.Errorf("expected TLS conn, got %T", conn)
	}

	// perform TLS handshake
	if err := tlsConn.Handshake(); err != nil {
		slog.Error(
			err,
			"failed to compete TLS handshake",
		)
		return nil, fmt.Errorf("ailed to compete TLS handshake")
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		slog.Error(
			err,
			"failed to SetDeadline",
		)
		return nil, fmt.Errorf("failed to SetDeadline")
	}

	// Handshake completed, add to Connections map; key is remoteAddr
	if !isConnectionExist(conn.RemoteAddr().String()) {
		var rconn *RelayConn
		state := tlsConn.ConnectionState()

		identifier, err := GetRemoteCertID(tlsConn)
		if err != nil {
			slog.Error(
				err,
				"error in getting client cert id",
			)
			return nil, fmt.Errorf("error in getting client cert id")
		}

		rconn = &RelayConn{
			Conn:       conn,
			Type:       srv.Type,
			ServerName: srv.ServerName,
			CertSNI:    state.ServerName,
			PeerID:     identifier,
			server:     srv,
		}

		addConnection(conn.RemoteAddr().String(), rconn)

		if rconn.Type == utils.DIALIN {
			// add to dialin pool
			key, err := srv.AddToDialinPool(rconn, conn.RemoteAddr().String())
			if err != nil {
				deleteConnections(conn.RemoteAddr().String())
				return nil, err
			}
			// send init handshake
			err = srv.sendDialinHandshake(key, slog)
			if err != nil {
				// handshake failed; delete added connection
				srv.DialinPool.DeleteConn(rconn.PeerID, rconn.CertSNI, conn.RemoteAddr().String())
				deleteConnections(conn.RemoteAddr().String())
				return nil, fmt.Errorf("failed to send dialin handshake")
			}
		} else {
			return nil, fmt.Errorf("unknown service type")
		}

		return srv, nil
	}

	return nil, fmt.Errorf("connection already exist %s", conn.RemoteAddr().String())

}

func (srv *Server) handleCDRelayConnection(ctx context.Context, conn net.Conn, tlsconfig *tls.Config) {
	defer conn.Close()
	var vhostName string
	hlog := slog.WithName("handleDialinConnection")

	vhostName = ""
	tlsVConn, ok := conn.(*vhost.TLSConn)
	if ok {
		vhostName = tlsVConn.Host()
	}

	if vhostName == "" {
		slog.Error(
			nil,
			"vhost name is empty",
		)
		return
	}

	// prcess the TLS layer
	_, err := srv.processCDRelayTLSState(ctx, tls.Server(conn, tlsconfig), hlog, tlsconfig)
	if err != nil {
		hlog.Error(
			err,
			"failed to process TLS state",
		)
		return
	}

}

func (srv *Server) handleDialinConnection(conn net.Conn, tlsconfig *tls.Config) {
	var vhostName string
	hlog := slog.WithName("handleDialinConnection")

	vhostName = ""
	tlsVConn, ok := conn.(*vhost.TLSConn)
	if ok {
		vhostName = tlsVConn.Host()
	}

	if vhostName == "" {
		slog.Error(
			nil,
			"vhost name is empty",
		)
		conn.Close()
		return
	}

	// prcess the TLS layer
	_, err := srv.processDialinTLSState(tls.Server(conn, tlsconfig), hlog, tlsconfig)
	if err != nil {
		hlog.Error(
			err,
			"failed to process TLS state",
		)
		conn.Close()
		return
	}

}

func cleanServerName(s string) string {
	return strings.ReplaceAll(s, "*", "star")
}

// StartServer starts server
func StartServer(ctx context.Context, log *relaylogger.RelayLog, auditPath string, exitChan chan<- bool) {
	var config *ServerConfig
	var err error
	slog = log.WithName("Server")

	config = &ServerConfig{}
	config.Relays = make(map[string]*Relay)
	config.Relays["kubectl"] = &Relay{
		Protocol:   "https",
		Addr:       fmt.Sprintf(":%d", utils.RelayUserPort),
		ServerName: utils.RelayUserHost,
		DialinSfx:  utils.RelayConnectorHost,
		RootCA:     utils.RelayUserCACert,
		ServerCRT:  utils.RelayUserCert,
		ServerKEY:  utils.RelayUserKey,
	}
	config.Dialins = make(map[string]*Dialin)
	config.Dialins["kubectl"] = &Dialin{
		Protocol:   "https",
		Addr:       fmt.Sprintf(":%d", utils.RelayConnectorPort),
		ServerName: utils.RelayConnectorHost,
		RootCA:     utils.RelayConnectorCACert,
		ServerCRT:  utils.RelayConnectorCert,
		ServerKEY:  utils.RelayConnectorKey,
	}
	config.AuditPath = auditPath

	utils.PeerCache, err = peerclient.InitPeerCache(nil)
	if err != nil {
		slog.Error(
			err,
			"failed to init peer cache",
		)
		return
	}

	loadRelayServers(config)

	err = sessions.InitUserSessionCache()
	if err != nil {
		slog.Error(
			err,
			"failed to init user session cache",
		)
		return
	}

	if err := proxy.InitUnixCacheRoundTripper(); err != nil {
		slog.Error(
			err,
			"failed to init unix cached round tripper",
		)
		return
	}

	if err := proxy.InitPeerCacheRoundTripper(); err != nil {
		slog.Error(
			err,
			"failed to init unix cached round tripper",
		)
		return
	}

	// Start listeners
	for _, listener := range ServerListeners {
		switch listener.Protocol {
		case utils.HTTPS:
			go listener.StartHTTPSListen(ctx)
		default:
			slog.Error(
				nil,
				"unknown protocol",
				"listener.Protocol", listener.Protocol,
				"Addr", listener.Addr,
			)
		}
	}

	go StartPeeringMgr(ctx, log, exitChan, config)

	StartDialinPoolMgr(ctx, log, exitChan)

	for {
		select {
		case <-ctx.Done():
			slog.Error(
				ctx.Err(),
				"stoping Server",
			)
			return
		}
	}
}

// StartCDServer starts server
func StartCDServer(ctx context.Context, log *relaylogger.RelayLog, auditPath string, exitChan chan<- bool) {
	var config *ServerConfig
	var err error
	slog = log.WithName("Server")

	config = &ServerConfig{}
	config.CDRelays = make(map[string]*Relay)
	config.CDRelays["cdrelay"] = &Relay{
		Protocol:   "https",
		Addr:       fmt.Sprintf(":%d", utils.CDRelayUserPort),
		ServerName: utils.CDRelayUserHost,
		DialinSfx:  utils.CDRelayConnectorHost,
		RootCA:     utils.CDRelayUserCACert,
		ServerCRT:  utils.CDRelayUserCert,
		ServerKEY:  utils.CDRelayUserKey,
	}
	config.Dialins = make(map[string]*Dialin)
	config.Dialins["cdrelay"] = &Dialin{
		Protocol:   "https",
		Addr:       fmt.Sprintf(":%d", utils.CDRelayConnectorPort),
		ServerName: utils.CDRelayConnectorHost,
		RootCA:     utils.CDRelayConnectorCACert,
		ServerCRT:  utils.CDRelayConnectorCert,
		ServerKEY:  utils.CDRelayConnectorKey,
	}
	config.AuditPath = auditPath

	utils.PeerCache, err = peerclient.InitPeerCache(nil)
	if err != nil {
		slog.Error(
			err,
			"failed to init peer cache",
		)
		return
	}

	loadRelayServers(config)

	err = sessions.InitUserSessionCache()
	if err != nil {
		slog.Error(
			err,
			"failed to init user session cache",
		)
		return
	}

	if err := proxy.InitUnixCacheRoundTripper(); err != nil {
		slog.Error(
			err,
			"failed to init unix cached round tripper",
		)
		return
	}

	if err := proxy.InitPeerCacheRoundTripper(); err != nil {
		slog.Error(
			err,
			"failed to init unix cached round tripper",
		)
		return
	}

	// Start listeners
	for _, listener := range ServerListeners {
		switch listener.Protocol {
		case utils.HTTPS:
			go listener.StartHTTPSListen(ctx)
		default:
			slog.Error(
				nil,
				"unknown protocol",
				"listener.Protocol", listener.Protocol,
				"Addr", listener.Addr,
			)
		}
	}

	go StartPeeringMgr(ctx, log, exitChan, config)

	StartDialinPoolMgr(ctx, log, exitChan)

	for {
		select {
		case <-ctx.Done():
			slog.Error(
				ctx.Err(),
				"stoping Server",
			)
			return
		}
	}
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
		"len", len(p),
		"wrote", n,
		"data", string(p),
		"err", err,
	)
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	// tolerate max read/write idle timeout
	fw.c.SetDeadline(time.Now().Add(utils.IdleTimeout))
	return
}

func isConnectionExist(key string) bool {
	connectionsRWMu.RLock()
	defer connectionsRWMu.RUnlock()
	_, ok := connections[key]
	return ok
}

func addConnection(key string, rconn *RelayConn) {
	connectionsRWMu.Lock()
	defer connectionsRWMu.Unlock()
	connections[key] = rconn
}

func deleteConnections(key string) {
	connectionsRWMu.Lock()
	defer connectionsRWMu.Unlock()
	delete(connections, key)
}

//DialinMetric for cluster connection
func DialinMetric(w http.ResponseWriter) {
	for _, dsrv := range Servers {
		if dsrv.Type == utils.DIALIN {
			// look dialin pool to find dilain connection key
			dsrv.DialinPool.GetDialinMetrics(w)
		}
	}
}
