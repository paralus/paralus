package service

import (
	"context"
	"time"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
)

// MetroService is the interface for metro operations
type MetroService interface {
	Close() error
	// create metro
	Create(ctx context.Context, metro *infrav3.Metro, oid uuid.NullUUID, pid uuid.NullUUID) (*infrav3.Metro, error)
	// get metro by id
	GetById(ctx context.Context, id uuid.UUID) (*infrav3.Metro, error)
	// get metro by name
	GetByName(ctx context.Context, name string, oid uuid.NullUUID, pid uuid.NullUUID) (*infrav3.Metro, error)
	// get metro id by name
	GetIDByName(ctx context.Context, name string, oid uuid.NullUUID, pid uuid.NullUUID) (uuid.UUID, error)
	// create or update metro
	Update(ctx context.Context, metro *infrav3.Metro, oid uuid.NullUUID, pid uuid.NullUUID) (*infrav3.Metro, error)
	// delete metro
	Delete(ctx context.Context, metro *infrav3.Metro, oid uuid.NullUUID, pid uuid.NullUUID) (*infrav3.Metro, error)
	// list metro
	List(ctx context.Context, oid uuid.NullUUID, pid uuid.NullUUID) (*[]infrav3.Metro, error)
}

// metroService implements MetroService
type metroService struct {
	dao pg.EntityDAO
}

// NewProjectService return new project service
func NewMetroService(db *bun.DB) MetroService {
	return &metroService{
		dao: pg.NewEntityDAO(db),
	}
}

func (s *metroService) Create(ctx context.Context, metro *infrav3.Metro, oid uuid.NullUUID, pid uuid.NullUUID) (*infrav3.Metro, error) {

	//convert v3 spec to internal models
	metrodb := models.Metro{
		Name:           metro.Name,
		CreatedAt:      time.Now(),
		ModifiedAt:     time.Now(),
		Trash:          false,
		Latitude:       metro.Latitude,
		Longitude:      metro.Longitude,
		City:           metro.City,
		State:          "", //metro.State,
		Country:        metro.Country,
		CountryCode:    "", //metro.CountryCode,
		StateCode:      "", //metro.StateCode,
		OrganizationId: oid.UUID,
		PartnerId:      pid.UUID,
	}
	_, err := s.dao.Create(ctx, &metrodb)
	if err != nil {
		return nil, err
	}

	return metro, nil

}

func (s *metroService) GetByName(ctx context.Context, name string, oid uuid.NullUUID, pid uuid.NullUUID) (*infrav3.Metro, error) {

	var metro infrav3.Metro

	entity, err := s.dao.GetByNamePartnerOrg(ctx, name, pid, oid, &models.Metro{})
	if err != nil {
		return nil, err
	}

	if metrodb, ok := entity.(*models.Metro); ok {

		metro := &infrav3.Metro{
			Name:        metrodb.Name,
			Country:     metrodb.Country,
			City:        metrodb.City,
			State:       metrodb.State,
			Latitude:    metrodb.Latitude,
			Longitude:   metrodb.Longitude,
			StateCode:   metrodb.StateCode,
			CountryCode: metrodb.CountryCode,
		}
		return metro, nil

	}
	return &metro, nil
}

func (s *metroService) GetById(ctx context.Context, id uuid.UUID) (*infrav3.Metro, error) {
	var metro infrav3.Metro

	entity, err := s.dao.GetByID(ctx, id, &models.Metro{})
	if err != nil {
		return nil, err
	}

	if metrodb, ok := entity.(*models.Metro); ok {

		metro := &infrav3.Metro{
			Name:        metrodb.Name,
			Country:     metrodb.Country,
			City:        metrodb.City,
			State:       metrodb.State,
			Latitude:    metrodb.Latitude,
			Longitude:   metrodb.Longitude,
			StateCode:   metrodb.StateCode,
			CountryCode: metrodb.CountryCode,
		}
		return metro, nil

	}
	return &metro, nil
}

func (s *metroService) Update(ctx context.Context, metro *infrav3.Metro, oid uuid.NullUUID, pid uuid.NullUUID) (*infrav3.Metro, error) {

	entity, err := s.dao.GetByNamePartnerOrg(ctx, metro.Name, pid, oid, &models.Project{})
	if err != nil {
		return metro, err
	}

	if metrodb, ok := entity.(*models.Metro); ok {
		//update metro details
		metrodb.City = metro.City
		metrodb.Country = metro.Country
		metrodb.State = metro.State
		metrodb.StateCode = metro.StateCode
		metrodb.CountryCode = metro.CountryCode
		metrodb.ModifiedAt = time.Now()

		_, err = s.dao.Update(ctx, metrodb.ID, metrodb)
		if err != nil {
			return metro, err
		}
	}

	return metro, nil
}

func (s *metroService) Delete(ctx context.Context, metro *infrav3.Metro, oid uuid.NullUUID, pid uuid.NullUUID) (*infrav3.Metro, error) {

	entity, err := s.dao.GetByNamePartnerOrg(ctx, metro.Name, pid, oid, &models.Metro{})
	if err != nil {
		return metro, err
	}
	if metrodb, ok := entity.(*models.Metro); ok {
		err = s.dao.Delete(ctx, metrodb.ID, metrodb)
		if err != nil {
			return metro, err
		}
	}

	return metro, nil
}

func (s *metroService) List(ctx context.Context, oid uuid.NullUUID, pid uuid.NullUUID) (*[]infrav3.Metro, error) {

	var metros []infrav3.Metro
	var metrodbs []models.Metro
	entities, err := s.dao.List(ctx, oid, pid, &metrodbs)
	if err != nil {
		return nil, err
	}

	if metrodbs, ok := entities.(*[]models.Metro); ok {
		for _, metrodb := range *metrodbs {

			metro := infrav3.Metro{
				Name:        metrodb.Name,
				City:        metrodb.City,
				State:       metrodb.State,
				Country:     metrodb.Country,
				Latitude:    metrodb.Latitude,
				Longitude:   metrodb.Longitude,
				StateCode:   metrodb.StateCode,
				CountryCode: metrodb.CountryCode,
			}
			metros = append(metros, metro)
		}
	}

	return &metros, nil
}

func (s *metroService) GetIDByName(ctx context.Context, name string, oid uuid.NullUUID, pid uuid.NullUUID) (uuid.UUID, error) {
	entity, err := s.dao.GetByNamePartnerOrg(ctx, name, pid, oid, &models.Metro{})
	if err != nil {
		return uuid.Nil, err
	}

	if metrodb, ok := entity.(*models.Metro); ok {
		return metrodb.ID, nil
	}
	return uuid.Nil, nil
}

func (s *metroService) Close() error {
	return s.dao.Close()
}
