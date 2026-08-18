# Test Pyramid guide for Paralus

## Why this matters here specifically

The ask isn't just "hit 70%" — it's "hit 70% by following the Test Pyramid." That means the bulk of the new coverage should come from fast, isolated unit tests. Integration tests (real Postgres/Elasticsearch/Kratos via testcontainers) should cover the seams between components — not be used as a shortcut to cover business logic that could be unit-tested. E2E tests should stay few and high-value.

Rough target shape for new test additions:

- **~70%** of new test cases: unit tests
- **~20%**: integration tests (real dependency wiring)
- **~10%**: e2e tests (full-stack critical paths)

If a proposed change is mostly integration or e2e tests, pause and ask whether the same behavior could be captured faster and more reliably as a unit test with a fake or sqlmock instead — don't just add whatever's requested without flagging an inverted pyramid.

## Worked examples by typical Paralus package shape

**RBAC / permission-check logic** (pure decision logic given roles/permissions as input)
→ Unit test. Table-driven: for each (subject, resource, action) combination, assert allow/deny. No DB needed if the check operates on already-loaded role data; if it queries the DB to resolve roles first, split into (a) unit test the decision function with fake loaded data, and (b) a separate sqlmock/integration test for the loading query.

**systemrpc handlers** (gRPC-style internal service calls)
→ Unit test the handler logic with a fake for whatever it depends on (e.g. `fakeClusterGetter`). Add one integration test per handler that talks to a real dependency, to catch wiring/serialization issues unit tests can't (e.g. wrong SQL, ES mapping mismatches) — not one per input variation.

**HTTP API handlers** (e.g. `internal/server`, REST endpoints)
→ Unit test with `httptest.NewRecorder()` + fakes for the service layer, covering status codes and error-body shape. Integration test only for a handler that's genuinely hard to unit test in isolation (e.g. one that does its own transaction management across Postgres and Elasticsearch).

**Cluster relay / audit logging**
→ Unit test the event-shaping logic. Integration test for "does an audit event actually land in Elasticsearch in the expected shape" — this is a good integration-test candidate since it's exactly the cross-component seam integration tests exist for.

**Login / SSO / cluster onboarding flows**
→ These are the e2e candidates: a small number of smoke tests that go through the real Kratos flow end-to-end against a running stack. Resist the urge to enumerate every edge case here — edge cases belong in unit tests on the underlying logic; e2e just proves the happy path (and maybe one or two critical failure paths, like "expired token is rejected") actually works wired together.

## Red flags to call out to the user

- A PR that adds 10 integration tests and 1 unit test for a package that's mostly pure logic — inverted pyramid, will make CI slow and flaky.
- Tests that only assert "no error returned" without checking the actual returned value/state — this pads coverage numbers without catching regressions.
- Integration tests with no build tag — these will run (and need Docker) on every `go test ./...` invocation, including fast unit-only CI jobs, and will break local `go test ./...` for anyone without Docker running.
- Coverage-driven tests with no meaningful assertions (calling a function just to "touch" the lines). Flag this rather than writing it, even if it would technically raise the percentage.
