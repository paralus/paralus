package sentry

import (
	"context"
	"fmt"

	grpctools "github.com/paralus/paralus/pkg/grpc"
	"github.com/paralus/paralus/pkg/pool"
	grpcpool "github.com/processout/grpc-go-pool"
	"google.golang.org/grpc/credentials"
)

// SentryClient is the interface for accessing all the RPCs
// exposed by Paralus Sentry
type SentryClient interface {
	Unhealthy()
	Close() error
	BootstrapServiceClient
	ClusterAuthorizationServiceClient
	KubeConfigServiceClient
}

// SentryAuthorizationClient is the interface for accessing all the RPCs
// exposed by Paralus Sentry for Authorization
type SentryAuthorizationClient interface {
	Unhealthy()
	Close() error
	ClusterAuthorizationServiceClient
	AuditInformationServiceClient
}

type sentryClient struct {
	*grpcpool.ClientConn
	*bootstrapServiceClient
	*clusterAuthorizationServiceClient
	*kubeConfigServiceClient
}

var _ SentryClient = (*sentryClient)(nil)

type sentryAuthorizationClient struct {
	*grpcpool.ClientConn
	*clusterAuthorizationServiceClient
	*auditInformationServiceClient
}

var _ SentryAuthorizationClient = (*sentryAuthorizationClient)(nil)

// SentryPool maintains pool of grpc connections to sentry service
type SentryPool interface {
	Close()
	NewClient(ctx context.Context) (SentryClient, error)
}

// SentryAuthorizationPool maintains pool of grpc connections to sentry
// authorization service
type SentryAuthorizationPool interface {
	Close()
	NewClient(ctx context.Context) (SentryAuthorizationClient, error)
}

// NewSentryPool new sentry pool
func NewSentryPool(addr string, maxConn int) SentryPool {
	return &sentryPool{
		GRPCPool: pool.NewGRPCPool(addr, maxConn, nil),
	}
}

type sentryPool struct {
	*pool.GRPCPool
}

func (p *sentryPool) Close() {
	if p.GRPCPool != nil {
		p.GRPCPool.Close()
	}
}

func (p *sentryPool) NewClient(ctx context.Context) (SentryClient, error) {
	cc, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}
	return &sentryClient{
		cc,
		&bootstrapServiceClient{cc.ClientConn},
		&clusterAuthorizationServiceClient{cc.ClientConn},
		&kubeConfigServiceClient{cc.ClientConn},
	}, nil
}

type sentryAuthorizationPool struct {
	*pool.GRPCPool
}

func (p *sentryAuthorizationPool) Close() {
	if p.GRPCPool != nil {
		p.GRPCPool.Close()
	}
}

func (p *sentryAuthorizationPool) NewClient(ctx context.Context) (SentryAuthorizationClient, error) {
	cc, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}
	return &sentryAuthorizationClient{
		cc,
		&clusterAuthorizationServiceClient{cc.ClientConn},
		&auditInformationServiceClient{cc.ClientConn},
	}, nil
}

type options struct {
	addr    string
	maxConn int
	cert    []byte
	key     []byte
	caCert  []byte
}

// Option is the functional argument for Pool options
type Option func(*options)

// WithAddr sets address of the pool
func WithAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

// WithMaxConn sets maximum number of connections of the pool
// if not set defaults to 10
func WithMaxConn(maxConn int) Option {
	return func(o *options) {
		o.maxConn = maxConn
	}
}

// WithClientCertPEM sets PEM encoded client cert for pool
func WithClientCertPEM(cert []byte) Option {
	return func(o *options) {
		o.cert = cert
	}
}

// WithClientKeyPEM sets PEM encoded client key for pool
func WithClientKeyPEM(key []byte) Option {
	return func(o *options) {
		o.key = key
	}
}

// WithCaCertPEM sets PEM encoded CA cert for pool
func WithCaCertPEM(caCert []byte) Option {
	return func(o *options) {
		o.caCert = caCert
	}
}

// NewSentryAuthorizationPool new sentry authorization pool
func NewSentryAuthorizationPool(opts ...Option) (SentryAuthorizationPool, error) {

	pOpts := options{}
	for _, opt := range opts {
		opt(&pOpts)
	}

	if pOpts.addr == "" {
		return nil, fmt.Errorf("addr cannot be empty")
	}

	if pOpts.maxConn <= 0 {
		pOpts.maxConn = 10
	}

	var creds credentials.TransportCredentials
	var err error

	if pOpts.cert != nil && pOpts.key != nil {
		creds, err = grpctools.NewClientTransportCredentials(pOpts.cert, pOpts.key, pOpts.caCert, pOpts.addr)
		if err != nil {
			return nil, err
		}
	}

	return &sentryAuthorizationPool{
		GRPCPool: pool.NewGRPCPool(pOpts.addr, pOpts.maxConn, creds),
	}, nil
}
