package rpcv3

import (
	"context"

	"github.com/RafayLabs/rcloud-base/pkg/pool"
	grpcpool "github.com/processout/grpc-go-pool"
)

// SchedulerClient is the interface for accessing all the RPCs
// exposed by Cluster Scheduler
type SchedulerClient interface {
	Unhealthy()
	Close() error
	ClusterClient
}

type schedulerClient struct {
	*grpcpool.ClientConn
	*clusterClient
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
		&clusterClient{cc.ClientConn},
	}, nil
}
