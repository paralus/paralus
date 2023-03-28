package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/internal/random"
	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/uptrace/bun"
)

func CreateOrUpdateBootstrapInfra(ctx context.Context, db bun.IDB, infra *models.BootstrapInfra) error {

	_, err := db.NewInsert().On("CONFLICT (name) DO UPDATE").
		Set("ca_cert = ?", infra.CaCert).
		Set("ca_key = ?", infra.CaKey).
		Set("modified_at = ?", time.Now()).
		Model(infra).
		Where("bi.name = ?", infra.Name).Exec(ctx)
	return err
}

func CreateOrUpdateBootstrapAgentTemplate(ctx context.Context, db bun.IDB, template *models.BootstrapAgentTemplate) error {

	_, err := db.NewInsert().On("CONFLICT (name) DO UPDATE").
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

func GetBootstrapAgentTemplateForToken(ctx context.Context, db bun.IDB, token string) (*models.BootstrapAgentTemplate, error) {
	var template models.BootstrapAgentTemplate
	err := db.NewSelect().Model(&template).Where("token = ?", token).Scan(ctx)
	return &template, err
}

func SelectBootstrapAgentTemplates(ctx context.Context, db bun.IDB, opts *commonv3.QueryOptions) (ret []models.BootstrapAgentTemplate, count int, err error) {
	q, err := query.Select(db.NewSelect().Model(&ret), opts)
	if err != nil {
		return
	}

	q = query.Paginate(q, opts)

	count, err = q.ScanAndCount(ctx)

	return
}

func DeleteBootstrapAgentTempate(ctx context.Context, db bun.IDB, opts *commonv3.QueryOptions, infraRef string) error {

	q, err := query.Delete(db.NewSelect().Model((*models.BootstrapAgentTemplate)(nil)), opts)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx)
	return err
}

func GetBootstrapAgent(ctx context.Context, db bun.IDB, templateRef string, opts *commonv3.QueryOptions) (*models.BootstrapAgent, error) {
	var ba models.BootstrapAgent
	q, err := query.Get(db.NewSelect().Model(&ba), opts)
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

func GetBootstrapAgents(ctx context.Context, db bun.IDB, opts *commonv3.QueryOptions, templateRef string) (ret []models.BootstrapAgent, count int, err error) {
	q, err := query.Get(db.NewSelect().Model(&ret), opts)
	if err != nil {
		return nil, 0, err
	}

	if templateRef != "-" {
		q = q.Where("template_ref = ?", templateRef)
	}

	count, err = q.ScanAndCount(ctx)

	return
}

func SelectBootstrapAgents(ctx context.Context, db bun.IDB, templateRef string, opts *commonv3.QueryOptions) (ret []models.BootstrapAgent, count int, err error) {

	q, err := query.Select(db.NewSelect().Model(&ret), opts)
	if err != nil {
		return
	}

	if templateRef != "-" {
		q = q.Where("template_ref = ?", templateRef)
	}

	count, err = q.ScanAndCount(ctx)

	return
}

func CreateBootstrapAgent(ctx context.Context, db bun.IDB, ba *models.BootstrapAgent) error {
	ba.TokenState = sentry.BootstrapAgentState_NotRegistered.String()
	_, err := Create(ctx, db, ba)
	return err
}

func RegisterBootstrapAgent(ctx context.Context, db bun.Tx, token, ip, fingerprint string) error {
	ba, err := getBootstrapAgentForToken(ctx, db, token)
	if err != nil {
		return err
	}

	bat, err := getBootstrapAgentTemplate(ctx, db, ba.TemplateRef)
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
		} else if ba.Fingerprint != fingerprint {
			return fmt.Errorf("fingerprint mismatch for token %s", token)
		}
	default:
		return fmt.Errorf("invalid token state %s", ba.TokenState)
	}

	_, err = db.NewUpdate().Model(ba).
		Set("token_state = ?", state).
		Set("fingerprint = ?", fingerprint).
		Set("ip_address = ?", ip).
		Where("token = ?", token).
		Exec(ctx)

	return err
}

func getBootstrapAgentForToken(ctx context.Context, db bun.IDB, token string) (*models.BootstrapAgent, error) {
	var ba models.BootstrapAgent
	err := db.NewSelect().Model(&ba).Where("token = ?", token).Scan(ctx)
	return &ba, err
}

