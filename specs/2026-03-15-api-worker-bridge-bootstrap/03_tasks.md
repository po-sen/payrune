---
doc: 03_tasks
spec_date: 2026-03-15
slug: api-worker-bridge-bootstrap
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-14-runtime-entrypoint-alignment
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a narrow package-boundary correction with no contract or runtime behavior change.
- Upstream dependencies (`depends_on`):
  - `2026-03-14-runtime-entrypoint-alignment`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The target move is localized and the desired boundary is clear.
  - What would trigger switching to Full mode:
    - Any broader restructuring of HTTP transport composition or worker runtime contracts.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Move the API worker request bridge into bootstrap.
- M2:
  - Remove the old inbound bridge package and verify imports/tests.

## Tasks (ordered)

1. T-001 - Move API worker bridge logic into bootstrap

   - Scope:
     - Copy or fold the request/response bridge logic into `internal/bootstrap` and update
       `api_worker.go` to use the local helper/types.
   - Output:
     - Updated bootstrap API worker orchestration and tests.
   - Linked requirements: FR-001, FR-003, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted Go tests for
           `./internal/bootstrap` and `./cmd/api-worker`.
     - [ ] Expected result: API worker behavior and JSON contract stay unchanged.
     - [ ] Logs/metrics to check (if applicable): none

2. T-002 - Remove the old inbound bridge package
   - Scope:
     - Delete the old `internal/adapters/inbound/http/cloudflare` package and verify that no
       imports remain.
   - Output:
     - Deleted bridge package and clean import graph.
   - Linked requirements: FR-002, FR-003, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): run `rg` for stale imports, `go list ./...`,
           and full `go test ./...`.
     - [ ] Expected result: no imports remain to the removed package and the repo remains green.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-002
- FR-003 -> T-001, T-002
- NFR-001 -> T-001, T-002
- NFR-002 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Move logic into bootstrap before deleting the old package.
- Rollback steps:
  - Restore the old bridge package if the bootstrap-local move causes regressions.

## Validation evidence

- `SPEC_DIR="specs/2026-03-15-api-worker-bridge-bootstrap" bash scripts/spec-lint.sh`
- `go fmt ./internal/bootstrap ./cmd/api-worker`
- `GOCACHE=/tmp/go-build go test ./internal/bootstrap ./cmd/api-worker`
- `go list ./...`
- `GOCACHE=/tmp/go-build go test ./...`
- `rg -n "internal/adapters/inbound/http/cloudflare|httpcloudflare" cmd internal -g '*.go'`
