package server

import (
	"context"
	"encoding/json"

	ec "github.com/RafayLabs/rcloud-base/pkg/common"
	q "github.com/RafayLabs/rcloud-base/pkg/service"
	v1 "github.com/RafayLabs/rcloud-base/proto/rpc/audit"
)

type relayAuditServer struct {
	rs *q.RelayAuditService
	al *q.AuditLogService
}

var _ v1.RelayAuditServer = (*relayAuditServer)(nil)

// NewAuditServer returns new placement server implementation
func NewRelayAuditServer(relayAuditService *q.RelayAuditService, relayCommandAuditService *q.AuditLogService) (v1.RelayAuditServer, error) {
	return &relayAuditServer{
		rs: relayAuditService,
		al: relayCommandAuditService,
	}, nil
}

func (r *relayAuditServer) GetRelayAPIAudit(ctx context.Context, req *v1.RelayAuditSearchRequest) (res *v1.RelayAuditSearchResponse, err error) {
	return r.rs.GetRelayAudit(req)
}

func (r *relayAuditServer) GetRelayAPIAuditByProjects(ctx context.Context, req *v1.RelayAuditSearchRequest) (res *v1.RelayAuditSearchResponse, err error) {
	return r.rs.GetRelayAuditByProjects(req)
}

func (r *relayAuditServer) GetRelayAudit(ctx context.Context, req *v1.RelayAuditSearchRequest) (res *v1.RelayAuditSearchResponse, err error) {
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
		res = &v1.RelayAuditSearchResponse{
			AuditType: ec.RelayCommandsAuditType,
			Result:    auditRes.Result,
		}
	}
	return
}

func (r *relayAuditServer) GetRelayAuditByProjects(ctx context.Context, req *v1.RelayAuditSearchRequest) (res *v1.RelayAuditSearchResponse, err error) {
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
		res = &v1.RelayAuditSearchResponse{
			AuditType: ec.RelayCommandsAuditType,
			Result:    auditRes.Result,
		}
	}

	return
}

func convertRelayToAuditSearchRequest(req *v1.RelayAuditSearchRequest) (*v1.AuditLogSearchRequest, error) {
	reqByte, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var res v1.AuditLogSearchRequest
	err = json.Unmarshal(reqByte, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
