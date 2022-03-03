package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"google.golang.org/grpc/peer"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

var (
	serverDefaultOpts = []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    time.Second * 30, // server initiated keep alive interval
			Timeout: time.Second * 30, // server initiated keep alive timeout
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             time.Minute, // server enforcement for client keep alive
			PermitWithoutStream: true,        // allow connection without any active ongoing streams
		}),
	}

	// ErrInvalidClient is returned when client cert is not present in peer context
	ErrInvalidClient = errors.New("client has not presented certificate")
)

// NewSecureServerWithPEM creates a secure gRPC service with give PEM encoded cert, key and ca
func NewSecureServerWithPEM(cert, key, ca []byte, opts ...grpc.ServerOption) (*grpc.Server, error) {
	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("invalid cert/key pair: %s", err)
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, fmt.Errorf("failed to append ca cert")
	}
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})
	opts = append(opts, serverDefaultOpts...)
	opts = append([]grpc.ServerOption{grpc.Creds(creds)}, opts...)

	return grpc.NewServer(opts...), nil
}

// NewSecureServer returns new grpc server given cert path, key path and ca path
func NewSecureServer(certPath, keyPath, caPath string, opts ...grpc.ServerOption) (*grpc.Server, error) {
	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load cert/key : %s", err)
	}
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read ca cert: %s", err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, fmt.Errorf("failed to append ca cert %s", "")
	}

	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	opts = append(opts, serverDefaultOpts...)
	opts = append([]grpc.ServerOption{grpc.Creds(creds)}, opts...)

	return grpc.NewServer(opts...), nil
}

// NewServer returns new unsecured grpc server
func NewServer(opts ...grpc.ServerOption) (*grpc.Server, error) {
	opts = append(opts, serverDefaultOpts...)
	return grpc.NewServer(opts...), nil
}

// GetClientName returns client CommonName from client cert
func GetClientName(ctx context.Context) (string, error) {
	if p, ok := peer.FromContext(ctx); ok {
		tlsInfo := p.AuthInfo.(credentials.TLSInfo)
		for _, c := range tlsInfo.State.PeerCertificates {
			return c.Subject.CommonName, nil
		}
	}
	return "", ErrInvalidClient

}

// GetClientOU returns client Organization Unit from client cert
func GetClientOU(ctx context.Context) (string, error) {
	if p, ok := peer.FromContext(ctx); ok {
		tlsInfo := p.AuthInfo.(credentials.TLSInfo)
		for _, c := range tlsInfo.State.PeerCertificates {
			if len(c.Subject.OrganizationalUnit) > 0 {
				return strings.Join(c.Subject.OrganizationalUnit, "-"), nil
			}
		}
	}
	return "", ErrInvalidClient
}
