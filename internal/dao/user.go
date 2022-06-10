package dao

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/uptrace/bun"
)

func GetUserType(ctx context.Context, db bun.IDB, id uuid.UUID) (string, error) {
	var user = models.KratosIdentities{}
	q := db.NewSelect().Model(&user)
	q.Relation("IdentityCredential").
		Relation("IdentityCredential.IdentityCredentialType")
	q.Where("id = ?", id)
	err := q.Scan(ctx)
	if err != nil {
		return "", err
	}
	return user.IdentityCredential.IdentityCredentialType.Name, nil
}

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
		ColumnExpr("authsrv_resourcerole.name as role, authsrv_project.name as project, namespace").
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

	if role != uuid.Nil {
		q.Where("role_id = ?", role)
	}
	if len(projects) != 0 {
		q.Where("project_id IN (?)", bun.In(projects))
	}

	if group != uuid.Nil {
		// If the group is not mapped to a project, we won't be able
		// to pick it from sentry table and that is why we have to do
		// this.
		gaccs := []models.GroupAccount{}
		subq := db.NewSelect().
			Model(&gaccs).
			ColumnExpr("DISTINCT account_id").
			Where("group_id = ?", group).
			Where("trash = ?", false)

		q = q.Where("account_id IN (?)", subq)
	}
	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}
	acc := []uuid.UUID{}
	for _, a := range p {
		acc = append(acc, a.AccountId)
	}
	return acc, nil

}

func listFilteredUsersQuery(
	q *bun.SelectQuery,
	fusers []uuid.UUID,
	query string,
	utype string,
	orderBy string,
	order string,
	limit int,
	offset int,
) *bun.SelectQuery {
	if utype != "" {
		q = q.Relation("IdentityCredential", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.ExcludeColumn("*")
		}).
			Relation("IdentityCredential.IdentityCredentialType", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.ExcludeColumn("*").Where("name = ?", utype)
			})
	}

	if len(fusers) > 0 {
		// filter with precomputed users if we have any
		q = q.Where("identities.id IN (?)", bun.In(fusers))
	}
	if query != "" {
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			q = q.Where("traits ->> 'email' ILIKE ?", "%"+query+"%") // XXX: ILIKE is not-standard sql
			q = q.WhereOr("traits ->> 'first_name' ILIKE ?", "%"+query+"%")
			q = q.WhereOr("traits ->> 'last_name' ILIKE ?", "%"+query+"%")
			return q
		})
	}
	if orderBy != "" && order != "" {
		q = q.Order("traits ->> '" + orderBy + "' " + order)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	return q
}

// ListFilteredUsers will return the list of users fileterd by query
func ListFilteredUsers(
	ctx context.Context,
	db bun.IDB,
	fusers []uuid.UUID,
	query string,
	utype string,
	orderBy string,
	order string,
	limit int,
	offset int,
) ([]models.KratosIdentities, error) {
	var users []models.KratosIdentities
	q := db.NewSelect().Model(&users)
	listFilteredUsersQuery(q, fusers, query, utype, orderBy, order, limit, offset)

	//restrict oidc users, this is required as kratos creates entry with credential type password for oidc users as well
	if utype == KratosPasswordType {
		var ssousers []models.KratosIdentities
		oq := db.NewSelect().Model(&ssousers).
			Relation("IdentityCredential", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.ExcludeColumn("*")
			}).
			Relation("IdentityCredential.IdentityCredentialType", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.ExcludeColumn("*").Where("name = ?", KratosOidcType)
			})
		if len(fusers) > 0 {
			// filter with precomputed users if we have any
			oq = oq.Where("identities.id IN (?)", bun.In(fusers))
		}
		q.Except(oq)
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ListFilteredUsersWithGroup is ListFilteredUsers but with Group fileter as well
func ListFilteredUsersWithGroup(
	ctx context.Context,
	db bun.IDB,
	fusers []uuid.UUID,
	group uuid.UUID,
	query string,
	utype string,
	orderBy string,
	order string,
	limit int,
	offset int,
) ([]models.KratosIdentities, error) {
	var users []models.KratosIdentities
	var gaccs []models.GroupAccount
	q := db.NewSelect().Model(&gaccs)
	q.Where("group_id = ?", group).Where("groupaccount.trash = ?", false)

	q = q.Relation("Account", func(q *bun.SelectQuery) *bun.SelectQuery {
		return listFilteredUsersQuery(q, fusers, query, utype, orderBy, order, limit, offset)
	})
	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}
	for _, ga := range gaccs {
		users = append(users, *ga.Account)
	}
	return users, nil
}

func GetUserNamesByIds(ctx context.Context, db bun.IDB, id []uuid.UUID, entity interface{}) ([]string, error) {
	names := []string{}
	if len(id) == 0 {
		return names, nil
	}
	err := db.NewSelect().ColumnExpr("traits ->> 'email' as name").Model(entity).
		Where("id = (?)", bun.In(id)).
		Scan(ctx, &names)
	if err != nil {
		return nil, err
	}
	return names, nil
}

func IsSSOAccount(ctx context.Context, db bun.IDB, id uuid.UUID) (bool, error) {
	var user models.KratosIdentities
	q := db.NewSelect().Model(&user)
	q = q.Relation("IdentityCredential").
		Relation("IdentityCredential.IdentityCredentialType", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("name = ?", KratosOidcType)
		})
	q = q.Where("identities.id = ?", id)
	err := q.Scan(ctx)
	if err != nil && err == sql.ErrNoRows {
		return false, nil
	} else if err == nil && user.ID != uuid.Nil {
		return true, nil
	}
	return false, err
}
