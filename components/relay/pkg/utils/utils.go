package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/dgraph-io/ristretto"
	"github.com/felixge/tcpkeepalive"
	"github.com/google/uuid"
)

// Known relay services
const (
	KUBECTL = "kubectl"
	KUBEWEB = "kubeweb"
	GENTCP  = "tcp"
	HTTPUP  = "httpupgrade"
)

// Known server types
const (
	RELAY        = "relay"
	CDRELAY      = "cdrelay"
	RELAYAGENT   = "relay-agent"
	CDRELAYAGENT = "cdrelay-agent"
	DIALIN       = "dialin"
	JoinString   = "--"
)

// Relay Network Types
const (
	KUBECTLCORE      = "rafay-core-relay-agent"
	KUBECTLDEDICATED = "rafay-non-core-relay-agent"
	CDAGENTCORE      = "rafay-core-cd-relay-agent"
)

//SNICertificate sni based certs
type SNICertificate struct {
	CertFile []byte
	KeyFile  []byte
}

//Relaynetwork configmap data
type Relaynetwork struct {
	Token         string `json:"token"`         // bootstrap agent token
	Addr          string `json:"addr"`          // bootstrap register host
	Domain        string `json:"endpoint"`      // dialout domain
	Name          string `json:"name"`          // network name
	TemplateToken string `json:"templateToken"` // bootstrap template token
	Upstream      string `json:"upstream"`      // upstream tcp service host:port
}

//RelayNetworkConfig config for relay agent
type RelayNetworkConfig struct {
	// Network configmap
	Network Relaynetwork
	// RelayAgentCert used for relay-agent client cert
	RelayAgentCert []byte
	// RelayAgentKey used for relay-agent client cert
	RelayAgentKey []byte
	// RelayAgentCACert used for relay-agent client cert
	RelayAgentCACert []byte
}

// Known protocol types.
const (
	HTTP  = "HTTP"
	HTTPS = "https"
	TCP   = "tcp"
	UNIX  = "unix"
)

// Known dialin types
const (
	KUBECTLDILAIN = "kubectldialin"
	KUBEWEBDIALIN = "kubewebdialin"
	PEERKEY       = "04112005676520746869732070617373776f726420746f206120736563726574"
)

const (
	//HeaderError ..
	HeaderError = "X-Error"
	//HeaderAction ...
	HeaderAction = "X-Action"
	//HeaderForwardedHost ..
	HeaderForwardedHost = "X-Forwarded-Host"
	//HeaderForwardedService ..
	HeaderForwardedService = "X-Forwarded-Service"

	//HeaderRafayUserName ..
	HeaderRafayUserName = "X-Rafay-User"
	//HeaderRafayNamespace ..
	HeaderRafayNamespace = "X-Rafay-Namespace"
	//HeaderRafayScope ..
	HeaderRafayScope = "X-Rafay-Scope"
	//HeaderRafayAllow ..
	HeaderRafayAllow = "X-Rafay-Allow"
	//HeaderRafayAuthZSA yaml contains service account
	HeaderRafayAuthZSA = "X-Rafay-AuthzSA"
	//HeaderRafayAuthZRole yaml contains role
	HeaderRafayAuthZRole = "X-Rafay-AuthzRole"
	//HeaderRafayAuthZRoleBinding yaml contains rolebinding
	HeaderRafayAuthZRoleBinding = "X-Rafay-AuthzRoleBinding"
	//HeaderRafayServiceAccountNoExpire don't expire service account
	HeaderRafayServiceAccountNoExpire = "X-Rafay-ServiceAccount-NoExpire"
	//HeaderClearSecret to clear the current secret cache of user
	HeaderClearSecret = "X-Rafay-Clear-Cache"
)

// Known actions.
const (
	ActionProxy = "proxy"

	// DefaultAuditPolicyPath default audit policy filter path
	// k8s audit need a file path
	DefaultAuditPolicyPath = "./relayaudit.yaml"

	//DefaultAuditPath defailt audit log files path
	DefaultAuditPath = "-" // - means standard out

	//RafayRelayServiceAccountNameSpace namespace used to create service account for relays
	RafayRelayServiceAccountNameSpace = "system-sa"
)

