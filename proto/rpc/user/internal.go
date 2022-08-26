package userv3

import (
	"context"

	"github.com/paralus/paralus/pkg/pool"
	grpcpool "github.com/processout/grpc-go-pool"
)

// UserClient is the interface for accessing all User & Group RPCs
// exposed by Paralus Base
type UGClient interface {
	Unhealthy()
	Close() error
	UserServiceClient
}

type ugClient struct {
	*grpcpool.ClientConn
	*userServiceClient
}

var _ UGClient = (*ugClient)(nil)

// UGPool maintains pool of grpc connections to system base services
type UGPool interface {
	Close()
	NewClient(ctx context.Context) (UGClient, error)
}

// NewUGPool new user group pool
func NewUGPool(addr string, maxConn int) UGPool {
	return &ugPool{
		GRPCPool: pool.NewGRPCPool(addr, maxConn, nil),
	}
}

type ugPool struct {
	*pool.GRPCPool
}

func (p *ugPool) Close() {
	if p.GRPCPool != nil {
		p.GRPCPool.Close()
	}
}

func (p *ugPool) NewClient(ctx context.Context) (UGClient, error) {
	cc, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}
	return &ugClient{
		cc,
		&userServiceClient{cc.ClientConn},
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
