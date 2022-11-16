package service

import (
	"context"
	"encoding/json"

	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	v1 "github.com/paralus/paralus/proto/rpc/audit"
	auditv1 "github.com/paralus/paralus/proto/types/audit"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/structpb"
)

type auditLogDatabaseService struct {
	db  *bun.DB
	tag string
}

func (a *auditLogDatabaseService) GetAuditLog(req *v1.GetAuditLogSearchRequest) (res *v1.GetAuditLogSearchResponse, err error) {
	project, err := getProjectFromUrlScope(req.GetMetadata().UrlScope)
	if err != nil {
		return nil, err
	}
	req.Filter.Projects = []string{project}
	return a.GetAuditLogByProjects(req)
}

func buildAggregators(aggr []models.AggregatorData) []*auditv1.GroupByType {
	var groups = make([]*auditv1.GroupByType, 0)
	for _, agg := range aggr {
		groups = append(groups, &auditv1.GroupByType{
			DocCount: int32(agg.Count),
			Key:      agg.Key,
		})
	}
	return groups
}

func buildDataSource(logs []models.AuditLog) (ds []*auditv1.DataSource) {
	for _, log := range logs {
		data := &auditv1.Data{}
		json.Unmarshal(log.Data, data)
		ds = append(ds, &auditv1.DataSource{
			XSource: &auditv1.DataSourceJSON{
				Json: data,
			},
		},
		)
	}
	return ds
}

func (a *auditLogDatabaseService) GetAuditLogByProjects(req *v1.GetAuditLogSearchRequest) (res *v1.GetAuditLogSearchResponse, err error) {
	err = validateQueryString(req.GetFilter().QueryString)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	auditLogs, err := dao.GetAuditLogs(ctx, a.db, a.tag, req.Filter)
	if err != nil {
		return nil, err
	}

	// aggregations
	_, err = dao.GetAuditLogAggregations(ctx, a.db, a.tag, "project", req.Filter)
	if err != nil {
		return nil, err
	}

	usernameAggr, err := dao.GetAuditLogAggregations(ctx, a.db, a.tag, "username", req.Filter)
	if err != nil {
		return nil, err
	}

	typeAggr, err := dao.GetAuditLogAggregations(ctx, a.db, a.tag, "type", req.Filter)
	if err != nil {
		return nil, err
	}

	response := &auditv1.AuditResponse{
		Aggregations: &auditv1.Aggregations{
			GroupByType: &auditv1.AggregatorGroup{
				Buckets: buildAggregators(typeAggr),
			},
			GroupByUsername: &auditv1.AggregatorGroup{
				Buckets: buildAggregators(usernameAggr),
			},
		},
		Hits: &auditv1.Hits{Hits: buildDataSource(auditLogs)},
	}

	var resMap map[string]interface{}
	data, _ := json.Marshal(response)
	json.Unmarshal(data, &resMap)

	result, _ := structpb.NewStruct(resMap)
	res = &v1.GetAuditLogSearchResponse{
		Result: result,
	}
	return res, nil
}