func getBootstrapAgentTemplate(ctx context.Context, db bun.IDB, name string) (*models.BootstrapAgentTemplate, error) {
	var template models.BootstrapAgentTemplate
	err := db.NewSelect().Model(&template).Where("name = ?", name).Scan(ctx)
	return &template, err
}

func DeleteBootstrapAgent(ctx context.Context, db bun.IDB, templateRef string, opts *commonv3.QueryOptions) error {

	dq := db.NewDelete().Model((*models.BootstrapAgent)(nil)).Where("name = ?", opts.ID)
	if templateRef != "" {
		dq = dq.Where("template_ref = ?", templateRef)
	}
	_, err := dq.Exec(ctx)
	return err
}

func UpdateBootstrapAgent(ctx context.Context, db bun.IDB, ba *models.BootstrapAgent, opts *commonv3.QueryOptions) error {
	_, err := db.NewUpdate().Model(ba).Where("id = ?", ba.ID).Returning("*").Exec(ctx)
	return err

}

func GetBootstrapAgentForToken(ctx context.Context, db bun.IDB, token string) (*models.BootstrapAgent, error) {
	var ba models.BootstrapAgent
	err := db.NewSelect().Model(&ba).Where("token = ?", token).Scan(ctx)
	return &ba, err
}

func GetBootstrapAgentTemplateForHost(ctx context.Context, db bun.IDB, host string) (*models.BootstrapAgentTemplate, error) {

	bat := models.BootstrapAgentTemplate{}
	err := db.NewSelect().Model(&bat).
		ColumnExpr("bat.*").
		Join("JOIN sentry_bootstrap_template_host as bth").
		JoinOn("bat.name = bth.name").
		JoinOn("bth.host = ?", host).
		Scan(ctx)

	return &bat, err
}

func GetBootstrapAgentCountForClusterID(ctx context.Context, db bun.IDB, clusterID string, orgID uuid.UUID) (int, error) {
	var ba []models.BootstrapAgent
	err := db.NewSelect().Model(&ba).
		Where("name = ?", clusterID).
		Where("organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		return 0, err
	}
	return len(ba), nil
}

func GetBootstrapAgentForClusterID(ctx context.Context, db bun.IDB, clusterID string, orgID uuid.UUID) (*models.BootstrapAgent, error) {
	var ba models.BootstrapAgent
	err := db.NewSelect().Model(&ba).
		Where("name = ?", clusterID).
		Where("organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &ba, nil
}

// updateBootstrapAgentDeleteAt builds query for deleting resource
func UpdateBootstrapAgentDeleteAt(ctx context.Context, db bun.IDB, templateRef string) error {
	var toBeDeletedAgent *models.BootstrapAgent
	_, err := GetX(ctx, db, "template_ref", templateRef, &toBeDeletedAgent)
	if err != nil {
		return err
	}

	if toBeDeletedAgent == nil {
		return errors.New("could not find bootstrap agent to delete")
	}

	opts := &commonv3.QueryOptions{}
	query.WithName(toBeDeletedAgent.Name)(opts)
	q, err := query.Update(db.NewUpdate().Model((*models.BootstrapAgent)(nil)), opts)
	if err != nil {
		return err
	}
	now := time.Now()
	q = q.Set("name = ?", fmt.Sprintf("%s-%s", opts.Name, random.NewRandomString(10)))
	q = q.Set("deleted_at = ?", &now)
	_, err = q.Exec(ctx)
	return err
}

func UpdateBootstrapAgentTempateDeleteAt(ctx context.Context, db bun.IDB, opts *commonv3.QueryOptions) error {
	q, err := query.Update(db.NewUpdate().Model((*models.BootstrapAgentTemplate)(nil)), opts)
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
func UpdateBootstrapInfraDeleteAt(ctx context.Context, db bun.IDB, opts *commonv3.QueryOptions) error {
	q, err := query.Update(db.NewUpdate().Model((*models.BootstrapInfra)(nil)), opts)
	if err != nil {
		return err
	}
	now := time.Now()
	q = q.Set("deleted_at = ?", &now)
	_, err = q.Exec(ctx)
	return err
}
