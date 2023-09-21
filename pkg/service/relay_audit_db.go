package service

import (
	"context"
	"encoding/json"

	"github.com/paralus/paralus/internal/dao"
	v1 "github.com/paralus/paralus/proto/rpc/audit"
	auditv1 "github.com/paralus/paralus/proto/types/audit"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/structpb"
)

type relayAuditDatabaseService struct {
	db  *bun.DB
	tag string
	aps AccountPermissionService
}

func (ra *relayAuditDatabaseService) GetRelayAudit(ctx context.Context, req *v1.RelayAuditRequest) (res *v1.RelayAuditResponse, err error) {
	if err != nil {
		return nil, err
	}
	project, err := getProjectFromUrlScope(req.GetMetadata().UrlScope)
	if err != nil {
		return nil, err
	}
	req.Filter.Projects = []string{project}
	return ra.GetRelayAuditByProjects(ctx, req)
}

func (ra *relayAuditDatabaseService) GetRelayAuditByProjects(ctx context.Context, req *v1.RelayAuditRequest) (res *v1.RelayAuditResponse, err error) {
	err = validateQueryString(req.GetFilter().QueryString)
	if err != nil {
		return &v1.RelayAuditResponse{}, err
	}

	//validate user authz with incoming request
	if len(req.GetFilter().GetProjects()) > 0 {
		if err := ValidateUserAuditReadRequest(ctx, req.GetFilter().GetProjects(), ra.db, true); err != nil {
			return nil, err
		}
	}

	auditLogs, err := dao.GetAuditLogs(ctx, ra.db, ra.tag, req.Filter)
	if err != nil {
		return nil, err
	}

	// aggregations
	projectAggr, err := dao.GetAuditLogAggregations(ctx, ra.db, ra.tag, "project", req.Filter)
	if err != nil {
		return nil, err
	}

	clusterAggr, err := dao.GetAuditLogAggregations(ctx, ra.db, ra.tag, "cluster", req.Filter)
	if err != nil {
		return nil, err
	}

	usernameAggr, err := dao.GetAuditLogAggregations(ctx, ra.db, ra.tag, "username", req.Filter)
	if err != nil {
		return nil, err
	}

	nsAggr, err := dao.GetAuditLogAggregations(ctx, ra.db, ra.tag, "namespace", req.Filter)
	if err != nil {
		return nil, err
	}

	kindAggr, err := dao.GetAuditLogAggregations(ctx, ra.db, ra.tag, "kind", req.Filter)
	if err != nil {
		return nil, err
	}

	methodAggr, err := dao.GetAuditLogAggregations(ctx, ra.db, ra.tag, "method", req.Filter)
	if err != nil {
		return nil, err
	}

	response := &auditv1.AuditResponse{
		Aggregations: &auditv1.Aggregations{
			GroupByProject: &auditv1.AggregatorGroup{
				Buckets: buildAggregators(projectAggr),
			},
			GroupByCluster: &auditv1.AggregatorGroup{
				Buckets: buildAggregators(clusterAggr),
			},
			GroupByUsername: &auditv1.AggregatorGroup{
				Buckets: buildAggregators(usernameAggr),
			},
			GroupByNamespace: &auditv1.AggregatorGroup{
				Buckets: buildAggregators(nsAggr),
			},
			GroupByKind: &auditv1.AggregatorGroup{
				Buckets: buildAggregators(kindAggr),
			},
			GroupByMethod: &auditv1.AggregatorGroup{
				Buckets: buildAggregators(methodAggr),
			},
		},
		Hits: &auditv1.Hits{Hits: buildDataSource(auditLogs)},
	}

	var resMap map[string]interface{}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, &resMap)

	result, _ := structpb.NewStruct(resMap)
	res = &v1.RelayAuditResponse{
		Result: result,
	}
	return res, nil
}
