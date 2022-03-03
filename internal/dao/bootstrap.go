package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/internal/random"
	"github.com/RafaySystems/rcloud-base/pkg/query"
	commonv3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	"github.com/RafaySystems/rcloud-base/proto/types/sentry"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// BootstrapDao is the interface for bootstrap operations
type BootstrapDao interface {
	CreateOrUpdateBootstrapInfra(ctx context.Context, infra *models.BootstrapInfra) error
	CreateOrUpdateBootstrapAgentTemplate(context.Context, *models.BootstrapAgentTemplate) error
	GetBootstrapAgentTemplateForToken(ctx context.Context, token string) (*models.BootstrapAgentTemplate, error)
	SelectBootstrapAgentTemplates(ctx context.Context, opts *commonv3.QueryOptions) (ret []models.BootstrapAgentTemplate, count int, err error)
	DeleteBootstrapAgentTempate(ctx context.Context, opts *commonv3.QueryOptions, infraRef string) error
	GetBootstrapAgents(ctx context.Context, opts *commonv3.QueryOptions, templateRef string) (ret []models.BootstrapAgent, count int, err error)
	CreateBootstrapAgent(ctx context.Context, ba *models.BootstrapAgent) error
	GetBootstrapAgent(ctx context.Context, templateRef string, opts *commonv3.QueryOptions) (*models.BootstrapAgent, error)
	SelectBootstrapAgents(ctx context.Context, templateRef string, opts *commonv3.QueryOptions) (ret []models.BootstrapAgent, count int, err error)
	RegisterBootstrapAgent(ctx context.Context, token string) error
	DeleteBootstrapAgent(ctx context.Context, templateRef string, opts *commonv3.QueryOptions) error
	UpdateBootstrapAgent(ctx context.Context, ba *models.BootstrapAgent, opts *commonv3.QueryOptions) error
	GetBootstrapAgentForToken(ctx context.Context, token string) (*models.BootstrapAgent, error)
	GetBootstrapAgentTemplateForHost(ctx context.Context, host string) (*models.BootstrapAgentTemplate, error)
	GetBootstrapAgentCountForClusterID(ctx context.Context, clusterID string, orgID uuid.UUID) (int, error)
	GetBootstrapAgentForClusterID(ctx context.Context, clusterID string, orgID uuid.UUID) (*models.BootstrapAgent, error)
	UpdateBootstrapAgentDeleteAt(ctx context.Context, templateRef string) error
	UpdateBootstrapAgentTempateDeleteAt(ctx context.Context, opts *commonv3.QueryOptions) error
	UpdateBootstrapInfraDeleteAt(ctx context.Context, opts *commonv3.QueryOptions) error
}

// bootstrapDao implements BootstrapDao
type bootstrapDao struct {
	bdao pg.EntityDAO
}

// BootstrapDao return new bootstrap dao
func NewBootstrapDao(edao pg.EntityDAO) BootstrapDao {
	return &bootstrapDao{
		bdao: edao,
	}
}

func (s *bootstrapDao) CreateOrUpdateBootstrapInfra(ctx context.Context, infra *models.BootstrapInfra) error {

	_, err := s.bdao.GetInstance().NewInsert().On("CONFLICT (name) DO UPDATE").
		Set("ca_cert = ?", infra.CaCert).
		Set("ca_key = ?", infra.CaKey).
		Set("modified_at = ?", time.Now()).
		Model(infra).
		Where("bi.name = ?", infra.Name).Exec(ctx)
	return err
}

func (s *bootstrapDao) CreateOrUpdateBootstrapAgentTemplate(ctx context.Context, template *models.BootstrapAgentTemplate) error {

	_, err := s.bdao.GetInstance().NewInsert().On("CONFLICT (name) DO UPDATE").
		Set("infra_ref = ?", template.InfraRef).
		Set("ignore_multiple_register = ?", template.IgnoreMultipleRegister).
		Set("auto_register = ?", template.AutoRegister).
		Set("incluster_template = ?", template.InclusterTemplate).
		Set("outofcluster_template = ?", template.OutofclusterTemplate).
		Set("hosts = ?", template.Hosts).
		Set("labels = ?", template.Labels).
		Set("annotations = ?", template.Annotations).
		Set("modified_at = ?", time.Now()).
		Model(template).
		Where("bat.name = ?", template.Name).Exec(ctx)
	return err
}

func (s *bootstrapDao) GetBootstrapAgentTemplateForToken(ctx context.Context, token string) (*models.BootstrapAgentTemplate, error) {
	var template models.BootstrapAgentTemplate
	err := s.bdao.GetInstance().NewSelect().Model(&template).Where("token = ?", token).Scan(ctx)
	return &template, err
}

