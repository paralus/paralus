package authv3

import (
	"github.com/RafaySystems/rcloud-base/components/common/pkg/pool"

	"context"

	grpcpool "github.com/processout/grpc-go-pool"
)

// AuthPoolClient is the interface for auth pool client
type AuthPoolClient interface {
	Unhealthy()
	Close() error
	AuthClient
}

type authPoolClient struct {
	*grpcpool.ClientConn
	*authClient
}

var _ AuthPoolClient = (*authPoolClient)(nil)

// AuthPool maintains pool of grpc connections to auth service
type AuthPool interface {
	Close()
	NewClient(ctx context.Context) (AuthPoolClient, error)
}

var _ AuthPool = (*authPool)(nil)

type authPool struct {
	*pool.GRPCPool
}

func (p *authPool) Close() {
	if p.GRPCPool != nil {
		p.GRPCPool.Close()
	}
}

func (p *authPool) NewClient(ctx context.Context) (AuthPoolClient, error) {
	cc, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}
	return &authPoolClient{
		cc,
		&authClient{cc.ClientConn},
	}, nil
}

// NewAuthPool returns auth pool
func NewAuthPool(addr string, maxConn int) AuthPool {
	return &authPool{
		GRPCPool: pool.NewGRPCPool(addr, maxConn, nil),
	}
}
