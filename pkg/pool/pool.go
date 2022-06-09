package pool

import (
	"context"
	"sync"
	"time"

	grpcpool "github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// grpc pool constants
const (
	DefaultMaxPoolConn = 20
)

// GRPCPool holds all the shared state for GRPC pool
type GRPCPool struct {
	capacity int
	creds    credentials.TransportCredentials
	addr     string
	*grpcpool.Pool
	m sync.Mutex
}

// NewGRPCPool returns new auth pool
func NewGRPCPool(addr string, maxConnections int, creds credentials.TransportCredentials) *GRPCPool {
	// min number of connections for grpc across all services is
	// set to 20; any service creating a connection pool size < 20
	// is now defaulted to 20
	if maxConnections < DefaultMaxPoolConn {
		maxConnections = DefaultMaxPoolConn
	}
	return &GRPCPool{addr: addr, capacity: maxConnections, creds: creds}
}

// GetConnection returns new connection from grpc pool
func (gp *GRPCPool) GetConnection(ctx context.Context) (*grpcpool.ClientConn, error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if gp.Pool == nil {
		gp.m.Lock()
		defer gp.m.Unlock()

		var pool *grpcpool.Pool
		var err error
		if gp.creds == nil {
			pool, err = newPool(gp.addr, gp.capacity)
		} else {
			pool, err = newSecurePool(gp.addr, gp.capacity, gp.creds)
		}
		if err != nil {
			return nil, err
		}

		gp.Pool = pool
	}
	cc, err := gp.Pool.Get(nCtx)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

// newPool returns new grpc connection pool for given host port and capacity
func newPool(target string, capacity int) (*grpcpool.Pool, error) {
	pool, err := grpcpool.New(func() (*grpc.ClientConn, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		cc, err := grpc.DialContext(
			ctx,
			target,
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                time.Second * 30,
				Timeout:             time.Second * 30,
				PermitWithoutStream: true,
			}),
		)
		return cc, err
	}, 1, capacity, time.Second*60)
	return pool, err
}

// newSecurePool returns new grpc connection pool for given host port and credentials
func newSecurePool(target string, capacity int, creds credentials.TransportCredentials) (*grpcpool.Pool, error) {
	pool, err := grpcpool.New(func() (*grpc.ClientConn, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		cc, err := grpc.DialContext(
			ctx,
			target,
			grpc.WithTransportCredentials(creds),
			grpc.WithBlock(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                time.Second * 30,
				Timeout:             time.Second * 30,
				PermitWithoutStream: true,
			}),
		)

		return cc, err
	}, 1, capacity, time.Second*60)
	return pool, err

}
