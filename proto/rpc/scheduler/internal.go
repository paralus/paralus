package rpcv3

import (
	"context"

	"github.com/paralus/paralus/pkg/pool"
	grpcpool "github.com/processout/grpc-go-pool"
)

// ClusterClient is the interface for accessing all the RPCs
// exposed by Cluster service
type ClusterClient interface {
	Unhealthy()
	Close() error
	ClusterServiceClient
}

type clusterClient struct {
	*grpcpool.ClientConn
	*clusterServiceClient
}

var _ ClusterClient = (*clusterClient)(nil)

// ClusterPool maintains pool of grpc connections to cluster service
type ClusterPool interface {
	Close()
	NewClient(ctx context.Context) (ClusterClient, error)
}

// NewClusterPool new cluster pool
func NewClusterPool(addr string, maxConn int) ClusterPool {
	return &clusterPool{
		GRPCPool: pool.NewGRPCPool(addr, maxConn, nil),
	}
}

type clusterPool struct {
	*pool.GRPCPool
}

func (p *clusterPool) Close() {
	if p.GRPCPool != nil {
		p.GRPCPool.Close()
	}
}

func (p *clusterPool) NewClient(ctx context.Context) (ClusterClient, error) {
	cc, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}
	return &clusterClient{
		cc,
		&clusterServiceClient{cc.ClientConn},
	}, nil
}