func (s *bootstrapDao) SelectBootstrapAgentTemplates(ctx context.Context, opts *commonv3.QueryOptions) (ret []models.BootstrapAgentTemplate, count int, err error) {
	q, err := query.Select(s.bdao.GetInstance().NewSelect().Model(&ret), opts)
	if err != nil {
		return
	}

	q = query.Paginate(q, opts)

	count, err = q.ScanAndCount(ctx)

	return
}

func (s *bootstrapDao) DeleteBootstrapAgentTempate(ctx context.Context, opts *commonv3.QueryOptions, infraRef string) error {

	q, err := query.Delete(s.bdao.GetInstance().NewSelect().Model((*models.BootstrapAgentTemplate)(nil)), opts)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx)
	return err
}

func (s *bootstrapDao) GetBootstrapAgent(ctx context.Context, templateRef string, opts *commonv3.QueryOptions) (*models.BootstrapAgent, error) {
	var ba models.BootstrapAgent
	q, err := query.Get(s.bdao.GetInstance().NewSelect().Model(&ba), opts)
	if err != nil {
		return nil, err
	}

	// - will be used to select all the agents
	// to cross template boundaries
	if templateRef != "-" {
		q = q.Where("template_ref = ?", templateRef)
	}

	err = q.Scan(ctx)

	return &ba, err
}

func (s *bootstrapDao) GetBootstrapAgents(ctx context.Context, opts *commonv3.QueryOptions, templateRef string) (ret []models.BootstrapAgent, count int, err error) {
	q, err := query.Get(s.bdao.GetInstance().NewSelect().Model(&ret), opts)
	if err != nil {
		return nil, 0, err
	}

	if templateRef != "-" {
		q = q.Where("template_ref = ?", templateRef)
	}

	count, err = q.ScanAndCount(ctx)

	return
}

func (s *bootstrapDao) SelectBootstrapAgents(ctx context.Context, templateRef string, opts *commonv3.QueryOptions) (ret []models.BootstrapAgent, count int, err error) {

	q, err := query.Select(s.bdao.GetInstance().NewSelect().Model(&ret), opts)
	if err != nil {
		return
	}

	if templateRef != "-" {
		q = q.Where("template_ref = ?", templateRef)
	}

	count, err = q.ScanAndCount(ctx)

	return
}

func (s *bootstrapDao) CreateBootstrapAgent(ctx context.Context, ba *models.BootstrapAgent) error {
	ba.TokenState = sentry.BootstrapAgentState_NotRegistered.String()
	_, err := s.bdao.Create(ctx, ba)
	return err
}

func (s *bootstrapDao) RegisterBootstrapAgent(ctx context.Context, token string) error {

	err := s.bdao.GetInstance().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		ba, err := s.getBootstrapAgentForToken(ctx, token)
		if err != nil {
			return err
		}

		bat, err := s.getBootstrapAgentTemplate(ctx, ba.TemplateRef)
		if err != nil {
			return err
		}

		state := sentry.BootstrapAgentState_NotApproved
		if bat.AutoApprove {
			state = sentry.BootstrapAgentState_Approved
		}

		switch ba.TokenState {
		case sentry.BootstrapAgentState_NotRegistered.String():
			ba.TokenState = sentry.BootstrapAgentState_Approved.String()
		case sentry.BootstrapAgentState_NotApproved.String(), sentry.BootstrapAgentState_Approved.String():
			if !bat.IgnoreMultipleRegister {
				return fmt.Errorf("cannot register token %s state is %s", token, ba.TokenState)
			}
		default:
			return fmt.Errorf("invalid token state %s", ba.TokenState)
		}

		_, err = s.bdao.GetInstance().NewUpdate().Model(ba).
			Set("token_state = ?", state).
			Where("token = ?", token).
			Exec(ctx)

		return err
	})

	return err
}

func (s *bootstrapDao) getBootstrapAgentForToken(ctx context.Context, token string) (*models.BootstrapAgent, error) {
	var ba models.BootstrapAgent
	err := s.bdao.GetInstance().NewSelect().Model(&ba).Where("token = ?", token).Scan(ctx)
	return &ba, err
}

func (s *bootstrapDao) getBootstrapAgentTemplate(ctx context.Context, name string) (*models.BootstrapAgentTemplate, error) {
	var template models.BootstrapAgentTemplate
	err := s.bdao.GetInstance().NewSelect().Model(&template).Where("name = ?", name).Scan(ctx)
	return &template, err
}

