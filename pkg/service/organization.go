package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	v3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	systemv3 "github.com/RafaySystems/rcloud-base/proto/types/systempb/v3"
	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	organizationKind     = "Organization"
	organizationListKind = "OrganizationList"
)

// OrganizationService is the interface for organization operations
type OrganizationService interface {
	// create organization
	Create(ctx context.Context, organization *systemv3.Organization) (*systemv3.Organization, error)
	// get organization by id
	GetByID(ctx context.Context, id string) (*systemv3.Organization, error)
	// get organization by id
	GetByName(ctx context.Context, name string) (*systemv3.Organization, error)
	// create or update organization
	Update(ctx context.Context, organization *systemv3.Organization) (*systemv3.Organization, error)
	// delete organization
	Delete(ctx context.Context, organization *systemv3.Organization) (*systemv3.Organization, error)
	// list organization
	List(ctx context.Context, organization *systemv3.Organization) (*systemv3.OrganizationList, error)
}

// organizationService implements OrganizationService
type organizationService struct {
	db *bun.DB
}

// NewOrganizationService return new organization service
func NewOrganizationService(db *bun.DB) OrganizationService {
	return &organizationService{db}
}

func (s *organizationService) Create(ctx context.Context, org *systemv3.Organization) (*systemv3.Organization, error) {

	var partner models.Partner
	_, err := pg.GetByName(ctx, s.db, org.Metadata.Partner, &partner)
	if err != nil {
		return &systemv3.Organization{}, err
	}

	//update default organization setting values
	org.Spec.Settings = &systemv3.OrganizationSettings{
		Lockout: &systemv3.Lockout{
			Enabled:   true,
			PeriodMin: 15,
			Attempts:  5,
		},
		IdleLogoutMin: 60,
	}
	sb, err := json.MarshalIndent(org.GetSpec().GetSettings(), "", "\t")
	if err != nil {
		return &systemv3.Organization{}, err
	}
	//convert v3 spec to internal models
	organization := models.Organization{
		Name:              org.GetMetadata().GetName(),
		Description:       org.GetMetadata().GetDescription(),
		Trash:             false,
		Settings:          json.RawMessage(sb),
		BillingAddress:    org.GetSpec().GetBillingAddress(),
		PartnerId:         partner.ID,
		Active:            org.GetSpec().GetActive(),
		Approved:          org.GetSpec().GetApproved(),
		Type:              org.GetSpec().GetType(),
		AddressLine1:      org.GetSpec().GetAddressLine1(),
		AddressLine2:      org.GetSpec().GetAddressLine2(),
		City:              org.GetSpec().GetCity(),
		Country:           org.GetSpec().GetCountry(),
		Phone:             org.GetSpec().GetPhone(),
		State:             org.GetSpec().GetState(),
		Zipcode:           org.GetSpec().GetZipcode(),
		IsPrivate:         org.GetSpec().GetIsPrivate(),
		IsTOTPEnabled:     org.GetSpec().GetIsTotpEnabled(),
		AreClustersShared: org.GetSpec().GetAreClustersShared(),
		CreatedAt:         time.Now(),
		ModifiedAt:        time.Now(),
	}
	entity, err := pg.Create(ctx, s.db, &organization)
	if err != nil {
		return &systemv3.Organization{}, err
	}

	if createdOrg, ok := entity.(*models.Organization); ok {
		//update v3 spec
		org.Metadata.Id = createdOrg.ID.String()
	}

	return org, nil
}

