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
	Create(context.Context, interface{}) (interface{}, error)
	// get entity by field
	GetX(context.Context, string, interface{}, interface{}) (interface{}, error)
	// get entity by multiple fields
	GetM(context.Context, map[string]interface{}, interface{}) (interface{}, error)
	// get entity by id
	GetByID(context.Context, uuid.UUID, interface{}) (interface{}, error)

	// get entity by name
	GetByName(context.Context, string, interface{}) (interface{}, error)
	// get entity by name partner and org
	GetByNamePartnerOrg(context.Context, string, uuid.NullUUID, uuid.NullUUID, interface{}) (interface{}, error)
	// get entity id by name
	GetIdByName(context.Context, string, interface{}) (interface{}, error)
	// get entity id by name partner and org
	GetIdByNamePartnerOrg(context.Context, string, uuid.NullUUID, uuid.NullUUID, interface{}) (interface{}, error)
	// get entity name by id
	GetNameById(context.Context, uuid.UUID, interface{}) (interface{}, error)
	//Update entity
	Update(context.Context, uuid.UUID, interface{}) (interface{}, error)
	// get entity by field
	UpdateX(context.Context, string, interface{}, interface{}) (interface{}, error)
	// delete entity by field
	DeleteX(context.Context, string, interface{}, interface{}) error
	// delete entity
	Delete(context.Context, uuid.UUID, interface{}) error
	// delete all items in table (for script)
	HardDeleteAll(context.Context, interface{}) error
	// get list of entities
	List(context.Context, uuid.NullUUID, uuid.NullUUID, interface{}) (interface{}, error)
	// get list of entities
	ListByProject(context.Context, uuid.NullUUID, uuid.NullUUID, uuid.NullUUID, interface{}) error
	// get list of entities without filtering
	ListAll(context.Context, interface{}) (interface{}, error)

	// lookup user by traits
	GetByTraits(ctx context.Context, name string, entity interface{}) (interface{}, error)
	// lookup user id by traits
	GetIdByTraits(ctx context.Context, name string, entity interface{}) (interface{}, error)

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

// M for multi ;)
func (dao *entityDAO) GetM(ctx context.Context, checks map[string]interface{}, entity interface{}) (interface{}, error) {
	// Can we get the checks directly from entity and create an upsert sort of func?
	q := dao.db.NewSelect().Model(entity)
	for field := range checks {
		q.Where(fmt.Sprintf("%s = ?", field), checks[field])
	}
	err := q.Scan(ctx)
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

func (dao *entityDAO) GetByNamePartnerOrg(ctx context.Context, name string, pid uuid.NullUUID, oid uuid.NullUUID, entity interface{}) (interface{}, error) {
	sq := dao.db.NewSelect().Model(entity)
	if oid.Valid {
		sq = sq.Where("organization_id = ?", oid)
	}
	if pid.Valid {
		sq = sq.Where("partner_id = ?", pid)
	}
	sq = sq.Where("name = ?", name).
		Where("trash = ?", false)

	err := sq.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (dao *entityDAO) GetIdByName(ctx context.Context, name string, entity interface{}) (interface{}, error) {
	err := dao.db.NewSelect().Column("id").Model(entity).
		Where("name = ?", name).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (dao *entityDAO) GetIdByNamePartnerOrg(ctx context.Context, name string, pid uuid.NullUUID, oid uuid.NullUUID, entity interface{}) (interface{}, error) {
	sq := dao.db.NewSelect().Column("id").Model(entity)
	if oid.Valid {
		sq = sq.Where("organization_id = ?", oid)
	}
	if pid.Valid {
		sq = sq.Where("partner_id = ?", pid)
	}
	sq = sq.Where("name = ?", name).
		Where("trash = ?", false)

	err := sq.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (dao *entityDAO) GetNameById(ctx context.Context, id uuid.UUID, entity interface{}) (interface{}, error) {
	err := dao.db.NewSelect().Column("name").Model(entity).
		Where("id = ?", id).
		Where("trash = ?", false).
		Scan(ctx)
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
	if _, err := dao.db.NewUpdate().Model(entity).Where("? = ?", bun.Ident(field), value).Exec(ctx); err != nil {
		return nil, err
	}
	return entity, nil
}

func (dao *entityDAO) Delete(ctx context.Context, id uuid.UUID, entity interface{}) error {
	_, err := dao.db.NewUpdate().
		Model(entity).
		Column("trash").
		Where("id  = ?", id).
		Set("trash = ?", true).
		Exec(ctx)
	return err
}

func (dao *entityDAO) DeleteX(ctx context.Context, field string, value interface{}, entity interface{}) error {
	_, err := dao.db.NewUpdate().
		Model(entity).
		Column("trash").
		Where("? = ?", bun.Ident(field), value).
		Set("trash = ?", true).
		Exec(ctx)
	return err
}

// HardDeleteAll deletes all records in a table (primarily for use in scripts)
func (dao *entityDAO) HardDeleteAll(ctx context.Context, entity interface{}) error {
	_, err := dao.db.NewDelete().
		Model(entity).
		Where("1  = 1"). // TODO: see how to remove this
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
	sq.Where("trash = false")
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
	sq.Where("trash = false")
	err := sq.Scan(ctx)
	return err
}

func (dao *entityDAO) ListAll(ctx context.Context, entities interface{}) (interface{}, error) {
	err := dao.db.NewSelect().Model(entities).Scan(ctx)
	return entities, err
}

func (dao *entityDAO) GetByTraits(ctx context.Context, name string, entity interface{}) (interface{}, error) {
	// TODO: better name and possibly pass in trait name
	err := dao.db.NewSelect().Model(entity).
		Where("traits ->> 'email' = ?", name).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (dao *entityDAO) GetIdByTraits(ctx context.Context, name string, entity interface{}) (interface{}, error) {
	// TODO: better name and possibly pass in trait name
	err := dao.db.NewSelect().Column("id").Model(entity).
		Where("traits ->> 'email' = ?", name).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (dao *entityDAO) GetInstance() *bun.DB {
	return dao.db
}
