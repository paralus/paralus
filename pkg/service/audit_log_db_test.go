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

const project = "projectone"

func TestAuditLog(t *testing.T) {

	testcases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "auditlog-lasthour",
			run:  testGetAuditLogForLastHour(audit.SYSTEM),
		},
		{
			name: "auditlog-lastday",
			run:  testGetAuditLogForDay(audit.SYSTEM),
		},
		{
			name: "kubectlcmd-lastday",
			run:  testGetKubectlCommands(audit.KUBECTL_CMD),
		},
	}

	t.Run("auditlogs", func(t *testing.T) {
		for _, testcase := range testcases {
			tc := testcase
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				tc.run(t)
			})
		}
	})
}

func testGetAuditLogForLastHour(tag string) func(t *testing.T) {
	return func(t *testing.T) {

		db, mock := getDB(t)
		defer db.Close()

		as, err := NewAuditLogDatabaseService(db, tag)
		if err != nil {
			t.Fatal(err)
		}

		timefrom := "1h"
		auditrecord := "{\"actor\":{\"account\":{\"username\":\"admin@paralus.local\"},\"groups\":[\"All Local Users\",\"Organization Admins\"],\"type\":\"USER\"},\"category\":\"AUDIT\",\"client\":{\"host\":\"console-ic-oss.dev.rafay-edge.net\",\"ip\":\"10.0.0.147:47204\",\"type\":\"BROWSER\",\"user_agent\":\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36\"},\"detail\":{\"message\":\"Cluster kind-local deleted\",\"meta\":{\"cluster_name\":\"kind-local\"}},\"origin\":\"core\",\"portal\":\"OPS\",\"project\":\"stage\",\"timestamp\":\"2022-11-21T09:48:30.597615647Z\",\"type\":\"cluster.delete.success\",\"version\":\"1.0\"}"

		req := &eventv1.GetAuditLogSearchRequest{
			Metadata: &commonv3.Metadata{
				UrlScope: "auditlogs/" + project,
			},
			Filter: &eventv1.AuditLogQueryFilter{
				Timefrom: "now-" + timefrom,
				Projects: []string{project},
			},
		}

		uuid := uuid.New().String()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "sap"."account_id", "sap"."project_id", "sap"."group_id", "sap"."role_id", "sap"."role_name", "sap"."organization_id", "sap"."partner_id", "sap"."is_global", "sap"."scope", "sap"."permission_name", "sap"."base_url", "sap"."urls" FROM "sentry_account_permission" AS "sap" WHERE (account_id = '` + uuid + `') AND (partner_id = '` + uuid + `') AND (lower(role_name) = 'admin') AND (lower(scope) = 'organization')`)).
			WillReturnRows(sqlmock.NewRows([]string{"account_id", "role_name", "scope"}).AddRow(uuid, "admin", "organization"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "auditlog"."tag", "auditlog"."time", "auditlog"."data" FROM "audit_logs" AS "auditlog" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now())`)).
			WillReturnRows(sqlmock.NewRows([]string{"tag", "time", "data"}).AddRow("system", time.Now(), auditrecord))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'project' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now()) GROUP BY data->>'project'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "cluster"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->'actor'->'account'->>'username' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now()) GROUP BY data->'actor'->'account'->>'username'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "username"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'type' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now()) GROUP BY data->>'type'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "type"))

		sd := commonv3.SessionData{
			Account:      uuid,
			Organization: uuid,
			Partner:      uuid,
			Username:     "user",
			Project: &commonv3.ProjectData{
				List: []*commonv3.ProjectRole{
					{
						Project: project,
					},
				},
			},
		}
		ctx := context.WithValue(context.Background(), common.SessionDataKey, &sd)
		res, err := as.GetAuditLogByProjects(ctx, req)
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

