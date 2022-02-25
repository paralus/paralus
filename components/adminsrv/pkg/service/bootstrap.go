package service

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/internal/dao"
	"github.com/RafaySystems/rcloud-base/components/adminsrv/internal/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/converter"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/query"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/sentry/cryptoutil"
	schedulerrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/scheduler"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/RafaySystems/rcloud-base/components/common/proto/types/sentry"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _log = log.GetLogger()
var KEKFunc cryptoutil.PasswordFunc

// BootstrapService is the interface for bootstrap operations
type BootstrapService interface {
	Close() error

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
	RegisterBootstrapAgent(ctx context.Context, token string) error
	DeleteBoostrapAgent(ctx context.Context, templateRef string, opts ...query.Option) error
	PatchBootstrapAgent(ctx context.Context, ba *sentry.BootstrapAgent, templateRef string, opts ...query.Option) error
	UpdateClusterStatus(ctx context.Context, clusterID string) error
}

// bootstrapService implements BootstrapService
type bootstrapService struct {
	dao  pg.EntityDAO
	bdao dao.BootstrapDao
	sp   schedulerrpc.SchedulerPool
}

// NewBootstrapService return new bootstrap service
func NewBootstrapService(db *bun.DB, pool schedulerrpc.SchedulerPool) BootstrapService {
	edao := pg.NewEntityDAO(db)
	return &bootstrapService{
		dao:  edao,
		bdao: dao.NewBootstrapDao(edao),
		sp:   pool,
	}
}

func (s *bootstrapService) PatchBootstrapInfra(ctx context.Context, infra *sentry.BootstrapInfra) error {
	return s.bdao.CreateOrUpdateBootstrapInfra(ctx, convertToInfraModel(infra))
}

func (s *bootstrapService) GetBootstrapInfra(ctx context.Context, name string) (*sentry.BootstrapInfra, error) {

	var bi models.BootstrapInfra
	_, err := s.dao.GetByName(ctx, name, &bi)
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

	return s.bdao.CreateOrUpdateBootstrapAgentTemplate(ctx, &templ)
}

func (s *bootstrapService) GetBootstrapAgentTemplate(ctx context.Context, agentType string) (*sentry.BootstrapAgentTemplate, error) {
	var template models.BootstrapAgentTemplate
	_, err := s.dao.GetByName(ctx, agentType, &template)
	if err != nil {
		return nil, err
	}

	return prepareTemplateResponse(&template), nil
}

func (s *bootstrapService) GetBootstrapAgentTemplateForToken(ctx context.Context, token string) (*sentry.BootstrapAgentTemplate, error) {
	bat, err := s.bdao.GetBootstrapAgentTemplateForToken(ctx, token)
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

	batl, count, err := s.bdao.SelectBootstrapAgentTemplates(ctx, queryOptions)
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
	return s.bdao.CreateBootstrapAgent(ctx, ba)
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
			Description: agent.DisplayName,
			ModifiedAt:  timestamppb.New(agent.ModifiedAt),
			Labels:      lbls,
			Annotations: ann,
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
		Kind: "BootstapAgentTemplate",
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

	agl, count, err := s.bdao.GetBootstrapAgents(ctx, queryOptions, templateRef)
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
	ba, err := s.bdao.GetBootstrapAgent(ctx, templateRef, queryOptions)
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

	agl, count, err := s.bdao.SelectBootstrapAgents(ctx, templateRef, queryOptions)
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

func (s *bootstrapService) RegisterBootstrapAgent(ctx context.Context, token string) error {
	return s.bdao.RegisterBootstrapAgent(ctx, token)
}

func (s *bootstrapService) DeleteBoostrapAgent(ctx context.Context, templateRef string, opts ...query.Option) error {
	queryOptions := &commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(queryOptions)
	}

	err := s.bdao.DeleteBootstrapAgent(ctx, templateRef, queryOptions)
	return err
}

