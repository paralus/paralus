package dao

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	"github.com/uptrace/bun"
)

func GetGroupPermissions(ctx context.Context, db bun.IDB, groupNames []string, orgID, partnerID uuid.UUID) ([]models.GroupPermission, error) {
	var gps []models.GroupPermission
	err := db.NewSelect().Model(&gps).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("group_name IN (?)", bun.In(groupNames)).
		Scan(ctx)
	return gps, err
}

func GetGroupProjectsByPermission(ctx context.Context, db bun.IDB, groupNames []string, orgID, partnerID uuid.UUID, permission string) ([]models.GroupPermission, error) {
	var gps []models.GroupPermission

	err := db.NewSelect().Model(&gps).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("group_name IN (?)", bun.In(groupNames)).
		Where("permission_name = ?", permission).
		Scan(ctx)

	return gps, err
}

func GetGroupPermissionsByProjectIDPermissions(ctx context.Context, db bun.IDB, groupNames []string, orgID, partnerID uuid.UUID, projects []string, permissions []string) ([]models.GroupPermission, error) {
	var gps []models.GroupPermission

	err := db.NewSelect().Model(&gps).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("group_name IN (?)", bun.In(groupNames)).
		Where("project_id IN (?)", bun.In(projects)).
		Where("permission_name IN (?)", bun.In(permissions)).
		Scan(ctx)

	return gps, err
}

func GetProjectByGroup(ctx context.Context, db bun.IDB, groupNames []string, orgID, partnerID uuid.UUID) ([]models.GroupPermission, error) {
	var gps []models.GroupPermission

	err := db.NewSelect().Model(&gps).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("group_name IN (?)", bun.In(groupNames)).
		DistinctOn("project_id", "project_name", "role_name", "scope", "group_name").
		Scan(ctx)

	return gps, err
}

func GetAccountPermissions(ctx context.Context, db bun.IDB, accountID, orgID, partnerID uuid.UUID) ([]models.AccountPermission, error) {
	var aps []models.AccountPermission

	err := db.NewSelect().Model(&aps).
		Where("account_id = ?", accountID).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Scan(ctx)

	return aps, err
}

func IsPartnerSuperAdmin(ctx context.Context, db bun.IDB, accountID, partnerID uuid.UUID) (isPartnerAdmin, isSuperAdmin bool, err error) {
	var aps []models.AccountPermission

	isSuperAdmin = false
	isPartnerAdmin = false

	err = db.NewSelect().Model(&aps).
		Where("account_id = ?", accountID).
		Where("partner_id = ?", partnerID).
		WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
			sq = sq.WhereOr("role_name = ?", "PARTNER_ADMIN").WhereOr("role_name = ?", "SUPER_ADMIN")
			return sq
		}).
		Scan(ctx)
	if err != nil {
		return isPartnerAdmin, isSuperAdmin, err
	}

	for _, ap := range aps {
		if ap.RoleName == "SUPER_ADMIN" {
			isSuperAdmin = true
		} else if ap.RoleName == "PARTNER_ADMIN" {
			isPartnerAdmin = true
		}
	}

	return isPartnerAdmin, isSuperAdmin, nil
}

func GetAccountProjectsByPermission(ctx context.Context, db bun.IDB, accountID, orgID, partnerID uuid.UUID, permission string) ([]models.AccountPermission, error) {
	var aps []models.AccountPermission

	err := db.NewSelect().Model(&aps).
		Where("account_id = ?", accountID).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("permission_name = ?", permission).
		Scan(ctx)

	return aps, err
}

func GetDefaultAccountProject(ctx context.Context, db bun.IDB, accountID uuid.UUID) (models.AccountPermission, error) {
	var aps models.AccountPermission

	err := db.NewSelect().Model(&aps).
		ColumnExpr("sap.*").
		Join("JOIN authsrv_project as proj").JoinOn("proj.id = sap.project_id").JoinOn("proj.default = ?", true).
		Where("account_id = ?", accountID).Limit(1).
		Scan(ctx)

	return aps, err
}

func GetAccountPermissionsByProjectIDPermissions(ctx context.Context, db bun.IDB, accountID, orgID, partnerID uuid.UUID, projects []uuid.UUID, permissions []string) ([]models.AccountPermission, error) {
	var aps []models.AccountPermission

	err := db.NewSelect().Model(&aps).
		Where("account_id = ?", accountID).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("project_id IN (?)", bun.In(projects)).
		Where("permission_name IN (?)", bun.In(permissions)).
		Scan(ctx)

	return aps, err
}

func GetSSOUsersGroupProjectRole(ctx context.Context, db bun.IDB, orgID uuid.UUID) ([]models.SSOAccountGroupProjectRole, error) {
	var ssos []models.SSOAccountGroupProjectRole

	err := db.NewSelect().Model(&ssos).
		Where("organization_id = ?", orgID).
		Scan(ctx)

	return ssos, err
}

