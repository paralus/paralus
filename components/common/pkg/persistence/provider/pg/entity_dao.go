package pg

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
)

// DAO is the interface for database operations
type EntityDAO interface {
	Close() error
	// create entity
	Create(ctx context.Context, entity interface{}) (interface{}, error)
	// get entity by field
	GetX(ctx context.Context, field string, value interface{}, entity interface{}) (interface{}, error)
	// get entity by id
	GetByID(ctx context.Context, id uuid.UUID, entity interface{}) (interface{}, error)
	// get entity by name
	GetByName(ctx context.Context, name string, entity interface{}) (interface{}, error)
	// get entity by name
	GetEntityByName(ctx context.Context, name string, oid uuid.NullUUID, pid uuid.NullUUID, entity interface{}) (interface{}, error)
	//Update entity
	Update(ctx context.Context, id uuid.UUID, entity interface{}) (interface{}, error)
	// get entity by field
	UpdateX(ctx context.Context, field string, value interface{}, entity interface{}) (interface{}, error)
	// delete entity
	Delete(ctx context.Context, id uuid.UUID, entity interface{}) error
	// get entity by field
	DeleteX(ctx context.Context, field string, value interface{}, entity interface{}) error
	// get list of entities
	List(ctx context.Context, partnerId uuid.NullUUID, organizationId uuid.NullUUID, entities interface{}) (interface{}, error)
	// get list of entities
	ListByProject(ctx context.Context, partnerId uuid.NullUUID, organizationId uuid.NullUUID, projectId uuid.NullUUID, entities interface{}) error
	//returns db object
	GetInstance() *bun.DB
}

type entityDAO struct {
	db *bun.DB
}

func (dao *entityDAO) Close() error {
	return dao.db.Close()
}

// NewEntityDao return new entity dao
func NewEntityDAO(db *bun.DB) EntityDAO {
	return &entityDAO{db}
}

func (dao *entityDAO) Create(ctx context.Context, entity interface{}) (interface{}, error) {

	if _, err := dao.db.NewInsert().Model(entity).Exec(ctx); err != nil {
		return nil, err
	}

	return entity, nil
}

func (dao *entityDAO) GetX(ctx context.Context, field string, value interface{}, entity interface{}) (interface{}, error) {

	err := dao.db.NewSelect().Model(entity).
		Where(fmt.Sprintf("%s = ?", field), value).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (dao *entityDAO) GetByID(ctx context.Context, id uuid.UUID, entity interface{}) (interface{}, error) {

	err := dao.db.NewSelect().Model(entity).
		Where("id = ?", id).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (dao *entityDAO) GetByName(ctx context.Context, name string, entity interface{}) (interface{}, error) {

	err := dao.db.NewSelect().Model(entity).
		Where("name = ?", name).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (dao *entityDAO) GetEntityByName(ctx context.Context, name string, oid uuid.NullUUID, pid uuid.NullUUID, entity interface{}) (interface{}, error) {

	sq := dao.db.NewSelect().Model(entity)
	if oid.Valid {
		sq = sq.Where("organization_id = ?", oid)
	}
	if pid.Valid {
		sq = sq.Where("partner_id = ?", pid)
	}
	sq = sq.Where("name = ?", name)

	err := sq.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (dao *entityDAO) Update(ctx context.Context, id uuid.UUID, entity interface{}) (interface{}, error) {
	if _, err := dao.db.NewUpdate().Model(entity).Where("id  = ?", id).Exec(ctx); err != nil {
		return nil, err
	}
	return entity, nil
}

func (dao *entityDAO) UpdateX(ctx context.Context, field string, value interface{}, entity interface{}) (interface{}, error) {
	if _, err := dao.db.NewUpdate().Model(entity).Where(fmt.Sprintf("%s = ?", field), value).Exec(ctx); err != nil {
		return nil, err
	}
	return entity, nil
}

func (dao *entityDAO) Delete(ctx context.Context, id uuid.UUID, entity interface{}) error {
	_, err := dao.db.NewDelete().
		Model(entity).
		Where("id  = ?", id).
		Exec(ctx)
	return err
}

func (dao *entityDAO) DeleteX(ctx context.Context, field string, value interface{}, entity interface{}) error {
	_, err := dao.db.NewDelete().
		Model(entity).
		Where(fmt.Sprintf("%s = ?", field), value).
		Exec(ctx)
	return err
}

func (dao *entityDAO) List(ctx context.Context, partnerId uuid.NullUUID, organizationId uuid.NullUUID, entities interface{}) (interface{}, error) {
	sq := dao.db.NewSelect().Model(entities)
	if partnerId.Valid {
		sq = sq.Where("partner_id = ?", partnerId)
	}
	if organizationId.Valid {
		sq = sq.Where("organization_id = ?", organizationId)
	}
	err := sq.Scan(ctx)
	return entities, err
}

func (dao *entityDAO) ListByProject(ctx context.Context, partnerId uuid.NullUUID, organizationId uuid.NullUUID, projectId uuid.NullUUID, entities interface{}) error {
	sq := dao.db.NewSelect().Model(entities)
	if partnerId.Valid {
		sq = sq.Where("partner_id = ?", partnerId)
	}
	if organizationId.Valid {
		sq = sq.Where("organization_id = ?", organizationId)
	}
	if projectId.Valid {
		sq = sq.Where("project_id = ?", projectId)
	}
	err := sq.Scan(ctx)
	return err
}

func (dao *entityDAO) GetInstance() *bun.DB {
	return dao.db
}
