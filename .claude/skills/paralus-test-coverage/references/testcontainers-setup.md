# Testcontainers & sqlmock patterns for Paralus

## 1. Unit test pattern: `go-sqlmock` for the repository layer

Use this when the code under test uses `database/sql` (directly, or through `sqlx`). Add `github.com/DATA-DOG/go-sqlmock` to `go.mod` if not present.

```go
package rolerepo

import (
    "testing"

    sqlmock "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGetRoleByID(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    repo := NewRoleRepository(db) // adapt to actual constructor

    tests := []struct {
        name    string
        id      string
        setup   func()
        wantErr bool
    }{
        {
            name: "role found",
            id:   "role-1",
            setup: func() {
                rows := sqlmock.NewRows([]string{"id", "name"}).AddRow("role-1", "admin")
                mock.ExpectQuery(`SELECT id, name FROM roles WHERE id = \$1`).
                    WithArgs("role-1").
                    WillReturnRows(rows)
            },
        },
        {
            name: "role not found",
            id:   "missing",
            setup: func() {
                mock.ExpectQuery(`SELECT id, name FROM roles WHERE id = \$1`).
                    WithArgs("missing").
                    WillReturnError(sql.ErrNoRows)
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            role, err := repo.GetByID(tt.id)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.id, role.ID)
            assert.NoError(t, mock.ExpectationsWereMet())
        })
    }
}
```

If the code uses raw `pgx` (`pgxpool.Pool`) instead of `database/sql`, `sqlmock` won't attach cleanly — flag this to the user and suggest `github.com/pashagolub/pgxmock` instead, same pattern.

## 2. Hand-written fakes (no testify/mock, no gomock)

```go
type ClusterGetter interface {
    GetCluster(ctx context.Context, id string) (*Cluster, error)
}

type fakeClusterGetter struct {
    getFn func(ctx context.Context, id string) (*Cluster, error)
}

func (f *fakeClusterGetter) GetCluster(ctx context.Context, id string) (*Cluster, error) {
    return f.getFn(ctx, id)
}

func TestAudit_RecordsAccess(t *testing.T) {
    fake := &fakeClusterGetter{
        getFn: func(_ context.Context, id string) (*Cluster, error) {
            return &Cluster{ID: id, Name: "prod"}, nil
        },
    }
    svc := NewAuditService(fake)
    // ... exercise svc, assert on result
}
```

## 3. Integration tests: shared `TestMain` harness

Put shared container bring-up in a `testhelpers` package so multiple `_integration_test.go` files can reuse it without each spinning duplicate containers.

```go
//go:build integration

package testhelpers

import (
    "context"
    "fmt"
    "testing"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/elasticsearch"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

type Harness struct {
    PostgresDSN string
    ESAddress   string
    pg          *postgres.PostgresContainer
    es          *elasticsearch.ElasticsearchContainer
}

func Start(ctx context.Context) (*Harness, error) {
    pg, err := postgres.Run(ctx, "postgres:16-alpine",
        postgres.WithDatabase("paralus_test"),
        postgres.WithUsername("paralus"),
        postgres.WithPassword("paralus"),
        postgres.BasicWaitStrategies(),
    )
    if err != nil {
        return nil, fmt.Errorf("start postgres: %w", err)
    }

    dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        return nil, err
    }

    es, err := elasticsearch.Run(ctx, "docker.elastic.co/elasticsearch/elasticsearch:8.9.0")
    if err != nil {
        return nil, fmt.Errorf("start elasticsearch: %w", err)
    }

    return &Harness{PostgresDSN: dsn, ESAddress: es.Settings.Address, pg: pg, es: es}, nil
}

func (h *Harness) Stop(ctx context.Context) {
    _ = testcontainers.TerminateContainer(h.pg)
    _ = testcontainers.TerminateContainer(h.es)
}
```

Kratos has **no official testcontainers-go module**. Use the generic container API against the public `oryd/kratos` image, pointing it at a config file checked into `testdata/`:

```go
kratosReq := testcontainers.ContainerRequest{
    Image:        "oryd/kratos:v1.0.0",
    ExposedPorts: []string{"4433/tcp", "4434/tcp"},
    Cmd:          []string{"serve", "--config", "/etc/config/kratos/kratos.yml"},
    Files: []testcontainers.ContainerFile{
        {HostFilePath: "testdata/kratos.yml", ContainerFilePath: "/etc/config/kratos/kratos.yml"},
    },
    WaitingFor: wait.ForHTTP("/health/ready").WithPort("4434/tcp"),
}
kratosContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
    ContainerRequest: kratosReq,
    Started:          true,
})
```

Example integration test consuming the harness:

```go
//go:build integration

package rolerepo_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/paralus/paralus/internal/testhelpers"
)

func TestRoleRepository_Integration(t *testing.T) {
    ctx := context.Background()
    h, err := testhelpers.Start(ctx)
    require.NoError(t, err)
    defer h.Stop(ctx)

    // connect real repo to h.PostgresDSN, run actual migrations, exercise real queries
}
```

## 4. Running each layer locally / in CI

```bash
# unit only (fast, no docker required)
go test ./... -race -coverprofile=cover_unit.out -covermode=atomic -coverpkg=./...

# integration (needs docker running)
go test -tags=integration ./... -coverprofile=cover_integration.out -covermode=atomic -coverpkg=./...

# e2e (needs a running Paralus stack)
go test -tags=e2e ./test/e2e/...
```

`go-test-coverage`'s `profile:` field accepts a comma-separated list, so combine unit + integration profiles: `profile: cover_unit.out,cover_integration.out`.
