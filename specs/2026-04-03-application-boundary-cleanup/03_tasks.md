---
doc: 03_tasks
spec_date: 2026-04-03
slug: application-boundary-cleanup
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-domain-model-boundary-cleanup
  - 2026-04-02-sweep-material-redesign
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
  - This is a refactor-only boundary cleanup with no schema change, no new integration, and no async
    flow redesign.
- Upstream dependencies (`depends_on`):
  - 2026-04-02-domain-model-boundary-cleanup
  - 2026-04-02-sweep-material-redesign
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The change is local to existing application and HTTP adapter contracts.
  - What would trigger switching to Full mode:
    - Any schema change, payload content change, or new collaborator layer.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not applicable. `04_test_plan.md` is produced for this spec.

## Milestones

- M1:
  - Remove persistence representation naming from application outbound contracts.
- M2:
  - Move HTTP JSON response shaping out of application DTOs and into HTTP controllers.

## Tasks (ordered)

1. T-001 - Clean outbound sweep-material contract
   - Scope:
     - Rename/remove `SweepMaterialJSON` from application outbound contracts and use cases.
     - Keep runtime behavior unchanged in deriver and store adapters.
   - Output:
     - Updated `internal/application/ports/outbound` and affected use cases/adapters with
       transport-neutral naming.
   - Linked requirements: FR-001 / FR-003 / NFR-001 / NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
           `rg -n "SweepMaterialJSON" internal/application`
     - [ ] Expected result:
           No runtime matches under `internal/application`
     - [ ] Logs/metrics to check (if applicable):
           None
2. T-002 - Remove HTTP JSON shaping from application DTOs
   - Scope:
     - Strip `json` tags and HTTP-only response types from `internal/application/dto`.
     - Add local HTTP response/error mapping in inbound controllers without changing external
       payloads.
   - Output:
     - Transport-agnostic application DTOs and HTTP adapter-owned response structs/mappers.
   - Linked requirements: FR-002 / FR-003 / NFR-002 / NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
           `rg -n 'json:\"' internal/application/dto`
     - [ ] Expected result:
           No matches
     - [ ] Logs/metrics to check (if applicable):
           None
3. T-003 - Regression validation
   - Scope:
     - Update tests for use cases and HTTP controllers to lock current behavior after boundary
       cleanup.
   - Output:
     - Passing application and controller regression tests plus spec lint.
   - Linked requirements: FR-001 / FR-002 / FR-003 / NFR-001 / NFR-002 / NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
           `go test ./internal/application/... ./internal/adapters/inbound/http/controllers/...`
     - [ ] Expected result:
           All targeted tests pass
     - [ ] Logs/metrics to check (if applicable):
           None

## Traceability (optional)

- FR-001 -> T-001, T-003
- FR-002 -> T-002, T-003
- FR-003 -> T-001, T-002, T-003
- NFR-001 -> T-001, T-003
- NFR-002 -> T-002, T-003
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag:
  - None
- Migration sequencing:
  - None
- Rollback steps:
  - Revert the refactor commit if controller payload compatibility regresses.

## Validation evidence

- `rg -n 'json:\"' internal/application/dto` returned no matches.
- `rg -n 'SweepMaterialJSON' internal/application` returned no matches.
- `go test ./internal/application/... ./internal/adapters/inbound/http/controllers/...` passed.
- `go list ./...` passed.
- `go test ./...` passed.
- `SPEC_DIR="specs/2026-04-03-application-boundary-cleanup" bash scripts/spec-lint.sh` passed.
- `bash scripts/precommit-run.sh` passed.
