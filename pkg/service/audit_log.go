package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	v1 "github.com/paralus/paralus/proto/rpc/audit"
	"github.com/uptrace/bun"
)

type AuditLogService interface {
	GetAuditLog(ctx context.Context, req *v1.GetAuditLogSearchRequest) (res *v1.GetAuditLogSearchResponse, err error)
	GetAuditLogByProjects(ctx context.Context, req *v1.GetAuditLogSearchRequest) (res *v1.GetAuditLogSearchResponse, err error)
}

// auditlogs permissions
const (
	OrgRelayAuditPermission     = "org.relayAudit.read"
	ProjectRelayAuditPermission = "project.relayAudit.read"
	ProjectAuditLogPermission   = "project.auditLog.read"
)

func NewAuditLogElasticSearchService(url string, auditPattern string, logPrefix string, db *bun.DB) (AuditLogService, error) {
	auditQuery, err := NewElasticSearchQuery(url, auditPattern, logPrefix)
	if err != nil {
		return nil, err
	}
	return &auditLogElasticSearchService{auditQuery: auditQuery, db: db}, nil
}

func NewAuditLogDatabaseService(db *bun.DB, tag string) (AuditLogService, error) {
	return &auditLogDatabaseService{db: db, tag: tag}, nil
}

type RelayAuditService interface {
	GetRelayAudit(ctx context.Context, req *v1.RelayAuditRequest) (res *v1.RelayAuditResponse, err error)
	GetRelayAuditByProjects(ctx context.Context, req *v1.RelayAuditRequest) (res *v1.RelayAuditResponse, err error)
}

func NewRelayAuditDatabaseService(db *bun.DB, tag string) (RelayAuditService, error) {
	return &relayAuditDatabaseService{db: db, tag: tag}, nil
}

func NewRelayAuditElasticSearchService(url string, auditPattern string, logPrefix string, db *bun.DB) (RelayAuditService, error) {
	relayQuery, err := NewElasticSearchQuery(url, auditPattern, logPrefix)
	if err != nil {
		return nil, err
	}
	return &relayAuditElasticSearchService{relayQuery: relayQuery, db: db}, nil
}

func ValidateUserAuditReadRequest(ctx context.Context, projects []string, db *bun.DB, isRelayAudit bool) error {
	var prerr error
	//validate user authz with incoming request
	sd, ok := GetSessionDataFromContext(ctx)
	if !ok {
		return errors.New("failed to get session data")
	}
	_log.Infow("fetching auditlogs", "account", sd)

	// MustParse panics on an empty string and crash-loops the pod. IdP sessions
	// and the documented lowercase metadata migration can leave partner or
	// organization empty (issue 344). Return an error instead.
	accountID, err := parseSessionUUID("account", sd.Account)
	if err != nil {
		return err
	}
	partnerID, err := parseSessionUUID("partner", sd.Partner)
	if err != nil {
		return err
	}
	organizationID, err := parseSessionUUID("organization", sd.Organization)
	if err != nil {
		return err
	}

	// let's check if user has organization scoped roles associated
	isOrgAdmin, err := dao.IsOrgAdmin(ctx, db, accountID, partnerID)
	if err != nil {
		return err
	}
	if isOrgAdmin {
		return prerr
	}

	isOrgReadOnly, err := dao.IsOrgReadOnly(ctx, db, accountID, organizationID, partnerID)
	if err != nil {
		return err
	}

	if isOrgReadOnly {
		return prerr
	}

	sap, err := dao.GetAccountPermissions(ctx, db, accountID, organizationID, partnerID)
	if err != nil {
		return err
	}
	for _, rproject := range projects {
		rprojectid, err := dao.GetProjectId(ctx, db, rproject)
		if err != nil {
			return err
		}
		available := false
		for _, ap := range sap {
			if rprojectid == ap.ProjectId {
				if isRelayAudit {
					if ap.PermissionName == ProjectRelayAuditPermission {
						available = true
						break
					}
				} else {
					if ap.PermissionName == ProjectAuditLogPermission {
						available = true
						break
					}
				}
			}
		}
		if !available {
			prerr = errors.Join(prerr, fmt.Errorf("not authorized for project %s", rproject))
		}
	}
	return prerr
}

func parseSessionUUID(field, value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid session %s: %w", field, err)
	}
	return id, nil
}
