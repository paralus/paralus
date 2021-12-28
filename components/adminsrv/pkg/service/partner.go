package service

import (
	"context"
	"strconv"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/internal/models"
	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/internal/persistence/provider/pg"
	systemv3 "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/types/systempb/v3"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	bun "github.com/uptrace/bun"
)

const (
	partnerKind = "Partner"
)

// PartnerService is the interface for partner operations
type PartnerService interface {
	Close() error
	// create partner
	Create(ctx context.Context, partner *systemv3.Partner) (*systemv3.Partner, error)
	// get partner by id
	GetByID(ctx context.Context, partnerId string) (*systemv3.Partner, error)
	// create or update partner
	Patch(ctx context.Context, partner *systemv3.Partner) error
	// delete pipeline
	Delete(ctx context.Context, partnerId string) error
}

// partnerService implements PartnerService
type partnerService struct {
	dao pg.PartnerDAO
}

// NewPartnerService return new partner service
func NewPartnerService(db *bun.DB) PartnerService {
	return &partnerService{
		dao: pg.NewPartnerDAO(db),
	}
}

func (s *partnerService) Create(ctx context.Context, p *systemv3.Partner) (*systemv3.Partner, error) {

	//convert v3 spec to internal models
	partner := models.Partner{
		Name:                      p.GetMetadata().GetName(),
		Description:               p.GetMetadata().GetDescription(),
		Trash:                     false,
		Settings:                  nil, // p.GetSettings(),
		Host:                      p.GetSpec().GetHost(),
		Domain:                    p.GetSpec().GetDomain(),
		TosLink:                   p.GetSpec().GetTosLink(),
		LogoLink:                  p.GetSpec().GetLogoLink(),
		NotificationEmail:         p.GetSpec().GetNotificationEmail(),
		HelpdeskEmail:             p.GetSpec().GetHelpdeskEmail(),
		ProductName:               p.GetSpec().GetProductName(),
		SupportTeamName:           p.GetSpec().GetSupportTeamName(),
		OpsHost:                   p.GetSpec().GetOpsHost(),
		FavIconLink:               p.GetSpec().GetFavIconLink(),
		IsTOTPEnabled:             p.GetSpec().GetIsTOTPEnabled(),
		IsSyntheticPartnerEnabled: false,
	}
	createdPartner, err := s.dao.CreatePartner(ctx, partner)
	if err != nil {
		return nil, err
	}

	//update v3 spec
	p.Metadata.Id = strconv.FormatInt(createdPartner.ID, 10)
	if p.Status != nil {
		p.Status = &v3.Status{
			ConditionType: "StatusOK",
		}
	}

	return p, nil

}

func (s *partnerService) GetByID(ctx context.Context, id string) (*systemv3.Partner, error) {

	p, err := s.dao.GetPartner(ctx, id)
	if err != nil {
		return nil, err
	}

	partner := &systemv3.Partner{
		ApiVersion: "core.rafay.dev/v3",
		Kind:       "Partner",
		Metadata: &v3.Metadata{
			Name:        p.Name,
			Description: p.Description,
			Id:          strconv.FormatInt(p.ID, 10),
			ModifiedAt:  nil, //p.ModifiedAt,
		},
		Spec: &systemv3.PartnerSpec{
			Host:              p.Host,
			Domain:            p.Domain,
			TosLink:           p.TosLink,
			LogoLink:          p.LogoLink,
			NotificationEmail: p.NotificationEmail,
			HelpdeskEmail:     p.HelpdeskEmail,
			ProductName:       p.ProductName,
			SupportTeamName:   p.SupportTeamName,
			OpsHost:           p.OpsHost,
			FavIconLink:       p.FavIconLink,
			IsTOTPEnabled:     p.IsTOTPEnabled,
			Settings:          nil, //p.Settings,
		},
		Status: &v3.Status{
			LastUpdated:     nil, //p.ModifiedAt,
			ConditionType:   "StatusOK",
			ConditionStatus: 2,
		},
	}

	return partner, nil
}

func (s *partnerService) Patch(ctx context.Context, partner *systemv3.Partner) error {
	return nil
}

func (s *partnerService) Delete(ctx context.Context, partnerId string) error {
	return nil
}

func (s *partnerService) Close() error {
	return s.dao.Close()
}
