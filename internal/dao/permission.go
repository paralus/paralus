package dao

import (
	"context"

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// PermissionDao is the interface for permission operations
type PermissionDao interface {
	GetGroupPermissions(ctx context.Context, groupNames []string, orgID, partnerID uuid.UUID) ([]models.GroupPermission, error)
	GetGroupProjectsByPermission(ctx context.Context, groupNames []string, orgID, partnerID uuid.UUID, permission string) ([]models.GroupPermission, error)
	GetGroupPermissionsByProjectIDPermissions(ctx context.Context, groupNames []string, orgID, partnerID uuid.UUID, projects []string, permissions []string) ([]models.GroupPermission, error)
	GetProjectByGroup(ctx context.Context, groupNames []string, orgID, partnerID uuid.UUID) ([]models.GroupPermission, error)
	GetAccountPermissions(ctx context.Context, accountID, orgID, partnerID uuid.UUID) ([]models.AccountPermission, error)
	IsPartnerSuperAdmin(ctx context.Context, accountID, partnerID uuid.UUID) (isPartnerAdmin, isSuperAdmin bool, err error)
	GetAccountProjectsByPermission(ctx context.Context, accountID, orgID, partnerID uuid.UUID, permission string) ([]models.AccountPermission, error)
	GetAccountPermissionsByProjectIDPermissions(ctx context.Context, accountID, orgID, partnerID uuid.UUID, projects []uuid.UUID, permissions []string) ([]models.AccountPermission, error)
	GetSSOUsersGroupProjectRole(ctx context.Context, orgID uuid.UUID) ([]models.SSOAccountGroupProjectRole, error)
	GetAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID uuid.UUID) ([]string, error)
	GetSSOAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID uuid.UUID) ([]string, error)
	IsOrgAdmin(ctx context.Context, accountID, partnerID uuid.UUID) (isOrgAdmin bool, err error)
	GetAccountBasics(ctx context.Context, accountID uuid.UUID) (*models.Account, error)
	GetAccountGroups(ctx context.Context, accountID uuid.UUID) ([]models.GroupAccount, error)
	GetDefaultUserGroup(ctx context.Context, orgID uuid.UUID) (*models.Group, error)
	GetDefaultUserGroupAccount(ctx context.Context, accountID, groupID uuid.UUID) (*models.GroupAccount, error)
	GetDefaultAccountProject(ctx context.Context, accountID uuid.UUID) (models.AccountPermission, error)
}

// permissionDao implements PermissionDao
type permissionDao struct {
	dao pg.EntityDAO
}

// PermissionDao return new permission dao
func NewPermissionDao(edao pg.EntityDAO) PermissionDao {
	return &permissionDao{
		dao: edao,
	}
}

func (s *permissionDao) GetGroupPermissions(ctx context.Context, groupNames []string, orgID, partnerID uuid.UUID) ([]models.GroupPermission, error) {
	var gps []models.GroupPermission
	err := s.dao.GetInstance().NewSelect().Model(&gps).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("group_name IN (?)", bun.In(groupNames)).
		Scan(ctx)
	return gps, err
}

func (s *permissionDao) GetGroupProjectsByPermission(ctx context.Context, groupNames []string, orgID, partnerID uuid.UUID, permission string) ([]models.GroupPermission, error) {
	var gps []models.GroupPermission

	err := s.dao.GetInstance().NewSelect().Model(&gps).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("group_name IN (?)", bun.In(groupNames)).
		Where("permission_name = ?", permission).
		Scan(ctx)

	return gps, err
}

func (s *permissionDao) GetGroupPermissionsByProjectIDPermissions(ctx context.Context, groupNames []string, orgID, partnerID uuid.UUID, projects []string, permissions []string) ([]models.GroupPermission, error) {
	var gps []models.GroupPermission

	err := s.dao.GetInstance().NewSelect().Model(&gps).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("group_name IN (?)", bun.In(groupNames)).
		Where("project_id IN (?)", bun.In(projects)).
		Where("permission_name IN (?)", bun.In(permissions)).
		Scan(ctx)

	return gps, err
}

func (s *permissionDao) GetProjectByGroup(ctx context.Context, groupNames []string, orgID, partnerID uuid.UUID) ([]models.GroupPermission, error) {
	var gps []models.GroupPermission

	err := s.dao.GetInstance().NewSelect().Model(&gps).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("group_name IN (?)", bun.In(groupNames)).
		DistinctOn("project_id", "project_name", "role_name", "scope", "group_name").
		Scan(ctx)

	return gps, err
}

func (a *permissionDao) GetAccountPermissions(ctx context.Context, accountID, orgID, partnerID uuid.UUID) ([]models.AccountPermission, error) {
	var aps []models.AccountPermission

	err := a.dao.GetInstance().NewSelect().Model(&aps).
		Where("account_id = ?", accountID).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Scan(ctx)

	return aps, err
}

