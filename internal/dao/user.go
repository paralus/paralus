package dao

import (
	"context"

	"github.com/RafayLabs/rcloud-base/internal/models"
	userv3 "github.com/RafayLabs/rcloud-base/proto/types/userpb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func GetGroups(ctx context.Context, db bun.IDB, id uuid.UUID) ([]models.Group, error) {
	var entities = []models.Group{}
	err := db.NewSelect().Model(&entities).
		Join(`JOIN authsrv_groupaccount ON authsrv_groupaccount.group_id="group".id`).
		Where("authsrv_groupaccount.account_id = ?", id).
		Where("authsrv_groupaccount.trash = ?", false).
		Scan(ctx)
	return entities, err
}

func GetUserRoles(ctx context.Context, db bun.IDB, id uuid.UUID) ([]*userv3.ProjectNamespaceRole, error) {
	// Could possibly union them later for some speedup
	// TODO filter by org and partner
	var r = []*userv3.ProjectNamespaceRole{}
	err := db.NewSelect().Table("authsrv_accountresourcerole").
		ColumnExpr("authsrv_resourcerole.name as role").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_accountresourcerole.role_id`).
		Where("authsrv_accountresourcerole.account_id = ?", id).
		Where("authsrv_resourcerole.trash = ?", false).
		Where("authsrv_accountresourcerole.trash = ?", false).
		Scan(ctx, &r)
	if err != nil {
		return nil, err
	}

	var pr = []*userv3.ProjectNamespaceRole{}
	err = db.NewSelect().Table("authsrv_projectaccountresourcerole").
		ColumnExpr("distinct authsrv_resourcerole.name as role, authsrv_project.name as project").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountresourcerole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountresourcerole.project_id`).
		Where("authsrv_projectaccountresourcerole.account_id = ?", id).
		Where("authsrv_project.trash = ?", false).
		Where("authsrv_resourcerole.trash = ?", false).
		Where("authsrv_projectaccountresourcerole.trash = ?", false).
		Scan(ctx, &pr)
	if err != nil {
		return nil, err
	}

	var pnr = []*userv3.ProjectNamespaceRole{}
	err = db.NewSelect().Table("authsrv_projectaccountnamespacerole").
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project, namespace_id as namespace").
		Join(`JOIN authsrv_resourcerole ON authsrv_resourcerole.id=authsrv_projectaccountnamespacerole.role_id`).
		Join(`JOIN authsrv_project ON authsrv_project.id=authsrv_projectaccountnamespacerole.project_id`). // also need a namespace join
		Where("authsrv_projectaccountnamespacerole.account_id = ?", id).
		Where("authsrv_project.trash = ?", false).
		Where("authsrv_resourcerole.trash = ?", false).
		Where("authsrv_projectaccountnamespacerole.trash = ?", false).
		Scan(ctx, &pnr)
	if err != nil {
		return nil, err
	}

	return append(append(r, pr...), pnr...), err
}

type userProjectnamesaceRole struct {
	AccountId uuid.UUID `bun:"account_id,type:uuid"`
	Role      string    `bun:"role,type:string"`
	Project   *string   `bun:"project,type:string"`
}

func GetQueryFilteredUsers(ctx context.Context, db bun.IDB, partner, org, group, role uuid.UUID, projects []uuid.UUID) ([]uuid.UUID, error) {
	p := []models.AccountPermission{}
	q := db.NewSelect().Model(&p).ColumnExpr("DISTINCT account_id")

	q.Where("partner_id = ?", partner).
		Where("organization_id = ?", org)

	if group != uuid.Nil {
		q.Where("group_id = ?", group)
	}
	if role != uuid.Nil {
		q.Where("role_id = ?", role)
	}
	if len(projects) != 0 {
		q.Where("project_id IN (?)", bun.In(projects))
	}
	q.Scan(ctx)

	acc := []uuid.UUID{}
	for _, a := range p {
		acc = append(acc, a.AccountId)
	}
	return acc, nil
}

// ListFilteredUsers will return the list of users fileterd by query
func ListFilteredUsers(ctx context.Context, db bun.IDB, users *[]models.KratosIdentities, fusers []uuid.UUID, query string, orderBy string, order string, limit int, offset int) (*[]models.KratosIdentities, error) {
	q := db.NewSelect().Model(users)
	if len(fusers) > 0 {
		// filter with precomputed users if we have any
		q.Where("id IN (?)", bun.In(fusers))
	}
	if query != "" {
		q.Where("traits ->> 'email' ILIKE ?", "%"+query+"%") // XXX: ILIKE is not-standard
		q.WhereOr("traits ->> 'first_name' ILIKE ?", "%"+query+"%")
		q.WhereOr("traits ->> 'last_name' ILIKE ?", "%"+query+"%")
	}
	if orderBy != "" && order != "" {
		q.Order("traits ->> '" + orderBy + "' " + order)
	}
	if limit != 0 || offset != 0 {
		q.Limit(limit)
		q.Offset(offset)
	}
	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
