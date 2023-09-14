package service

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/audit"
	"github.com/paralus/paralus/pkg/common"
	eventv1 "github.com/paralus/paralus/proto/rpc/audit"
	auditv1 "github.com/paralus/paralus/proto/types/audit"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
)

func TestRelayAuditLog(t *testing.T) {

	testcases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "auditlog-kubectlapi-kind",
			run:  testGetRelayAuditLogForKind(audit.KUBECTL_API),
		},
		{
			name: "auditlog-kubectlapi-cluster",
			run:  testGetRelayAuditLogForCluster(audit.KUBECTL_API),
		},
	}

	t.Run("relayauditlogs", func(t *testing.T) {
		for _, testcase := range testcases {
			tc := testcase
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				tc.run(t)
			})
		}
	})
}

func testGetRelayAuditLogForCluster(tag string) func(t *testing.T) {
	return func(t *testing.T) {

		db, mock := getDB(t)
		defer db.Close()

		ras, err := NewRelayAuditDatabaseService(db, tag)
		if err != nil {
			t.Fatal(err)
		}

		timefrom := "1h"
		auditrecord := "{\"av\":\"flowcontrol.apiserver.k8s.io/v1beta2\",\"cn\":\"kind-2\",\"d\": 0.492472386,\"id\":\"cduajnic6p9tna60re3g\",\"k\":\"\",\"m\": \"GET\", \"n\": \"\", \"ns\": \"\", \"o\": \"9fcdf482-6191-44f1-987a-8469addf2566\", \"p\": \"ba184458-b899-4cf3-99fa-d77a21578ede\", \"q\": \"timeout=32s\", \"ra\": \"10.0.0.147\", \"sc\": 200, \"st\": \"browser shell\", \"ts\": \"2022-11-22T10:52:14.987Z\", \"un\": \"admin@paralus.local\", \"url\": \"/apis/flowcontrol.apiserver.k8s.io/v1beta2\", \"w\": 819 }\""

		req := &eventv1.RelayAuditRequest{
			Metadata: &commonv3.Metadata{
				UrlScope: "auditlogs/" + project,
			},
			Filter: &eventv1.RelayAuditQueryFilter{
				Cluster:  "kind-2",
				Timefrom: "now-" + timefrom,
			},
		}

		uuid := uuid.New().String()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "sap"."account_id", "sap"."project_id", "sap"."group_id", "sap"."role_id", "sap"."role_name", "sap"."organization_id", "sap"."partner_id", "sap"."is_global", "sap"."scope", "sap"."permission_name", "sap"."base_url", "sap"."urls" FROM "sentry_account_permission" AS "sap" WHERE (account_id = '` + uuid + `') AND (partner_id = '` + uuid + `') AND (lower(role_name) = 'admin') AND (lower(scope) = 'organization')`)).
			WillReturnRows(sqlmock.NewRows([]string{"account_id", "role_name", "scope"}).AddRow(uuid, "admin", "organization"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "auditlog"."tag", "auditlog"."time", "auditlog"."data" FROM "audit_logs" AS "auditlog" WHERE (tag = '` + tag + `') AND (data->>'cn' = '` + req.Filter.Cluster + `') AND (time between now() - interval '` + timefrom + `' and now())`)).
			WillReturnRows(sqlmock.NewRows([]string{"tag", "time", "data"}).AddRow(tag, time.Now(), auditrecord))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'pr' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'cn' = '` + req.Filter.Cluster + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'pr'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "project"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'cn' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'cn' = '` + req.Filter.Cluster + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'cn'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "cluster"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'un' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'cn' = '` + req.Filter.Cluster + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'un'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "username"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'n' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'cn' = '` + req.Filter.Cluster + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'n'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "namespace"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'k' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'cn' = '` + req.Filter.Cluster + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'k'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "kind"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'m' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'cn' = '` + req.Filter.Cluster + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'m'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "method"))

		sd := commonv3.SessionData{
			Account:      uuid,
			Organization: uuid,
			Partner:      uuid,
			Username:     "user",
		}
		ctx := context.WithValue(context.Background(), common.SessionDataKey, &sd)
		res, err := ras.GetRelayAudit(ctx, req)
		if err != nil {
			t.Fatal("could not get audit logs:", err)
		}

		var response auditv1.AuditResponse
		data, err := json.Marshal(res.Result)
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal(data, &response)
		if err != nil {
			t.Fatal(err)
		}

		if len(response.GetHits().Hits) != 1 {
			t.Fail()
		}
	}
}