var (
	// LogLevel loglevel set from commadline
	LogLevel int
	// Mode relay/relay-agent
	Mode string
	// ClusterID unique id of the cluster
	ClusterID string
	// AgentID unique id for cd agent
	AgentID string
	// ExitChan trigger this channel to exit
	ExitChan = make(chan bool)
	// TerminateChan trigger this channel to exit
	TerminateChan = make(chan bool)
	// IdleTimeout is the maximum amount of time to wait for the
	// next read/write before closing connection.
	IdleTimeout = 5 * time.Minute
	// DefaultTimeout specifies a general purpose timeout.
	DefaultTimeout = 5 * time.Minute
	// DefaultPingTimeout specifies a ping timeout.
	DefaultPingTimeout = 5 * time.Second

	// DefaultKeepAliveIdleTime specifies how long connection can be idle
	// before sending keepalive message.
	DefaultKeepAliveIdleTime = 15 * time.Second

	// DefaultKeepAliveCount specifies maximal number of keepalive messages
	// sent before marking connection as dead.
	DefaultKeepAliveCount = 3

	// DefaultKeepAliveInterval specifies how often retry sending keepalive
	// messages when no response is received.
	DefaultKeepAliveInterval = 5 * time.Second

	//DefaultMuxTimeout specifies vmux timeout
	DefaultMuxTimeout = 10 * time.Second

	//UNIXSOCKET prefix path for unix socket
	UNIXSOCKET = "/tmp/relay-unix-" // need to change this from tmp to appropriate path after integration

	//UNIXAGENTSOCKET prefix path for unix socket
	UNIXAGENTSOCKET = "/tmp/relay-agent-unix-" // need to change this from tmp to appropriate path after integration

	//ProxyProtocolSize Default PROXY PROTO buffer size
	ProxyProtocolSize = 1024

	//RelayUUID runtime Unique ID for relay
	RelayUUID string

	//RelayIPFromConfig IP address of the relay for peering
	RelayIPFromConfig string

	//PeerCache stores peer dialin info
	PeerCache *ristretto.Cache

	//ServiceAccountCache stores service account, role, role binding in relay-agetn in connector
	ServiceAccountCache *ristretto.Cache

	//ServiceAccountCacheDefaultExpiry default expiry
	ServiceAccountCacheDefaultExpiry = 600 * time.Second

	//PeerCacheDefaultExpiry default expiry
	PeerCacheDefaultExpiry = 600 * time.Second
	//PeerHelloInterval heartbeat interval
	PeerHelloInterval = 60 * time.Second
	//PeerServiceURI is the URI to join peering service
	PeerServiceURI string
	//PeerCertificate used for peering service communication
	PeerCertificate []byte
	//PeerPrivateKey used for peering service communication
	PeerPrivateKey []byte
	//PeerCACertificate used for peering service communication
	PeerCACertificate []byte

	//RelayUserCert used for user/peer communication
	RelayUserCert []byte
	//RelayUserKey used for user/peer communication
	RelayUserKey []byte
	//RelayUserCACert used for user/peer communication
	RelayUserCACert []byte

	//RelayUserPort user facing seerver port
	RelayUserPort int
	// RelayUserHost user facing seerver host (domain)
	RelayUserHost string
	//RelayConnectorCert used for relay-connector termination
	RelayConnectorCert []byte
	//RelayConnectorKey used for relay-connector termination
	RelayConnectorKey []byte
	//RelayConnectorCACert used for relay-connector termination
	RelayConnectorCACert []byte
	// RelayConnectorHost connector facing server host (domain)
	RelayConnectorHost string
	// RelayConnectorPort connector facing server port
	RelayConnectorPort int

	// CDRelayUserCert used for client/peer communication
	CDRelayUserCert []byte
	// CDRelayUserKey used for client/peer communication
	CDRelayUserKey []byte
	// CDRelayUserCACert used for client/peer communication
	CDRelayUserCACert []byte
	// CDRelayUserHost client facing server host
	CDRelayUserHost string
	// CDRelayUserPort client facing server port
	CDRelayUserPort int
	// CDRelayConnectorCert used for cd-relay-connector termination
	CDRelayConnectorCert []byte
	// CDRelayConnectorKey used for cd-relay-connector termination
	CDRelayConnectorKey []byte
	// CDRelayConnectorCACert used for cd-relay-connector termination
	CDRelayConnectorCACert []byte
	// CDRelayConnectorHost connector facing server host (domain)
	CDRelayConnectorHost string
	// CDRelayConnectorPort connector facing server port
	CDRelayConnectorPort int

	//RelayNetworks list of relaynemtworks from configmap
	RelayNetworks []Relaynetwork
	// RelayAgentConfig map of relay agent configurations
	RelayAgentConfig map[string]RelayNetworkConfig

	// MaxDials max connections dialed
	MaxDials = 10

	// MinDials minimum connections dialed
	MinDials = 8

	//PODNAME name of the pod
	PODNAME string

	// DialoutProxy setting used while connecting to relay IP:PORT or HOST:PORT format
	DialoutProxy = ""

	// DialoutProxyAuth Proxy-Authorization header base64 encoded value of user:password
	DialoutProxyAuth = ""

	// DefaultTCPUpstream default TCP upstream
	DefaultTCPUpstream = "127.0.0.1:16001"

	// ScalingStreamsThreshold concurrent streams count to trigger scaling
	ScalingStreamsThreshold = 400

	// ScalingStreamsRateThreshold new streams rate to trigger scaling
	ScalingStreamsRateThreshold = 200

	// MaxScaleMultiplier multiplier to limit max scaled connections
	MaxScaleMultiplier = 3

	// HealingInterval time to close idle scaled connection
	HealingInterval = 24 // Hour
)

