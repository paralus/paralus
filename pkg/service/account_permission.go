package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AccountPermissionService is the interface for account permission operations
type AccountPermissionService interface {
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
	IsSSOAccount(ctx context.Context, accountID string) (bool, error)
}

// accountPermissionService implements AccountPermissionService
type accountPermissionService struct {
	db *bun.DB
}

// NewKubeconfigRevocation return new kubeconfig revocation service
func NewAccountPermissionService(db *bun.DB) AccountPermissionService {
	return &accountPermissionService{db}
}

func (a *accountPermissionService) GetAccountPermissions(ctx context.Context, accountID string, orgID, partnerID string) ([]sentry.AccountPermission, error) {
	aps, err := dao.GetAccountPermissions(ctx, a.db, uuid.MustParse(accountID), uuid.MustParse(orgID), uuid.MustParse(partnerID))
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
	return dao.IsPartnerSuperAdmin(ctx, a.db, uuid.MustParse(accountID), uuid.MustParse(partnerID))
}

func (a *accountPermissionService) GetAccountProjectsByPermission(ctx context.Context, accountID, orgID, partnerID string, permission string) ([]sentry.AccountPermission, error) {
	aps, err := dao.GetAccountProjectsByPermission(ctx, a.db, uuid.MustParse(accountID), uuid.MustParse(orgID), uuid.MustParse(partnerID), permission)
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
	projids := make([]uuid.UUID, len(projects))
	for _, proj := range projects {
		id, err := uuid.Parse(proj)
		if err == nil {
			projids = append(projids, id)
		}
	}
	aps, err := dao.GetAccountPermissionsByProjectIDPermissions(ctx, a.db, uuid.MustParse(accountID), uuid.MustParse(orgID), uuid.MustParse(partnerID), projids, permissions)
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
	return dao.GetAccountBasics(ctx, a.db, uuid.MustParse(accountID))
}

func (a *accountPermissionService) GetAccountGroups(ctx context.Context, accountID string) ([]string, error) {
	ag, err := dao.GetAccountGroups(ctx, a.db, uuid.MustParse(accountID))
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
func (a *accountPermissionService) GetSSOAccount(ctx context.Context, accountID, orgID ctypesv2.ParalusID) (*typesv2.SSOAccountData, error) {
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

// TODO: this needs to be revisited as sso users for oidc are stored in identities by kratos
func (a *accountPermissionService) GetSSOUsersGroupProjectRole(ctx context.Context, orgID string) ([]sentry.SSOAccountGroupProjectRoleData, error) {
	ssos, err := dao.GetSSOUsersGroupProjectRole(ctx, a.db, uuid.MustParse(orgID))
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
	usernames, err := dao.GetAcccountsWithApprovalPermission(ctx, a.db, uuid.MustParse(orgID), uuid.MustParse(partnerID))
	if err != nil {
		return nil, err
	}
	return usernames, nil
}

func (a *accountPermissionService) GetSSOAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID string) ([]string, error) {
	usernames, err := dao.GetSSOAcccountsWithApprovalPermission(ctx, a.db, uuid.MustParse(orgID), uuid.MustParse(partnerID))
	if err != nil {
		return nil, err
	}
	return usernames, nil
}

func (a *accountPermissionService) IsOrgAdmin(ctx context.Context, accountID, partnerID string) (isOrgAdmin bool, err error) {
	return dao.IsOrgAdmin(ctx, a.db, uuid.MustParse(accountID), uuid.MustParse(partnerID))
}

func (a *accountPermissionService) IsAccountActive(ctx context.Context, accountID, orgID string) (bool, error) {
	active := false

	group, err := dao.GetDefaultUserGroup(ctx, a.db, uuid.MustParse(orgID))
	if err != nil {
		return false, err
	}
	ga, err := dao.GetDefaultUserGroupAccount(ctx, a.db, uuid.MustParse(accountID), group.ID)
	if err != nil {
		return active, err
	}
	return ga.Active, nil
}

func (a *accountPermissionService) IsSSOAccount(ctx context.Context, accountID string) (bool, error) {
	return dao.IsSSOAccount(ctx, a.db, uuid.MustParse(accountID))
}

func prepareAccountPermissionResponse(aps models.AccountPermission) sentry.AccountPermission {
	var urls []*sentry.PermissionURL
	if aps.Urls != nil {
		json.Unmarshal(aps.Urls, &urls)
	}
	return sentry.AccountPermission{
		AccountID:      aps.AccountId.String(),
		ProjectID:      aps.ProjectId.String(),
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
