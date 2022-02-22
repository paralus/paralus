package tunnel

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
)

type handShakeMsg struct {
	ServiceName string
	Protocol    string
	Host        string
}

var (
	//PeerProbeChanel to push cluster connection probe.
	//The RPC picks the sni and sends to core
	PeerProbeChanel = make(chan string, 256)
	emptyID         [32]byte
)

//ServerTLSConfigFromBytes prepare a tls config from cert,key,rootCA
func ServerTLSConfigFromBytes(certList []utils.SNICertificate, rootCAs []string, nextprotos ...string) (*tls.Config, error) {
	var err error

	// load certs
	config := &tls.Config{}
	config.Certificates = make([]tls.Certificate, len(certList))

	for i, v := range certList {
		config.Certificates[i], err = tls.X509KeyPair(v.CertFile, v.KeyFile)
		if err != nil {
			return nil, err
		}
	}

	config.BuildNameToCertificate()

	// load rootCAs for client authentication
	clientAuth := tls.RequireAndVerifyClientCert
	var roots *x509.CertPool
	if len(rootCAs) > 0 {
		roots = x509.NewCertPool()

		for _, rootCA := range rootCAs {
			rootPEM := []byte(rootCA)
			if len(rootPEM) > 0 {
				if ok := roots.AppendCertsFromPEM(rootPEM); !ok {
					return nil, err
				}
			}
		}

		clientAuth = tls.RequireAndVerifyClientCert
	}

	config.ClientAuth = clientAuth
	config.ClientCAs = roots
	config.SessionTicketsDisabled = true
	config.MinVersion = tls.VersionTLS12
	config.CipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256}
	config.PreferServerCipherSuites = true
	if len(nextprotos) > 0 {
		config.NextProtos = nextprotos
	} else {
		config.NextProtos = []string{"h1", "h2"}
	}

	return config, nil
}

//ClientTLSConfigFromBytes sets tls config
func ClientTLSConfigFromBytes(tlsCrt []byte, tlsKey []byte, rootPEM []byte, addr string) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(tlsCrt, tlsKey)
	if err != nil {
		return nil, err
	}

	var roots *x509.CertPool
	if len(rootPEM) > 0 {
		roots = x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM(rootPEM); !ok {
			return nil, err
		}
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		ServerName:             host,
		Certificates:           []tls.Certificate{cert},
		InsecureSkipVerify:     roots == nil,
		RootCAs:                roots,
		SessionTicketsDisabled: true,
		MinVersion:             tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		PreferServerCipherSuites: true,
	}, nil
}

//ClientTLSConfig sets tls config
func ClientTLSConfig(tlsCrt string, tlsKey string, rootCA string, addr string) (*tls.Config, error) {

	cert, err := tls.LoadX509KeyPair(tlsCrt, tlsKey)
	if err != nil {
		return nil, err
	}

	var roots *x509.CertPool
	if rootCA != "" {
		roots = x509.NewCertPool()
		rootPEM, err := ioutil.ReadFile(rootCA)
		if err != nil {
			return nil, err
		}
		if ok := roots.AppendCertsFromPEM(rootPEM); !ok {
			return nil, err
		}
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		ServerName:             host,
		Certificates:           []tls.Certificate{cert},
		InsecureSkipVerify:     roots == nil,
		RootCAs:                roots,
		SessionTicketsDisabled: true,
		MinVersion:             tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		PreferServerCipherSuites: true,
	}, nil
}

//GetRemoteCertID extract peer ID
func GetRemoteCertID(conn *tls.Conn) (string, error) {
	// Try a TLS connection over the given connection. We explicitly perform
	// the handshake, since we want to maintain the invariant that, if this
	// function returns successfully, then the connection should be valid
	// and verified.
	if err := conn.Handshake(); err != nil {
		return "", err
	}

	cs := conn.ConnectionState()

	// We should have exactly one peer certificate.
	certs := cs.PeerCertificates
	if cl := len(certs); cl != 1 {
		return "", fmt.Errorf("expecting 1 peer certificate got %d", cl)
	}

	// Get remote cert's ID.
	remoteCert := certs[0]
	//remoteID := New(remoteCert.Raw)

	return remoteCert.Subject.CommonName, nil
}

func SendPeerProbe(chnl chan<- string, clustersni string) {
	attempt := 0
	select {
	case chnl <- clustersni:
		return
	default:
		attempt++
		if attempt > 3 {
			return
		}
		time.Sleep(5 * time.Second)
	}
}
