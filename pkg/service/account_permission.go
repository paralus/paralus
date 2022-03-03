package service

import (
	"context"
	"encoding/json"

	"github.com/RafaySystems/rcloud-base/internal/dao"
	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/proto/types/sentry"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AccountPermissionService is the interface for account permission operations
type AccountPermissionService interface {
	Close() error
	GetAccountPermissions(ctx context.Context, accountID string, orgID, partnerID string) ([]sentry.AccountPermission, error)
	IsPartnerSuperAdmin(ctx context.Context, accountID, partnerID string) (isPartnerAdmin, isSuperAdmin bool, err error)
	GetAccountProjectsByPermission(ctx context.Context, accountID, orgID, partnerID string, permission string) ([]sentry.AccountPermission, error)
	GetAccountPermissionsByProjectIDPermissions(ctx context.Context, accountID, orgID, partnerID string, projects, permissions []string) ([]sentry.AccountPermission, error)
	GetAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID string) ([]string, error)
	GetSSOAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID string) ([]string, error)
	IsOrgAdmin(ctx context.Context, accountID, partnerID string) (isOrgAdmin bool, err error)
	GetAccount(ctx context.Context, accountID string) (*models.Account, error)
	GetAccountGroups(ctx context.Context, accountID string) ([]string, error)
	IsAccountActive(ctx context.Context, accountID, orgID string) (bool, error)
}

// accountPermissionService implements AccountPermissionService
type accountPermissionService struct {
	dao  pg.EntityDAO
	pdao dao.PermissionDao
}

// NewKubeconfigRevocation return new kubeconfig revocation service
func NewAccountPermissionService(db *bun.DB) AccountPermissionService {
	edao := pg.NewEntityDAO(db)
	return &accountPermissionService{
		dao:  edao,
		pdao: dao.NewPermissionDao(edao),
	}
}

func (s *accountPermissionService) Close() error {
	return s.dao.Close()
}

func (a *accountPermissionService) GetAccountPermissions(ctx context.Context, accountID string, orgID, partnerID string) ([]sentry.AccountPermission, error) {
	aps, err := a.pdao.GetAccountPermissions(ctx, uuid.MustParse(accountID), uuid.MustParse(orgID), uuid.MustParse(partnerID))
	if err != nil {
		return nil, err
	}
	accountPermissions := []sentry.AccountPermission{}
	for _, ap := range aps {
		accountPermissions = append(accountPermissions, prepareAccountPermissionResponse(ap))
	}

	return accountPermissions, nil
}

func (a *accountPermissionService) IsPartnerSuperAdmin(ctx context.Context, accountID, partnerID string) (isPartnerAdmin, isSuperAdmin bool, err error) {
	return a.pdao.IsPartnerSuperAdmin(ctx, uuid.MustParse(accountID), uuid.MustParse(partnerID))
}

func (a *accountPermissionService) GetAccountProjectsByPermission(ctx context.Context, accountID, orgID, partnerID string, permission string) ([]sentry.AccountPermission, error) {
	aps, err := a.pdao.GetAccountProjectsByPermission(ctx, uuid.MustParse(accountID), uuid.MustParse(orgID), uuid.MustParse(partnerID), permission)
	if err != nil {
		return nil, err
	}
	accountPermissions := []sentry.AccountPermission{}
	for _, ap := range aps {
		accountPermissions = append(accountPermissions, prepareAccountPermissionResponse(ap))
	}

	return accountPermissions, nil
}

func (a *accountPermissionService) GetAccountPermissionsByProjectIDPermissions(ctx context.Context, accountID, orgID, partnerID string, projects, permissions []string) ([]sentry.AccountPermission, error) {
	aps, err := a.pdao.GetAccountPermissionsByProjectIDPermissions(ctx, uuid.MustParse(accountID), uuid.MustParse(orgID), uuid.MustParse(partnerID), projects, permissions)
	if err != nil {
		return nil, err
	}
	accountPermissions := []sentry.AccountPermission{}
	for _, ap := range aps {
		accountPermissions = append(accountPermissions, prepareAccountPermissionResponse(ap))
	}

	return accountPermissions, nil
}

func (a *accountPermissionService) GetAccount(ctx context.Context, accountID string) (*models.Account, error) {
	return a.pdao.GetAccountBasics(ctx, uuid.MustParse(accountID))
}

func (a *accountPermissionService) GetAccountGroups(ctx context.Context, accountID string) ([]string, error) {
	ag, err := a.pdao.GetAccountGroups(ctx, uuid.MustParse(accountID))
	if err != nil {
		return nil, err
	}
	groups := make([]string, 0)
	for _, grp := range ag {
		groups = append(groups, grp.Name)
	}
	return groups, nil
}

