package server

import (
	"context"

	configrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/config"
	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/sentry/authz"
	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/service"
)

type clusterAuthzServer struct {
	bs  service.BootstrapService
	aps service.AccountPermissionService
	gps service.GroupPermissionService
	krs service.KubeconfigRevocationService
	kcs service.KubectlClusterSettingsService
	kss service.KubeconfigSettingService
	//apn   models.AccountProjectNamespaceService
	cPool configrpc.ConfigPool
}

// GetUserAuthorization return authorization profile of user for a given cluster
func (s *clusterAuthzServer) GetUserAuthorization(ctx context.Context, req *sentryrpc.GetUserAuthorizationRequest) (*sentryrpc.GetUserAuthorizationResponse, error) {
	resp, err := authz.GetAuthorization(ctx, req, s.bs, s.aps, s.gps, s.krs, s.kcs, s.kss, s.cPool)
	if err != nil {
		_log.Errorw("error getting auth profile", "req", req, "error", err.Error())
		return nil, err
	}
	return resp, nil
}

// NewClusterAuthzServer returns New ClusterAuthzServer
func NewClusterAuthzServer(bs service.BootstrapService, aps service.AccountPermissionService, gps service.GroupPermissionService, krs service.KubeconfigRevocationService, kcs service.KubectlClusterSettingsService, kss service.KubeconfigSettingService, cPool configrpc.ConfigPool) sentryrpc.ClusterAuthorizationServer {
	return &clusterAuthzServer{
		bs:    bs,
		aps:   aps,
		gps:   gps,
		krs:   krs,
		kcs:   kcs,
		kss:   kss,
		cPool: cPool,
	}
}
