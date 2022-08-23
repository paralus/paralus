package rpcv3

import (
	"context"

	"github.com/paralus/paralus/pkg/pool"
	grpcpool "github.com/processout/grpc-go-pool"
)

// SchedulerClient is the interface for accessing all the RPCs
// exposed by Cluster Scheduler
type SchedulerClient interface {
	Unhealthy()
	Close() error
	ClusterServiceClient
}

type schedulerClient struct {
	*grpcpool.ClientConn
	*clusterServiceClient
}

var _ SchedulerClient = (*schedulerClient)(nil)

// SchedulerPool maintains pool of grpc connections to scheduler service
type SchedulerPool interface {
	Close()
	NewClient(ctx context.Context) (SchedulerClient, error)
}

// NewSchedulerPool new scheduler pool
func NewSchedulerPool(addr string, maxConn int) SchedulerPool {
	return &schedulerPool{
		GRPCPool: pool.NewGRPCPool(addr, maxConn, nil),
	}
}

type schedulerPool struct {
	*pool.GRPCPool
}

func (p *schedulerPool) Close() {
	if p.GRPCPool != nil {
		p.GRPCPool.Close()
	}
}

func (p *schedulerPool) NewClient(ctx context.Context) (SchedulerClient, error) {
	cc, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}
	return &schedulerClient{
		cc,
		&clusterServiceClient{cc.ClientConn},
	}, nil
}
