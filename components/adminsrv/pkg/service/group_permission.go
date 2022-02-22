package service

import (
	"context"
	"encoding/json"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/internal/dao"
	"github.com/RafaySystems/rcloud-base/components/adminsrv/internal/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/components/common/proto/types/sentry"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// GroupPermissionService is the interface for group permission operations
type GroupPermissionService interface {
	Close() error
	GetGroupPermissions(ctx context.Context, groupNames []string, orgID, partnerID string) ([]sentry.GroupPermission, error)
	GetGroupProjectsByPermission(ctx context.Context, groupNames []string, orgID, partnerID string, permission string) ([]sentry.GroupPermission, error)
	GetGroupPermissionsByProjectIDPermissions(ctx context.Context, groupNames []string, orgID, partnerID string, projects []string, permissions []string) ([]sentry.GroupPermission, error)
	GetProjectByGroup(ctx context.Context, groupNames []string, orgID, partnerID string) ([]sentry.GroupPermission, error)
}

// groupPermissionService implements GroupPermissionService
type groupPermissionService struct {
	dao  pg.EntityDAO
	pdao dao.PermissionDao
}

// NewKubeconfigRevocation return new kubeconfig revocation service
func NewGroupPermissionService(db *bun.DB) GroupPermissionService {
	edao := pg.NewEntityDAO(db)
	return &groupPermissionService{
		dao:  edao,
		pdao: dao.NewPermissionDao(edao),
	}
}

func (s *groupPermissionService) Close() error {
	return s.dao.Close()
}

func (s *groupPermissionService) GetGroupPermissions(ctx context.Context, groupNames []string, orgID, partnerID string) ([]sentry.GroupPermission, error) {
	gps, err := s.pdao.GetGroupPermissions(ctx, groupNames, uuid.MustParse(orgID), uuid.MustParse(partnerID))
	if err != nil {
		return nil, err
	}
	groupPermissions := []sentry.GroupPermission{}
	for _, gp := range gps {
		groupPermissions = append(groupPermissions, prepareGroupPermissionResponse(gp))
	}

	return groupPermissions, nil
}

func (a *groupPermissionService) GetGroupProjectsByPermission(ctx context.Context, groupNames []string, orgID, partnerID string, permission string) ([]sentry.GroupPermission, error) {
	aps, err := a.pdao.GetGroupProjectsByPermission(ctx, groupNames, uuid.MustParse(orgID), uuid.MustParse(partnerID), permission)
	if err != nil {
		return nil, err
	}
	groupPermissions := []sentry.GroupPermission{}
	for _, ap := range aps {
		groupPermissions = append(groupPermissions, prepareGroupPermissionResponse(ap))
	}

	return groupPermissions, nil
}

func (s *groupPermissionService) GetGroupPermissionsByProjectIDPermissions(ctx context.Context, groupNames []string, orgID, partnerID string, projects []string, permissions []string) ([]sentry.GroupPermission, error) {
	gps, err := s.pdao.GetGroupPermissionsByProjectIDPermissions(ctx, groupNames, uuid.MustParse(orgID), uuid.MustParse(partnerID), projects, permissions)
	if err != nil {
		return nil, err
	}
	groupPermissions := []sentry.GroupPermission{}
	for _, ap := range gps {
		groupPermissions = append(groupPermissions, prepareGroupPermissionResponse(ap))
	}

	return groupPermissions, nil
}

func (s *groupPermissionService) GetProjectByGroup(ctx context.Context, groupNames []string, orgID, partnerID string) ([]sentry.GroupPermission, error) {
	gps, err := s.pdao.GetProjectByGroup(ctx, groupNames, uuid.MustParse(orgID), uuid.MustParse(partnerID))
	if err != nil {
		return nil, err
	}
	groupPermissions := []sentry.GroupPermission{}
	for _, ap := range gps {
		groupPermissions = append(groupPermissions, prepareGroupPermissionResponse(ap))
	}

	return groupPermissions, nil
}

func prepareGroupPermissionResponse(gps models.GroupPermission) sentry.GroupPermission {
	var urls []*sentry.PermissionURL
	if gps.Urls != nil {
		json.Unmarshal(gps.Urls, &urls)
	}
	return sentry.GroupPermission{
		GroupID:        gps.GroupId.String(),
		ProjectID:      gps.ProjecttId,
		OrganizationID: gps.OrganizationId.String(),
		PartnerID:      gps.PartnerId.String(),
		GroupName:      gps.GroupName,
		RoleName:       gps.RoleName,
		IsGlobal:       gps.IsGlobal,
		Scope:          gps.Scope,
		PermissionName: gps.PermissionName,
		BaseURL:        gps.BaseUrl,
		Urls:           urls,
		ProjectName:    gps.ProjectName,
	}
}
