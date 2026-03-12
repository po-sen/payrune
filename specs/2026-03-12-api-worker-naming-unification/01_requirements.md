---
doc: 01_requirements
spec_date: 2026-03-12
slug: api-worker-naming-unification
mode: Full
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Functional Requirements

### FR-001 - API service naming uses `api`

- The repository MUST rename runtime/deployment references to the main HTTP service from `app` to
  `api`.
- Acceptance criteria:
  - [ ] `cmd/api/main.go` exists and old `cmd/app/main.go` no longer exists.
  - [ ] `build/api/Dockerfile` exists and old `build/app/Dockerfile` no longer exists.
  - [ ] Compose service key is `api` instead of `app`.
  - [ ] Operational docs/spec text no longer instruct operators to use `app` for the API runtime.

### FR-002 - Tracked underscore directories are renamed to hyphenated paths

- The repository MUST rename tracked directory paths that use `_` so they use `-` instead.
- Acceptance criteria:
  - [ ] `cmd/api-worker/` exists and `cmd/api_worker/` no longer exists.
  - [ ] `cmd/poller-worker/` exists and `cmd/poller_worker/` no longer exists.
  - [ ] `cmd/webhook-dispatcher/` exists and `cmd/webhook_dispatcher/` no longer exists.
  - [ ] `cmd/fake-webhook-receiver/` exists and `cmd/fake_webhook_receiver/` no longer exists.
  - [ ] `internal/application/use-cases/` exists and
        `internal/application/use_cases/` no longer exists.
  - [ ] `internal/domain/value-objects/` exists and
        `internal/domain/value_objects/` no longer exists.

### FR-003 - Internal package-oriented directories use compact names

- Selected internal directories MUST be renamed so directory names, import paths, and package names
  stay aligned with the intended compact package naming style.
- Acceptance criteria:
  - [ ] `internal/application/usecases/` exists and `internal/application/use-cases/` no longer
        exists.
  - [ ] `internal/domain/valueobjects/` exists and `internal/domain/value-objects/` no longer
        exists.
  - [ ] `internal/application/ports/inbound/` exists and `internal/application/ports/in/` no
        longer exists.
  - [ ] `internal/application/ports/outbound/` exists and `internal/application/ports/out/` no
        longer exists.
  - [ ] Package names and import paths align with `usecases`, `valueobjects`, `inbound`, and
        `outbound`.

### FR-004 - References align with renamed tracked paths

- Build scripts, script filenames, imports, deploy docs, and current Cloudflare references MUST
  align with the renamed tracked paths.
- Acceptance criteria:
  - [ ] Go build paths and imports reference the renamed tracked paths.
  - [ ] Cloudflare helper script filenames under `scripts/` use `-`, not `_`.
  - [ ] Current docs/spec references point to the renamed `cmd/`, `internal/`, and `scripts/`
        paths where applicable.
  - [ ] Repo runtime/deployment docs do not point to stale underscored tracked paths.

### FR-005 - Current docs/specs are updated consistently

- Current docs/specs and `AGENTS.md` referencing the renamed runtime paths MUST be updated to the
  new names.
- Acceptance criteria:
  - [ ] Current Cloudflare specs point to the renamed `cmd/`, `internal/`, and `scripts/` paths.
  - [ ] `AGENTS.md` points to `usecases`, `valueobjects`, `ports/inbound`, and `ports/outbound`
        where it describes this repository's local structure.
  - [ ] Existing repo docs/specs touched by this rename no longer reference stale tracked paths.
  - [ ] Historical specs that are not required for this change are left untouched.

## Non-functional Requirements

### NFR-001 - No behavioral change

- This rename MUST NOT change API, poller, webhook, or migration behavior.
- Verification:
  - Relevant targeted tests pass after rename.

### NFR-002 - Mechanical rename remains verifiable

- The rename MUST be verifiable with deterministic repo-wide search checks.
- Verification:
  - `rg` checks for stale names return no matches for the targeted tracked directory paths.

### NFR-003 - Spec compliance

- This spec and affected current Cloudflare specs MUST pass `spec-lint`.
- Verification:
  - `SPEC_DIR="specs/2026-03-12-api-worker-naming-unification" bash scripts/spec-lint.sh`
  - related current Cloudflare specs also pass after updates
