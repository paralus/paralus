package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/common"
	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func performClusterBasicChecks(t *testing.T, cluster *infrav3.Cluster, puuid string) {
	if cluster.GetMetadata().GetName() != "cluster-"+puuid {
		t.Error("invalid name returned")
	}
}

func TestCreateCluster(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	downloadData := &common.DownloadData{
		ControlAddr:     "localhost:5002",
		APIAddr:         "localhost:8000",
		RelayAgentImage: "paralus/relay:latest",
	}

	ps := NewClusterService(db, downloadData, NewBootstrapService(db), getLogger())

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
	mock.ExpectExec(`INSERT INTO "cluster_project_cluster"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

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

	downloadData := &common.DownloadData{
		ControlAddr:     "localhost:5002",
		APIAddr:         "localhost:8000",
		RelayAgentImage: "paralus/relay:latest",
	}

	ps := NewClusterService(db, downloadData, NewBootstrapService(db), getLogger())

	puuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id"`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))

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

	downloadData := &common.DownloadData{
		ControlAddr:     "localhost:5002",
		APIAddr:         "localhost:8000",
		RelayAgentImage: "paralus/relay:latest",
	}

	ps := NewClusterService(db, downloadData, NewBootstrapService(db), getLogger())

	puuid := uuid.New().String()
	cuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id"`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))

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

	downloadData := &common.DownloadData{
		ControlAddr:     "localhost:5002",
		APIAddr:         "localhost:8000",
		RelayAgentImage: "paralus/relay:latest",
	}

	ps := NewClusterService(db, downloadData, NewBootstrapService(db), getLogger())

	puuid := uuid.New().String()
	cuuid := uuid.New().String()

	// mock.ExpectQuery(`SELECT "project"."id"`).WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))

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

	downloadData := &common.DownloadData{
		ControlAddr:     "localhost:5002",
		APIAddr:         "localhost:8000",
		RelayAgentImage: "paralus/relay:latest",
	}

	ps := NewClusterService(db, downloadData, NewBootstrapService(db), getLogger())

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

func TestListClusterNoProject(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()
	ps := NewClusterService(db, &common.DownloadData{}, NewBootstrapService(db), getLogger())

	pruuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "project"."id", "project"."name"`).
		WithArgs().WillReturnError(sql.ErrNoRows)

	expect := status.Error(codes.NotFound, "no clusters found")

	qo := commonv3.QueryOptions{
		Project: pruuid,
	}
	_, err := ps.List(context.Background(), query.WithOptions(&qo))
	if err == nil {
		t.Errorf("expect error %s, got no error", expect.Error())
	}
	if err.Error() != expect.Error() {
		t.Errorf("expect error: %s, got error: %s", expect.Error(), err.Error())
	}
}
