# rcloud-base
This repository contains all the rcloud-system components that are the backbone for ztka and gitops.

## Prerequisites 
1) Postgres database
2) [`Ory Kratos`](https://www.ory.sh/kratos) - API for user management
3) We use [`Casbin`](https://casbin.org) - An authorization library that supports access control models like ACL, RBAC, ABAC

## Setting up the database
You can use the [`bitnami charts for postgresql`](https://github.com/bitnami/charts/tree/master/bitnami/postgresql/#installing-the-chart)

### Create the initial db/user

Scripts for `admindb`:

``` sql
create database admindb;
CREATE ROLE admindbuser WITH LOGIN PASSWORD '<your_password>';
GRANT ALL PRIVILEGES ON DATABASE admindb to admindbuser;
```

Now in the newly created db:

``` sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
grant execute on function uuid_generate_v4() to admindbuser;
```

Scripts for `clusterdb`:

``` sql
create database clusterdb;
CREATE ROLE clusterdbuser WITH LOGIN PASSWORD '<your_password>';
GRANT ALL PRIVILEGES ON DATABASE clusterdb to clusterdbuser;
```

Now in the newly created db:

``` sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
grant execute on function uuid_generate_v4() to clusterdbuser;
```


*This will grant the necessary permission to the newly created user to run uuid_generate_v4()*

### Run application migrations

We use [`golang-migrate`](https://github.com/golang-migrate/migrate) to perform migrations.

#### Install [`golang-migrate`](https://github.com/golang-migrate/migrate)

``` shell
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

*`-tags 'postgres'` is important as otherwise it compiles without postgres support*

You can refer to the [guide](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) for full details.

#### Run migrations

Example for `admindb`:

``` shell
export POSTGRESQL_URL='postgres://<user>:<pass>@<host>:<port>/admindb?sslmode=disable'
migrate -path ./persistence/migrations/admindb -database "$POSTGRESQL_URL" up
```

See [cli-usage](https://github.com/golang-migrate/migrate#cli-usage) for more info.