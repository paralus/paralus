---
name: paralus-test-coverage
description: Use this skill whenever the user is working on the Paralus project (github.com/paralus/paralus, the Go core repo for RBAC/SSO/audit access management) and wants to improve test coverage, add unit/integration/e2e tests, set up testcontainers, or wire up a GitHub Actions coverage gate. Trigger this for requests like "add tests for this package", "improve coverage", "set up integration tests", "follow the Test Pyramid", "add a testcontainers test for Postgres/Elasticsearch/Kratos", or "add a coverage GitHub Action" — even if the user doesn't mention Paralus by name but is clearly working in this repo's Go codebase (imports from github.com/paralus/paralus, uses Kratos/pgx/Elasticsearch). Always consult this skill before writing any *_test.go file or CI workflow in this repo so the Test Pyramid ratio, mocking conventions, and 70% coverage gate stay consistent across the whole codebase.
---

# Paralus Test Coverage & Test Pyramid

Paralus (github.com/paralus/paralus) is the Go core for a Kubernetes access-management platform: RBAC, SSO, audit logs. Per its CONTRIBUTING.md, local dev depends on **Postgres**, **Elasticsearch**, and **Kratos** (identity server), run via docker-compose. Keep that dependency set in mind — it's exactly what integration tests need to spin up for real.

The goal for this repo: **70% overall statement coverage**, reached by following the **Test Pyramid** (lots of unit tests, a meaningful layer of integration tests against real dependencies via testcontainers, a thin layer of e2e tests) — not by writing shallow tests that only pad the number.

## Conventions to follow exactly (confirmed with the maintainer)

| Layer | Build tag | File suffix | Tooling |
|---|---|---|---|
| Unit | none | `_test.go` | stdlib `testing`, `testify/assert`/`require` for assertions only |
| Integration | `//go:build integration` | `_integration_test.go` | stdlib `testing` + `testcontainers-go` (real Postgres, Elasticsearch, Kratos) |
| E2E | `//go:build e2e` | `_e2e_test.go` | stdlib `testing`, hits a running Paralus stack's real API |

**Mocking — do not deviate:**
- No `testify/mock`, no `gomock`/`mockgen`.
- Define a small interface at the point of use and write a minimal hand-written fake struct implementing it (e.g. `type fakeUserGetter struct{ getFn func(...) (...) }`).
- For the DB/repository layer, use `github.com/DATA-DOG/go-sqlmock` in unit tests. Check the file first for whether it uses `database/sql`, `sqlx`, or `pgx` — sqlmock wraps `database/sql`, so if the repo uses raw `pgx`, say so and ask before assuming sqlmock will work cleanly, since pgx-native code sometimes needs `pgxmock` instead.

## Workflow when asked to improve coverage for a file/package

1. **Read the actual code first** (`view`/`bash_tool cat`). Identify exported functions, branches, and error-return paths. Don't guess at signatures.
2. **Check for an existing `_test.go`** in the same package — extend it, don't create a duplicate or a competing test file.
3. **Classify each function**:
   - Pure logic / transforms → unit test only.
   - Repository / DB query methods → unit test with `go-sqlmock` **and**, if the user wants integration coverage too, a companion `_integration_test.go` using testcontainers.
   - Cross-service flows (e.g. RBAC check touching Postgres + Kratos) → integration test.
   - A critical user-facing path (login, cluster onboarding, role assignment) → e2e candidate, but keep the e2e suite small — this is the top of the pyramid, not the base.
4. **Write table-driven unit tests**: happy path, at least one error path per returned error, boundary/edge cases (empty input, nil, zero values). Use subtests (`t.Run`) with descriptive names.
5. For DB-backed code, follow `references/testcontainers-setup.md` for the sqlmock unit pattern and the testcontainers integration pattern (Postgres, Elasticsearch, and a generic-container recipe for Kratos, which has no official testcontainers module).
6. **Never claim a coverage percentage you haven't measured.** If you have the repo checked out in the sandbox, actually run:
   ```bash
   go test ./... -coverprofile=cover.out -covermode=atomic -coverpkg=./...
   go tool cover -func=cover.out | tail -1
   ```
   and report the real number. If you don't have the repo available, say so explicitly and give the user the exact commands to run themselves instead of inventing a figure.
7. When asked for CI, use `assets/ci-coverage.yml` and `assets/testcoverage.yml` as the starting point (see below) rather than writing a workflow from scratch.

## GitHub Actions coverage gate

Use **`vladopajic/go-test-coverage`** (confirmed choice) as the threshold-gate action, config-driven via `.testcoverage.yml`.

Two things to get right, and both were explicitly requested by the maintainer:
- **70% total** project coverage (`threshold.total: 70`).
- **Fail the PR if a changed file drops below 70%.**

Important nuance to explain to the user (don't silently pick one): `go-test-coverage`'s `threshold.file` setting applies to **every** file in the repo, not just files changed in the PR — so setting `threshold.file: 70` repo-wide will fail on *any* under-covered file, old or new, which is stricter than "only changed files." The tool's `diff` block instead compares the PR's total coverage against a stored base-branch breakdown and can fail if coverage *drops*, but that's a total-coverage delta, not a per-changed-file floor. There is no built-in "70% on changed files only" mode. Present the tradeoff plainly:
- **Option A (simplest, matches "file threshold" literally):** `threshold.file: 70` + `threshold.total: 70`, and use `override` rules to temporarily exempt known-low legacy files while they're being brought up to par.
- **Option B (matches "don't regress" intent more precisely):** `threshold.total: 70` plus the `diff` block with `base-breakdown-file-name` generated on `main` and a small negative `diff.threshold` (e.g. `-0.5`) so PRs can't meaningfully drop total coverage, combined with `go-coverage-report`-style PR annotations if per-line-changed granularity is wanted (that's a separate action, not built into `go-test-coverage`).

Default to **Option A** unless the user says they want strict diff-based gating — it's simpler to reason about and was the literal ask. Mention Option B as the more precise alternative.

Starter files: copy `assets/ci-coverage.yml` to `.github/workflows/test-coverage.yml` and `assets/testcoverage.yml` to `.testcoverage.yml` in the target repo, then adjust the Go version and any package-specific `override`/`exclude` paths (e.g. generated protobuf/`.pb.go` files should be excluded, not tested).

## Reference files

- `references/testcontainers-setup.md` — ready-to-adapt Go code for Postgres, Elasticsearch, and Kratos containers, a shared `TestMain` harness pattern, and the `go-sqlmock` unit-test pattern for the repository layer.
- `references/test-pyramid-guide.md` — pyramid ratio guidance and worked examples of what counts as unit vs. integration vs. e2e for typical Paralus packages (RBAC checks, systemrpc handlers, HTTP API handlers, cluster-relay logic).
- `assets/ci-coverage.yml` — GitHub Actions workflow: unit job (every push/PR), integration job (testcontainers, needs Docker — `ubuntu-latest` runners have it preinstalled), coverage-gate job.
- `assets/testcoverage.yml` — `go-test-coverage` config template implementing Option A above.

Read the relevant reference file before writing integration tests or CI YAML — don't reconstruct the testcontainers API from memory, since module APIs (e.g. `postgres.Run` vs the deprecated `postgres.RunContainer`) change between versions.