func testGetAuditLogForDay(tag string) func(t *testing.T) {
	return func(t *testing.T) {

		db, mock := getDB(t)
		defer db.Close()

		as, err := NewAuditLogDatabaseService(db, tag)
		if err != nil {
			t.Fatal(err)
		}

		timefrom := "1d"
		auditrecord := "{\"actor\":{\"account\":{\"username\":\"admin@paralus.local\"},\"groups\":[\"All Local Users\",\"Organization Admins\"],\"type\":\"USER\"},\"category\":\"AUDIT\",\"client\":{\"host\":\"console-ic-oss.dev.rafay-edge.net\",\"ip\":\"10.0.0.147:47204\",\"type\":\"BROWSER\",\"user_agent\":\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36\"},\"detail\":{\"message\":\"Cluster kind-local deleted\",\"meta\":{\"cluster_name\":\"kind-local\"}},\"origin\":\"core\",\"portal\":\"OPS\",\"project\":\"stage\",\"timestamp\":\"2022-11-21T09:48:30.597615647Z\",\"type\":\"cluster.delete.success\",\"version\":\"1.0\"}"
		auditrecordtwo := "{\"actor\":{\"account\":{\"username\":\"admin@paralus.local\"},\"groups\":[\"All Local Users\",\"Organization Admins\"],\"type\":\"USER\"},\"category\":\"AUDIT\",\"client\":{\"host\":\"console-ic-oss.dev.rafay-edge.net\",\"ip\":\"10.0.0.147:47204\",\"type\":\"BROWSER\",\"user_agent\":\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36\"},\"detail\":{\"message\":\"Cluster kind-local deleted\",\"meta\":{\"cluster_name\":\"kind-local\"}},\"origin\":\"core\",\"portal\":\"OPS\",\"project\":\"stage\",\"timestamp\":\"2022-11-21T09:48:30.597615647Z\",\"type\":\"cluster.delete.success\",\"version\":\"1.0\"}"

		req := &eventv1.GetAuditLogSearchRequest{
			Metadata: &commonv3.Metadata{
				UrlScope: "auditlogs/" + project,
			},
			Filter: &eventv1.AuditLogQueryFilter{
				Timefrom: "now-" + timefrom,
			},
		}

		uuid := uuid.New().String()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "sap"."account_id", "sap"."project_id", "sap"."group_id", "sap"."role_id", "sap"."role_name", "sap"."organization_id", "sap"."partner_id", "sap"."is_global", "sap"."scope", "sap"."permission_name", "sap"."base_url", "sap"."urls" FROM "sentry_account_permission" AS "sap" WHERE (account_id = '` + uuid + `') AND (partner_id = '` + uuid + `') AND (lower(role_name) = 'admin') AND (lower(scope) = 'organization')`)).
			WillReturnRows(sqlmock.NewRows([]string{"account_id", "role_name", "scope"}).AddRow(uuid, "admin", "organization"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "auditlog"."tag", "auditlog"."time", "auditlog"."data" FROM "audit_logs" AS "auditlog" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now())`)).
			WillReturnRows(sqlmock.NewRows([]string{"tag", "time", "data"}).AddRow(tag, time.Now(), auditrecord).AddRow(tag, time.Now(), auditrecordtwo))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'project' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now()) GROUP BY data->>'project'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "project"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->'actor'->'account'->>'username' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now()) GROUP BY data->'actor'->'account'->>'username'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "username"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'type' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now()) GROUP BY data->>'type'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "type"))

		sd := commonv3.SessionData{
			Organization: uuid,
			Partner:      uuid,
			Account:      uuid,
			Username:     "user",
			Project: &commonv3.ProjectData{
				List: []*commonv3.ProjectRole{
					{
						Project: project,
					},
				},
			},
		}
		ctx := context.WithValue(context.Background(), common.SessionDataKey, &sd)
		res, err := as.GetAuditLog(ctx, req)
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

