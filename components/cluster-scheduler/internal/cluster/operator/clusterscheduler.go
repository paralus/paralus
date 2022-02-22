package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/common"
	rgrpc "github.com/RafaySystems/rcloud-base/components/common/pkg/grpc"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/scheduler"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var (
	clusterClient rpcv3.ClusterClient = nil
	_log                              = log.GetLogger()
)

func NewClusterGRPCClient() *grpc.ClientConn {
	clusterSchedulerHost := viper.GetString(common.ClusterSchedulerHost)
	clusterSchedulerPort := viper.GetInt(common.ClusterSchedulerPort)

	_log.Infow("Creating the cluster scheduler grpc client ",
		"clusterSchedulerHost", clusterSchedulerHost,
		"clusterSchedulerPort", clusterSchedulerPort)

	if clusterSchedulerHost == "" || clusterSchedulerPort == 0 {
		_log.Infow("Cluster scheduler is not configured")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	grpc, err := rgrpc.NewGrpcClientClientConn(ctx,
		clusterSchedulerHost,
		clusterSchedulerPort)

	if err != nil {
		_log.Errorw("Error while creating grpc client", "Error", err)
		return nil
	}

	return grpc
}

func NewClusterClient() (rpcv3.ClusterClient, error) {
	if clusterClient != nil {
		return clusterClient, nil
	}

	if grpc := NewClusterGRPCClient(); grpc != nil {
		clusterClient = rpcv3.NewClusterClient(grpc)
	}

	// Still nil? At least notify the user that a client could not be procued
	if clusterClient == nil {
		return nil, fmt.Errorf("error getting handle to cluster grpc client")
	}

	return clusterClient, nil
}