func (s *bootstrapService) PatchBootstrapAgent(ctx context.Context, ba *sentry.BootstrapAgent, templateRef string, opts ...query.Option) error {

	queryOptions := &commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(queryOptions)
	}

	err := s.dao.GetInstance().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		bdb, err := s.bdao.GetBootstrapAgent(ctx, templateRef, queryOptions)
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
		return s.bdao.UpdateBootstrapAgent(ctx, bdb, queryOptions)
	})
	return err
}

func (s *bootstrapService) GetBootstrapAgentForToken(ctx context.Context, token string) (*sentry.BootstrapAgent, error) {
	ba, err := s.bdao.GetBootstrapAgentForToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return prepareAgentResponse(ba), nil
}

func (s *bootstrapService) GetBootstrapAgentTemplateForHost(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error) {
	bat, err := s.bdao.GetBootstrapAgentTemplateForHost(ctx, host)
	if err != nil {
		return nil, err
	}

	return prepareTemplateResponse(bat), nil

}

func (s *bootstrapService) GetBootstrapAgentCountForClusterID(ctx context.Context, clusterID string, orgID string) (int, error) {
	count, err := s.bdao.GetBootstrapAgentCountForClusterID(ctx, clusterID, uuid.MustParse(orgID))
	if err != nil {
		return 0, err
	}
	if count <= 0 {
		return 0, fmt.Errorf("invalid request")
	}
	return count, nil
}

func (s *bootstrapService) GetBootstrapAgentForClusterID(ctx context.Context, clusterID string, orgID string) (*sentry.BootstrapAgent, error) {
	ba, err := s.bdao.GetBootstrapAgentForClusterID(ctx, clusterID, uuid.MustParse(orgID))
	if err != nil || ba == nil {
		return nil, err
	}
	return prepareAgentResponse(ba), nil
}

func getCertAndKey(providedCert, providedKey, password string) ([]byte, []byte, error) {
	var cert, key []byte

	cert = []byte(providedCert)
	key = []byte(providedKey)

	block, _ := pem.Decode([]byte(providedCert))
	if block == nil {
		_log.Infow("failed in pem decode cert", "cert", providedCert)
		return nil, nil, errors.New("failed to pem decode cert")
	}
	_, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		_log.Infow("could not parse cert", "cert", providedCert)
		return nil, nil, errors.Wrapf(err, "could not parse cert")
	}

	if password != "" {
		kekFunc := func() ([]byte, error) {
			return []byte(password), nil
		}
		key, err = reEncryptKeyWithSystemKek([]byte(providedKey), kekFunc, KEKFunc)
		if err != nil {
			_log.Infow("could not encode privatekey with system kek", "error", err)
			return nil, nil, errors.Wrapf(err, "could not encode privatekey with system kek")
		}
	} else {
		key, err = reEncryptKeyWithSystemKek([]byte(providedKey), cryptoutil.NoPassword, KEKFunc)
		if err != nil {
			key, err = reEncryptKeyWithSystemKek([]byte(providedKey), KEKFunc, KEKFunc)
			if err != nil {
				_log.Infow("could not encode privatekey with system kek", "error", err)
				return nil, nil, errors.Wrapf(err, "could not encode privatekey with system kek")
			}
		}
	}

	return cert, key, nil
}

func reEncryptKeyWithSystemKek(privKey []byte, f, f1 cryptoutil.PasswordFunc) ([]byte, error) {
	pk, err := cryptoutil.DecodePrivateKey(privKey, f)
	if err != nil {
		_log.Infow("could not decode PEM block for provided key", "error", err)
		return nil, err
	}

	pk1, err := cryptoutil.EncodePrivateKey(pk, f1)
	if err != nil {
		_log.Infow("could not encode PEM block for provided key", "error", err)
		return nil, err
	}

	return pk1, nil
}

