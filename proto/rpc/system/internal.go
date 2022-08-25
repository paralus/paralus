package systemv3

import (
	"context"

	"github.com/paralus/paralus/pkg/pool"
	grpcpool "github.com/processout/grpc-go-pool"
)

// SystemClient is the interface for accessing all the RPCs
// exposed by Paralus Base
type SystemClient interface {
	Unhealthy()
	Close() error
	ProjectServiceClient
	OrganizationServiceClient
	PartnerServiceClient
}

type systemClient struct {
	*grpcpool.ClientConn
	*projectServiceClient
	*organizationServiceClient
	*partnerServiceClient
}

var _ SystemClient = (*systemClient)(nil)

// SystemPool maintains pool of grpc connections to system base services
type SystemPool interface {
	Close()
	NewClient(ctx context.Context) (SystemClient, error)
}

// NewSystemPool new system pool
func NewSystemPool(addr string, maxConn int) SystemPool {
	return &systemPool{
		GRPCPool: pool.NewGRPCPool(addr, maxConn, nil),
	}
}

type systemPool struct {
	*pool.GRPCPool
}

func (p *systemPool) Close() {
	if p.GRPCPool != nil {
		p.GRPCPool.Close()
	}
}

func (p *systemPool) NewClient(ctx context.Context) (SystemClient, error) {
	cc, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}
	return &systemClient{
		cc,
		&projectServiceClient{cc.ClientConn},
		&organizationServiceClient{cc.ClientConn},
		&partnerServiceClient{cc.ClientConn},
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