func (a *permissionDao) IsPartnerSuperAdmin(ctx context.Context, accountID, partnerID uuid.UUID) (isPartnerAdmin, isSuperAdmin bool, err error) {
	var aps []models.AccountPermission

	isSuperAdmin = false
	isPartnerAdmin = false

	err = a.dao.GetInstance().NewSelect().Model(&aps).
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

func (a *permissionDao) GetAccountProjectsByPermission(ctx context.Context, accountID, orgID, partnerID uuid.UUID, permission string) ([]models.AccountPermission, error) {
	var aps []models.AccountPermission

	err := a.dao.GetInstance().NewSelect().Model(&aps).
		Where("account_id = ?", accountID).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("permission_name = ?", permission).
		Scan(ctx)

	return aps, err
}

func (a *permissionDao) GetDefaultAccountProject(ctx context.Context, accountID uuid.UUID) (models.AccountPermission, error) {
	var aps models.AccountPermission

	err := a.dao.GetInstance().NewSelect().Model(&aps).
		ColumnExpr("sap.*").
		Join("JOIN authsrv_project as proj").JoinOn("proj.id = sap.project_id").JoinOn("proj.default = ?", true).
		Where("account_id = ?", accountID).Limit(1).
		Scan(ctx)

	return aps, err
}

func (a *permissionDao) GetAccountPermissionsByProjectIDPermissions(ctx context.Context, accountID, orgID, partnerID uuid.UUID, projects []uuid.UUID, permissions []string) ([]models.AccountPermission, error) {
	var aps []models.AccountPermission

	err := a.dao.GetInstance().NewSelect().Model(&aps).
		Where("account_id = ?", accountID).
		Where("organization_id = ?", orgID).
		Where("partner_id = ?", partnerID).
		Where("project_id IN (?)", bun.In(projects)).
		Where("permission_name IN (?)", bun.In(permissions)).
		Scan(ctx)

	return aps, err
}

func (a *permissionDao) GetSSOUsersGroupProjectRole(ctx context.Context, orgID uuid.UUID) ([]models.SSOAccountGroupProjectRole, error) {
	var ssos []models.SSOAccountGroupProjectRole

	err := a.dao.GetInstance().NewSelect().Model(&ssos).
		Where("organization_id = ?", orgID).
		Scan(ctx)

	return ssos, err
}

func (a *permissionDao) GetAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID uuid.UUID) ([]string, error) {
	// TODO: remove this from here once Account is structured in types.proto
	type accountPermission struct {
		bun.BaseModel `bun:"table:sentry_account_permission,alias:sap"`
		Username      string `json:"username,omitempty"`
		*models.AccountPermission
	}
	var aps []accountPermission
	err := a.dao.GetInstance().NewSelect().Model(&aps).
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

func (a *permissionDao) GetSSOAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID uuid.UUID) ([]string, error) {
	var ssoaps []models.SSOAccountGroupProjectRole
	err := a.dao.GetInstance().NewSelect().Model(&ssoaps).
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

func (a *permissionDao) IsOrgAdmin(ctx context.Context, accountID, partnerID uuid.UUID) (isOrgAdmin bool, err error) {
	var aps []models.AccountPermission

	isOrgAdmin = false

	err = a.dao.GetInstance().NewSelect().Model(&aps).
		Where("account_id = ?", accountID).
		Where("partner_id = ?", partnerID).
		Where("role_name = ?", "ADMIN").
		Where("scope = ?", "ORGANIZATION").
		Scan(ctx)
	if err != nil {
		return isOrgAdmin, err
	}

	for _, ap := range aps {
		if ap.RoleName == "ADMIN" && ap.Scope == "ORGANIZATION" {
			isOrgAdmin = true
			break
		}
	}

	return isOrgAdmin, nil
}

func (a *permissionDao) GetAccountBasics(ctx context.Context, accountID uuid.UUID) (*models.Account, error) {
	var acc models.Account

	err := a.dao.GetInstance().NewSelect().Model(&acc).
		Column("identities.id", "traits", "state").
		ColumnExpr("max(ks.authenticated_at) as lastlogin").
		ColumnExpr("identities.traits -> 'email' as username").
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

func (a *permissionDao) GetAccountGroups(ctx context.Context, accountID uuid.UUID) ([]models.GroupAccount, error) {
	var ga []models.GroupAccount

	err := a.dao.GetInstance().NewSelect().Model(&ga).
		Where("account_id = ?", accountID).
		Where("trash = ?", false).
		Where("active = ?", true).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return ga, nil
}

func (a *permissionDao) GetDefaultUserGroup(ctx context.Context, orgID uuid.UUID) (*models.Group, error) {
	var g models.Group
	err := a.dao.GetInstance().NewSelect().Model(&g).
		Where("organization_id = ?", orgID).
		Where("type = ?", "DEFAULT_USERS").
		Where("trash = ?", false).
		Scan(ctx)
	return &g, err
}

func (a *permissionDao) GetDefaultUserGroupAccount(ctx context.Context, accountID, groupID uuid.UUID) (*models.GroupAccount, error) {
	var ga models.GroupAccount
	err := a.dao.GetInstance().NewSelect().Model(&ga).
		Where("account_id = ?", accountID).
		Where("group_id = ?", groupID).
		Where("trash = ?", false).
		Scan(ctx)
	return &ga, err
}
