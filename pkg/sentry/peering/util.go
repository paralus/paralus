package peering

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
)

// ClientTLSConfig sets tls config
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
