package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	bun "github.com/uptrace/bun"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PartnerService is the interface for partner operations
type PartnerService interface {
	// create partner
	Create(ctx context.Context, partner *systemv3.Partner) (*systemv3.Partner, error)
	// get partner by id
	GetByID(ctx context.Context, partnerId string) (*systemv3.Partner, error)
	// get partner by id
	GetByName(ctx context.Context, name string) (*systemv3.Partner, error)
	// create or update partner
	Update(ctx context.Context, partner *systemv3.Partner) (*systemv3.Partner, error)
	// delete partner
	Delete(ctx context.Context, partner *systemv3.Partner) (*systemv3.Partner, error)
	// list partner
	GetOnlyPartner(ctx context.Context) (*systemv3.Partner, error)
}

// partnerService implements PartnerService
type partnerService struct {
	db *bun.DB
	al *zap.Logger
}

// NewPartnerService return new partner service
func NewPartnerService(db *bun.DB, al *zap.Logger) PartnerService {
	return &partnerService{db, al}
}

func (s *partnerService) Create(ctx context.Context, partner *systemv3.Partner) (*systemv3.Partner, error) {

	var sb []byte
	if partner.GetSpec().GetSettings() != nil {
		sb = json.RawMessage(partner.GetSpec().GetSettings().String())
	}
	//convert v3 spec to internal models
	part := models.Partner{
		Name:                      partner.GetMetadata().GetName(),
		Description:               partner.GetMetadata().GetDescription(),
		Trash:                     false,
		Settings:                  sb,
		Host:                      partner.GetSpec().GetHost(),
		Domain:                    partner.GetSpec().GetDomain(),
		TosLink:                   partner.GetSpec().GetTosLink(),
		LogoLink:                  partner.GetSpec().GetLogoLink(),
		NotificationEmail:         partner.GetSpec().GetNotificationEmail(),
		HelpdeskEmail:             partner.GetSpec().GetHelpdeskEmail(),
		ProductName:               partner.GetSpec().GetProductName(),
		SupportTeamName:           partner.GetSpec().GetSupportTeamName(),
		OpsHost:                   partner.GetSpec().GetOpsHost(),
		FavIconLink:               partner.GetSpec().GetFavIconLink(),
		IsTOTPEnabled:             partner.GetSpec().GetIsTOTPEnabled(),
		IsSyntheticPartnerEnabled: false,
		CreatedAt:                 time.Now(),
		ModifiedAt:                time.Now(),
	}
	entity, err := dao.Create(ctx, s.db, &part)
	if err != nil {
		return &systemv3.Partner{}, err
	}

	if createdPartner, ok := entity.(*models.Partner); ok {
		//update v3 spec
		partner.Metadata.Id = createdPartner.ID.String()
		partner.Metadata.ModifiedAt = timestamppb.New(createdPartner.ModifiedAt)

		CreatePartnerAuditEvent(ctx, s.al, AuditActionCreate, partner.GetMetadata().GetName(), createdPartner.ID)
	}

	return partner, nil

}

func (s *partnerService) GetByID(ctx context.Context, id string) (*systemv3.Partner, error) {

	partner := &systemv3.Partner{
		ApiVersion: apiVersion,
		Kind:       partnerKind,
		Metadata: &v3.Metadata{
			Id: id,
		},
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return &systemv3.Partner{}, err
	}
	entity, err := dao.GetByID(ctx, s.db, uid, &models.Partner{})
	if err != nil {
		return &systemv3.Partner{}, err
	}

	if part, ok := entity.(*models.Partner); ok {

		partner.Metadata = &v3.Metadata{
			Name:        part.Name,
			Description: part.Description,
			ModifiedAt:  timestamppb.New(part.ModifiedAt),
		}
		partner.Spec = &systemv3.PartnerSpec{
			Host:              part.Host,
			Domain:            part.Domain,
			TosLink:           part.TosLink,
			LogoLink:          part.LogoLink,
			NotificationEmail: part.NotificationEmail,
			HelpdeskEmail:     part.HelpdeskEmail,
			ProductName:       part.ProductName,
			SupportTeamName:   part.SupportTeamName,
			OpsHost:           part.OpsHost,
			FavIconLink:       part.FavIconLink,
			IsTOTPEnabled:     part.IsTOTPEnabled,
			Settings:          nil, //TODO
		}

		return partner, nil

	} else {
		partner := &systemv3.Partner{
			ApiVersion: apiVersion,
			Kind:       partnerKind,
			Metadata: &v3.Metadata{
				Id: id,
			},
			Status: &v3.Status{
				ConditionStatus: v3.ConditionStatus_StatusNotSet,
				Reason:          "Unable to fetch partner information",
				LastUpdated:     timestamppb.Now(),
			},
		}

		return partner, nil
	}

}

