package service

import (
	"bytes"
	"encoding/json"

	v1 "github.com/RafayLabs/rcloud-base/proto/rpc/audit"
	"google.golang.org/protobuf/types/known/structpb"
)

type RelayAuditService struct {
	relayQuery *ElasticSearchQuery
}

func NewRelayAuditService(url string, auditPattern string, logPrefix string) (*RelayAuditService, error) {
	relayQuery, err := NewElasticSearchQuery(url, auditPattern, logPrefix)
	if err != nil {
		return nil, err
	}
	return &RelayAuditService{relayQuery: relayQuery}, nil
}

func (ra *RelayAuditService) GetRelayAudit(req *v1.RelayAuditSearchRequest) (res *v1.RelayAuditSearchResponse, err error) {
	if err != nil {
		return nil, err
	}
	projectID, err := getPrjectIdFromUrlScope(req.GetMetadata().UrlScope)
	if err != nil {
		return nil, err
	}
	req.Filter.ProjectIds = []string{projectID}
	return ra.GetRelayAuditByProjects(req)
}

func (ra *RelayAuditService) GetRelayAuditByProjects(req *v1.RelayAuditSearchRequest) (res *v1.RelayAuditSearchResponse, err error) {
	err = validateQueryString(req.GetFilter().QueryString)
	oid := req.GetMetadata().GetOrganization()
	pid := req.GetMetadata().GetPartner()
	if err != nil {
		return &v1.RelayAuditSearchResponse{}, err
	}
	var buf bytes.Buffer
	var r map[string]interface{}
	res = &v1.RelayAuditSearchResponse{}
	//Handle defaults value
	query := map[string]interface{}{
		"_source": true,
		"size":    500,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"o": oid,
						},
					},
					{
						"term": map[string]interface{}{
							"p": pid,
						},
					},
				},
			},
		},
		"sort": map[string]interface{}{
			"ts": map[string]interface{}{
				"order": "desc",
			},
		},
		"aggs": map[string]interface{}{
			"group_by_username": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "un",
				},
			},
			"group_by_cluster": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "cn",
				},
			},
			"group_by_namespace": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "ns",
				},
			},
			"group_by_kind": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "k",
				},
			},
			"group_by_method": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "m",
				},
			},
		},
	}

	//aggregations
	agg, _ := query["aggs"].(map[string]interface{})
	if req.GetFilter().DashboardData {
		agg["group_by_cluster"] = map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "cn",
				"size":  1000,
			},
			"aggs": map[string]interface{}{
				"group_by_username": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "un",
						"size":  1000,
					},
				},
				"group_by_namespace": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "ns",
						"size":  1000,
					},
				},
				//----------Trend not added in current Dashboard - might be useful later------
				//"group_by_time": map[string]interface{}{
				//	"date_histogram": map[string]interface{}{
				//		"field":    "ts",
				//		"interval": "1h",
				//	},
				//},
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
	if req.Filter.Timefrom != "" {
		b["filter"] = map[string]interface{}{
			"range": map[string]interface{}{
				"ts": map[string]interface{}{
					"gte": req.Filter.Timefrom,
					"lt":  "now",
				},
			},
		}
	}
	// User
	if req.Filter.User != "" {
		t := map[string]interface{}{
			"term": map[string]interface{}{
				"un": req.Filter.User,
			},
		}
		m = append(m, t)
	}
	// Cluster
	if req.Filter.Cluster != "" {
		t := map[string]interface{}{
			"term": map[string]interface{}{
				"cn": req.Filter.Cluster,
			},
		}
		m = append(m, t)
	}
	// Namespace
	if req.Filter.Namespace != "" {
		t := map[string]interface{}{
			"term": map[string]interface{}{
				"ns": req.Filter.Namespace,
			},
		}
		m = append(m, t)
	}
	// Kind
	if req.Filter.Kind != "" {
		t := map[string]interface{}{
			"term": map[string]interface{}{
				"k": req.Filter.Kind,
			},
		}
		m = append(m, t)
	}
	// Method
	if req.Filter.Method != "" {
		t := map[string]interface{}{
			"term": map[string]interface{}{
				"m": req.Filter.Method,
			},
		}
		m = append(m, t)
	}
	// ProjectIds - [project_id not present ES for relay audit logs currently
	// so in case of dashboard filtering with cluster name(belonging to the project)]
	if len(req.Filter.ProjectIds) > 0 {
		if req.GetFilter().DashboardData &&
			req.Filter.ClusterNames != nil && len(req.Filter.ClusterNames) > 0 {
			t := map[string]interface{}{
				"terms": map[string]interface{}{
					"cn": req.Filter.ClusterNames,
				},
			}
			m = append(m, t)
		} else {

			t := map[string]interface{}{
				"terms": map[string]interface{}{
					"project_id": req.Filter.ProjectIds,
				},
			}
			m = append(m, t)
		}
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

	r, err = ra.relayQuery.Handle(buf)
	if err != nil {
		return res, err
	}
	if r == nil {
		return res, nil
	}
	raw, err := structpb.NewStruct(r)
	if err != nil {
		return res, err
	}
	res = &v1.RelayAuditSearchResponse{Result: raw}

	return res, nil
}
