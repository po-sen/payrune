---
doc: 04_test_plan
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

# Test Plan

## Scope

- Verify runtime/deployment naming rename only.
- Verify build/deploy/import references resolve to the new names.
- Verify stale tracked directory names are removed from current operational docs/specs.

## Test cases

- TC-001 (FR-001, NFR-001)

  - Check `cmd/api` and `build/api` exist.
  - Check old `cmd/app` and `build/app` paths are gone.

- TC-002 (FR-002, FR-003, NFR-001)

  - Check hyphenated directories exist:
    - `cmd/api-worker`
    - `cmd/poller-worker`
    - `cmd/webhook-dispatcher`
    - `cmd/fake-webhook-receiver`
  - Check compact package-oriented directories exist:
    - `internal/application/usecases`
    - `internal/domain/valueobjects`
    - `internal/application/ports/inbound`
    - `internal/application/ports/outbound`
  - Run targeted Go tests for renamed command packages.
  - Verify build/import references point at the new command paths.

- TC-003 (FR-004, FR-005, NFR-002)

  - Run repo-wide `rg` for stale names:
    - `cmd/app`
    - `build/app`
    - `services.app`
    - `cmd/api_worker`
    - `cmd/poller_worker`
    - `cmd/webhook_dispatcher`
    - `cmd/fake_webhook_receiver`
    - `internal/application/use_cases`
    - `internal/domain/value_objects`
    - `internal/application/use-cases`
    - `internal/domain/value-objects`
    - `internal/application/ports/in`
    - `internal/application/ports/out`
    - `scripts/build-cf-api_worker-wasm.sh`
    - `scripts/build-cf-poller_worker-wasm.sh`
    - `scripts/cf-api_worker-deploy.sh`
    - `scripts/cf-api_worker-delete.sh`
    - `scripts/cf-poller_worker-deploy.sh`
    - `scripts/cf-poller_worker-delete.sh`
    - `use_cases` in `AGENTS.md`
    - `value_objects` in `AGENTS.md`
    - `ports/in` in `AGENTS.md`
    - `ports/out` in `AGENTS.md`

- TC-004 (NFR-003)
  - Run:
    - `SPEC_DIR="specs/2026-03-12-api-worker-naming-unification" bash scripts/spec-lint.sh`
    - related current Cloudflare specs lint

## Commands

```bash
SPEC_DIR="specs/2026-03-12-api-worker-naming-unification" bash scripts/spec-lint.sh
```
