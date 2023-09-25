package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/converter"
	"github.com/paralus/paralus/pkg/query"
	"github.com/paralus/paralus/pkg/sentry/cryptoutil"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var KEKFunc cryptoutil.PasswordFunc

// BootstrapService is the interface for bootstrap operations
type BootstrapService interface {
	// bootstrap infra methods
	PatchBootstrapInfra(ctx context.Context, infra *sentry.BootstrapInfra) error
	GetBootstrapInfra(ctx context.Context, name string) (*sentry.BootstrapInfra, error)
	// bootstrap template methods
	PatchBootstrapAgentTemplate(ctx context.Context, template *sentry.BootstrapAgentTemplate) error
	GetBootstrapAgentTemplate(ctx context.Context, name string) (*sentry.BootstrapAgentTemplate, error)
	GetBootstrapAgentTemplateForToken(ctx context.Context, token string) (*sentry.BootstrapAgentTemplate, error)
	GetBootstrapAgentTemplateForHost(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error)
	SelectBootstrapAgentTemplates(ctx context.Context, opts ...query.Option) (*sentry.BootstrapAgentTemplateList, error)
	// bootstrap agent methods
	CreateBootstrapAgent(ctx context.Context, agent *sentry.BootstrapAgent) error
	GetBootstrapAgent(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error)
	GetBootstrapAgents(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgentList, error)
	GetBootstrapAgentForToken(ctx context.Context, token string) (*sentry.BootstrapAgent, error)
	GetBootstrapAgentCountForClusterID(ctx context.Context, clusterID string, orgID string) (int, error)
	GetBootstrapAgentForClusterID(ctx context.Context, clusterID string, orgID string) (*sentry.BootstrapAgent, error)
	SelectBootstrapAgents(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgentList, error)
	RegisterBootstrapAgent(ctx context.Context, token, ip, fingerprint string) error
	DeleteBootstrapAgent(ctx context.Context, templateRef string, opts ...query.Option) error
	PatchBootstrapAgent(ctx context.Context, ba *sentry.BootstrapAgent, templateRef string, opts ...query.Option) error
}

// bootstrapService implements BootstrapService
type bootstrapService struct {
	db *bun.DB
}

// NewBootstrapService return new bootstrap service
func NewBootstrapService(db *bun.DB) BootstrapService {
	return &bootstrapService{db}
}

func (s *bootstrapService) PatchBootstrapInfra(ctx context.Context, infra *sentry.BootstrapInfra) error {
	return dao.CreateOrUpdateBootstrapInfra(ctx, s.db, convertToInfraModel(infra))
}

func (s *bootstrapService) GetBootstrapInfra(ctx context.Context, name string) (*sentry.BootstrapInfra, error) {

	var bi models.BootstrapInfra
	_, err := dao.GetByName(ctx, s.db, name, &bi)
	if err != nil {
		return nil, err
	}
	return prepareInfraResponse(&bi), nil
}

func (s *bootstrapService) PatchBootstrapAgentTemplate(ctx context.Context, template *sentry.BootstrapAgentTemplate) error {
	templ := models.BootstrapAgentTemplate{
		Name:                   template.Metadata.Name,
		DisplayName:            template.Metadata.DisplayName,
		InfraRef:               template.Spec.InfraRef,
		ModifiedAt:             time.Now(),
		Labels:                 converter.ConvertToJsonRawMessage(template.Metadata.Labels),
		Annotations:            converter.ConvertToJsonRawMessage(template.Metadata.Annotations),
		AutoRegister:           template.Spec.AutoRegister,
		AutoApprove:            template.Spec.AutoApprove,
		TemplateType:           sentry.BootstrapAgentTemplateType_name[int32(template.Spec.TemplateType)],
		IgnoreMultipleRegister: template.Spec.IgnoreMultipleRegister,
		InclusterTemplate:      template.Spec.InClusterTemplate,
		OutofclusterTemplate:   template.Spec.OutOfClusterTemplate,
		Token:                  template.Spec.Token,
		Hosts:                  converter.ConvertToJsonRawMessage(template.Spec.Hosts),
		CreatedAt:              time.Now(),
	}

	return dao.CreateOrUpdateBootstrapAgentTemplate(ctx, s.db, &templ)
}

