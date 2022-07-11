We 💚 Opensource!

Yes, because we feel that it’s the best way to build and improve a product. It allows people like you from across the globe to contribute and improve a product over time. And we’re super happy to see that you’d like to contribute to Paralus.

We are always on the lookout for anything that can improve the product. Be it feature requests, issues/bugs, code or content, we’d love to see what you’ve got to make this better. If you’ve got anything exciting and would love to contribute, this is the right place to begin your journey as a contributor to Paralus and the larger open source community.

## How to get started?

The easiest way to start is to look at existing issues and see if there’s something there that you’d like to work on. You can filter issues with the label “[Good first issue](https://github.com/paralus/paralus/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)” which are relatively self sufficient issues and great for first time contributors.

Once you decide on an issue, please comment on it so that all of us know that you’re on it.

If you’re looking to add a new feature, [raise a new issue](https://github.com/paralus/paralus/issues/new) and start a discussion with the community. Engage with the maintainers of the project and work your way through.

You'll need to perform the following tasks in order to submit your changes:

- Fork the Paralus repository.
- Create a branch for your changes.
- Add commits to that branch.
- Open a PR to share your contribution.

Below are all the details you need to know about the `Paralus` repo and get started with the development.

# Paralus

This repository contains all the core system components that are the backbone for Paralus.

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
Postgres db, Kratos etc. for core.

_This will start up postgres and elasticsearch as well as kratos and
run the kratos migrations. It will also run all the necessary
migrations. It also starts up a mail slurper for you to use Kratos._

```bash
docker-compose --env-file ./env.example up -d
```

Start core:

```bash
go run github.com/paralus/paralus
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

#### Start Paralus

Start Paralus:

```bash
go run github.com/paralus/paralus
```

#### Updating Proto Files?

- Make sure you have [`buf`](https://github.com/bufbuild/buf) installed
- Install dependencies:
```
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
- Run `make build-proto` to regenerate proto artifacts

## Need Help?

If you are interested to contribute to core but are stuck with any of the steps, feel free to reach out to us. Please [create an issue](https://github.com/paralus/paralus/issues/new) in this repository describing your issue and we'll take it up from there.
