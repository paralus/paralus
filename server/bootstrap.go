package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/paralus/paralus/pkg/service"

	"github.com/paralus/paralus/pkg/gateway"
	"github.com/paralus/paralus/pkg/log"
	"github.com/paralus/paralus/pkg/query"
	sentryrpc "github.com/paralus/paralus/proto/rpc/sentry"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/paralus/paralus/proto/types/sentry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/paralus/paralus/pkg/sentry/cryptoutil"
	"github.com/paralus/paralus/pkg/sentry/util"
)

var _log = log.GetLogger()

type RelayNetworkDownloadData struct {
	SentryAddr     string
	PeerHost       string
	PeerToken      string
	UserToken      string
	ConnectorToken string
	RelayImage     string
	UserHostPort   string
	RelayHostPort  string
	RelayNetworkID string
}

type RelayAgentDownloadData struct {
	ClusterID     string
	SentryAddr    string
	Token         string
	RelayImage    string
	HostName      string
	TemplateName  string
	TemplateToken string
}

type bootstrapServer struct {
	bs       service.BootstrapService
	passFunc cryptoutil.PasswordFunc
	cs       service.ClusterService
}

var _ sentryrpc.BootstrapServiceServer = (*bootstrapServer)(nil)

func (s *bootstrapServer) GetBootstrapInfra(ctx context.Context, in *sentry.BootstrapInfra) (*sentry.BootstrapInfra, error) {
	return s.bs.GetBootstrapInfra(ctx, in.Metadata.Name)
}

func (s *bootstrapServer) PatchBootstrapInfra(ctx context.Context, in *sentry.BootstrapInfra) (*sentry.BootstrapInfra, error) {
	err := s.bs.PatchBootstrapInfra(ctx, in)
	return in, err
}

func (s *bootstrapServer) PatchBootstrapAgentTemplate(ctx context.Context, in *sentry.BootstrapAgentTemplate) (*sentry.BootstrapAgentTemplate, error) {
	err := s.bs.PatchBootstrapAgentTemplate(ctx, in)
	return in, err
}

func (s *bootstrapServer) GetBootstrapAgentTemplate(ctx context.Context, in *sentry.BootstrapAgentTemplate) (*sentry.BootstrapAgentTemplate, error) {
	return s.bs.GetBootstrapAgentTemplate(ctx, in.Metadata.Name)
}

func (s *bootstrapServer) CreateBootstrapAgent(ctx context.Context, in *sentry.BootstrapAgent) (*sentry.BootstrapAgent, error) {
	err := s.bs.CreateBootstrapAgent(ctx, in)

	return in, err
}

func (s *bootstrapServer) GetBootstrapAgent(ctx context.Context, in *sentry.BootstrapAgent) (ret *sentry.BootstrapAgent, err error) {
	templateRef, err := util.GetTemplateScope(in.Spec.TemplateRef)
	if err != nil {
		return
	}

	ret, err = s.bs.GetBootstrapAgent(ctx, templateRef, query.WithMeta(in.Metadata))
	if err == sql.ErrNoRows {
		err = status.Error(codes.NotFound, err.Error())
	}
	return
}

func (s *bootstrapServer) GetBootstrapAgentTemplates(ctx context.Context, qo *commonv3.QueryOptions) (*sentry.BootstrapAgentTemplateList, error) {
	return s.bs.SelectBootstrapAgentTemplates(ctx, query.WithOptions(qo))
}

func (s *bootstrapServer) GetBootstrapAgents(ctx context.Context, in *sentryrpc.GetBootstrapAgentsRequest) (ret *sentry.BootstrapAgentList, err error) {
	templateRef, err := util.GetTemplateScope(in.TemplateScope)
	if err != nil {
		return
	}

	ret, err = s.bs.SelectBootstrapAgents(ctx, templateRef, query.WithOptions(in.Opts))
	if err != nil {
		return nil, err
	}

	return
}

func (s *bootstrapServer) DeleteBootstrapAgent(ctx context.Context, in *sentry.BootstrapAgent) (ret *sentryrpc.DeleteBootstrapAgentResponse, err error) {
	templateRef, err := util.GetTemplateScope(in.Spec.TemplateRef)
	if err != nil {
		return
	}

	err = s.bs.DeleteBootstrapAgent(ctx, templateRef, query.WithMeta(in.Metadata))
	if err == sql.ErrNoRows {
		err = status.Error(codes.NotFound, err.Error())
	}

	return &sentryrpc.DeleteBootstrapAgentResponse{}, err
}

func (s *bootstrapServer) UpdateBootstrapAgent(ctx context.Context, in *sentry.BootstrapAgent) (ret *sentry.BootstrapAgent, err error) {

	templateRef, err := util.GetTemplateScope(in.Spec.TemplateRef)
	if err != nil {
		return
	}

	err = s.bs.PatchBootstrapAgent(ctx, in, templateRef, query.WithMeta(in.Metadata))
	if err == sql.ErrNoRows {
		err = status.Error(codes.NotFound, err.Error())
	}

	ret = in

	return
}

