package service

import (
	"bytes"
	"encoding/json"
	"testing"

	v1 "github.com/paralus/paralus/proto/rpc/audit"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
)

type md struct {
	Source []string `json:"_source"`
	Aggs   struct {
		GroupByProject struct {
			Aggs struct {
				GroupByType struct {
					Terms struct {
						Field string `json:"field"`
						Size  int    `json:"size"`
					} `json:"terms"`
				} `json:"group_by_type"`
				GroupByUsername struct {
					Terms struct {
						Field string `json:"field"`
						Size  int    `json:"size"`
					} `json:"terms"`
				} `json:"group_by_username"`
			} `json:"aggs"`
			Terms struct {
				Field string `json:"field"`
				Size  int    `json:"size"`
			} `json:"terms"`
		} `json:"group_by_project"`
		GroupByType struct {
			Terms struct {
				Field string `json:"field"`
			} `json:"terms"`
		} `json:"group_by_type"`
		GroupByUsername struct {
			Terms struct {
				Field string `json:"field"`
			} `json:"terms"`
		} `json:"group_by_username"`
	} `json:"aggs"`
	Query struct {
		Bool struct {
			Filter struct {
				Range struct {
					JSONTimestamp struct {
						Gte string `json:"gte"`
						Lt  string `json:"lt"`
					} `json:"json.timestamp"`
				} `json:"range"`
			} `json:"filter"`
			Must []struct {
				Term struct {
					JSONCategory string `json:"json.category"`
				} `json:"term,omitempty"`
				Terms struct {
					JSONProject []string `json:"json.project"`
				} `json:"terms,omitempty"`
				QueryString struct {
					Query string `json:"query"`
				} `json:"query_string,omitempty"`
			} `json:"must"`
		} `json:"bool"`
	} `json:"query"`
	Size int `json:"size"`
	Sort struct {
		JSONTimestamp struct {
			Order string `json:"order"`
		} `json:"json.timestamp"`
	} `json:"sort"`
}

type mockElasticSearchQuery struct {
	msg []bytes.Buffer
}

func (m *mockElasticSearchQuery) Handle(msg bytes.Buffer) (map[string]interface{}, error) {
	m.msg = append(m.msg, msg)
	return map[string]interface{}{}, nil
}

func TestGetAuditLogByProjectsSimple(t *testing.T) {
	esq := &mockElasticSearchQuery{}
	al := &AuditLogService{auditQuery: esq}
	req := v1.AuditLogSearchRequest{
		Filter: &v1.AuditLogQueryFilter{
			QueryString:   "query-string",
			Projects:      []string{"project-one", "project-two"},
			Timefrom:      "now-1h",
			Type:          "fake-type",
			User:          "fake-user",
			Client:        "fake-client",
			DashboardData: true,
		},
	}
	_, err := al.GetAuditLogByProjects(&req)
	if err != nil {
		t.Error("unable to get audit logs")
	}
	if len(esq.msg) != 1 {
		t.Fatalf("incorrect number of searches; expected '%v', got '%v'", 1, len(esq.msg))
	}
	m := &md{}
	err = json.Unmarshal(esq.msg[0].Bytes(), m)
	if err != nil {
		t.Fatal("unable to unmarshall es request")
	}

	if m.Source[0] != "json" {
		t.Errorf("incorrect source; expected '%v', got '%v'", "json", m.Source[0])
	}

	if m.Aggs.GroupByType.Terms.Field != "json.type" {
		t.Errorf("incorrect group_by_type field; expected '%v', got '%v'", "json.type", m.Aggs.GroupByType.Terms.Field)
	}

	if m.Aggs.GroupByUsername.Terms.Field != "json.actor.account.username" {
		t.Errorf("incorrect group_by_type field; expected '%v', got '%v'", "json.actor.account.username", m.Aggs.GroupByUsername.Terms.Field)
	}

	// FIXME: will this cause issues as json has no ordering?
	if m.Query.Bool.Must[0].Term.JSONCategory != "AUDIT" {
		t.Errorf("incorrect category; expected '%v', got '%v'", "AUDIT", m.Query.Bool.Must[0].Term.JSONCategory)
	}

	// TODO: add checks for user, client and type
	if m.Query.Bool.Must[5].QueryString.Query != "query-string" {
		t.Errorf("incorrect category; expected '%v', got '%v'", "query-string", m.Query.Bool.Must[5].QueryString.Query)
	}

	if len(m.Query.Bool.Must[4].Terms.JSONProject) != 2 {
		t.Errorf("incorrect number of project filters; expected '%v', got '%v'", 2, len(m.Query.Bool.Must[4].Terms.JSONProject))
	}

	if m.Query.Bool.Must[4].Terms.JSONProject[0] != "project-one" {
		t.Errorf("incorrect number of project filters; expected '%v', got '%v'", "project-one", m.Query.Bool.Must[4].Terms.JSONProject[0])
	}

	if m.Query.Bool.Filter.Range.JSONTimestamp.Gte != "now-1h" {
		t.Errorf("incorrect time range; expected '%v', got '%v'", "now-1h", m.Query.Bool.Filter.Range.JSONTimestamp.Gte)
	}
	if m.Query.Bool.Filter.Range.JSONTimestamp.Lt != "now" {
		t.Errorf("incorrect time range; expected '%v', got '%v'", "now", m.Query.Bool.Filter.Range.JSONTimestamp.Lt)
	}
}

