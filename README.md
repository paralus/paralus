# Rcloud Base

This repository contains all the rcloud-system components that are the backbone for ztka and gitops.

## Prerequisites

- [Postgres](https://github.com/postgres/postgres): Primary database
- [Ory Kratos](https://www.ory.sh/kratos): API for user management
- [Elasticsearch](https://www.elastic.co/elasticsearch/): Storage for audit logs

> You can use the
> [bitnami/charts](https://github.com/bitnami/charts/tree/master/bitnami/postgresql/#installing-the-chart)
> for postgres and
> [elastic/helm-charts](https://github.com/elastic/helm-charts) for
> elasticsearch.

## Development setup

### Using `docker-compose`

Run following Docker Compose command to setup all requirements like
Postgres db, Kratos etc. for the rcloud-base.

_This will start up postgres and elasticsearch as well as kratos and
run the kratos migrations. It will also run all the necessary
migrations. It also starts up a mail slurper for you to use Kratos._

```bash
docker-compose --env-file ./env.example up -d
```

Start rcloud-base:

```bash
go run github.com/RafayLabs/rcloud-base
```

### Manual

#### Start databases

##### Postgres

```bash
docker run --network host \
    --env POSTGRES_HOST_AUTH_METHOD=trust \
    -v pgdata:/var/lib/postgresql/data \
    -it postgres
```

##### Elasticsearch

```bash
docker run --network host \
    -v elastic-data:/usr/share/elasticsearch/data \
    -e "discovery.type=single-node" \
    -e "xpack.security.enabled=false" \
    -it docker.elastic.co/elasticsearch/elasticsearch:8.0.0
```

#### Create the initial db and user

```sql
create database admindb;
CREATE ROLE admindbuser WITH LOGIN PASSWORD '<your_password>';
GRANT ALL PRIVILEGES ON DATABASE admindb to admindbuser;
```

#### Ory Kratos

Install Ory Kratos using the [installation
guide](https://www.ory.sh/docs/kratos/install) from Kratos
documentation.

Perform the Kratos migrations:

```bash
export DSN='postgres://<user>:<pass>@<host>:<port>/admindb?sslmode=disable'
kratos -c <kratos-config> migrate sql -e --yes
```

Start the Ory Kratos server using kratos config provided in
[_kratos](./_kratos) directory.

#### Run application migrations

We use [`golang-migrate`](https://github.com/golang-migrate/migrate) to perform migrations.

##### Install [`golang-migrate`](https://github.com/golang-migrate/migrate)

```shell
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

_`-tags 'postgres'` is important as otherwise it compiles without postgres support_

You can refer to the [guide](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) for full details.

##### Run migrations

_It is required to perform Kratos migrations before this step._

```shell
export POSTGRESQL_URL='postgres://<user>:<pass>@<host>:<port>/admindb?sslmode=disable'
migrate -path ./persistence/migrations/admindb -database "$POSTGRESQL_URL" up
```

See [cli-usage](https://github.com/golang-migrate/migrate#cli-usage) for more info.

#### Start application

Start rcloud-base:

```bash
go run github.com/RafayLabs/rcloud-base
```
