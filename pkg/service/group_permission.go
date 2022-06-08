package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/uptrace/bun"
)

// GroupPermissionService is the interface for group permission operations
type GroupPermissionService interface {
	GetGroupPermissions(ctx context.Context, groupNames []string, orgID, partnerID string) ([]sentry.GroupPermission, error)
	GetGroupProjectsByPermission(ctx context.Context, groupNames []string, orgID, partnerID string, permission string) ([]sentry.GroupPermission, error)
	GetGroupPermissionsByProjectIDPermissions(ctx context.Context, groupNames []string, orgID, partnerID string, projects []string, permissions []string) ([]sentry.GroupPermission, error)
	GetProjectByGroup(ctx context.Context, groupNames []string, orgID, partnerID string) ([]sentry.GroupPermission, error)
}

// groupPermissionService implements GroupPermissionService
type groupPermissionService struct {
	db *bun.DB
}

// NewKubeconfigRevocation return new kubeconfig revocation service
func NewGroupPermissionService(db *bun.DB) GroupPermissionService {
	return &groupPermissionService{db}
}

func (s *groupPermissionService) GetGroupPermissions(ctx context.Context, groupNames []string, orgID, partnerID string) ([]sentry.GroupPermission, error) {
	gps, err := dao.GetGroupPermissions(ctx, s.db, groupNames, uuid.MustParse(orgID), uuid.MustParse(partnerID))
	if err != nil {
		return nil, err
	}
	groupPermissions := []sentry.GroupPermission{}
	for _, gp := range gps {
		groupPermissions = append(groupPermissions, prepareGroupPermissionResponse(gp))
	}

	return groupPermissions, nil
}

func (s *groupPermissionService) GetGroupProjectsByPermission(ctx context.Context, groupNames []string, orgID, partnerID string, permission string) ([]sentry.GroupPermission, error) {
	aps, err := dao.GetGroupProjectsByPermission(ctx, s.db, groupNames, uuid.MustParse(orgID), uuid.MustParse(partnerID), permission)
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
	gps, err := dao.GetGroupPermissionsByProjectIDPermissions(ctx, s.db, groupNames, uuid.MustParse(orgID), uuid.MustParse(partnerID), projects, permissions)
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
	gps, err := dao.GetProjectByGroup(ctx, s.db, groupNames, uuid.MustParse(orgID), uuid.MustParse(partnerID))
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
