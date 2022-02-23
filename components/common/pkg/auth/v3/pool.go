package authv3

import (
	"github.com/RafaySystems/rcloud-base/components/common/pkg/pool"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/v3"

	"context"

	grpcpool "github.com/processout/grpc-go-pool"
)

// AuthPoolClient is the interface for auth pool client
type AuthPoolClient interface {
	Unhealthy()
	Close() error
	rpcv3.AuthClient
}

type authPoolClient struct {
	*grpcpool.ClientConn
	rpcv3.AuthClient
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
		rpcv3.NewAuthClient(cc.ClientConn),
	}, nil
}

// NewAuthPool returns auth pool
func NewAuthPool(addr string, maxConn int) AuthPool {
	return &authPool{
		GRPCPool: pool.NewGRPCPool(addr, maxConn, nil),
	}
}
