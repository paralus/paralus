package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/RafayLabs/rcloud-base/internal/dao"
	"github.com/RafayLabs/rcloud-base/internal/models"
	"github.com/RafayLabs/rcloud-base/pkg/audit"
	systemv3 "github.com/RafayLabs/rcloud-base/proto/types/systempb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	AuditActionCreate = "create"
	AuditActionDelete = "delete"
	AuditActionUpdate = "update"
)

func CreateUserAuditEvent(ctx context.Context, db bun.IDB, action string, name string, id uuid.UUID, rolesBefore, rolesAfter, groupsBefore, groupsAfter []uuid.UUID) {
	sd, ok := GetSessionDataFromContext(ctx)
	if !ok {
		_log.Warn("unable to create audit event: could not fetch info from context")
		return
	}

	detail := &audit.EventDetail{
		Message: fmt.Sprintf("User %s %sd", name, action),
		Meta: map[string]string{
			"account_id": id.String(),
			"username":   name,
		},
	}
	if err := audit.CreateV1Event(sd, detail, fmt.Sprintf("user.%s.success", action), ""); err != nil {
		_log.Warn("unable to create audit event", err)
	}

	cr, _, dr := diffu(rolesBefore, rolesAfter)
	ncr, err := dao.GetNamesByIds(ctx, db, cr, &models.Role{})
	if err != nil {
		_log.Warn("unable to create audit event", err)
	}
	ndr, err := dao.GetNamesByIds(ctx, db, dr, &models.Role{})
	if err != nil {
		_log.Warn("unable to create audit event", err)
	}
	for _, r := range ncr {
		detail := &audit.EventDetail{
			Message: fmt.Sprintf("Role %s added to user %s", r, name),
			Meta: map[string]string{
				"username":   name,
				"roles_name": r, // TODO: add info like namespace and project
			},
		}
		// user.role.created is user.project.created in rcloud
		if err := audit.CreateV1Event(sd, detail, "user.role.created", ""); err != nil {
			_log.Warn("unable to create audit event", err)
		}
	}

	for _, r := range ndr {
		detail := &audit.EventDetail{
			Message: fmt.Sprintf("Role %s deleted from user %s", r, name),
			Meta: map[string]string{
				"username":  name,
				"role_name": r,
			},
		}
		if err := audit.CreateV1Event(sd, detail, "user.role.deleted", ""); err != nil {
			_log.Warn("unable to create audit event", err)
		}
	}

	cg, _, dg := diffu(groupsBefore, rolesAfter)
	ncg, err := dao.GetNamesByIds(ctx, db, cg, &models.Group{})
	if err != nil {
		_log.Warn("unable to create audit event", err)
	}
	ndg, err := dao.GetNamesByIds(ctx, db, dg, &models.Group{})
	if err != nil {
		_log.Warn("unable to create audit event", err)
	}
	for _, g := range ncg {
		detail := &audit.EventDetail{
			Message: fmt.Sprintf("User %s added to group %s", name, g),
			Meta: map[string]string{
				"username":   name,
				"group_name": g,
			},
		}
		// user.role.created is user.project.created in rcloud
		if err := audit.CreateV1Event(sd, detail, "user.group.created", ""); err != nil {
			_log.Warn("unable to create audit event", err)
		}
	}

	for _, g := range ndg {
		detail := &audit.EventDetail{
			Message: fmt.Sprintf("User %s added to group %s", name, g),
			Meta: map[string]string{
				"username":   name,
				"group_name": g,
			},
		}
		if err := audit.CreateV1Event(sd, detail, "user.group.deleted", ""); err != nil {
			_log.Warn("unable to create audit event", err)
		}
	}
}

