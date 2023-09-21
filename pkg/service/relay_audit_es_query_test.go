package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/common"
	v1 "github.com/paralus/paralus/proto/rpc/audit"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
)

type rmd struct {
	Source []string `json:"_source"`
	Aggs   struct {
		GroupByCluster struct {
			Aggs struct {
				GroupByNamespace struct {
					Terms struct {
						Field string `json:"field"`
						Size  int    `json:"size"`
					} `json:"terms"`
				} `json:"group_by_namespace"`
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
		} `json:"group_by_cluster"`
		GroupByKind struct {
			Terms struct {
				Field string `json:"field"`
			} `json:"terms"`
		} `json:"group_by_kind"`
		GroupByMethod struct {
			Terms struct {
				Field string `json:"field"`
			} `json:"terms"`
		} `json:"group_by_method"`
		GroupByNamespace struct {
			Terms struct {
				Field string `json:"field"`
			} `json:"terms"`
		} `json:"group_by_namespace"`
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
					JSONTs struct {
						Gte string `json:"gte"`
						Lt  string `json:"lt"`
					} `json:"json.ts"`
				} `json:"range"`
			} `json:"filter"`
			Must []struct {
				Term struct {
					JSONUn string `json:"json.un"`
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
		JSONTs struct {
			Order string `json:"order"`
		} `json:"json.ts"`
	} `json:"sort"`
}

func TestGetRelayAuditLogByProjectsSimple(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	esq := &mockElasticSearchQuery{}
	al := &relayAuditElasticSearchService{relayQuery: esq, db: db}
	req := v1.RelayAuditRequest{
		Filter: &v1.RelayAuditQueryFilter{
			QueryString:   "query-string",
			Projects:      []string{"project-one", "project-two"},
			Timefrom:      "now-1h",
			Type:          "test-type",
			User:          "test-user",
			Client:        "test-client",
			Cluster:       "test-cluster",
			Namespace:     "test-namespace",
			Kind:          "test-kind",
			Method:        "test-method",
			DashboardData: true,
		},
	}
	uuid := uuid.New().String()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "sap"."account_id", "sap"."project_id", "sap"."group_id", "sap"."role_id", "sap"."role_name", "sap"."organization_id", "sap"."partner_id", "sap"."is_global", "sap"."scope", "sap"."permission_name", "sap"."base_url", "sap"."urls" FROM "sentry_account_permission" AS "sap" WHERE (account_id = '` + uuid + `') AND (partner_id = '` + uuid + `') AND (lower(role_name) = 'admin') AND (lower(scope) = 'organization')`)).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "role_name", "scope"}).AddRow(uuid, "admin", "organization"))

	sd := v3.SessionData{
		Account:      uuid,
		Organization: uuid,
		Partner:      uuid,
		Username:     "user",
	}
	ctx := context.WithValue(context.Background(), common.SessionDataKey, &sd)
	_, err := al.GetRelayAuditByProjects(ctx, &req)
	if err != nil {
		t.Error("unable to get audit logs")
	}
	if len(esq.msg) != 1 {
		t.Fatalf("incorrect number of searches; expected '%v', got '%v'", 1, len(esq.msg))
	}
	m := &rmd{}
	err = json.Unmarshal(esq.msg[0].Bytes(), m)
	if err != nil {
		t.Fatal("unable to unmarshall es request")
	}
	expected := `{"_source":["json"],"aggs":{"group_by_cluster":{"aggs":{"group_by_namespace":{"terms":{"field":"json.ns","size":1000}},"group_by_username":{"terms":{"field":"json.un","size":1000}}},"terms":{"field":"json.cn","size":1000}},"group_by_kind":{"terms":{"field":"json.k"}},"group_by_method":{"terms":{"field":"json.m"}},"group_by_namespace":{"terms":{"field":"json.ns"}},"group_by_username":{"terms":{"field":"json.un"}}},"query":{"bool":{"filter":{"range":{"json.ts":{"gte":"now-1h","lt":"now"}}},"must":[{"term":{"json.un":"test-user"}},{"term":{"json.cn":"test-cluster"}},{"term":{"json.ns":"test-namespace"}},{"term":{"json.k":"test-kind"}},{"term":{"json.m":"test-method"}},{"terms":{"json.project":["project-one","project-two"]}},{"query_string":{"query":"query-string"}}]}},"size":0,"sort":{"json.ts":{"order":"desc"}}}`
	if strings.TrimSpace(esq.msg[0].String()) != expected {
		t.Errorf("incorrect es query; expected '%v', got '%v'", expected, strings.TrimSpace(esq.msg[0].String()))
	}
}

func TestGetRelayAuditLogByProjectsNoProject(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	esq := &mockElasticSearchQuery{}
	al := &relayAuditElasticSearchService{relayQuery: esq, db: db}
	req := v1.RelayAuditRequest{
		Metadata: &v3.Metadata{UrlScope: "url/project"},
		Filter: &v1.RelayAuditQueryFilter{
			QueryString: "query-string",
		},
	}
	uuid := uuid.New().String()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "sap"."account_id", "sap"."project_id", "sap"."group_id", "sap"."role_id", "sap"."role_name", "sap"."organization_id", "sap"."partner_id", "sap"."is_global", "sap"."scope", "sap"."permission_name", "sap"."base_url", "sap"."urls" FROM "sentry_account_permission" AS "sap" WHERE (account_id = '` + uuid + `') AND (partner_id = '` + uuid + `') AND (lower(role_name) = 'admin') AND (lower(scope) = 'organization')`)).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "role_name", "scope"}).AddRow(uuid, "admin", "organization"))

	sd := v3.SessionData{
		Account:      uuid,
		Organization: uuid,
		Partner:      uuid,
		Username:     "user",
	}
	ctx := context.WithValue(context.Background(), common.SessionDataKey, &sd)
	_, err := al.GetRelayAudit(ctx, &req)
	if err != nil {
		t.Error("unable to get audit logs", err)
	}
	if len(esq.msg) != 1 {
		t.Fatalf("incorrect number of searches; expected '%v', got '%v'", 1, len(esq.msg))
	}
	m := &rmd{}
	err = json.Unmarshal(esq.msg[0].Bytes(), m)
	if err != nil {
		t.Fatal("unable to unmarshall es request")
	}

	expected := `{"_source":["json"],"aggs":{"group_by_cluster":{"terms":{"field":"json.cn"}},"group_by_kind":{"terms":{"field":"json.k"}},"group_by_method":{"terms":{"field":"json.m"}},"group_by_namespace":{"terms":{"field":"json.ns"}},"group_by_username":{"terms":{"field":"json.un"}}},"query":{"bool":{"must":[{"terms":{"json.project":["project"]}},{"query_string":{"query":"query-string"}}]}},"size":500,"sort":{"json.ts":{"order":"desc"}}}`
	if strings.TrimSpace(esq.msg[0].String()) != expected {
		t.Errorf("incorrect es query; expected '%v', got '%v'", expected, strings.TrimSpace(esq.msg[0].String()))
	}
}