func TestGetAuditLogByProjectsNoProject(t *testing.T) {
	esq := &mockElasticSearchQuery{}
	al := &AuditLogService{auditQuery: esq}
	req := v1.AuditLogSearchRequest{
		Metadata: &v3.Metadata{UrlScope: "url/project"},
		Filter: &v1.AuditLogQueryFilter{
			QueryString: "query-string",
		},
	}
	_, err := al.GetAuditLog(&req)
	if err != nil {
		t.Error("unable to get audit logs", err)
	}
	if len(esq.msg) != 1 {
		t.Fatalf("incorrect number of searches; expected '%v', got '%v'", 1, len(esq.msg))
	}
	m := &md{}
	err = json.Unmarshal(esq.msg[0].Bytes(), m)
	if err != nil {
		t.Fatal("unable to unmarshall es request")
	}

	if m.Source[0] != "json" {
		t.Errorf("incorrect source; expected '%v', got '%v'", "json", m.Source[0])
	}

	if m.Aggs.GroupByType.Terms.Field != "json.type" {
		t.Errorf("incorrect group_by_type field; expected '%v', got '%v'", "json.type", m.Aggs.GroupByType.Terms.Field)
	}

	if m.Aggs.GroupByUsername.Terms.Field != "json.actor.account.username" {
		t.Errorf("incorrect group_by_type field; expected '%v', got '%v'", "json.actor.account.username", m.Aggs.GroupByUsername.Terms.Field)
	}

	// FIXME: will this cause issues as json has no ordering?
	if m.Query.Bool.Must[0].Term.JSONCategory != "AUDIT" {
		t.Errorf("incorrect category; expected '%v', got '%v'", "AUDIT", m.Query.Bool.Must[0].Term.JSONCategory)
	}

	if m.Query.Bool.Must[2].QueryString.Query != "query-string" {
		t.Errorf("incorrect category; expected '%v', got '%v'", "query-string", m.Query.Bool.Must[2].QueryString.Query)
	}

	if len(m.Query.Bool.Must[1].Terms.JSONProject) != 1 {
		t.Errorf("incorrect number of project filters; expected '%v', got '%v'", 1, len(m.Query.Bool.Must[1].Terms.JSONProject))
	}

	if m.Query.Bool.Must[1].Terms.JSONProject[0] != "project" {
		t.Errorf("incorrect number of project filters; expected '%v', got '%v'", "project", m.Query.Bool.Must[1].Terms.JSONProject[0])
	}
}
