package server

import (
	"context"
	"encoding/json"

	ec "github.com/paralus/paralus/pkg/common"
	q "github.com/paralus/paralus/pkg/service"
	v1 "github.com/paralus/paralus/proto/rpc/audit"
)

type relayAuditServer struct {
	rs *q.RelayAuditService
	al *q.AuditLogService
}

var _ v1.RelayAuditServiceServer = (*relayAuditServer)(nil)

// NewAuditServer returns new placement server implementation
func NewRelayAuditServer(relayAuditService *q.RelayAuditService, relayCommandAuditService *q.AuditLogService) (v1.RelayAuditServiceServer, error) {
	return &relayAuditServer{
		rs: relayAuditService,
		al: relayCommandAuditService,
	}, nil
}

func (r *relayAuditServer) GetRelayAPIAudit(ctx context.Context, req *v1.RelayAuditRequest) (res *v1.RelayAuditResponse, err error) {
	return r.rs.GetRelayAudit(req)
}

func (r *relayAuditServer) GetRelayAPIAuditByProjects(ctx context.Context, req *v1.RelayAuditRequest) (res *v1.RelayAuditResponse, err error) {
	return r.rs.GetRelayAuditByProjects(req)
}

func (r *relayAuditServer) GetRelayAudit(ctx context.Context, req *v1.RelayAuditRequest) (res *v1.RelayAuditResponse, err error) {
	if req.AuditType == ec.RelayAPIAuditType {
		res, err = r.rs.GetRelayAudit(req)
		if err != nil {
			return nil, err
		}
		res.AuditType = ec.RelayAPIAuditType
	} else {
		auditReq, err := convertRelayToAuditSearchRequest(req)
		if err != nil {
			return nil, err
		}
		auditRes, err := r.al.GetAuditLog(auditReq)
		if err != nil {
			return nil, err
		}
		res = &v1.RelayAuditResponse{
			AuditType: ec.RelayCommandsAuditType,
			Result:    auditRes.Result,
		}
	}
	return
}

func (r *relayAuditServer) GetRelayAuditByProjects(ctx context.Context, req *v1.RelayAuditRequest) (res *v1.RelayAuditResponse, err error) {
	if req.AuditType == ec.RelayAPIAuditType {
		res, err = r.rs.GetRelayAuditByProjects(req)
		if err != nil {
			return nil, err
		}
		res.AuditType = ec.RelayAPIAuditType
	} else {
		auditReq, err := convertRelayToAuditSearchRequest(req)
		if err != nil {
			return nil, err
		}
		auditRes, err := r.al.GetAuditLogByProjects(auditReq)
		if err != nil {
			return nil, err
		}
		res = &v1.RelayAuditResponse{
			AuditType: ec.RelayCommandsAuditType,
			Result:    auditRes.Result,
		}
	}

	return
}

func convertRelayToAuditSearchRequest(req *v1.RelayAuditRequest) (*v1.GetAuditLogSearchRequest, error) {
	reqByte, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var res v1.GetAuditLogSearchRequest
	err = json.Unmarshal(reqByte, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
