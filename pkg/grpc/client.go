package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

var (
	clientDefaultOpts = []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 30,
			Timeout:             time.Second * 30,
			PermitWithoutStream: true,
		}),
	}
)

// NewSecureClientConn returns new grpc client connection given server host, server port and transport credentials
func NewSecureClientConn(ctx context.Context, addr string, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	nctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	opts = append(opts, clientDefaultOpts...)
	opts = append([]grpc.DialOption{grpc.WithTransportCredentials(creds)}, opts...)

	return grpc.DialContext(nctx, addr, opts...)
}

// NewClientConn returns new grpc client connection given server host and server port
func NewClientConn(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	nctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	opts = append(opts, clientDefaultOpts...)
	opts = append([]grpc.DialOption{grpc.WithInsecure()}, opts...)

	return grpc.DialContext(nctx, addr, opts...)
}

func newClientTLSConfig(certPEM []byte, keyPEM []byte, caCertPEM []byte, addr string) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	var roots *x509.CertPool
	if len(caCertPEM) > 0 {
		roots = x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM(caCertPEM); !ok {
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

// NewClientTransportCredentials returns grpc client transport credentials
func NewClientTransportCredentials(cert, key, caCert []byte, addr string) (credentials.TransportCredentials, error) {
	tlsConfig, err := newClientTLSConfig(cert, key, caCert, addr)
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(tlsConfig), nil
}

func NewGrpcClientClientConn(ctx context.Context, serverHost string, serverPort int) (*grpc.ClientConn, error) {
	cc, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s:%d", serverHost, serverPort),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 30,
			Timeout:             time.Second * 30,
			PermitWithoutStream: true,
		}),
	)

	return cc, err
}

// NewGrpcClientClientConn returns new grpc client connection given server host and server port
func NewGrpcClientClientConnWithTimeout(ctx context.Context, serverHost string, serverPort int, timeoutInMins time.Duration) (*grpc.ClientConn, error) {
	// resolve every 30 seconds
	//resolver, _ := naming.NewDNSResolverWithFreq(time.Second * 30)
	//serverBalancer := grpc.RoundRobin(resolver)
	to := time.Duration(timeoutInMins)
	cc, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s:%d", serverHost, serverPort),
		grpc.WithInsecure(),
		//grpc.WithTransportCredentials(creds),
		//grpc.WithBackoffConfig(grpc.DefaultBackoffConfig),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 30,
			Timeout:             time.Minute * to,
			PermitWithoutStream: true,
		}),
		//grpc.WithBalancer(serverBalancer),
	)

	return cc, err
}