func (s *bootstrapDao) DeleteBootstrapAgent(ctx context.Context, templateRef string, opts *commonv3.QueryOptions) error {

	dq := s.bdao.GetInstance().NewDelete().Model((*models.BootstrapAgent)(nil)).Where("name = ?", opts.ID)
	if templateRef != "" {
		dq = dq.Where("template_ref = ?", templateRef)
	}
	_, err := dq.Exec(ctx)
	return err
}

func (s *bootstrapDao) UpdateBootstrapAgent(ctx context.Context, ba *models.BootstrapAgent, opts *commonv3.QueryOptions) error {
	_, err := s.bdao.GetInstance().NewUpdate().Model(ba).Where("id = ?", ba.ID).Returning("*").Exec(ctx)
	return err

}

func (s *bootstrapDao) GetBootstrapAgentForToken(ctx context.Context, token string) (*models.BootstrapAgent, error) {
	var ba models.BootstrapAgent
	err := s.bdao.GetInstance().NewSelect().Model(&ba).Where("token = ?", token).Scan(ctx)
	return &ba, err
}

func (s *bootstrapDao) GetBootstrapAgentTemplateForHost(ctx context.Context, host string) (*models.BootstrapAgentTemplate, error) {

	bat := models.BootstrapAgentTemplate{}
	err := s.bdao.GetInstance().NewSelect().Model(&bat).
		ColumnExpr("bat.*").
		Join("JOIN sentry_bootstrap_template_host as bth").
		JoinOn("bat.name = bth.name").
		JoinOn("bth.host = ?", host).
		Scan(ctx)

	return &bat, err
}

func (s *bootstrapDao) GetBootstrapAgentCountForClusterID(ctx context.Context, clusterID string, orgID uuid.UUID) (int, error) {
	var ba []models.BootstrapAgent
	err := s.bdao.GetInstance().NewSelect().Model(&ba).
		Where("name = ?", clusterID).
		Where("organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		return 0, err
	}
	return len(ba), nil
}

func (s *bootstrapDao) GetBootstrapAgentForClusterID(ctx context.Context, clusterID string, orgID uuid.UUID) (*models.BootstrapAgent, error) {
	var ba models.BootstrapAgent
	err := s.bdao.GetInstance().NewSelect().Model(&ba).
		Where("name = ?", clusterID).
		Where("organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &ba, nil
}

// updateBootstrapAgentDeleteAt builds query for deleting resource
func (s *bootstrapDao) UpdateBootstrapAgentDeleteAt(ctx context.Context, templateRef string) error {
	var toBeDeletedAgent *models.BootstrapAgent
	_, err := s.bdao.GetX(ctx, "template_ref", templateRef, &toBeDeletedAgent)
	if err != nil {
		return err
	}

	if toBeDeletedAgent == nil {
		return errors.New("could not find bootstrap agent to delete")
	}

	opts := &commonv3.QueryOptions{}
	query.WithName(toBeDeletedAgent.Name)(opts)
	q, err := query.Update(s.bdao.GetInstance().NewUpdate().Model((*models.BootstrapAgent)(nil)), opts)
	if err != nil {
		return err
	}
	now := time.Now()
	q = q.Set("name = ?", fmt.Sprintf("%s-%s", opts.Name, random.NewRandomString(10)))
	q = q.Set("deleted_at = ?", &now)
	_, err = q.Exec(ctx)
	return err
}

func (s *bootstrapDao) UpdateBootstrapAgentTempateDeleteAt(ctx context.Context, opts *commonv3.QueryOptions) error {
	q, err := query.Update(s.bdao.GetInstance().NewUpdate().Model((*models.BootstrapAgentTemplate)(nil)), opts)
	if err != nil {
		return err
	}
	now := time.Now()
	q = q.Set("name = ?", fmt.Sprintf("%s-%s", opts.Name, random.NewRandomString(10)))
	q = q.Set("deleted_at = ?", &now)
	_, err = q.Exec(ctx)
	return err
}

// updateBootstrapInfraDeleteAt builds query for deleting resource
func (s *bootstrapDao) UpdateBootstrapInfraDeleteAt(ctx context.Context, opts *commonv3.QueryOptions) error {
	q, err := query.Update(s.bdao.GetInstance().NewUpdate().Model((*models.BootstrapInfra)(nil)), opts)
	if err != nil {
		return err
	}
	now := time.Now()
	q = q.Set("deleted_at = ?", &now)
	_, err = q.Exec(ctx)
	return err
}