func (s *organizationService) GetByID(ctx context.Context, id string) (*systemv3.Organization, error) {

	organization := &systemv3.Organization{
		ApiVersion: apiVersion,
		Kind:       organizationKind,
		Metadata: &v3.Metadata{
			Id: id,
		},
		Spec: &systemv3.OrganizationSpec{},
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return &systemv3.Organization{}, err
	}
	entity, err := pg.GetByID(ctx, s.db, uid, &models.Organization{})
	if err != nil {
		return &systemv3.Organization{}, err
	}

	if org, ok := entity.(*models.Organization); ok {

		var partner models.Partner
		_, err := pg.GetByID(ctx, s.db, org.PartnerId, &partner)
		if err != nil {
			return &systemv3.Organization{}, err
		}

		organization, err = prepareOrganizationResponse(organization, org, partner.Name)
		if err != nil {
			return &systemv3.Organization{}, err
		}

		return organization, nil

	} else {
		organization := &systemv3.Organization{
			ApiVersion: apiVersion,
			Kind:       organizationKind,
			Status: &v3.Status{
				ConditionType:   "Describe",
				ConditionStatus: v3.ConditionStatus_StatusNotSet,
				Reason:          "Unable to fetch organization information",
				LastUpdated:     timestamppb.Now(),
			},
		}

		return organization, nil
	}

}

func (s *organizationService) GetByName(ctx context.Context, name string) (*systemv3.Organization, error) {

	organization := &systemv3.Organization{
		ApiVersion: apiVersion,
		Kind:       organizationKind,
		Metadata: &v3.Metadata{
			Name: name,
		},
	}
	entity, err := pg.GetByName(ctx, s.db, name, &models.Organization{})
	if err != nil {
		return &systemv3.Organization{}, err
	}

	if org, ok := entity.(*models.Organization); ok {

		var partner models.Partner
		_, err := pg.GetByID(ctx, s.db, org.PartnerId, &partner)
		if err != nil {
			return &systemv3.Organization{}, err
		}

		organization, err = prepareOrganizationResponse(organization, org, partner.Name)
		if err != nil {
			return &systemv3.Organization{}, err
		}
	}

	return organization, nil
}

func (s *organizationService) Update(ctx context.Context, organization *systemv3.Organization) (*systemv3.Organization, error) {

	entity, err := pg.GetByName(ctx, s.db, organization.Metadata.Name, &models.Organization{})
	if err != nil {
		return &systemv3.Organization{}, err
	}

	if org, ok := entity.(*models.Organization); ok {

		sb, err := json.MarshalIndent(organization.GetSpec().GetSettings(), "", "\t")
		if err != nil {
			return &systemv3.Organization{}, err
		}

		//update organization details
		org.Name = organization.GetMetadata().GetName()
		org.Description = organization.GetMetadata().GetDescription()
		org.ModifiedAt = time.Now()
		org.Trash = false
		org.Settings = json.RawMessage(sb)
		org.BillingAddress = organization.GetSpec().GetBillingAddress()
		org.Active = organization.GetSpec().GetActive()
		org.Approved = organization.GetSpec().GetApproved()
		org.Type = organization.GetSpec().GetType()
		org.AddressLine1 = organization.GetSpec().GetAddressLine1()
		org.AddressLine2 = organization.GetSpec().GetAddressLine2()
		org.City = organization.GetSpec().GetCity()
		org.Country = organization.GetSpec().GetCountry()
		org.Phone = organization.GetSpec().GetPhone()
		org.State = organization.GetSpec().GetState()
		org.Zipcode = organization.GetSpec().GetZipcode()
		org.IsPrivate = organization.GetSpec().GetIsPrivate()
		org.IsTOTPEnabled = organization.GetSpec().GetIsTotpEnabled()
		org.AreClustersShared = organization.GetSpec().GetAreClustersShared()

		_, err = pg.Update(ctx, s.db, org.ID, org)
		if err != nil {
			return &systemv3.Organization{}, err
		}
	}

	return organization, nil
}

func (s *organizationService) Delete(ctx context.Context, organization *systemv3.Organization) (*systemv3.Organization, error) {

	entity, err := pg.GetByName(ctx, s.db, organization.Metadata.Name, &models.Organization{})
	if err != nil {
		return &systemv3.Organization{}, err
	}

	if org, ok := entity.(*models.Organization); ok {
		org.Trash = true
		_, err := pg.Update(ctx, s.db, org.ID, org)
		if err != nil {
			return &systemv3.Organization{}, err
		}

		//update v3 status
		organization.Metadata.Name = org.Name
		organization.Metadata.ModifiedAt = timestamppb.New(org.ModifiedAt)
	}
	return organization, nil

}

