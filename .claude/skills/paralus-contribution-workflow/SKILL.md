---
name: paralus-contribution-workflow
description: Use this skill whenever contributing to ANY repo under the paralus GitHub org — core Paralus, Relay, Relay Server, Relay Agent, CLI (pctl), Dashboard, or Prompt — and preparing to fork, branch, commit, or open a pull request. Covers the universal OSS contribution workflow: forking, branching off main, issue etiquette, Conventional Commit format, PR scoping, and the PR template checklist. Trigger for "how do I raise a PR for paralus", "what commit message format does this repo use", "should I split this into multiple PRs", "am I following the contribution guidelines", or "I forked and cloned, now what" — even without naming a specific repo. Does NOT cover language-specific test/build/CI mechanics — those live in repo-specific skills like paralus-test-coverage. If none exists for the repo in question, say so and offer to build one instead of improvising.
---

# Paralus Contribution Workflow

Paralus is a Kubernetes access-management platform split across several repos under the `paralus` GitHub org, each with its own stack:

- **Paralus** (core) — Go, RBAC/SSO/audit, depends on Postgres, Kratos, Elasticsearch.
- **Relay** — Relay Server + Relay Agent, Go, gRPC/mTLS.
- **CLI** (`pctl`) — Go, built with Cobra.
- **Dashboard** — the web UI, a different (likely JS/TS) stack from the Go repos.
- **Prompt** — web-based kubectl client built on `kubeprompt`.

Each repo has its own build tooling, test framework, and CI, so don't assume Go conventions apply to Dashboard, or vice versa. What *is* shared across all of them is the open-source contribution process below, confirmed against the core Paralus repo's `CONTRIBUTING.md` and PR template. Treat it as the default for any `paralus`-org repo, but always check that specific repo's own `CONTRIBUTING.md` first — if it says something different, the repo's own file wins, and it's worth flagging the discrepancy to the user rather than silently overriding it.

## The workflow

**1. Fork, then branch — never commit on `main`.**
Work happens on `<your-fork>/main` only insofar as it stays in sync with upstream. Real work goes on a feature branch cut from `main`:
```
git checkout -b <type>/<short-description>
```
If you find uncommitted work already sitting on `main` in a fork, that's a signal to move it to a branch immediately (`git checkout -b <branch>` carries uncommitted changes with it) before committing anything.

**2. Check for an issue before writing code.**
Paralus asks contributors to comment on an existing issue to claim it, or open a new one to discuss the change, before submitting a PR — especially for anything beyond a trivial fix. For a first-time contributor this matters more, not less: it avoids duplicate work and gives a maintainer a chance to weigh in on approach before a large diff exists. Look for "good first issue" labels when picking initial work.

**3. Commit messages: Conventional Commits.**
Format: `<type>(<scope>): <subject>`, e.g. `test(dao): add unit tests for user repository`, `fix(auth): handle expired token refresh`, `ci: add coverage gate workflow`. Common types: `feat`, `fix`, `test`, `ci`, `docs`, `chore`, `refactor`. Keep the scope specific (a package or component name), not the whole repo.

**4. Format and lint before pushing — using that repo's own tools.**
Don't reuse Go commands on a JS repo or vice versa. Check the repo's `CONTRIBUTING.md`, `Makefile`, or `package.json` scripts for the actual commands (for the Go repos this is typically `gofmt -s -l .` and `golangci-lint run`). If a repo-specific skill exists for this (e.g. `paralus-test-coverage` for core), defer to it for the exact commands and conventions rather than guessing.

**5. Keep PRs reviewable — split by area, not one mega-diff.**
A PR touching dozens of files across unrelated packages is a heavy ask for a volunteer maintainer to review, especially from a new contributor with no track record yet. Prefer several small, focused PRs (one per package or feature area) over a single sprawling one. Each should stand on its own: buildable, testable, and understandable in isolation.

**6. Sync with upstream before opening.**
Rebase or merge the latest upstream `main` into your branch before opening the PR, so CI reflects the current state of the target branch, not a stale fork.

**7. Verify CI is green on your fork first.**
Push the branch and confirm the repo's CI workflow passes on your fork before opening the PR against upstream — catching a formatting or lint failure yourself is faster than waiting for a maintainer's CI run to catch it.

**8. Fill out the PR template completely.**
Paralus repos use a PR template with a "what does this change" section, a "depends on other PRs/issues" section, and a checklist:
- Read and followed `CONTRIBUTING.md`
- Added tests for the PR
- Formatted the code (language-appropriate)
- Updated documentation if applicable
- Updated `CHANGELOG.md`

Don't leave checklist items unchecked without explanation — if one genuinely doesn't apply (e.g. no doc changes needed), say so briefly rather than leaving it blank.

## What this skill does not cover

Anything specific to a language or repo's internals — test pyramid conventions, mocking rules, coverage thresholds, testcontainers setup, framework-specific linting — belongs in a skill scoped to that repo (the core Go repo already has one: `paralus-test-coverage`). If you're working in Relay, CLI, Dashboard, or Prompt and no equivalent skill exists, say so explicitly rather than inventing Go-specific conventions for a repo that might use a completely different stack, and offer to help build one for that repo using the same interview-and-draft approach.