func (s *bootstrapService) GetBootstrapAgentTemplate(ctx context.Context, agentType string) (*sentry.BootstrapAgentTemplate, error) {
	var template models.BootstrapAgentTemplate
	_, err := dao.GetByName(ctx, s.db, agentType, &template)
	if err != nil {
		return nil, err
	}

	return prepareTemplateResponse(&template), nil
}

func (s *bootstrapService) GetBootstrapAgentTemplateForToken(ctx context.Context, token string) (*sentry.BootstrapAgentTemplate, error) {
	bat, err := dao.GetBootstrapAgentTemplateForToken(ctx, s.db, token)
	if err != nil {
		return nil, err
	}
	return prepareTemplateResponse(bat), nil
}

func (s *bootstrapService) SelectBootstrapAgentTemplates(ctx context.Context, opts ...query.Option) (*sentry.BootstrapAgentTemplateList, error) {
	queryOptions := &commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(queryOptions)
	}

	batl, count, err := dao.SelectBootstrapAgentTemplates(ctx, s.db, queryOptions)
	if err != nil {
		return nil, err
	}

	ret := &sentry.BootstrapAgentTemplateList{Metadata: &commonv3.ListMetadata{
		Count: int64(count),
	}}
	for _, bat := range batl {
		ret.Items = append(ret.Items, prepareTemplateResponse(&bat))
	}

	return ret, nil
}

func (s *bootstrapService) CreateBootstrapAgent(ctx context.Context, agent *sentry.BootstrapAgent) error {
	ba := convertToAgentModel(agent)
	ba.CreatedAt = time.Now()
	return dao.CreateBootstrapAgent(ctx, s.db, ba)
}

func convertToAgentModel(agent *sentry.BootstrapAgent) *models.BootstrapAgent {
	agentMdl := &models.BootstrapAgent{
		Name:        agent.Metadata.Name,
		TemplateRef: agent.Spec.TemplateRef,
		AgentMode:   agent.Spec.AgentMode.String(),
		DisplayName: agent.Metadata.DisplayName,
		Labels:      converter.ConvertToJsonRawMessage(agent.Metadata.Labels),
		Annotations: converter.ConvertToJsonRawMessage(agent.Metadata.Annotations),
		Token:       agent.Spec.Token,
	}
	if orgId, err := uuid.Parse(agent.Metadata.Organization); err == nil {
		agentMdl.OrganizationId = orgId
	}
	if partId, err := uuid.Parse(agent.Metadata.Partner); err == nil {
		agentMdl.PartnerId = partId
	}
	if projId, err := uuid.Parse(agent.Metadata.Project); err == nil {
		agentMdl.ProjectId = projId
	}
	return agentMdl
}

func convertToInfraModel(infra *sentry.BootstrapInfra) *models.BootstrapInfra {
	return &models.BootstrapInfra{
		Name:        infra.Metadata.Name,
		ModifiedAt:  time.Now(),
		CaCert:      infra.Spec.CaCert,
		CaKey:       infra.Spec.CaKey,
		DisplayName: infra.Metadata.DisplayName,
		Labels:      converter.ConvertToJsonRawMessage(infra.Metadata.Labels),
		Annotations: converter.ConvertToJsonRawMessage(infra.Metadata.Annotations),
	}
}

func prepareAgentResponse(agent *models.BootstrapAgent) *sentry.BootstrapAgent {
	var lbls map[string]string
	if agent.Labels != nil {
		json.Unmarshal(agent.Labels, &lbls)
	}
	var ann map[string]string
	if agent.Annotations != nil {
		json.Unmarshal(agent.Annotations, &ann)
	}
	ba := &sentry.BootstrapAgent{
		Kind: "BootstrapAgent",
		Metadata: &commonv3.Metadata{
			Name:        agent.Name,
			DisplayName: agent.DisplayName,
			Description: agent.DisplayName,
			ModifiedAt:  timestamppb.New(agent.ModifiedAt),
			Labels:      lbls,
			Annotations: ann,
			Project:     agent.ProjectId.String(),
		},
		Spec: &sentry.BootstrapAgentSpec{
			Token:       agent.Token,
			TemplateRef: agent.TemplateRef,
			AgentMode:   sentry.BootstrapAgentMode(sentry.BootstrapAgentMode_value[agent.AgentMode]),
		},
		Status: &sentry.BootStrapAgentStatus{
			TokenState:    sentry.BootstrapAgentState(sentry.BootstrapAgentMode_value[agent.TokenState]),
			IpAddress:     agent.IPAddress,
			LastCheckedIn: timestamppb.New(agent.LastCheckedIn),
			Fingerprint:   agent.Fingerprint,
		},
	}
	return ba
}

