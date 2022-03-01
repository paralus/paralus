package service

import (
	"context"
	"fmt"
	"testing"

	goruntime "runtime"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/bootstrapper"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/query"
	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

func getDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal("unable to create sqlmock:", err)
	}
	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
	return db, mock
}

func performClusterBasicChecks(t *testing.T, cluster *infrav3.Cluster, puuid string) {
	if cluster.GetMetadata().GetName() != "cluster-"+puuid {
		t.Error("invalid name returned")
	}
}

func TestCreateCluster(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	sentryPool := sentryrpc.NewSentryPool("localhost:10000", 5*goruntime.NumCPU())

	downloadData := &bootstrapper.DownloadData{
		ControlAddr:     "localhost:5002",
		APIAddr:         "localhost:8000",
		RelayAgentImage: "rafaysystems/relay:latest",
	}

	ps := NewClusterService(db, db, downloadData, sentryPool)
	defer ps.Close()

	puuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id", "project"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "project-"+puuid))
	mock.ExpectQuery(`SELECT "cluster"."id", "cluster"."organization_id"`).
		WithArgs().WillReturnError(fmt.Errorf("sql: no rows in result set"))

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "cluster_tokens"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cuuid))
	mock.ExpectQuery(`INSERT INTO "cluster_clusters"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cuuid))
	mock.ExpectExec(`UPDATE "cluster_clusters"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectExec(`INSERT INTO "cluster_project_cluster"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	cluster := &infrav3.Cluster{
		Metadata: &v3.Metadata{Id: cuuid, Name: "cluster-" + cuuid, Organization: "orgname", Project: "project-" + puuid},
		Spec: &infrav3.ClusterSpec{
			ClusterType: "imported",
		},
	}
	cluster, err := ps.Create(context.Background(), cluster)
	if err != nil {
		t.Fatal("could not create cluster:", err)
	}
	performClusterBasicChecks(t, cluster, cuuid)
}

func TestUpdateCluster(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	sentryPool := sentryrpc.NewSentryPool("localhost:10000", 5*goruntime.NumCPU())

	downloadData := &bootstrapper.DownloadData{
		ControlAddr:     "localhost:5002",
		APIAddr:         "localhost:8000",
		RelayAgentImage: "rafaysystems/relay:latest",
	}

	ps := NewClusterService(db, db, downloadData, sentryPool)
	defer ps.Close()

	puuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "cluster"."id", "cluster"."organization_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cuuid))

	mock.ExpectExec(`UPDATE "cluster_clusters"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	cluster := &infrav3.Cluster{
		Metadata: &v3.Metadata{Id: cuuid, Name: "cluster-" + cuuid, Organization: "orgname", Project: "project-" + puuid},
		Spec: &infrav3.ClusterSpec{
			ClusterType: "imported",
		},
	}
	cluster, err := ps.Update(context.Background(), cluster)
	if err != nil {
		t.Fatal("could not update cluster:", err)
	}
	performClusterBasicChecks(t, cluster, cuuid)
}

func TestSelectCluster(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	sentryPool := sentryrpc.NewSentryPool("localhost:10000", 5*goruntime.NumCPU())

	downloadData := &bootstrapper.DownloadData{
		ControlAddr:     "localhost:5002",
		APIAddr:         "localhost:8000",
		RelayAgentImage: "rafaysystems/relay:latest",
	}

	ps := NewClusterService(db, db, downloadData, sentryPool)
	defer ps.Close()

	puuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "cluster"."id", "cluster"."organization_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cuuid))

	mock.ExpectQuery(`SELECT "projectcluster"."cluster_id", "projectcluster"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cuuid))

	cluster := &infrav3.Cluster{
		Metadata: &v3.Metadata{Id: cuuid, Name: "cluster-" + cuuid, Organization: "orgname", Project: "project-" + puuid},
		Spec: &infrav3.ClusterSpec{
			ClusterType: "imported",
		},
	}
	cluster, err := ps.Select(context.Background(), cluster, false)
	if err != nil {
		t.Fatal("could not fetch cluster:", err)
	}
	performClusterBasicChecks(t, cluster, cuuid)
}

func TestGetCluster(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	sentryPool := sentryrpc.NewSentryPool("localhost:10000", 5*goruntime.NumCPU())

	downloadData := &bootstrapper.DownloadData{
		ControlAddr:     "localhost:5002",
		APIAddr:         "localhost:8000",
		RelayAgentImage: "rafaysystems/relay:latest",
	}

	ps := NewClusterService(db, db, downloadData, sentryPool)
	defer ps.Close()

	puuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "cluster"."id", "cluster"."organization_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cuuid))

	mock.ExpectQuery(`SELECT "projectcluster"."cluster_id", "projectcluster"."project_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cuuid))

	qo := commonv3.QueryOptions{
		ClusterID: cuuid,
		Name:      "cluster-" + cuuid,
		Project:   puuid,
	}
	cluster, err := ps.Get(context.Background(), query.WithOptions(&qo))
	if err != nil {
		t.Fatal("could not fetch cluster:", err)
	}
	performClusterBasicChecks(t, cluster, cuuid)
}

func TestListCluster(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	sentryPool := sentryrpc.NewSentryPool("localhost:10000", 5*goruntime.NumCPU())

	downloadData := &bootstrapper.DownloadData{
		ControlAddr:     "localhost:5002",
		APIAddr:         "localhost:8000",
		RelayAgentImage: "rafaysystems/relay:latest",
	}

	ps := NewClusterService(db, db, downloadData, sentryPool)
	defer ps.Close()

	puuid := uuid.New().String()
	ouuid := uuid.New().String()
	pruuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id", "project"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "partner_id"}).AddRow(cuuid, ouuid, puuid))

	mock.ExpectQuery(`SELECT "cluster"."id", "cluster"."organization_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "partner_id"}).AddRow(cuuid, ouuid, puuid))

	mock.ExpectQuery(`SELECT "projectcluster"."project_id", "projectcluster"."cluster_id"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"cluster_id", "project_id"}).AddRow(cuuid, puuid))

	mock.ExpectQuery(`SELECT "partner"."id", "partner"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(puuid, "partner-"+puuid))

	qo := commonv3.QueryOptions{
		Project: pruuid,
	}
	_, err := ps.List(context.Background(), query.WithOptions(&qo))
	if err != nil {
		t.Fatal("could not fetch cluster:", err)
	}
}