/*TODO: to revisit if we end up using sso with crewjam
func (a *accountPermissionService) GetSSOAccount(ctx context.Context, accountID, orgID ctypesv2.RafayID) (*typesv2.SSOAccountData, error) {
	var sso ssoAccountData

	err := a.db.WithContext(ctx).Model(&sso).
		Where("id = ?", accountID).
		Where("organization_id = ?", orgID).
		Where("trash = ?", false).
		Select()
	if err != nil {
		return nil, err
	}
	return sso.SSOAccountData, nil
}
*/

//TODO: this needs to be revisited as sso users for oidc are stored in identities by kratos
func (a *accountPermissionService) GetSSOUsersGroupProjectRole(ctx context.Context, orgID string) ([]sentry.SSOAccountGroupProjectRoleData, error) {
	ssos, err := a.pdao.GetSSOUsersGroupProjectRole(ctx, uuid.MustParse(orgID))
	if err != nil {
		return nil, err
	}

	ssoUsers := []sentry.SSOAccountGroupProjectRoleData{}
	for _, sso := range ssos {
		ssoUsers = append(ssoUsers, prepareSSOAccountGroupProjectRoleData(sso))
	}
	return ssoUsers, nil
}

func (a *accountPermissionService) GetAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID string) ([]string, error) {
	usernames, err := a.pdao.GetAcccountsWithApprovalPermission(ctx, uuid.MustParse(orgID), uuid.MustParse(partnerID))
	if err != nil {
		return nil, err
	}
	return usernames, nil
}

func (a *accountPermissionService) GetSSOAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID string) ([]string, error) {
	usernames, err := a.pdao.GetSSOAcccountsWithApprovalPermission(ctx, uuid.MustParse(orgID), uuid.MustParse(partnerID))
	if err != nil {
		return nil, err
	}
	return usernames, nil
}

func (a *accountPermissionService) IsOrgAdmin(ctx context.Context, accountID, partnerID string) (isOrgAdmin bool, err error) {
	return a.pdao.IsOrgAdmin(ctx, uuid.MustParse(accountID), uuid.MustParse(partnerID))
}

func (a *accountPermissionService) IsAccountActive(ctx context.Context, accountID, orgID string) (bool, error) {
	active := false

	group, err := a.pdao.GetDefaultUserGroup(ctx, uuid.MustParse(orgID))
	if err != nil {
		return false, err
	}
	ga, err := a.pdao.GetDefaultUserGroupAccount(ctx, uuid.MustParse(accountID), group.ID)
	if err != nil {
		return active, err
	}
	return ga.Active, nil
}

/*
func (a *accountPermissionService) GetSSOAccounts(ctx context.Context, orgID ctypesv2.RafayID) ([]typesv2.SSOAccountData, error) {
	var ssoAccounts []ssoAccountData

	err := a.db.WithContext(ctx).Model(&ssoAccounts).
		Where("organization_id = ?", orgID).
		Where("trash = ?", false).
		Select()
	if err != nil {
		return nil, err
	}
	ssoAccountUsers := []typesv2.SSOAccountData{}
	for _, sso := range ssoAccounts {
		ssoAccountUsers = append(ssoAccountUsers, *sso.SSOAccountData)
	}
	return ssoAccountUsers, nil

}
*/
func prepareAccountPermissionResponse(aps models.AccountPermission) sentry.AccountPermission {
	var urls []*sentry.PermissionURL
	if aps.Urls != nil {
		json.Unmarshal(aps.Urls, &urls)
	}
	return sentry.AccountPermission{
		AccountID:      aps.AccountId.String(),
		ProjectID:      aps.ProjecttId,
		OrganizationID: aps.OrganizationId.String(),
		PartnerID:      aps.PartnerId.String(),
		RoleName:       aps.RoleName,
		IsGlobal:       aps.IsGlobal,
		Scope:          aps.Scope,
		PermissionName: aps.PermissionName,
		BaseURL:        aps.BaseUrl,
		Urls:           urls,
	}
}

func prepareSSOAccountGroupProjectRoleData(data models.SSOAccountGroupProjectRole) sentry.SSOAccountGroupProjectRoleData {
	return sentry.SSOAccountGroupProjectRoleData{
		Id:                    data.Id.String(),
		UserName:              data.Username,
		RoleName:              data.RoleName,
		ProjectID:             data.ProjectId,
		ProjectName:           data.ProjectName,
		Group:                 data.GroupName,
		AccountOrganizationID: data.AccountOrganizationId.String(),
		OrganizationID:        data.OrganizationId.String(),
		PartnerID:             data.PartnerId.String(),
		Scope:                 data.Scope,
		LastLogin:             timestamppb.New(data.LastLogin),
		CreatedAt:             timestamppb.New(data.CreatedAt),
		FirstName:             data.FirstName,
		LastName:              data.LastName,
		Phone:                 data.Phone,
		Name:                  data.Name,
		LastLogout:            timestamppb.New(data.LastLogin),
	}
}
