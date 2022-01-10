package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/internal/models"
	systemv3 "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/types/systempb/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	apiVersion  = "system.k8smgmt.io/v3"
	partnerKind = "Partner"
)

// PartnerService is the interface for partner operations
type PartnerService interface {
	Close() error
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
}

// partnerService implements PartnerService
type partnerService struct {
	dao pg.EntityDAO
}

// NewPartnerService return new partner service
func NewPartnerService(db *bun.DB) PartnerService {
	return &partnerService{
		dao: pg.NewEntityDAO(db),
	}
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
	entity, err := s.dao.Create(ctx, &part)
	if err != nil {
		partner.Status = &v3.Status{
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
		}
		return partner, err
	}

	if createdPartner, ok := entity.(*models.Partner); ok {
		//update v3 spec
		partner.Metadata.Id = createdPartner.ID.String()
		partner.Metadata.ModifiedAt = timestamppb.New(createdPartner.ModifiedAt)
		if partner.Status != nil {
			partner.Status = &v3.Status{
				ConditionStatus: v3.ConditionStatus_StatusOK,
				LastUpdated:     timestamppb.Now(),
			}
		}
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
		partner.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
			LastUpdated:     timestamppb.Now(),
		}
		return partner, err
	}
	entity, err := s.dao.GetByID(ctx, uid, &models.Partner{})
	if err != nil {
		partner.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
			LastUpdated:     timestamppb.Now(),
		}
		return partner, err
	}

	if part, ok := entity.(*models.Partner); ok {

		partner.Metadata = &v3.Metadata{
			Name:        part.Name,
			Description: part.Description,
			Id:          part.ID.String(),
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
		partner.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusOK,
			LastUpdated:     timestamppb.New(part.ModifiedAt),
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

	entity, err := s.dao.GetByName(ctx, name, &models.Partner{})
	if err != nil {
		partner.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
			LastUpdated:     timestamppb.Now(),
		}
		return partner, err
	}

	if part, ok := entity.(*models.Partner); ok {

		partner.Metadata = &v3.Metadata{
			Name:        part.Name,
			Description: part.Description,
			Id:          part.ID.String(),
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
		partner.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusOK,
			LastUpdated:     timestamppb.New(part.ModifiedAt),
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

	id, _ := uuid.Parse(partner.Metadata.Id)
	entity, err := s.dao.GetByID(ctx, id, &models.Partner{})
	if err != nil {
		partner.Status = &v3.Status{
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
			LastUpdated:     timestamppb.Now(),
		}
		return partner, err
	}

	var sb []byte
	if partner.GetSpec().GetSettings() != nil {
		sb = json.RawMessage(partner.GetSpec().GetSettings().String())
	}

	if part, ok := entity.(*models.Partner); ok {
		//update partner details
		part.ID = id
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
		_, err = s.dao.Update(ctx, id, part)
		if err != nil {
			partner.Status = &v3.Status{
				ConditionStatus: v3.ConditionStatus_StatusFailed,
				Reason:          err.Error(),
				LastUpdated:     timestamppb.Now(),
			}
			return partner, err
		}

		//update metadata and status
		partner.Metadata.ModifiedAt = timestamppb.New(part.ModifiedAt)
		partner.Status = &v3.Status{
			ConditionStatus: v3.ConditionStatus_StatusOK,
			LastUpdated:     timestamppb.Now(),
		}

	}

	return partner, nil
}

func (s *partnerService) Delete(ctx context.Context, partner *systemv3.Partner) (*systemv3.Partner, error) {
	id, err := uuid.Parse(partner.Metadata.Id)
	if err != nil {
		partner.Status = &v3.Status{
			ConditionType:   "Delete",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
			LastUpdated:     timestamppb.Now(),
		}
		return partner, err
	}
	entity, err := s.dao.GetByID(ctx, id, &models.Partner{})
	if err != nil {
		partner.Status = &v3.Status{
			ConditionType:   "Delete",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
			LastUpdated:     timestamppb.Now(),
		}
		return partner, err
	}

	if part, ok := entity.(*models.Partner); ok {
		err = s.dao.Delete(ctx, id, part)
		if err != nil {
			partner.Status = &v3.Status{
				ConditionType:   "Delete",
				ConditionStatus: v3.ConditionStatus_StatusFailed,
				Reason:          err.Error(),
				LastUpdated:     timestamppb.Now(),
			}
			return partner, err
		}
		//update status
		if partner != nil {
			partner.Metadata.Id = part.ID.String()
			partner.Metadata.Name = part.Name
			partner.Status = &v3.Status{
				ConditionStatus: v3.ConditionStatus_StatusOK,
				ConditionType:   "Delete",
			}
		}
		return partner, nil
	}

	return partner, nil

}

func (s *partnerService) Close() error {
	return s.dao.Close()
}
