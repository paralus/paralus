package testutil

import (
	"database/sql"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

var (
	_log = log.GetLogger()
)

// test db constants
const (
	ClusterDB         = "clusterdb"
	ClusterDBUser     = "clusterdbuser"
	ClusterDBPassword = "clusterdbpassword"
)

type queryHook struct {
}

// GetDB returns testdb
func GetDB() *bun.DB {

	dsn := "postgres://" + ClusterDBUser + ":" + ClusterDBPassword + "@localhost:5432" + "/" + ClusterDB + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	return db
}
