package pg

import (
	"context"

	models "github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/internal/models"
	bun "github.com/uptrace/bun"
)

// PartnerDAO is the interface for partner database operations
type PartnerDAO interface {
	Close() error
	// create partner
	CreatePartner(ctx context.Context, part models.Partner) (*models.Partner, error)
	// get partner by id
	GetPartner(ctx context.Context, id string) (*models.Partner, error)
	// delete pipeline
	DeletePartner(ctx context.Context, partnerId string) error
}

type partnerDAO struct {
	db *bun.DB
}

func (dao *partnerDAO) Close() error {
	return dao.db.Close()
}

// NewPartnerService return new partner service
func NewPartnerDAO(db *bun.DB) PartnerDAO {
	return &partnerDAO{db}
}

func (dao *partnerDAO) CreatePartner(ctx context.Context, part models.Partner) (*models.Partner, error) {
	var newPartner models.Partner

	if _, err := dao.db.NewInsert().Model(&part).Exec(ctx); err != nil {
		return nil, err
	}

	newPartner = models.Partner{
		ID: part.ID,
	}
	if err := dao.db.NewSelect().Model(&newPartner).WherePK().Scan(ctx); err != nil {
		return nil, err
	}

	return &newPartner, nil
}

func (dao *partnerDAO) GetPartner(ctx context.Context, id string) (*models.Partner, error) {
	var p models.Partner

	err := dao.db.NewSelect().Model(&p).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (dao *partnerDAO) DeletePartner(ctx context.Context, id string) error {
	var part models.Partner
	_, err := dao.db.NewDelete().
		Model(&part).
		Where("id  = ?", id).
		Exec(ctx)
	return err
}
