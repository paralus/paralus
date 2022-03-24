package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/RafayLabs/rcloud-base/pkg/query"
	sentryrpc "github.com/RafayLabs/rcloud-base/proto/rpc/sentry"

	"github.com/RafayLabs/rcloud-base/pkg/sentry/kubeconfig"
	"github.com/RafayLabs/rcloud-base/pkg/service"
)

type auditInfoServer struct {
	bs  service.BootstrapService
	aps service.AccountPermissionService
}

var _ sentryrpc.AuditInformationServer = (*auditInfoServer)(nil)

// NewAuditInfoServer returns new Audit Information Server
func NewAuditInfoServer(bs service.BootstrapService, aps service.AccountPermissionService) sentryrpc.AuditInformationServer {
	return &auditInfoServer{bs: bs, aps: aps}
}

func (s *auditInfoServer) LookupUser(ctx context.Context, req *sentryrpc.LookupUserRequest) (*sentryrpc.LookupUserResponse, error) {

	attrs := kubeconfig.GetCNAttributes(req.UserCN)

	_log.Infow("lookupUser", "attrs", attrs)

	if attrs.SystemUser {
		return &sentryrpc.LookupUserResponse{
			UserName:       attrs.Username,
			AccountID:      attrs.AccountID,
			IsSSO:          kubeconfig.GetStringFromBool(attrs.IsSSO),
			OrganizationID: attrs.OrganizationID,
			PartnerID:      attrs.PartnerID,
			SessionType:    kubeconfig.GetSessionTypeString(attrs.SessionType),
		}, nil
	}

	accountID := attrs.AccountID

	var userName string
	account, err := s.aps.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	userName = account.Username

	return &sentryrpc.LookupUserResponse{
		UserName:       userName,
		AccountID:      attrs.AccountID,
		IsSSO:          kubeconfig.GetStringFromBool(attrs.IsSSO),
		OrganizationID: attrs.OrganizationID,
		PartnerID:      attrs.PartnerID,
		SessionType:    kubeconfig.GetSessionTypeString(attrs.SessionType),
	}, nil
}

func (s *auditInfoServer) LookupCluster(ctx context.Context, req *sentryrpc.LookupClusterRequest) (*sentryrpc.LookupClusterResponse, error) {

	strs := strings.SplitN(req.ClusterSNI, ".", 2)
	if len(strs) != 2 {
		return nil, fmt.Errorf("invalid cluster SNI %s", req.ClusterSNI)
	}

	clusterID, relayHost := strs[0], strs[1]

	relayHost = fmt.Sprintf("*.%s", relayHost)

	bat, err := s.bs.GetBootstrapAgentTemplateForHost(ctx, relayHost)
	if err != nil {
		return nil, err
	}

	ba, err := s.bs.GetBootstrapAgent(ctx, bat.Metadata.Labels["rafay.dev/connectorAgentTemplate"], query.WithName(clusterID), query.WithIgnoreScopeDefault(), query.WithDeleted())
	if err != nil {
		_log.Infow("unable to get bootstrap agent", "req", req, "error", err)
		return nil, err
	}

	return &sentryrpc.LookupClusterResponse{
		Name: ba.Metadata.Labels["rafay.dev/clusterName"],
	}, nil
}
