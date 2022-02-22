package server

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/internal/fixtures"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/bootstrapper"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/credentials"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/service"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/query"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/scheduler"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrapbv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
)

type clusterServer struct {
	service.ClusterService
	credentials.Signer
	downloadData bootstrapper.DownloadData
}

// NewClusterServer returns new cluster server implementation
func NewClusterServer(es service.ClusterService, signer credentials.Signer, data *bootstrapper.DownloadData) rpcv3.ClusterServer {
	return &clusterServer{
		ClusterService: es,
		Signer:         signer,
		downloadData:   *data,
	}
}

func (s *clusterServer) CreateCluster(ctx context.Context, e *infrapbv3.Cluster) (*infrapbv3.Cluster, error) {
	edge, err := s.Create(ctx, e)
	if err != nil {
		return nil, err
	}
	return edge, nil
}

func (s *clusterServer) GetClusters(ctx context.Context, e *commonv3.QueryOptions) (*infrapbv3.ClusterList, error) {
	clusters, err := s.List(ctx, query.WithOptions(e))
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

func (s *clusterServer) GetCluster(ctx context.Context, e *infrapbv3.Cluster) (*infrapbv3.Cluster, error) {
	cluster, err := s.Select(ctx, e, true)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (s *clusterServer) DeleteCluster(ctx context.Context, e *infrapbv3.Cluster) (*rpcv3.DeleteClusterResponse, error) {
	err := s.Delete(ctx, e)
	if err != nil {
		return nil, err
	}
	return &rpcv3.DeleteClusterResponse{}, nil
}

func (s *clusterServer) UpdateCluster(ctx context.Context, e *infrapbv3.Cluster) (*infrapbv3.Cluster, error) {
	edge, err := s.Update(ctx, e)
	if err != nil {
		return nil, err
	}
	return edge, nil
}

func (s *clusterServer) RegisterCluster(ctx context.Context, rcr *rpcv3.RegisterClusterRequest) (*rpcv3.RegisterClusterResponse, error) {
	cert, err := s.Sign(rcr.SigningRequest)
	if err != nil {
		return nil, err
	}

	_, err = s.Register(ctx, rcr.Token)
	if err != nil {
		return nil, err
	}

	return &rpcv3.RegisterClusterResponse{
		Certificate:   cert,
		CaCertificate: s.GetCACert(),
	}, nil
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
		DownloadData bootstrapper.DownloadData
		Cluster      *infrapbv3.Cluster
	}{
		s.downloadData, c,
	})

	if err != nil {
		return nil, err
	}

	return &commonv3.HttpBody{
		ContentType: "application/x-rafay-yaml",
		Data:        bb.Bytes(),
	}, nil
}

func (s *clusterServer) UpdateClusterStatus(ctx context.Context, cluster *infrapbv3.Cluster) (*infrapbv3.Cluster, error) {
	err := s.UpdateClusterConditionStatus(ctx, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}
