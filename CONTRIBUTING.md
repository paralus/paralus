We 💚 Opensource!

Yes, because we feel that it’s the best way to build and improve a product. It allows people like you from across the globe to contribute and improve a product over time. And we’re super happy to see that you’d like to contribute to ZTKA.

We are always on the lookout for anything that can improve the product. Be it feature requests, issues/bugs, code or content, we’d love to see what you’ve got to make this better. If you’ve got anything exciting and would love to contribute, this is the right place to begin your journey as a contributor to ZTKA and the larger open source community.

**How to get started?**

The easiest way to start is to look at existing issues and see if there’s something there that you’d like to work on. You can filter issues with the label “Good first issue” which are relatively self sufficient issues and great for first time contributors.

Once you decide on an issue, please comment on it so that all of us know that you’re on it.

If you’re looking to add a new feature, raise a new issue and start a discussion with the community. Engage with the maintainers of the project and work your way through.

Below are all the details you need to know about the `RCloud Base` repo and get started with the development.

## RCloud Base

This repository contains all the rcloud-system components that are the backbone for ztka and gitops.

### Prerequisites

- Postgres: Primary database
- Ory Kratos: API for user management
- Elasticsearch: Storage for audit logs

You can use the bitnami/charts for postgres and elastic/helm-charts for elasticsearch.

### Development setup

#### Using docker-compose

Run following Docker Compose command to setup all requirements like Postgres db, Kratos etc. for the rcloud-base.

This will start up postgres and elasticsearch as well as kratos and run the kratos migrations. It will also run all the necessary migrations. It also starts up a mail slurper for you to use Kratos.

`docker-compose up -d`

Start rcloud-base:

`go run github.com/RafayLabs/rcloud-base`

**Manual**

**Start databases**

**Postgres**

```bash
docker run --network host \
    --env POSTGRES_HOST_AUTH_METHOD=trust \
    -v pgdata:/var/lib/postgresql/data \
    -it postgres
```

**Elasticsearch**

```bash
docker run --network host \
    -v elastic-data:/usr/share/elasticsearch/data \
    -e "discovery.type=single-node" \
    -e "xpack.security.enabled=false" \
    -it docker.elastic.co/elasticsearch/elasticsearch:8.0.0
```

**Create the initial db/user**

Scripts for admindb:

```SQL
create database admindb;
CREATE ROLE admindbuser WITH LOGIN PASSWORD '<your_password>';
GRANT ALL PRIVILEGES ON DATABASE admindb to admindbuser;
```

Now in the newly created db:

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
grant execute on function uuid_generate_v4() to admindbuser;
```

Scripts for clusterdb:

```sql
CREATE database clusterdb;
CREATE ROLE clusterdbuser WITH LOGIN PASSWORD '<your_password>';
GRANT ALL PRIVILEGES ON DATABASE clusterdb to clusterdbuser;
```

Now in the newly created db:

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
grant execute on function uuid_generate_v4() to clusterdbuser;
```

This will grant the necessary permission to the newly created user to run uuid_generate_v4()

**Run application migrations**

We use golang-migrate to perform migrations.
Install golang-migrate

`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

-tags `postgres` is important as otherwise it compiles without postgres support

You can refer to the guide for full details.
Run migrations

Example for admindb:

```bash
export POSTGRESQL_URL='postgres://<user>:<pass>@<host>:<port>/admindb?sslmode=disable'
migrate -path ./persistence/migrations/admindb -database "$POSTGRESQL_URL" up
```

See cli-usage for more info.

### Development setup

Copy env.example to .env:

`cp env.example .env`

Run following Docker Compose command to setup all requirements like Postgres db, Kratos etc. for the rcloud-base:

`docker-compose up -d`

Start rcloud-base server:

`go run github.com/RafayLabs/rcloud-base`

**Code Structure**

The following section lists out the code structure for each of the 4 repos. Mention the folder structure along with its importance and what it is for.

```
components
├── adminsrv
│   ├── proto
│   ├── server
│   │   └── organization.go
│   │   └── project.go
│   ├── internal
│   ├── pkg
│   │   └── service
│   ├── Dockerfile.adminsrv
│   └── main.go
│   └── buf.yaml
│   └── buf.gen.yaml
├── authz
│   ├── proto
│   ├── server
│   │   └── authz.go
│   ├── pkg
│   │   └── service
│   ├── Dockerfile.authz
│   └── main.go
│   └── buf.yaml
│   └── buf.gen.yaml
├── usermgmt
│   ├── proto
│   ├── server
│   │   └── user.go
│   │   └── role.go
│   ├── internal
│   ├── pkg
│   │   └── service
│   ├── Dockerfile.usermgmt
│   └── main.go
│   └── buf.yaml
│   └── buf.gen.yaml
├── cluster-scheduler
│   ├── proto
│   ├── server
│   │   └── cluster.go
│   ├── internal
│   ├── pkg
│   │   └── service
│   ├── Dockerfile.cluster-scheduler
│   └── main.go
│   └── buf.yaml
│   └── buf.gen.yaml
├── common
│   ├── proto
│   ├── pkg
│   └── buf.yaml
│   └── buf.gen.yaml
```

**Need Help?**

We’re there for you - the best part of being a part of an open source community. If you are stuck somewhere or are facing an issue or just don’t know how to get started, feel free to let us know.

You can reach out to us via our Slack Channel, Twitter, Discord etc.