//CountWriter to measure bytes
type CountWriter struct {
	W     io.Writer
	Count int64
}

// ControlMessage is sent from server to client before streaming data. It's
// used to inform client about the data and action to take. Based on that client
// routes requests to backend services.
type ControlMessage struct {
	Action           string
	ForwardedHost    string
	ForwardedService string
	RemoteAddr       string
	RafayUserName    string
	RafayNamespace   string
	RafayScope       string
	RafayAllow       string
	RafayAuthz       string
}

// ProxyConfig configs for the proxy
type ProxyConfig struct {
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

//ProxyProtocolMessage used across dialin unix socket
type ProxyProtocolMessage struct {
	DialinKey string
	UserName  string
	SNI       string
}

type ServiceAccountCacheObject struct {
	RafayAuthzSA       string
	RafayAuthzRole     string
	RafayAuthzRoleBind string
	Md5sum             string
	Key                string
}

//OnEvict cache on eviction call back function
type OnEvict = func(item *ristretto.Item)

//Fatal to exit the program
func Fatal(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprint(os.Stderr, "\n")
	os.Exit(1)
}

//CloneHeader clone http headers
func CloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

//SetXForwardedFor ...
func SetXForwardedFor(h http.Header, remoteAddr string) {
	clientIP, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior, ok := h["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		h.Set("X-Forwarded-For", clientIP)
	}
}

//SetXRAYUUID ...
func SetXRAYUUID(h http.Header) {
	var reluuid string

	reluuid = RelayUUID
	// If this isn't the first relay retain prior
	// UUIDs information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := h["X-Rafay-XRAY-RELAYUUID"]; ok {
		reluuid = strings.Join(prior, ", ") + ", " + reluuid
	}
	h.Set("X-Rafay-XRAY-RELAYUUID", reluuid)
}

//SetXForwardedRafay set rafay headers
func SetXForwardedRafay(h http.Header, msg *ControlMessage) {
	h.Set("X-Rafay-User", msg.RafayUserName)
	h.Set("X-Rafay-Namespace", msg.RafayNamespace)
	h.Set("X-Rafay-Scope", msg.RafayScope)
	h.Set("X-Rafay-Allow", msg.RafayAllow)
}

