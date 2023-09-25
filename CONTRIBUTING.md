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

#### Create the initial db and user

```sql
create database <db_name>;
CREATE ROLE <db_user> WITH LOGIN PASSWORD '<your_password>';
GRANT ALL PRIVILEGES ON DATABASE <db_name> to <db_user>;
```

#### Ory Kratos

Install Ory Kratos using the [installation
guide](https://www.ory.sh/docs/kratos/install) from Kratos
documentation.

Perform the Kratos migrations:

```bash
export DSN='postgres://<db_user>:<db_password>@<host>:<port>/<db_name>?sslmode=disable'
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
export POSTGRESQL_URL='postgres://<db_user>:<db_password>@<host>:<port>/<db_name>?sslmode=disable'
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

# Commit Message Guidelines

Paralus uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## Commit Message Format

Each commit message consists of a **header**, a **body** and a **footer**. The header has a special format that includes a type, a scope and a subject:

```
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

Samples:

```
build(deps): update kratos sdk to v0.11.0
```

```
feat: support SAML based IdP

New support for SAML based IdP. More deatils about how to configure SAML IdP is availble at paralus.io/docs/saml

Closes #111
```

```
fix(core): incorrect status code for user API

Fix incorrect 500 HTTP status code for user GET API request with invalid parameters.

Closes #89
```

## Message Header

The header is mandatory and the scope of the header is optional.

Any line of the commit message cannot be longer 100 characters! This allows the message to be easier to read on GitHub as well as in various git tools.

### Type

Must be one of the following:

- **build**: Changes that affect the build system or external dependencies (example scopes: go, npm).
- **chore**: Routing changes such as version update in docs, update changelog.
- **ci**: Changes to CI configuration files and scripts.
- **docs**: Documentation changes.
- **feat**: A new feature.
- **fix**: A bug fix.
- **perf**: A code that improves performance.
- **refactor**: A code change that neither fixes a bug nor adds a feature.
- **revert**: Reverts a previous commit.
- **style**: Changes that do not affect the meaning of the code (white-space, formatting etc).
- **test**: Adding missing tests or correcting existing tests.

### Scope

Scope is an optional in commit message. The following is the list of supported scopes:
  ```
  TBD
  ```

### Subject

The subject contains a succinct description of the change:

  - use the imperative, present tense: "change" not "changed" nor "changes"
  - don't capitalize the first letter
  - no dot (.) at the end
  
## Message Body

Just as in the subject, use the imperative, present tense: "change" not "changed" nor "changes". The body should include the motivation for the change and contrast this with previous behavior.

## Message Footer

The footer should contain any information about Breaking Changes and is also the place to reference GitHub issues that this commit **Closes**.

**Breaking Changes** should start with the word `BREAKING CHANGE: ` with a space. The rest of the commit message is then the description of the change, justification and migration notes.

Closed bugs should be listed on a separate line in the footer prefixed with "Closes" keyword like this:

```
Closes #177
```

or in case of multiple issues:

```
Closes #177, #200, #251
```

# DCO Sign off

All authors to the project retain copyright to their work. However, to ensure
that they are only submitting work that they have rights to, we are requiring
everyone to acknowledge this by signing their work.

Any copyright notices in this repo should specify the authors as "the
paralus contributors".

To sign your work, just add a line like this at the end of your commit message:

```
Signed-off-by: Joe Bloggs <joe@example.com>
```

This can easily be done with the `--signoff` option to `git commit`.
You can also mass sign-off a whole PR with `git rebase --signoff master`, replacing
`master` with the branch you are creating a pull request against, if not master.

By doing this you state that you can certify the following (from https://developercertificate.org/):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

# Need Help?

If you are interested to contribute to core but are stuck with any of the steps, feel free to reach out to us. Please [create an issue](https://github.com/paralus/paralus/issues/new) in this repository describing your issue and we'll take it up from there.