func testGetKubectlCommands(tag string) func(t *testing.T) {
	return func(t *testing.T) {

		db, mock := getDB(t)
		defer db.Close()

		as, err := NewAuditLogDatabaseService(db, tag)
		if err != nil {
			t.Fatal(err)
		}

		timefrom := "1d"
		auditrecord := "{\"actor\":{\"account\":{\"username\":\"admin@paralus.local\"},\"groups\":[\"All Local Users\",\"Organization Admins\"],\"type\":\"USER\"},\"category\":\"AUDIT\",\"client\":{\"host\":\"console-ic-oss.dev.rafay-edge.net\",\"ip\":\"122.50.194.215\",\"type\":\"BROWSER\",\"user_agent\":\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36\"},\"detail\":{\"message\":\"kubectl get all\",\"meta\":{\"cluster_name\":\"kind-local\"}},\"origin\":\"cluster\",\"portal\":\"ADMIN\",\"project\":\"stage\",\"timestamp\":\"2022-11-21T09:46:30.854455438Z\",\"type\":\"kubectl.command.detail\",\"version\":\"1.0\"}"
		auditrecordtwo := "{\"actor\":{\"account\":{\"username\":\"admin@paralus.local\"},\"groups\":[\"All Local Users\",\"Organization Admins\"],\"type\":\"USER\"},\"category\":\"AUDIT\",\"client\":{\"host\":\"console-ic-oss.dev.rafay-edge.net\",\"ip\":\"122.50.194.215\",\"type\":\"BROWSER\",\"user_agent\":\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36\"},\"detail\":{\"message\":\"kubectl get all\",\"meta\":{\"cluster_name\":\"kind-local\"}},\"origin\":\"cluster\",\"portal\":\"ADMIN\",\"project\":\"stage\",\"timestamp\":\"2022-11-21T09:46:30.854455438Z\",\"type\":\"kubectl.command.detail\",\"version\":\"1.0\"}"

		req := &eventv1.GetAuditLogSearchRequest{
			Metadata: &commonv3.Metadata{
				UrlScope: "auditlogs/" + project,
			},
			Filter: &eventv1.AuditLogQueryFilter{
				Timefrom: "now-" + timefrom,
			},
		}

		uuid := uuid.New().String()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "sap"."account_id", "sap"."project_id", "sap"."group_id", "sap"."role_id", "sap"."role_name", "sap"."organization_id", "sap"."partner_id", "sap"."is_global", "sap"."scope", "sap"."permission_name", "sap"."base_url", "sap"."urls" FROM "sentry_account_permission" AS "sap" WHERE (account_id = '` + uuid + `') AND (partner_id = '` + uuid + `') AND (lower(role_name) = 'admin') AND (lower(scope) = 'organization')`)).
			WillReturnRows(sqlmock.NewRows([]string{"account_id", "role_name", "scope"}).AddRow(uuid, "admin", "organization"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "auditlog"."tag", "auditlog"."time", "auditlog"."data" FROM "audit_logs" AS "auditlog" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now())`)).
			WillReturnRows(sqlmock.NewRows([]string{"tag", "time", "data"}).AddRow(tag, time.Now(), auditrecord).AddRow(tag, time.Now(), auditrecordtwo))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'project' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now()) GROUP BY data->>'project'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "project"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->'actor'->'account'->>'username' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now()) GROUP BY data->'actor'->'account'->>'username'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "username"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) as count, data->>'type' as key FROM "audit_logs" WHERE (tag = '` + tag + `') AND (data->>'project' = '` + project + `') AND (time between now() - interval '` + timefrom + `' and now()) GROUP BY data->>'type'`)).
			WillReturnRows(sqlmock.NewRows([]string{"count", "key"}).AddRow(1, "type"))

		sd := commonv3.SessionData{
			Organization: uuid,
			Partner:      uuid,
			Account:      uuid,
			Username:     "user",
			Project: &commonv3.ProjectData{
				List: []*commonv3.ProjectRole{
					{
						Project: project,
					},
				},
			},
		}
		ctx := context.WithValue(context.Background(), common.SessionDataKey, &sd)
		res, err := as.GetAuditLog(ctx, req)
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