func CreateGroupAuditEvent(ctx context.Context, db bun.IDB, action string, name string, id uuid.UUID, usersBefore, usersAfter, rolesBefore, rolesAfter []uuid.UUID) {
	sd, ok := GetSessionDataFromContext(ctx)
	if !ok {
		_log.Warn("unable to create audit event: could not fetch info from context")
		return
	}

	detail := &audit.EventDetail{
		Message: fmt.Sprintf("Group %s %sd", name, action),
		Meta: map[string]string{
			"group_id":   id.String(),
			"group_name": name,
		},
	}
	if err := audit.CreateV1Event(sd, detail, fmt.Sprintf("group.%s.success", action), ""); err != nil {
		_log.Warn("unable to create audit event", err)
	}

	cu, _, du := diffu(usersBefore, usersAfter)

	cun, err := dao.GetUserNamesByIds(ctx, db, cu, &models.KratosIdentities{})
	if err != nil {
		_log.Warn("unable to create audit event", err)
	}
	dun, err := dao.GetUserNamesByIds(ctx, db, du, &models.KratosIdentities{})
	if err != nil {
		_log.Warn("unable to create audit event", err)
	}

	for _, u := range cun {
		detail := &audit.EventDetail{
			Message: fmt.Sprintf("User %s added to group %s", u, name),
			Meta: map[string]string{
				"group_id":   id.String(),
				"group_name": name,
				"username":   u,
			},
		}
		if err := audit.CreateV1Event(sd, detail, "group.user.created", ""); err != nil {
			_log.Warn("unable to create audit event", err)
		}
	}

	for _, u := range dun {
		detail := &audit.EventDetail{
			Message: fmt.Sprintf("User %s deleted from group %s", u, name),
			Meta: map[string]string{
				"group_id":   id.String(),
				"group_name": name,
				"username":   u,
			},
		}
		if err := audit.CreateV1Event(sd, detail, "group.user.deleted", ""); err != nil {
			_log.Warn("unable to create audit event", err)
		}
	}

	cr, _, dr := diffu(rolesBefore, rolesAfter)
	ncr, err := dao.GetNamesByIds(ctx, db, cr, &models.Role{})
	if err != nil {
		_log.Warn("unable to create audit event", err)
	}
	ndr, err := dao.GetNamesByIds(ctx, db, dr, &models.Role{})
	if err != nil {
		_log.Warn("unable to create audit event", err)
	}
	for _, r := range ncr {
		detail := &audit.EventDetail{
			Message: fmt.Sprintf("Role %s added to group %s", r, name),
			Meta: map[string]string{
				"group_id":   id.String(),
				"group_name": name,
				"roles_name": r, // TODO: add info like namespace and project
			},
		}
		// group.role.created is group.project.created in rcloud
		if err := audit.CreateV1Event(sd, detail, "group.role.created", ""); err != nil {
			_log.Warn("unable to create audit event", err)
		}
	}

	for _, r := range ndr {
		detail := &audit.EventDetail{
			Message: fmt.Sprintf("Role %s deleted from group %s", r, name),
			Meta: map[string]string{
				"group_id":   id.String(),
				"group_name": name,
				"role_name":  r,
			},
		}
		if err := audit.CreateV1Event(sd, detail, "group.role.deleted", ""); err != nil {
			_log.Warn("unable to create audit event", err)
		}
	}

}

func CreateRoleAuditEvent(ctx context.Context, action string, name string, id uuid.UUID, permissions []string) {
	sd, ok := GetSessionDataFromContext(ctx)
	if !ok {
		_log.Warn("unable to create audit event: could not fetch info from context")
		return
	}

	detail := &audit.EventDetail{
		Message: fmt.Sprintf("Role %s %sd", name, action),
		Meta: map[string]string{
			"role_id":     id.String(),
			"role_name":   name,
			"permissions": strings.Join(permissions, ","), // TODO: Should we split it into individual ones?
		},
	}
	if err := audit.CreateV1Event(sd, detail, fmt.Sprintf("role.%s.success", action), ""); err != nil {
		_log.Warn("unable to create audit event", err)
	}
}

func CreateProjectAuditEvent(ctx context.Context, action string, name string, id uuid.UUID) {
	sd, ok := GetSessionDataFromContext(ctx)
	if !ok {
		_log.Warn("unable to create audit event: could not fetch info from context")
		return
	}

	detail := &audit.EventDetail{
		Message: fmt.Sprintf("Project %s %sd", name, action),
		Meta: map[string]string{
			"project_id":   id.String(),
			"project_name": name,
		},
	}
	if err := audit.CreateV1Event(sd, detail, fmt.Sprintf("project.%s.success", action), id.String()); err != nil {
		_log.Warn("unable to create audit event", err)
	}
}

