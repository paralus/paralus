package config

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/pool"
	grpcpool "github.com/processout/grpc-go-pool"
)

// ConfigClient is the interface which exposes all the RPCs
// to internal service
type ConfigClient interface {
	Unhealthy()
	Close() error
	OverrideClient
}

type configClient struct {
	*grpcpool.ClientConn
	*overrideClient
}

var _ ConfigClient = (*configClient)(nil)

// ConfigPool maintains pool of grpc connections to config service
type ConfigPool interface {
	Close()
	NewClient(ctx context.Context) (ConfigClient, error)
}

// NewConfigPool new config pool
func NewConfigPool(addr string, maxConn int) ConfigPool {
	return &configPool{
		GRPCPool: pool.NewGRPCPool(addr, maxConn, nil),
	}
}

type configPool struct {
	*pool.GRPCPool
}

func (p *configPool) Close() {
	if p.GRPCPool != nil {
		p.GRPCPool.Close()
	}
}

func (p *configPool) NewClient(ctx context.Context) (ConfigClient, error) {
	cc, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}
	return &configClient{
		cc,
		&overrideClient{cc: cc},
	}, nil
}