func (s *bootstrapServer) RegisterBootstrapAgent(ctx context.Context, in *sentryrpc.RegisterAgentRequest) (resp *sentryrpc.RegisterAgentResponse, err error) {
	_log.Infow("received agent register", "request", *in)

	resp = &sentryrpc.RegisterAgentResponse{}

	token, err := util.GetTemplateScope(in.TemplateToken)
	if err != nil {
		_log.Error(err.Error())
		return
	}

	var template *sentry.BootstrapAgentTemplate
	// bypass for auto registering core relay
	if (token == "-" || token == "cd-relay") && !gateway.IsGatewayRequest(ctx) {
		template, err = s.bs.GetBootstrapAgentTemplate(ctx, in.TemplateName)
		if err != nil {
			_log.Error(err.Error())
			return
		}
	} else {
		template, err = s.bs.GetBootstrapAgentTemplateForToken(ctx, token)
		if err != nil {
			_log.Error(err.Error())
			return
		}
	}

	infra, err := s.bs.GetBootstrapInfra(ctx, template.Spec.InfraRef)
	if err != nil {
		_log.Error(err.Error())
		return
	}

	var signer cryptoutil.Signer
	var opts []cryptoutil.Option

	opts = append(opts, cryptoutil.WithCAKeyDecrypt(s.passFunc))

	// only add altname for server or mixed templates
	if template.Spec.TemplateType == sentry.BootstrapAgentTemplateType_Server || template.Spec.TemplateType == sentry.BootstrapAgentTemplateType_Mixed {
		for _, host := range template.Spec.Hosts {
			h, _ := util.ParseAddr(host.Host)
			opts = append(opts, cryptoutil.WithAltName(h))
		}
	}

	if template.Spec.TemplateType == sentry.BootstrapAgentTemplateType_Client {
		opts = append(opts, cryptoutil.WithClient(), cryptoutil.WithCSRSubjectValidate(cryptoutil.CNShouldBe(in.Token)))
	} else if template.Spec.TemplateType == sentry.BootstrapAgentTemplateType_Server {
		opts = append(opts, cryptoutil.WithServer())
	} else if template.Spec.TemplateType == sentry.BootstrapAgentTemplateType_Mixed {
		opts = append(opts, cryptoutil.WithServer(), cryptoutil.WithClient())
	}

	signer, err = cryptoutil.NewSigner([]byte(infra.Spec.CaCert), []byte(infra.Spec.CaKey), opts...)
	if err != nil {
		_log.Errorw("error getting cert signer", "error", err.Error())
		return
	}

	signed, err := signer.Sign(in.Csr)
	if err != nil {
		_log.Error(err.Error())
		return
	}

	var agent *sentry.BootstrapAgent
	agent, err = s.bs.GetBootstrapAgentForToken(ctx, in.Token)

	// if agent is not found and template has auto register
	if err == sql.ErrNoRows && template.Spec.AutoRegister {
		agent = &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{
				Name: in.Name,
			}, Spec: &sentry.BootstrapAgentSpec{Token: in.Token,
				TemplateRef: template.Metadata.Name,
			},
		}

		err = s.bs.CreateBootstrapAgent(ctx, agent)
		if err != nil {
			_log.Error(err.Error())
			return
		}
	} else {
		if err != nil {
			//agent is nil
			_log.Error(err.Error())
			return
		}
	}

	if agent.Spec.TemplateRef != template.Metadata.Name {
		err = fmt.Errorf("token %s cannot be registered for template %s", in.Token, in.TemplateToken)
		_log.Error(err.Error())
		return
	}

	err = s.bs.RegisterBootstrapAgent(ctx, in.Token, in.IpAddress, in.Fingerprint)
	if err != nil {
		_log.Error(err.Error())
		return
	}

	resp.Certificate = signed
	resp.CaCertificate = []byte(infra.Spec.CaCert)

	if template.Metadata.Name == "paralus-core-relay-agent" {
		_log.Info("updating cluster status for :: ", agent.Metadata.Name)
		err = s.updateClusterStatus(ctx, agent.Metadata.Name, agent.Metadata.Project)
	}

	return
}

func (s *bootstrapServer) updateClusterStatus(ctx context.Context, clusterID, projectID string) error {
	cluster := &infrav3.Cluster{
		Metadata: &commonv3.Metadata{
			Id:      clusterID,
			Project: projectID,
		},
		Spec: &infrav3.ClusterSpec{
			ClusterData: &infrav3.ClusterData{
				ClusterStatus: &infrav3.ClusterStatus{
					Conditions: []*infrav3.ClusterCondition{
						{
							Type:        infrav3.ClusterConditionType_ClusterCheckIn,
							Status:      commonv3.ParalusConditionStatus_Success,
							LastUpdated: timestamppb.Now(),
							Reason:      "Relay agent established connection.",
						},
						{
							Type:        infrav3.ClusterConditionType_ClusterRegister,
							Status:      commonv3.ParalusConditionStatus_Success,
							LastUpdated: timestamppb.Now(),
							Reason:      "Relay agent established connection.",
						},
						{
							Type:        infrav3.ClusterConditionType_ClusterReady,
							Status:      commonv3.ParalusConditionStatus_Success,
							LastUpdated: timestamppb.Now(),
							Reason:      "Relay agent established connection.",
						},
					},
				},
			},
		},
	}
	return s.cs.UpdateClusterConditionStatus(ctx, cluster)
}

func (s *bootstrapServer) GetBootstrapAgentConfig(ctx context.Context, in *sentry.BootstrapAgent) (*commonv3.HttpBody, error) {
	return nil, nil
}

// NewBootstrapServer return new bootstrap server
func NewBootstrapServer(bs service.BootstrapService, f cryptoutil.PasswordFunc, cs service.ClusterService) sentryrpc.BootstrapServiceServer {
	return &bootstrapServer{bs, f, cs}
}