func CreateOrganizationAuditEvent(ctx context.Context, action string, name string, id uuid.UUID, settingsBefore, settingsAfter *systemv3.OrganizationSettings) {
	sd, ok := GetSessionDataFromContext(ctx)
	if !ok {
		_log.Warn("unable to create audit event: could not fetch info from context")
		return
	}

	detail := &audit.EventDetail{
		Message: fmt.Sprintf("Organization %s %sd", name, action),
		Meta: map[string]string{
			"organization_id":   id.String(),
			"organization_name": name,
		},
	}
	if err := audit.CreateV1Event(sd, detail, fmt.Sprintf("organization.%s.success", action), ""); err != nil {
		_log.Warn("unable to create audit event", err)
	}

	if settingsBefore == nil && settingsAfter == nil {
		return
	}

	bavail := settingsBefore != nil && settingsAfter != nil
	if !bavail || settingsBefore.IdleLogoutMin != settingsAfter.IdleLogoutMin {
		detail := &audit.EventDetail{
			Message: fmt.Sprintf("Idel logout settings updated for organization %s", name),
			Meta: map[string]string{
				"organization_id":   id.String(),
				"organization_name": name,
			},
		}

		if settingsAfter != nil {
			detail.Meta = map[string]string{
				"organization_id":   id.String(),
				"organization_name": name,
				"idle_logout_min":   string(settingsAfter.IdleLogoutMin),
			}
		}

		if err := audit.CreateV1Event(sd, detail, "organization.idle.timeout.settings.updated", ""); err != nil {
			_log.Warn("unable to create audit event", err)
		}
	}

	bavail = bavail && settingsBefore.Lockout != nil && settingsAfter.Lockout != nil

	if !bavail ||
		settingsBefore.Lockout.Enabled != settingsAfter.Lockout.Enabled ||
		settingsBefore.Lockout.PeriodMin != settingsAfter.Lockout.PeriodMin ||
		settingsBefore.Lockout.Attempts != settingsAfter.Lockout.Attempts {

		enabled := "false"
		detail := &audit.EventDetail{
			Message: fmt.Sprintf("Lockout settings updated for organization %s", name),
			Meta: map[string]string{
				"organization_id":   id.String(),
				"organization_name": name,
			},
		}

		if settingsAfter != nil && settingsAfter.Lockout != nil {
			if settingsAfter.Lockout.Enabled {
				enabled = "true"
			}
			detail.Meta = map[string]string{
				"organization_id":    id.String(),
				"organization_name":  name,
				"lockout_enabled":    enabled,
				"lockout_period_min": string(settingsAfter.Lockout.PeriodMin),
				"lockout_attempts":   string(settingsAfter.Lockout.Attempts),
			}
		}

		if err := audit.CreateV1Event(sd, detail, "organization.lockout.settings.updated", ""); err != nil {
			_log.Warn("unable to create audit event", err)
		}
	}
}

func CreateIdpAuditEvent(ctx context.Context, action string, name string, id uuid.UUID) {
	sd, ok := GetSessionDataFromContext(ctx)
	if !ok {
		_log.Warn("unable to create audit event: could not fetch info from context")
		return
	}

	detail := &audit.EventDetail{
		Message: fmt.Sprintf("Idp %s %sd", name, action),
		Meta: map[string]string{
			"idp_id":   id.String(),
			"idp_name": name,
		},
	}
	// TODO: it is idp.config.created in rcloud
	if err := audit.CreateV1Event(sd, detail, fmt.Sprintf("idp.%s.success", action), ""); err != nil {
		_log.Warn("unable to create audit event", err)
	}
}

func CreateApiKeyAuditEvent(ctx context.Context, action string, id string) {
	sd, ok := GetSessionDataFromContext(ctx)
	if !ok {
		_log.Warn("unable to create audit event: could not fetch info from context")
		return
	}

	detail := &audit.EventDetail{
		Message: fmt.Sprintf("ApiKey %s %sd", id, action),
		Meta: map[string]string{
			"apikey": id,
		},
	}
	if err := audit.CreateV1Event(sd, detail, fmt.Sprintf("apikey.%s.success", action), ""); err != nil {
		_log.Warn("unable to create audit event", err)
	}
}

func CreateClusterAuditEvent(ctx context.Context, action string, name string, id uuid.UUID) {
	sd, ok := GetSessionDataFromContext(ctx)
	if !ok {
		_log.Warn("unable to create audit event: could not fetch info from context")
		return
	}

	detail := &audit.EventDetail{
		Message: fmt.Sprintf("Cluster %s %sd", name, action),
		Meta: map[string]string{
			"cluster_id":   id.String(),
			"cluster_name": name,
		},
	}
	if err := audit.CreateV1Event(sd, detail, fmt.Sprintf("cluster.%s.success", action), ""); err != nil {
		_log.Warn("unable to create audit event", err)
	}
}

// TODO: figure out how this is to be added
func CreateLocationAuditEvent(ctx context.Context, action string, name string, id uuid.UUID) {
	sd, ok := GetSessionDataFromContext(ctx)
	if !ok {
		_log.Warn("unable to create audit event: could not fetch info from context")
		return
	}

	detail := &audit.EventDetail{
		Message: fmt.Sprintf("Location %s %sd", name, action),
		Meta: map[string]string{
			"location_id":   id.String(),
			"location_name": name,
		},
	}
	if err := audit.CreateV1Event(sd, detail, fmt.Sprintf("location.%s.success", action), ""); err != nil {
		_log.Warn("unable to create audit event", err)
	}
}