//UnSetXForwardedRafay set rafay headers
func UnSetXForwardedRafay(h http.Header) {
	h.Del("X-Rafay-Scope")
	h.Del("X-Rafay-Allow")
	h.Del("X-Forwarded-For")
	h.Del("X-Rafay-XRAY-RELAYUUID")
	h.Del("X-Rafay-Audit")
	h.Del("X-Rafay-Cluster-Id")
	h.Del("X-Rafay-Cluster-Servername")
	h.Del("X-Rafay-Peer-Hash")
	h.Del("X-Rafay-Peer-Nonce")
	h.Del("X-Rafay-Sessionkey")
	h.Del("X-Rafay-User-Cert-Issued")
}

//Transfer transfer by io.Copy
func Transfer(dst io.Writer, src io.Reader, tlog *relaylogger.RelayLog, direction string) {
	n, err := io.Copy(dst, src)
	if err != nil {
		if !strings.Contains(err.Error(), "context canceled") && !strings.Contains(err.Error(), "CANCEL") && !strings.Contains(err.Error(), "i/o timeout") && !strings.Contains(err.Error(), "closed network") {
			tlog.Error(
				err,
				"io.Copy error",
				"direction", direction,
			)
		}
	}

	tlog.Debug(
		"action transferred",
		"bytes", n,
		"in direction", direction,
	)
}

// WriteToHeader writes ControlMessage to HTTP header.
func WriteToHeader(h http.Header, c *ControlMessage) {
	h.Set(HeaderAction, string(c.Action))
	h.Set(HeaderForwardedHost, c.ForwardedHost)
	h.Set(HeaderForwardedService, c.ForwardedService)
	h.Set(HeaderRafayUserName, c.RafayUserName)
	h.Set(HeaderRafayNamespace, c.RafayNamespace)
	h.Set(HeaderRafayScope, c.RafayScope)
	h.Set(HeaderRafayAllow, c.RafayAllow)
}

func (cw *CountWriter) Write(p []byte) (n int, err error) {
	n, err = cw.W.Write(p)
	cw.Count += int64(n)
	return
}

//CopyHeader copy header
func CopyHeader(dst, src http.Header) {
	for k, v := range src {
		vv := make([]string, len(v))
		copy(vv, v)
		dst[k] = vv
	}
}

// ReadControlMessage reads ControlMessage from HTTP headers.
func ReadControlMessage(r *http.Request) (*ControlMessage, error) {
	msg := ControlMessage{
		Action:           r.Header.Get(HeaderAction),
		ForwardedHost:    r.Header.Get(HeaderForwardedHost),
		ForwardedService: r.Header.Get(HeaderForwardedService),
		RafayUserName:    r.Header.Get(HeaderRafayUserName),
		RafayNamespace:   r.Header.Get(HeaderRafayNamespace),
		RafayScope:       r.Header.Get(HeaderRafayScope),
		RafayAllow:       r.Header.Get(HeaderRafayAllow),
		RemoteAddr:       r.RemoteAddr,
	}

	var missing []string

	if msg.Action == "" {
		missing = append(missing, HeaderAction)
	}
	if msg.ForwardedHost == "" {
		missing = append(missing, HeaderForwardedHost)
	}
	if msg.ForwardedService == "" {
		missing = append(missing, HeaderForwardedService)
	}

	if len(missing) != 0 {
		return nil, fmt.Errorf("missing headers: %s", missing)
	}

	return &msg, nil
}

//FlushWriter flush writer
type FlushWriter struct {
	W io.Writer
}

func (fw FlushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.W.Write(p)
	if f, ok := fw.W.(http.Flusher); ok {
		f.Flush()
	}
	return
}

//KeepAlive set keepalive
func KeepAlive(conn net.Conn) error {
	return tcpkeepalive.SetKeepAlive(conn, DefaultKeepAliveIdleTime, DefaultKeepAliveCount, DefaultKeepAliveInterval)
}

//GenUUID generates a google UUID
func GenUUID() {
	id := uuid.New()
	RelayUUID = id.String()
}

