package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/internal/constants"
	"github.com/RafaySystems/rcloud-base/pkg/sentry/util"
	"github.com/RafaySystems/rcloud-base/pkg/service"
	sentryrpc "github.com/RafaySystems/rcloud-base/proto/rpc/sentry"
	"github.com/RafaySystems/rcloud-base/proto/types/sentry"
	"google.golang.org/grpc/metadata"
)

type kubectlClusterSettingsServer struct {
	bs  service.BootstrapService
	kcs service.KubectlClusterSettingsService
}

var _ sentryrpc.KubectlClusterSettingsServer = (*kubectlClusterSettingsServer)(nil)

func (s *kubectlClusterSettingsServer) GetKubectlClusterSettings(ctx context.Context, in *sentryrpc.GetKubectlClusterSettingsRequest) (*sentryrpc.GetKubectlClusterSettingsResponse, error) {
	opts := in.Opts

	clusterID, err := util.GetClusterScope(opts.UrlScope)
	if err != nil {
		_log.Infow("get kubectl cluster settings failed to get clusterID", "opts", opts)
		return nil, err
	}

	cnt, err := s.bs.GetBootstrapAgentCountForClusterID(ctx, clusterID, opts.Organization)
	if err != nil {
		_log.Infow("get kubectl cluster settings invalid request", "opts", opts, "cluster", clusterID)
		return nil, err
	}

	_log.Infow("get kubectl cluster settings ", "cnt", cnt, "opts", opts, "clusterID", clusterID)

	kc, err := s.kcs.Get(ctx, opts.Organization, clusterID)
	if err == constants.ErrNotFound {
		return &sentryrpc.GetKubectlClusterSettingsResponse{DisableWebKubectl: false, DisableCLIKubectl: false}, nil
	} else if err != nil {
		return nil, err
	}
	return &sentryrpc.GetKubectlClusterSettingsResponse{DisableWebKubectl: kc.DisableWebKubectl, DisableCLIKubectl: kc.DisableCLIKubectl}, nil
}

func (s *kubectlClusterSettingsServer) UpdateKubectlClusterSettings(ctx context.Context, in *sentryrpc.UpdateKubectlClusterSettingsRequest) (*sentryrpc.UpdateKubectlClusterSettingsResponse, error) {
	var clusterName, userAgent, host, remoteAddr string
	opts := in.Opts

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ua := md.Get("rafay-gateway-user-agent")
		if len(ua) > 0 {
			userAgent = ua[0]
		}
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		h := md.Get("rafay-gateway-host")
		if len(h) > 0 {
			host = h[0]
		}
	}

	remoteAddr = "127.0.0.1" //default
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ra := md.Get("rafay-gateway-remote-addr")
		if len(ra) > 0 {
			remoteAddr = ra[0]
		}
	}

	clusterID, err := util.GetClusterScope(opts.UrlScope)
	if err != nil {
		_log.Infow("update kubectl cluster settings failed to get clusterID", "opts", opts)
		return nil, err
	}

	_log.Infow("update kubectl cluster settings ", "opts", opts, "clusterID", clusterID)

	_, err = s.bs.GetBootstrapAgentCountForClusterID(ctx, clusterID, opts.Organization)
	if err != nil {
		_log.Infow("update kubectl cluster settings invalid request", "opts", opts, "cluster", clusterID)
		return nil, err
	}

	clusterName = ""
	ba, _ := s.bs.GetBootstrapAgentForClusterID(ctx, clusterID, opts.Organization)
	if ba != nil {
		clusterName = ba.Metadata.Labels["rafay.dev/clusterName"]
	}

	err = s.kcs.Patch(ctx, &sentry.KubectlClusterSettings{
		Name:              clusterID,
		OrganizationID:    opts.Organization,
		PartnerID:         opts.Partner,
		DisableWebKubectl: in.DisableWebKubectl,
		DisableCLIKubectl: in.DisableCLIKubectl,
	})
	if err != nil {
		return nil, err
	}

	_log.Infow("updated kubectl cluster setting with values ", clusterName, userAgent, host, remoteAddr)

	/*TODO: to be done with events
	partnerID := opts.Partner
	orgIDString := opts.Organization
	kubectlSettingEvent("cluster.kubectl.setting", clusterID, orgIDString, partnerID, opts.Username, opts.AccountID.String(), clusterName, userAgent, host, remoteAddr, opts.Groups, in.DisableWebKubectl, in.DisableCLIKubectl)
	*/

	return &sentryrpc.UpdateKubectlClusterSettingsResponse{}, nil
}

// NewKubectlClusterSettingsServer returns new kubectl cluster setting server
func NewKubectlClusterSettingsServer(bs service.BootstrapService, kcs service.KubectlClusterSettingsService) sentryrpc.KubectlClusterSettingsServer {
	return &kubectlClusterSettingsServer{bs, kcs}
}
