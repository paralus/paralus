package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
)

const (
	KratosPasswordType = "password"
	KratosOidcType     = "oidc"
)

func Create(ctx context.Context, db bun.IDB, entity interface{}) (interface{}, error) {
	if _, err := db.NewInsert().Model(entity).Exec(ctx); err != nil {
		return nil, err
	}

	return entity, nil
}

func GetX(ctx context.Context, db bun.IDB, field string, value interface{}, entity interface{}) (interface{}, error) {
	err := db.NewSelect().Model(entity).
		Where(fmt.Sprintf("%s = ?", field), value).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// M for multi ;)
func GetM(ctx context.Context, db bun.IDB, checks map[string]interface{}, entity interface{}) (interface{}, error) {
	// Can we get the checks directly from entity and create an upsert sort of func?
	q := db.NewSelect().Model(entity)
	for field := range checks {
		q.Where(fmt.Sprintf("%s = ?", field), checks[field])
	}
	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func GetByID(ctx context.Context, db bun.IDB, id uuid.UUID, entity interface{}) (interface{}, error) {
	err := db.NewSelect().Model(entity).
		Where("id = ?", id).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func GetByName(ctx context.Context, db bun.IDB, name string, entity interface{}) (interface{}, error) {
	err := db.NewSelect().Model(entity).
		Where("name = ?", name).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func GetByNamePartnerOrg(ctx context.Context, db bun.IDB, name string, pid uuid.NullUUID, oid uuid.NullUUID, entity interface{}) (interface{}, error) {
	sq := db.NewSelect().Model(entity)
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

func GetIdByName(ctx context.Context, db bun.IDB, name string, entity interface{}) (interface{}, error) {
	return GetAttributesByName(ctx, db, name, entity, "id")
}

func GetAttributesByName(ctx context.Context, db bun.IDB, name string, entity interface{}, columns ...string) (interface{}, error) {
	err := db.NewSelect().Column(columns...).Model(entity).
		Where("name = ?", name).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func GetIdByNamePartnerOrg(ctx context.Context, db bun.IDB, name string, pid uuid.NullUUID, oid uuid.NullUUID, entity interface{}) (interface{}, error) {
	sq := db.NewSelect().Column("id").Model(entity)
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

func GetNameById(ctx context.Context, db bun.IDB, id uuid.UUID, entity interface{}) (interface{}, error) {
	err := db.NewSelect().Column("name").Model(entity).
		Where("id = ?", id).
		Where("trash = ?", false).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func GetNamesByIds(ctx context.Context, db bun.IDB, id []uuid.UUID, entity interface{}) ([]string, error) {
	names := []string{}
	if len(id) == 0 {
		return names, nil
	}
	err := db.NewSelect().Column("name").Model(entity).
		Where("id = (?)", bun.In(id)).
		Where("trash = ?", false).
		Scan(ctx, &names)
	if err != nil {
		return nil, err
	}

	return names, nil
}

func Update(ctx context.Context, db bun.IDB, id uuid.UUID, entity interface{}) (interface{}, error) {
	if _, err := db.NewUpdate().Model(entity).Where("id  = ?", id).Exec(ctx); err != nil {
		return nil, err
	}
	return entity, nil
}

func UpdateX(ctx context.Context, db bun.IDB, field string, value interface{}, entity interface{}) (interface{}, error) {
	if _, err := db.NewUpdate().Model(entity).Where("? = ?", bun.Ident(field), value).Exec(ctx); err != nil {
		return nil, err
	}
	return entity, nil
}

func Delete(ctx context.Context, db bun.IDB, id uuid.UUID, entity interface{}) error {
	_, err := db.NewUpdate().
		Model(entity).
		Column("trash").
		Where("id  = ?", id).
		Where("trash = false").
		Set("trash = ?", true).
		Exec(ctx)
	return err
}

func DeleteX(ctx context.Context, db bun.IDB, field string, value interface{}, entity interface{}) error {
	_, err := db.NewUpdate().
		Model(entity).
		Column("trash").
		Where("? = ?", bun.Ident(field), value).
		Where("trash = false").
		Set("trash = ?", true).
		Exec(ctx)
	return err
}

// DeleteR delete and returns the changed items
func DeleteR(ctx context.Context, db bun.IDB, id uuid.UUID, entity interface{}) error {
	_, err := db.NewUpdate().
		Model(entity).
		Column("trash").
		Where("id  = ?", id).
		Where("trash = false").
		Set("trash = ?", true).
		Returning("*").
		Exec(ctx)
	return err
}

// DeleteXR delete with selector and returns the changed items
func DeleteXR(ctx context.Context, db bun.IDB, field string, value interface{}, entity interface{}) error {
	_, err := db.NewUpdate().
		Model(entity).
		Column("trash").
		Where("? = ?", bun.Ident(field), value).
		Where("trash = false").
		Set("trash = ?", true).
		Returning("*").
		Exec(ctx)
	return err
}

// HardDeleteAll deletes all records in a table (primarily for use in scripts)
func HardDeleteAll(ctx context.Context, db bun.IDB, entity interface{}) error {
	_, err := db.NewDelete().
		Model(entity).
		Where("1 = 1"). // TODO: see how to remove this
		Exec(ctx)
	return err
}

func List(ctx context.Context, db bun.IDB, partnerId uuid.NullUUID, organizationId uuid.NullUUID, entities interface{}) (interface{}, error) {
	sq := db.NewSelect().Model(entities)
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

// TODO: Should we simplify this (less args)?
func ListFiltered(ctx context.Context, db bun.IDB,
	partnerId uuid.NullUUID, organizationId uuid.NullUUID,
	projectId uuid.NullUUID, entities interface{},
	query string, orderBy string, order string,
	limit int, offset int) (interface{}, error) {
	sq := db.NewSelect().Model(entities)
	if query != "" {
		sq = sq.Where("name ILIKE ?", "%"+query+"%") // XXX: ILIKE is not-standard
	}
	if partnerId.Valid {
		sq = sq.Where("partner_id = ?", partnerId)
	}
	if organizationId.Valid {
		sq = sq.Where("organization_id = ?", organizationId)
	}
	if projectId.Valid {
		sq = sq.Where("project_id = ?", projectId)
	}
	sq = sq.Where("trash = ?", false)
	if orderBy != "" && order != "" {
		sq.Order(orderBy + " " + order)
	}
	if limit > 0 {
		sq.Limit(limit)
	}
	if offset > 0 {
		sq.Offset(offset)
	}
	err := sq.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func ListByProject(ctx context.Context, db bun.IDB, partnerId uuid.NullUUID, organizationId uuid.NullUUID, projectId uuid.NullUUID, entities interface{}) error {
	sq := db.NewSelect().Model(entities)
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

func ListAll(ctx context.Context, db bun.IDB, entities interface{}) (interface{}, error) {
	err := db.NewSelect().Model(entities).Scan(ctx)
	return entities, err
}

func GetUserByEmail(ctx context.Context, db bun.IDB, name string, entity interface{}) (interface{}, error) {
	err := db.NewSelect().Model(entity).
		Where("traits ->> 'email' = ?", name).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func GetUserFullByEmail(ctx context.Context, db bun.IDB, name string, entity interface{}) (interface{}, error) {
	err := db.NewSelect().Model(entity).
		Where("traits ->> 'email' = ?", name).
		Relation("IdentityCredential").
		Relation("IdentityCredential.IdentityCredentialType").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func GetUserIdByEmail(ctx context.Context, db bun.IDB, name string, entity interface{}) (interface{}, error) {
	err := db.NewSelect().Column("id").Model(entity).
		Where("traits ->> 'email' = ?", name).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func GetUserLastAuthTime(ctx context.Context, db bun.IDB, userId uuid.UUID) (time.Time, error) {
	var result time.Time
	query := `select max(authenticated_at) from sessions where identity_id = ?`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := db.QueryRowContext(ctx, query, userId).Scan(&result)
	if err != nil {
		switch {
		case err.Error() == `sql: Scan error on column index 0, name "max": unsupported Scan, storing driver.Value type <nil> into type *time.Time`:
			return time.Time{}, nil
		default:
			return time.Time{}, err
		}
	}
	return result, nil
}