//GetRelayIP get relay IP address
func GetRelayIP() string {
	if RelayIPFromConfig == "" {
		name, err := os.Hostname()
		if err == nil {
			addr, err := net.LookupIP(name)
			if err == nil {
				return addr[0].String()
			}
		} else {
			return ""
		}
	}
	return RelayIPFromConfig
}

// GetRelayIPPort get relay IP:PORT of user facing server
func GetRelayIPPort() string {
	if RelayIPFromConfig == "" {
		return GetRelayIP() + ":" + strconv.Itoa(RelayUserPort)
	}
	return RelayIPFromConfig
}

//CheckRelayLoops :does XRAY UUDI already present in header?
func CheckRelayLoops(h http.Header) bool {
	if uuidHdr, ok := h["X-Rafay-XRAY-RELAYUUID"]; ok {
		allIds := strings.Join(uuidHdr, ", ")
		matched, err := regexp.Match(RelayUUID, []byte(allIds))
		if err == nil && matched {
			return true
		}
	}
	return false
}

// InitCache initialize the cache to store dialin cluster-connection
// information of peers. When a dialin miss happens look into this cache
// to find the peer IP address to forward the user connection.
func InitCache(evict OnEvict) (*ristretto.Cache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // Num keys to track frequency of (10M).
		MaxCost:     1 << 30, // Maximum cost of cache (1GB).
		BufferItems: 64,      // Number of keys per Get buffer.
		OnEvict:     evict,
	})
}

// InsertCache inserts the value to cache
func InsertCache(cache *ristretto.Cache, expiry time.Duration, key, value interface{}) bool {
	return cache.SetWithTTL(key, value, 100, expiry)
}

// GetCache get value from cache
func GetCache(cache *ristretto.Cache, key interface{}) (interface{}, bool) {
	return cache.Get(key)

}

// DeleteCache delete value from cache
func DeleteCache(cache *ristretto.Cache, key interface{}) {
	cache.Del(key)
}

//WriteFile overwrite if exist
func WriteFile(filename, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	file.WriteString(data)
	return nil
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// SplitHostPort separates host and port. If the port is not valid, it returns
// the entire input as host, and it doesn't check the validity of the host.
func SplitHostPort(hostport string) (host, port string) {
	host = hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

// IsHTTPS returns true if port is 443
func IsHTTPS(addr string) bool {
	_, port, err := net.SplitHostPort(addr)
	if err == nil {
		if port == "443" {
			return true
		}
	}
	return false
}

// PeerSetHeaderNonce header
func PeerSetHeaderNonce(h http.Header) error {
	key, _ := hex.DecodeString(PEERKEY)

	data := h.Get("X-Rafay-XRAY-RELAYUUID")
	if data == "" {
		return fmt.Errorf("no X-Rafay-XRAY-RELAYUUID header")
	}
	value := []byte(data)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	ciphertext := aesgcm.Seal(nil, nonce, value, nil)
	hash := hex.EncodeToString(ciphertext)

	peerNonce := hex.EncodeToString(nonce)
	h.Set("X-Rafay-Peer-Nonce", peerNonce)
	h.Set("X-Rafay-Peer-Hash", hash)
	return nil
}

// CheckPeerHeaders validates upstreams request
func CheckPeerHeaders(h http.Header) bool {
	key, _ := hex.DecodeString(PEERKEY)

	hash := h.Get("X-Rafay-Peer-Hash")
	if hash == "" {
		return false
	}

	peerNonce := h.Get("X-Rafay-Peer-Nonce")
	if peerNonce == "" {
		return false
	}

	expected := h.Get("X-Rafay-XRAY-RELAYUUID")
	if expected == "" {
		return false
	}

	ciphertext, _ := hex.DecodeString(hash)
	block, err := aes.NewCipher(key)
	if err != nil {
		return false
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return false
	}

	nonce, _ := hex.DecodeString(peerNonce)
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return false
	}

	if string(plaintext) == expected {
		return true
	}

	return false
}
