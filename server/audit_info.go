package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/paralus/paralus/pkg/query"
	sentryrpc "github.com/paralus/paralus/proto/rpc/sentry"

	"github.com/paralus/paralus/pkg/sentry/kubeconfig"
	"github.com/paralus/paralus/pkg/service"
)

type auditInfoServer struct {
	bs  service.BootstrapService
	aps service.AccountPermissionService
	prs service.ProjectService
}

var _ sentryrpc.AuditInformationServiceServer = (*auditInfoServer)(nil)

// NewAuditInfoServer returns new Audit Information Server
func NewAuditInfoServer(bs service.BootstrapService, aps service.AccountPermissionService, prs service.ProjectService) sentryrpc.AuditInformationServiceServer {
	return &auditInfoServer{bs: bs, aps: aps, prs: prs}
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

	ba, err := s.bs.GetBootstrapAgent(ctx, bat.Metadata.Labels["paralus.dev/connectorAgentTemplate"], query.WithName(clusterID), query.WithIgnoreScopeDefault(), query.WithDeleted())
	if err != nil {
		_log.Infow("unable to get bootstrap agent", "req", req, "error", err)
		return nil, err
	}

	project, err := s.prs.GetByID(ctx, ba.Metadata.Project)
	if err != nil {
		_log.Warnw("unable to get project name", "id", ba.Metadata.Project, "error", err)
		return nil, err
	}

	_log.Infow("project name in lookup cluster", "project", project.Metadata.Name)

	return &sentryrpc.LookupClusterResponse{
		Name:    ba.Metadata.Labels["paralus.dev/clusterName"],
		Project: project.Metadata.Name,
	}, nil
}
