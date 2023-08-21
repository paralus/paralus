package server

import (
	"context"

	q "github.com/paralus/paralus/pkg/service"
	v1 "github.com/paralus/paralus/proto/rpc/audit"
)

type auditLogServer struct {
	as q.AuditLogService
}

var _ v1.AuditLogServiceServer = (*auditLogServer)(nil)

// NewAuditServer returns new placement server implementation
func NewAuditLogServer(auditLogService q.AuditLogService) (v1.AuditLogServiceServer, error) {
	return &auditLogServer{as: auditLogService}, nil
}

func (a *auditLogServer) GetAuditLog(ctx context.Context, req *v1.GetAuditLogSearchRequest) (res *v1.GetAuditLogSearchResponse, err error) {
	return a.as.GetAuditLog(ctx, req)
}

func (a *auditLogServer) GetAuditLogByProjects(ctx context.Context, req *v1.GetAuditLogSearchRequest) (res *v1.GetAuditLogSearchResponse, err error) {
	return a.as.GetAuditLogByProjects(ctx, req)
}