func testGetRelayAuditLogForKind(tag string) func(t *testing.T) {
	return func(t *testing.T) {

		db, mock := getDB(t)
		defer db.Close()

		as, err := NewRelayAuditDatabaseService(db, tag)
		if err != nil {
			t.Fatal(err)
		}

		timefrom := "1d"
		auditrecord := "{\"av\":\"flowcontrol.apiserver.k8s.io/v1beta2\",\"cn\":\"kind-2\",\"d\": 0.492472386,\"id\":\"cduajnic6p9tna60re3g\",\"k\":\"namespace\",\"m\": \"GET\", \"n\": \"\", \"ns\": \"\", \"o\": \"9fcdf482-6191-44f1-987a-8469addf2566\", \"p\": \"ba184458-b899-4cf3-99fa-d77a21578ede\", \"q\": \"timeout=32s\", \"ra\": \"10.0.0.147\", \"sc\": 200, \"st\": \"browser shell\", \"ts\": \"2022-11-22T10:52:14.987Z\", \"un\": \"admin@paralus.local\", \"url\": \"/apis/flowcontrol.apiserver.k8s.io/v1beta2\", \"w\": 819 }\""
		auditrecordtwo := "{\"av\":\"flowcontrol.apiserver.k8s.io/v1beta2\",\"cn\":\"kind-2\",\"d\": 0.492472386,\"id\":\"cduajnic6p9tna60re3g\",\"k\":\"namespace\",\"m\": \"GET\", \"n\": \"\", \"ns\": \"\", \"o\": \"9fcdf482-6191-44f1-987a-8469addf2566\", \"p\": \"ba184458-b899-4cf3-99fa-d77a21578ede\", \"q\": \"timeout=32s\", \"ra\": \"10.0.0.147\", \"sc\": 200, \"st\": \"browser shell\", \"ts\": \"2022-11-22T10:52:14.987Z\", \"un\": \"admin@paralus.local\", \"url\": \"/apis/flowcontrol.apiserver.k8s.io/v1beta2\", \"w\": 819 }\""

		req := &eventv1.RelayAuditRequest{
			Metadata: &commonv3.Metadata{
				UrlScope: "auditlogs/" + project,
			},
			Filter: &eventv1.RelayAuditQueryFilter{
				Kind:     "namespace",
				Timefrom: "now-" + timefrom,
			},
		}

		uuid := uuid.New().String()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "sap"."account_id", "sap"."project_id", "sap"."group_id", "sap"."role_id", "sap"."role_name", "sap"."organization_id", "sap"."partner_id", "sap"."is_global", "sap"."scope", "sap"."permission_name", "sap"."base_url", "sap"."urls" FROM "sentry_account_permission" AS "sap" WHERE (account_id = '` + uuid + `') AND (partner_id = '` + uuid + `') AND (lower(role_name) = 'admin') AND (lower(scope) = 'organization')`)).
			WillReturnRows(sqlmock.NewRows([]string{"account_id", "role_name", "scope"}).AddRow(uuid, "admin", "organization"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "auditlog"."tag", "auditlog"."time", "auditlog"."data" FROM "audit_logs" AS "auditlog" WHERE (tag = '` + tag + `') AND (data->>'k' = '` + req.Filter.Kind + `') AND (time between now() - interval '` + timefrom + `' and now())`)).
			WillReturnRows(sqlmock.NewRows([]string{"tag", "time", "data"}).AddRow(tag, time.Now(), auditrecord).AddRow(tag, time.Now(), auditrecordtwo))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'pr' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'k' = '` + req.Filter.Kind + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'pr'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "project"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'cn' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'k' = '` + req.Filter.Kind + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'cn'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "cluster"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'un' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'k' = '` + req.Filter.Kind + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'un'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "username"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'n' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'k' = '` + req.Filter.Kind + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'n'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "namespace"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'k' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'k' = '` + req.Filter.Kind + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'k'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "kind"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'m' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'k' = '` + req.Filter.Kind + `') AND (time between now() - interval '` + timefrom + `' and now()) AND (data->>'pr' = '` + project + `') GROUP BY data->>'m'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "method"))

		sd := commonv3.SessionData{
			Account:      uuid,
			Organization: uuid,
			Partner:      uuid,
			Username:     "user",
		}
		ctx := context.WithValue(context.Background(), common.SessionDataKey, &sd)
		res, err := as.GetRelayAudit(ctx, req)
		if err != nil {
			t.Fatal("could not get audit logs:", err)
		}

		var response auditv1.AuditResponse
		data, err := json.Marshal(res.Result)
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal(data, &response)
		if err != nil {
			t.Fatal(err)
		}

		if len(response.GetHits().Hits) != 2 {
			t.Fail()
		}
	}
}
