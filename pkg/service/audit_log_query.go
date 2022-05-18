package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	v1 "github.com/RafayLabs/rcloud-base/proto/rpc/audit"
	"google.golang.org/protobuf/types/known/structpb"
)

type AuditLogService struct {
	auditQuery *ElasticSearchQuery
}

func NewAuditLogService(url string, auditPattern string, logPrefix string) (*AuditLogService, error) {
	auditQuery, err := NewElasticSearchQuery(url, auditPattern, logPrefix)
	if err != nil {
		return nil, err
	}
	return &AuditLogService{auditQuery: auditQuery}, nil
}

func (a *AuditLogService) GetAuditLog(req *v1.AuditLogSearchRequest) (res *v1.AuditLogSearchResponse, err error) {
	if err != nil {
		return nil, err
	}
	project, err := getPrjectFromUrlScope(req.GetMetadata().UrlScope)
	if err != nil {
		return nil, err
	}
	req.Filter.Projects = []string{project}
	return a.GetAuditLogByProjects(req)
}

func validateQueryString(queryString string) error {
	if strings.Contains(queryString, "*") {
		return fmt.Errorf("'*' is not supported in search query")
	}
	if len(queryString) > 0 && len(queryString) < 3 {
		return fmt.Errorf("search string has to be atleast 3 characters")
	}
	return nil
}

func getPrjectFromUrlScope(urlScope string) (string, error) {
	s := strings.Split(urlScope, "/")
	if len(s) != 2 {
		_log.Errorw("Unable to retrieve project from urlScope", "urlScope", urlScope)
		return "", fmt.Errorf("unable to retrieve project from urlScope")
	}
	return s[1], nil
}

func (a *AuditLogService) GetAuditLogByProjects(req *v1.AuditLogSearchRequest) (res *v1.AuditLogSearchResponse, err error) {
	err = validateQueryString(req.GetFilter().QueryString)
	if err != nil {
		return nil, err
	}
	res = &v1.AuditLogSearchResponse{
		Result: &structpb.Struct{},
	}
	var buf bytes.Buffer
	var r map[string]interface{}
	query := map[string]interface{}{
		"_source": []string{"json"},
		"size":    500,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"json.category": "AUDIT",
						},
					},
				},
			},
		},
		"sort": map[string]interface{}{
			"json.timestamp": map[string]interface{}{
				"order": "desc",
			},
		},
		"aggs": map[string]interface{}{
			"group_by_username": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "json.actor.account.username",
				},
			},
			"group_by_type": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "json.type",
				},
			},
		},
	}

	//aggregations
	agg, _ := query["aggs"].(map[string]interface{})
	if req.GetFilter().DashboardData {
		agg["group_by_project"] = map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "json.project",
				"size":  1000,
			},
			"aggs": map[string]interface{}{
				"group_by_username": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "json.actor.account.username",
						"size":  1000,
					},
				},
				"group_by_type": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "json.type",
						"size":  1000,
					},
				},
			},
		}
	}

	// Filters
	q, _ := query["query"].(map[string]interface{})
	b, _ := q["bool"].(map[string]interface{})
	m, _ := b["must"].([]map[string]interface{})

	//Results not required in case of dashboard - only aggregations required
	if req.GetFilter().DashboardData {
		query["size"] = 0
	}
	//	Add timefrom filter
	if req.GetFilter().Timefrom != "" {
		b["filter"] = map[string]interface{}{
			"range": map[string]interface{}{
				"json.timestamp": map[string]interface{}{
					"gte": req.GetFilter().Timefrom,
					"lt":  "now",
				},
			},
		}
	}
	// Add type
	if req.GetFilter().Type != "" {
		t := map[string]interface{}{
			"term": map[string]interface{}{
				"json.type": req.GetFilter().Type,
			},
		}
		m = append(m, t)
	}
	// Add user
	if req.GetFilter().User != "" {
		t := map[string]interface{}{
			"term": map[string]interface{}{
				"json.actor.account.username": req.GetFilter().User,
			},
		}
		m = append(m, t)
	}
	// Add user
	if req.GetFilter().Client != "" {
		t := map[string]interface{}{
			"term": map[string]interface{}{
				"json.client.type": req.GetFilter().Client,
			},
		}
		m = append(m, t)
	}
	// Project
	if len(req.GetFilter().Projects) > 0 {
		t := map[string]interface{}{
			"terms": map[string]interface{}{
				"json.project": req.GetFilter().Projects,
			},
		}
		m = append(m, t)
	}
	// query string
	if req.GetFilter().QueryString != "" {
		q := map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": req.GetFilter().QueryString,
			},
		}
		m = append(m, q)
	}
	b["must"] = m
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		_log.Errorw("Error encoding query:", " err", err)
		return res, err
	}
	_log.Debug("Executing Query: ", q)
	r, err = a.auditQuery.Handle(buf)
	if err != nil {
		return res, err
	}
	if r == nil {
		return res, nil
	}
	// raw, err := json.Marshal(r)
	// if err != nil {
	// 	return res, err
	// }
	raw, err := structpb.NewStruct(r)
	if err != nil {
		return res, err
	}
	res = &v1.AuditLogSearchResponse{Result: raw}
	return res, nil
}