func prepareInfraResponse(infra *models.BootstrapInfra) *sentry.BootstrapInfra {
	var lbls map[string]string
	if infra.Labels != nil {
		json.Unmarshal(infra.Labels, &lbls)
	}
	var ann map[string]string
	if infra.Annotations != nil {
		json.Unmarshal(infra.Annotations, &ann)
	}
	bi := &sentry.BootstrapInfra{
		Kind: "BootstrapInfra",
		Metadata: &commonv3.Metadata{
			Name:        infra.Name,
			DisplayName: infra.DisplayName,
			ModifiedAt:  timestamppb.New(infra.ModifiedAt),
			Labels:      lbls,
			Annotations: ann,
		},
		Spec: &sentry.BootstrapInfraSpec{
			CaCert: infra.CaCert,
			CaKey:  infra.CaKey,
		},
	}
	return bi
}

func prepareTemplateResponse(template *models.BootstrapAgentTemplate) *sentry.BootstrapAgentTemplate {
	var lbls map[string]string
	if template.Labels != nil {
		json.Unmarshal(template.Labels, &lbls)
	}
	var ann map[string]string
	if template.Annotations != nil {
		json.Unmarshal(template.Annotations, &ann)
	}
	var hosts []*sentry.BootstrapTemplateHost
	if template.Hosts != nil {
		json.Unmarshal(template.Hosts, &hosts)
	}
	templResp := sentry.BootstrapAgentTemplate{
		Kind: "BootstrapAgentTemplate",
		Metadata: &commonv3.Metadata{
			Name:        template.Name,
			DisplayName: template.DisplayName,
			Labels:      lbls,
			Annotations: ann,
			ModifiedAt:  timestamppb.New(template.ModifiedAt),
		},
		Spec: &sentry.BootstrapAgentTemplateSpec{
			InfraRef:               template.InfraRef,
			AutoRegister:           template.AutoRegister,
			AutoApprove:            template.AutoApprove,
			IgnoreMultipleRegister: template.IgnoreMultipleRegister,
			TemplateType:           sentry.BootstrapAgentTemplateType(sentry.BootstrapAgentTemplateType_value[template.TemplateType]),
			Token:                  template.Token,
			Hosts:                  hosts,
			InClusterTemplate:      template.InclusterTemplate,
			OutOfClusterTemplate:   template.OutofclusterTemplate,
		},
	}
	return &templResp
}

func (s *bootstrapService) GetBootstrapAgents(ctx context.Context, templateRef string, opts ...query.Option) (ret *sentry.BootstrapAgentList, err error) {
	queryOptions := &commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(queryOptions)
	}

	agl, count, err := dao.GetBootstrapAgents(ctx, s.db, queryOptions, templateRef)
	if err != nil {
		return nil, err
	}

	ret = new(sentry.BootstrapAgentList)
	ret.Metadata = &commonv3.ListMetadata{
		Count: int64(count),
	}
	for _, ag := range agl {
		ret.Items = append(ret.Items, prepareAgentResponse(&ag))
	}
	return
}

func (s *bootstrapService) GetBootstrapAgent(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error) {
	queryOptions := &commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(queryOptions)
	}
	ba, err := dao.GetBootstrapAgent(ctx, s.db, templateRef, queryOptions)
	if err != nil {
		return nil, err
	}

	return prepareAgentResponse(ba), nil
}

func (s *bootstrapService) SelectBootstrapAgents(ctx context.Context, templateRef string, opts ...query.Option) (ret *sentry.BootstrapAgentList, err error) {
	queryOptions := &commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(queryOptions)
	}

	agl, count, err := dao.SelectBootstrapAgents(ctx, s.db, templateRef, queryOptions)
	if err != nil {
		return nil, err
	}

	ret = new(sentry.BootstrapAgentList)
	ret.Metadata = &commonv3.ListMetadata{
		Count: int64(count),
	}
	for _, ag := range agl {
		ret.Items = append(ret.Items, prepareAgentResponse(&ag))
	}

	return
}

func (s *bootstrapService) RegisterBootstrapAgent(ctx context.Context, token, ip, fingerprint string) error {
	err := s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		return dao.RegisterBootstrapAgent(ctx, tx, token, ip, fingerprint)
	})
	return err
}