func GetAcccountsWithApprovalPermission(ctx context.Context, db bun.IDB, orgID, partnerID uuid.UUID) ([]string, error) {
	// TODO: remove this from here once Account is structured in types.proto
	type accountPermission struct {
		bun.BaseModel `bun:"table:sentry_account_permission,alias:sap"`
		Username      string `json:"username,omitempty"`
		*models.AccountPermission
	}
	var aps []accountPermission
	err := db.NewSelect().Model(&aps).
		ColumnExpr("ki.traits -> 'email'").
		DistinctOn("ki.traits -> 'email'").
		Join("INNER JOIN identities as ki ON ?TableAlias.account_id = ki.id").
		Where("?TableAlias.organization_id = ?", orgID).
		Where("?TableAlias.partner_id = ?", partnerID).
		WhereGroup("grp", func(sq *bun.SelectQuery) *bun.SelectQuery {
			sq = sq.WhereOr("?TableAlias.role_name = ?", "ADMIN").WhereOr("?TableAlias.role_name = ?", "PROJECT_ADMIN")
			return sq
		}).
		Where("ki.state = ?", "ACTIVE").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	var usernames []string
	for _, ap := range aps {
		usernames = append(usernames, ap.Username)
	}

	return usernames, nil
}

func GetSSOAcccountsWithApprovalPermission(ctx context.Context, db bun.IDB, orgID, partnerID uuid.UUID) ([]string, error) {
	var ssoaps []models.SSOAccountGroupProjectRole
	err := db.NewSelect().Model(&ssoaps).
		Where("?TableAlias.organization_id = ?", orgID).
		Where("?TableAlias.partner_id = ?", partnerID).
		WhereGroup("grp", func(sq *bun.SelectQuery) *bun.SelectQuery {
			sq = sq.WhereOr("?TableAlias.role_name = ?", "ADMIN").WhereOr("?TableAlias.role_name = ?", "PROJECT_ADMIN")
			return sq
		}).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	visited := make(map[string]bool)
	var usernames []string
	for _, ssoap := range ssoaps {
		if _, ok := visited[ssoap.Username]; !ok {
			usernames = append(usernames, ssoap.Username)
			visited[ssoap.Username] = true
		}
	}
	return usernames, nil
}

func IsOrgAdmin(ctx context.Context, db bun.IDB, accountID, partnerID uuid.UUID) (isOrgAdmin bool, err error) {
	var aps []models.AccountPermission

	isOrgAdmin = false

	err = db.NewSelect().Model(&aps).
		Where("account_id = ?", accountID).
		Where("partner_id = ?", partnerID).
		Where("lower(role_name) = ?", "admin").
		Where("lower(scope) = ?", "organization").
		Scan(ctx)
	if err != nil {
		return isOrgAdmin, err
	}

	for _, ap := range aps {
		if strings.ToLower(ap.RoleName) == "admin" && strings.ToLower(ap.Scope) == "organization" {
			isOrgAdmin = true
			break
		}
	}

	return isOrgAdmin, nil
}

func IsOrgReadOnly(ctx context.Context, db bun.IDB, accountID, organizationID uuid.UUID, partnerID uuid.UUID) (isOrgReadOnly bool, err error) {
	var aps []models.AccountPermission

	isOrgReadOnly = false

	err = db.NewSelect().Model(&aps).
		Where("account_id = ?", accountID).
		Where("organization_id = ?", organizationID).
		Where("partner_id = ?", partnerID).
		Where("lower(role_name) = ?", "admin_read_only").
		Where("lower(scope) = ?", "organization").
		Scan(ctx)
	if err != nil {
		return isOrgReadOnly, err
	}

	for _, ap := range aps {
		if strings.ToLower(ap.RoleName) == "admin_read_only" && strings.ToLower(ap.Scope) == "organization" {
			isOrgReadOnly = true
			break
		}
	}

	return isOrgReadOnly, nil
}

func GetAccountBasics(ctx context.Context, db bun.IDB, accountID uuid.UUID) (*models.Account, error) {
	var acc models.Account

	err := db.NewSelect().Model(&acc).
		Column("identities.id", "traits", "state").
		ColumnExpr("max(ks.authenticated_at) as lastlogin").
		ColumnExpr("identities.traits ->> 'email' as username").
		Join("INNER JOIN sessions as ks ON identities.id = ks.identity_id").
		Where("identities.id = ?", accountID).
		Where("ks.active = ?", true).
		Group("identities.id").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func GetAccountGroups(ctx context.Context, db bun.IDB, accountID uuid.UUID) ([]models.GroupAccount, error) {
	var ga []models.GroupAccount

	err := db.NewSelect().Model(&ga).
		Where("account_id = ?", accountID).
		Where("trash = ?", false).
		Where("active = ?", true).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return ga, nil
}

func GetDefaultUserGroup(ctx context.Context, db bun.IDB, orgID uuid.UUID) (*models.Group, error) {
	var g models.Group
	err := db.NewSelect().Model(&g).
		Where("organization_id = ?", orgID).
		Where("type = ?", "DEFAULT_USERS").
		Where("trash = ?", false).
		Scan(ctx)
	return &g, err
}

func GetDefaultUserGroupAccount(ctx context.Context, db bun.IDB, accountID, groupID uuid.UUID) (*models.GroupAccount, error) {
	var ga models.GroupAccount
	err := db.NewSelect().Model(&ga).
		Where("account_id = ?", accountID).
		Where("group_id = ?", groupID).
		Where("trash = ?", false).
		Scan(ctx)
	return &ga, err
}
