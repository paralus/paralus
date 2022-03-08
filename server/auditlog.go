package server

import (
	"context"

	q "github.com/RafaySystems/rcloud-base/pkg/service"
	v1 "github.com/RafaySystems/rcloud-base/proto/rpc/audit"
)

type auditLogServer struct {
	as *q.AuditLogService
}

var _ v1.AuditLogServer = (*auditLogServer)(nil)

// NewAuditServer returns new placement server implementation
func NewAuditLogServer(auditLogService *q.AuditLogService) (v1.AuditLogServer, error) {
	return &auditLogServer{as: auditLogService}, nil
}

func (a *auditLogServer) GetAuditLog(ctx context.Context, req *v1.AuditLogSearchRequest) (res *v1.AuditLogSearchResponse, err error) {
	return a.as.GetAuditLog(req)
}

func (a *auditLogServer) GetAuditLogByProjects(ctx context.Context, req *v1.AuditLogSearchRequest) (res *v1.AuditLogSearchResponse, err error) {
	return a.as.GetAuditLogByProjects(req)
}