func (s *bootstrapService) DeleteBootstrapAgent(ctx context.Context, templateRef string, opts ...query.Option) error {
	queryOptions := &commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(queryOptions)
	}

	err := dao.DeleteBootstrapAgent(ctx, s.db, templateRef, queryOptions)
	return err
}

func (s *bootstrapService) PatchBootstrapAgent(ctx context.Context, ba *sentry.BootstrapAgent, templateRef string, opts ...query.Option) error {

	queryOptions := &commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(queryOptions)
	}

	err := s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		bdb, err := dao.GetBootstrapAgent(ctx, s.db, templateRef, queryOptions)
		if err != nil {
			return err
		}
		if bdb.TokenState > sentry.BootstrapAgentState_NotSet.String() {
			bdb.TokenState = ba.Status.TokenState.String()
		}
		if ba.Status != nil {
			if ba.Status.IpAddress != "" {
				bdb.IPAddress = ba.Status.IpAddress
			} else {
				bdb.IPAddress = ""
			}
			if !ba.Status.LastCheckedIn.AsTime().IsZero() {
				bdb.LastCheckedIn = ba.Status.LastCheckedIn.AsTime()
			}
			if ba.Status.Fingerprint != "" {
				bdb.Fingerprint = ba.Status.Fingerprint
			} else {
				bdb.Fingerprint = ""
			}
		}
		bdb.ModifiedAt = time.Now()
		bdb.DisplayName = ba.Metadata.DisplayName
		return dao.UpdateBootstrapAgent(ctx, s.db, bdb, queryOptions)
	})
	return err
}

func (s *bootstrapService) GetBootstrapAgentForToken(ctx context.Context, token string) (*sentry.BootstrapAgent, error) {
	ba, err := dao.GetBootstrapAgentForToken(ctx, s.db, token)
	if err != nil {
		return nil, err
	}
	return prepareAgentResponse(ba), nil
}

func (s *bootstrapService) GetBootstrapAgentTemplateForHost(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error) {
	bat, err := dao.GetBootstrapAgentTemplateForHost(ctx, s.db, host)
	if err != nil {
		return nil, err
	}

	return prepareTemplateResponse(bat), nil

}

func (s *bootstrapService) GetBootstrapAgentCountForClusterID(ctx context.Context, clusterID string, orgID string) (int, error) {
	count, err := dao.GetBootstrapAgentCountForClusterID(ctx, s.db, clusterID, uuid.MustParse(orgID))
	if err != nil {
		return 0, err
	}
	if count <= 0 {
		return 0, fmt.Errorf("invalid request")
	}
	return count, nil
}

func (s *bootstrapService) GetBootstrapAgentForClusterID(ctx context.Context, clusterID string, orgID string) (*sentry.BootstrapAgent, error) {
	ba, err := dao.GetBootstrapAgentForClusterID(ctx, s.db, clusterID, uuid.MustParse(orgID))
	if err != nil || ba == nil {
		return nil, err
	}
	return prepareAgentResponse(ba), nil
}

func (s *bootstrapService) GetRelayAgent(ctx context.Context, ClusterScope string, opts ...query.Option) (*sentry.BootstrapAgent, error) {
	queryOptions := &commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(queryOptions)
	}

	bal, err := s.SelectBootstrapAgents(ctx, queryOptions.Name,
		query.WithOrganizationID(queryOptions.Organization),
		query.WithPartnerID(queryOptions.Partner),
	)
	if err != nil {
		_log.Infow("failed to get default bootstrap agent list", "cluster", ClusterScope, "error", err)
		return nil, err
	}

	if bal != nil && bal.Metadata.Count > 0 {
		var ba sentry.BootstrapAgent
		found := false
		// match labels
		for _, b := range bal.Items {
			_log.Infow("match", "ClusterScope", ClusterScope, "DisplayName", b.Metadata.DisplayName)
			if "cluster/"+b.Metadata.DisplayName == ClusterScope {
				found = true
				ba = *b
				break
			}
		}
		if found {
			// found bootstrap relay agent for cluster as per the association
			return &ba, nil
		} else {
			_log.Infow("did not find relay bootstrap agent for", "cluster", ClusterScope, "template", queryOptions.Name)
		}
	}
	_log.Infow("did not find relay bootstrap agent for", "cluster", ClusterScope, "template", queryOptions.Name)
	return nil, fmt.Errorf("failed to get relay agent")
}