func (s *bootstrapService) reconcileBootstrapInfra(
	ctx context.Context,
	infraRef, relayNetworkName, providedCert, providedKey, password string,
	selfSigned bool,
) ([]byte, []byte, error) {
	var updateInfra bool
	existingBootstrapInfra, err := s.GetBootstrapInfra(ctx, infraRef)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not fetch existing boostrap infra")
	}
	var cert, key []byte
	if !selfSigned {
		if providedCert == "" {
			return cert, key, errors.New("provided certificate is empty")
		}

		if providedKey == "" {
			return cert, key, errors.New("provided key is empty")
		}
		cert, key, err = getCertAndKey(providedCert, providedKey, password)
		if err != nil {
			_log.Infow("could not parse cert and key", "error", err)
			return cert, key, errors.Wrapf(err, "could not parse cert and key")
		}
	} else {
		// Rafay selfsigned usecase
		if existingBootstrapInfra.Spec.CaCert != "" {
			// Check exisitng cert is rafay generated
			parsedCert, err := x509.ParseCertificate([]byte(existingBootstrapInfra.Spec.CaCert))
			if err == nil {
				if parsedCert.Issuer.CommonName == relayNetworkName &&
					parsedCert.Issuer.OrganizationalUnit[0] == "Rafay Sentry" &&
					parsedCert.Issuer.Organization[0] == "Rafay Systems Inc" &&
					parsedCert.Issuer.Locality[0] == "Sunnyvale" &&
					parsedCert.Issuer.Province[0] == "California" {
					// cert is generated by rafay keep using it
					cert = []byte(existingBootstrapInfra.Spec.CaCert)
					key = []byte(existingBootstrapInfra.Spec.CaKey)
				}
			}
		}
		if len(cert) == 0 || len(key) == 0 {
			// Generate new slefsigned key pair
			cert, key, err = cryptoutil.GenerateCA(pkix.Name{
				CommonName:         relayNetworkName,
				Country:            []string{"USA"},
				Organization:       []string{"Rafay Systems Inc"},
				OrganizationalUnit: []string{"Rafay Sentry"},
				Province:           []string{"California"},
				Locality:           []string{"Sunnyvale"},
			}, KEKFunc)

			if err != nil {
				_log.Infow("failed to generate infra CA for server", "error", err)
				return cert, key, errors.Wrap(err, "failed to generate infra CA for server")
			}
		}
	}
	if existingBootstrapInfra.Spec.CaCert != string(cert) {
		_log.Infof("reconciling boostrap infra %s ca cert", infraRef)
		existingBootstrapInfra.Spec.CaCert = string(cert)
		updateInfra = true
	}
	if existingBootstrapInfra.Spec.CaKey != string(key) {
		_log.Infof("reconciling boostrap infra %s key", infraRef)
		existingBootstrapInfra.Spec.CaKey = string(key)
		updateInfra = true
	}

	if updateInfra {
		_log.Infow("reconcileBootstrapInfra", "update-existingBootstrapInfra", existingBootstrapInfra.Metadata.Name)
		if err := s.PatchBootstrapInfra(ctx, existingBootstrapInfra); err != nil {
			_log.Infow("reconcileBootstrapInfra", "could not patch boostrap infra error", err)
			return cert, key, errors.Wrapf(err, "could not patch boostrap infra %s", infraRef)
		}
	}
	return cert, key, nil
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

func (s *bootstrapService) UpdateClusterStatus(ctx context.Context, clusterID string) error {
	client, err := s.sp.NewClient(ctx)
	if err != nil {
		_log.Infow("unable to establish connection with scheduler")
		return err
	}
	cluster := &infrav3.Cluster{
		Metadata: &commonv3.Metadata{
			Id: clusterID,
		},
		Spec: &infrav3.ClusterSpec{
			ClusterData: &infrav3.ClusterData{
				ClusterStatus: &infrav3.ClusterStatus{
					Conditions: []*infrav3.ClusterCondition{
						{
							Type:        infrav3.ClusterConditionType_ClusterCheckIn,
							Status:      commonv3.RafayConditionStatus_Success,
							LastUpdated: timestamppb.Now(),
							Reason:      "Relay agent established connection.",
						},
						{
							Type:        infrav3.ClusterConditionType_ClusterRegister,
							Status:      commonv3.RafayConditionStatus_Success,
							LastUpdated: timestamppb.Now(),
							Reason:      "Relay agent established connection.",
						},
					},
				},
			},
		},
	}
	_, err = client.UpdateClusterStatus(ctx, cluster)
	return err
}

func (s *bootstrapService) Close() error {
	return s.dao.Close()
}
