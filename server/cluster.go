package server

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/paralus/paralus/internal/cluster/fixtures"
	"github.com/paralus/paralus/pkg/common"
	"github.com/paralus/paralus/pkg/query"
	"github.com/paralus/paralus/pkg/service"
	rpcv3 "github.com/paralus/paralus/proto/rpc/scheduler"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrapbv3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type clusterServer struct {
	service.ClusterService
	downloadData common.DownloadData
}

// NewClusterServer returns new cluster server implementation
func NewClusterServer(es service.ClusterService, data *common.DownloadData) rpcv3.ClusterServiceServer {
	return &clusterServer{
		ClusterService: es,
		downloadData:   *data,
	}
}

func updateClusterStatus(req *infrapbv3.Cluster, resp *infrapbv3.Cluster, err error) *infrapbv3.Cluster {
	if err != nil {
		req.Status = &v3.Status{
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return req
	}
	resp.Status = &v3.Status{ConditionStatus: v3.ConditionStatus_StatusOK}
	return resp
}

func (s *clusterServer) CreateCluster(ctx context.Context, req *infrapbv3.Cluster) (*infrapbv3.Cluster, error) {
	resp, err := s.Create(ctx, req)
	return updateClusterStatus(req, resp, err), err
}

func (s *clusterServer) GetClusters(ctx context.Context, req *commonv3.QueryOptions) (*infrapbv3.ClusterList, error) {
	return s.List(ctx, query.WithOptions(req))
}

func (s *clusterServer) GetCluster(ctx context.Context, req *infrapbv3.Cluster) (*infrapbv3.Cluster, error) {
	resp, err := s.Select(ctx, req, true)
	return updateClusterStatus(req, resp, err), err
}

func (s *clusterServer) DeleteCluster(ctx context.Context, e *infrapbv3.Cluster) (*rpcv3.DeleteClusterResponse, error) {
	err := s.Delete(ctx, e)
	if err != nil {
		return nil, err
	}
	return &rpcv3.DeleteClusterResponse{}, nil
}

func (s *clusterServer) UpdateCluster(ctx context.Context, req *infrapbv3.Cluster) (*infrapbv3.Cluster, error) {
	resp, err := s.Update(ctx, req)
	return updateClusterStatus(req, resp, err), err
}

func (s *clusterServer) DownloadCluster(ctx context.Context, cluster *infrapbv3.Cluster) (*commonv3.HttpBody, error) {
	c, err := s.Select(ctx, cluster, true)
	if err != nil {
		return nil, err
	}
	bb := new(bytes.Buffer)

	if c.Spec.ProxyConfig != nil {
		if c.Spec.ProxyConfig.BootstrapCA != "" {
			c.Spec.ProxyConfig.BootstrapCA = base64.StdEncoding.EncodeToString([]byte(c.Spec.ProxyConfig.BootstrapCA))
		}
	}

	err = fixtures.DownloadTemplate.Execute(bb, struct {
		DownloadData common.DownloadData
		Cluster      *infrapbv3.Cluster
	}{
		s.downloadData, c,
	})

	if err != nil {
		return nil, err
	}

	return &commonv3.HttpBody{
		ContentType: "application/x-paralus-yaml",
		Data:        bb.Bytes(),
	}, nil
}

func (s *clusterServer) UpdateClusterStatus(ctx context.Context, request *rpcv3.UpdateClusterStatusRequest) (*rpcv3.UpdateClusterStatusResponse, error) {
	err := s.UpdateClusterConditionStatus(ctx, &infrav3.Cluster{
		Metadata: request.Metadata,
		Spec: &infrav3.ClusterSpec{
			ClusterData: &infrav3.ClusterData{
				ClusterStatus: request.ClusterStatus,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return &rpcv3.UpdateClusterStatusResponse{}, nil
}

func (s *clusterServer) GetClusterStatus(ctx context.Context, request *rpcv3.GetClusterStatusRequest) (*rpcv3.GetClusterStatusResponse, error) {
	cluster, err := s.Get(ctx, query.WithMeta(request.Metadata))
	if err != nil {
		return nil, err
	}
	return &rpcv3.GetClusterStatusResponse{
		Metadata:      cluster.GetMetadata(),
		ClusterStatus: cluster.GetSpec().GetClusterData().GetClusterStatus(),
	}, nil
}