func (s *partnerService) GetByName(ctx context.Context, name string) (*systemv3.Partner, error) {

	partner := &systemv3.Partner{
		ApiVersion: apiVersion,
		Kind:       partnerKind,
		Metadata: &v3.Metadata{
			Name: name,
		},
	}

	entity, err := dao.GetByName(ctx, s.db, name, &models.Partner{})
	if err != nil {
		return &systemv3.Partner{}, err
	}

	if part, ok := entity.(*models.Partner); ok {

		partner.Metadata = &v3.Metadata{
			Name:        part.Name,
			Id:          part.ID.String(),
			Description: part.Description,
			ModifiedAt:  timestamppb.New(part.ModifiedAt),
		}
		partner.Spec = &systemv3.PartnerSpec{
			Host:              part.Host,
			Domain:            part.Domain,
			TosLink:           part.TosLink,
			LogoLink:          part.LogoLink,
			NotificationEmail: part.NotificationEmail,
			HelpdeskEmail:     part.HelpdeskEmail,
			ProductName:       part.ProductName,
			SupportTeamName:   part.SupportTeamName,
			OpsHost:           part.OpsHost,
			FavIconLink:       part.FavIconLink,
			IsTOTPEnabled:     part.IsTOTPEnabled,
			Settings:          nil, //TODO
		}

		return partner, nil
	} else {
		partner := &systemv3.Partner{
			ApiVersion: apiVersion,
			Kind:       partnerKind,
			Metadata: &v3.Metadata{
				Name: name,
			},
			Status: &v3.Status{
				ConditionType:   "Describe",
				ConditionStatus: v3.ConditionStatus_StatusNotSet,
				Reason:          "Unable to fetch partner information",
				LastUpdated:     timestamppb.Now(),
			},
		}

		return partner, nil
	}

}

func (s *partnerService) Update(ctx context.Context, partner *systemv3.Partner) (*systemv3.Partner, error) {

	entity, err := dao.GetByName(ctx, s.db, partner.Metadata.Name, &models.Partner{})
	if err != nil {
		return &systemv3.Partner{}, err
	}

	var sb []byte
	if partner.GetSpec().GetSettings() != nil {
		sb = json.RawMessage(partner.GetSpec().GetSettings().String())
	}

	if part, ok := entity.(*models.Partner); ok {
		//update partner details
		part.Name = partner.GetMetadata().Name
		part.Description = partner.GetMetadata().GetDescription()
		part.Settings = sb
		part.Host = partner.GetSpec().GetHost()
		part.Domain = partner.GetSpec().GetDomain()
		part.TosLink = partner.GetSpec().GetTosLink()
		part.LogoLink = partner.GetSpec().GetLogoLink()
		part.NotificationEmail = partner.GetSpec().GetNotificationEmail()
		part.HelpdeskEmail = partner.GetSpec().GetHelpdeskEmail()
		part.ProductName = partner.GetSpec().GetProductName()
		part.SupportTeamName = partner.GetSpec().GetSupportTeamName()
		part.OpsHost = partner.GetSpec().GetOpsHost()
		part.FavIconLink = partner.GetSpec().GetFavIconLink()
		part.IsTOTPEnabled = partner.GetSpec().GetIsTOTPEnabled()
		part.ModifiedAt = time.Now()

		//Update the partner details
		_, err = dao.Update(ctx, s.db, part.ID, part)
		if err != nil {
			return &systemv3.Partner{}, err
		}

		//update metadata and status
		partner.Metadata.ModifiedAt = timestamppb.New(part.ModifiedAt)

		CreatePartnerAuditEvent(ctx, s.al, AuditActionUpdate, partner.GetMetadata().GetName(), part.ID)

	}

	return partner, nil
}

func (s *partnerService) Delete(ctx context.Context, partner *systemv3.Partner) (*systemv3.Partner, error) {
	entity, err := dao.GetByName(ctx, s.db, partner.Metadata.Name, &models.Partner{})
	if err != nil {
		return &systemv3.Partner{}, err
	}

	if part, ok := entity.(*models.Partner); ok {
		err := dao.Delete(ctx, s.db, part.ID, part)
		if err != nil {
			return &systemv3.Partner{}, err
		}

		CreatePartnerAuditEvent(ctx, s.al, AuditActionDelete, partner.GetMetadata().GetName(), part.ID)
		return partner, nil
	}

	return partner, nil

}

func (s *partnerService) GetOnlyPartner(ctx context.Context) (partner *systemv3.Partner, err error) {
	var partners []models.Partner
	entities, err := dao.ListAll(ctx, s.db, &partners)
	if err != nil {
		return nil, err
	}
	if pts, ok := entities.(*[]models.Partner); ok {
		for _, part := range *pts {
			partner = &systemv3.Partner{
				Metadata: &v3.Metadata{
					Name:        part.Name,
					Id:          part.ID.String(),
					Description: part.Description,
					ModifiedAt:  timestamppb.New(part.ModifiedAt),
				},
				Spec: &systemv3.PartnerSpec{
					Host:              part.Host,
					Domain:            part.Domain,
					TosLink:           part.TosLink,
					LogoLink:          part.LogoLink,
					NotificationEmail: part.NotificationEmail,
					HelpdeskEmail:     part.HelpdeskEmail,
					ProductName:       part.ProductName,
					SupportTeamName:   part.SupportTeamName,
					OpsHost:           part.OpsHost,
					FavIconLink:       part.FavIconLink,
					IsTOTPEnabled:     part.IsTOTPEnabled,
					Settings:          nil,
				},
				Status: &v3.Status{
					ConditionType:   "Describe",
					ConditionStatus: v3.ConditionStatus_StatusOK,
					LastUpdated:     timestamppb.New(part.ModifiedAt),
				},
			}

			return partner, nil
		}
	}
	return partner, err
}