func (s *organizationService) List(ctx context.Context, organization *systemv3.Organization) (*systemv3.OrganizationList, error) {

	var organizations []*systemv3.Organization
	organinzationList := &systemv3.OrganizationList{
		ApiVersion: apiVersion,
		Kind:       organizationListKind,
		Metadata: &v3.ListMetadata{
			Count: 0,
		},
	}
	if len(organization.Metadata.Partner) > 0 {
		var partner models.Partner
		_, err := pg.GetByName(ctx, s.db, organization.Metadata.Partner, &partner)
		if err != nil {
			return &systemv3.OrganizationList{}, err
		}

		var orgs []models.Organization
		entities, err := pg.List(ctx, s.db, uuid.NullUUID{UUID: partner.ID, Valid: true}, uuid.NullUUID{UUID: uuid.Nil}, &orgs)
		if err != nil {
			return &systemv3.OrganizationList{}, err
		}
		if orgs, ok := entities.(*[]models.Organization); ok {
			for _, org := range *orgs {
				var settings systemv3.OrganizationSettings
				err := json.Unmarshal(org.Settings, &settings)
				if err != nil {
					return &systemv3.OrganizationList{}, err
				}
				organization.Metadata = &v3.Metadata{
					Name:        org.Name,
					Description: org.Description,
					Partner:     partner.Name,
					ModifiedAt:  timestamppb.New(org.ModifiedAt),
				}
				organization.Spec = &systemv3.OrganizationSpec{
					BillingAddress:    org.BillingAddress,
					Active:            org.Active,
					Approved:          org.Approved,
					Type:              org.Type,
					AddressLine1:      org.AddressLine1,
					AddressLine2:      org.AddressLine2,
					City:              org.City,
					Country:           org.Country,
					Phone:             org.Phone,
					State:             org.State,
					Zipcode:           org.Zipcode,
					IsPrivate:         org.IsPrivate,
					IsTotpEnabled:     org.IsTOTPEnabled,
					AreClustersShared: org.AreClustersShared,
					Settings:          &settings,
				}
				organizations = append(organizations, organization)
			}

			//update the list metadata and items response
			organinzationList.Metadata = &v3.ListMetadata{
				Count: int64(len(organizations)),
			}
			organinzationList.Items = organizations
		}

	} else {
		return organinzationList, fmt.Errorf("missing partner in metadata")
	}
	return organinzationList, nil
}

func prepareOrganizationResponse(organization *systemv3.Organization, org *models.Organization, partnerName string) (*systemv3.Organization, error) {

	var settings systemv3.OrganizationSettings
	if org.Settings != nil {
		err := json.Unmarshal(org.Settings, &settings)
		if err != nil {
			return &systemv3.Organization{}, err
		}
	}

	organization.Metadata = &v3.Metadata{
		Name:        org.Name,
		Id:          org.ID.String(),
		Description: org.Description,
		Partner:     partnerName,
		ModifiedAt:  timestamppb.New(org.ModifiedAt),
	}
	organization.Spec = &systemv3.OrganizationSpec{
		BillingAddress:    org.BillingAddress,
		Active:            org.Active,
		Approved:          org.Approved,
		Type:              org.Type,
		AddressLine1:      org.AddressLine1,
		AddressLine2:      org.AddressLine2,
		City:              org.City,
		Country:           org.Country,
		Phone:             org.Phone,
		State:             org.State,
		Zipcode:           org.Zipcode,
		IsPrivate:         org.IsPrivate,
		IsTotpEnabled:     org.IsTOTPEnabled,
		AreClustersShared: org.AreClustersShared,
		Settings:          &settings,
	}

	return organization, nil
}
