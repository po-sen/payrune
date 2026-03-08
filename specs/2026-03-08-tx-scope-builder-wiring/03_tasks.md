---
doc: 03_tasks
spec_date: 2026-03-08
slug: tx-scope-builder-wiring
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-07-architecture-naming-refactor
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Tx Scope Builder Wiring - Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a small behavior-preserving wiring refactor with no schema change, new integration, or API contract change.
- Upstream dependencies (`depends_on`):
  - `2026-03-07-architecture-naming-refactor`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The change only simplifies wiring placement for a fixed Postgres transaction scope.
  - What would trigger switching to Full mode:
    - A new transaction model, persistence contract change, or broader DI redesign.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable; `04_test_plan.md` is included.

## Milestones

- M1:
  - Update the spec for the tx-scope builder refactor.
- M2:
  - Remove tx-scope builder/factory injection and verify behavior stays unchanged.

## Tasks (ordered)

1. T-001 - Finalize the tx-scope builder refactor spec
   - Scope:
     - Capture the exact refactor boundary and behavior-preserving constraints.
   - Output:
     - `specs/2026-03-08-tx-scope-builder-wiring/*.md`
   - Linked requirements: FR-001, FR-002, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-08-tx-scope-builder-wiring" bash scripts/spec-lint.sh`
     - [x] Expected result: spec lint passes and documents consistently describe the wiring move.
     - [x] Logs/metrics to check (if applicable): none
2. T-002 - Move production and test tx-scope builder wiring
   - Scope:
     - Simplify `UnitOfWork` so it assembles the fixed Postgres tx scope itself.
     - Remove tx-scope builder/factory injection from containers.
     - Update unit tests accordingly.
   - Output:
     - Behavior-preserving tx-scope wiring with clearer composition-root ownership.
   - Linked requirements: FR-001, FR-002, NFR-001, NFR-002, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/adapters/outbound/persistence/postgres ./internal/infrastructure/di -count=1`
     - [x] Expected result: transaction wiring tests pass and callback scope contents remain intact.
     - [x] Logs/metrics to check (if applicable): none
3. T-003 - Final validation and spec sync
   - Scope:
     - Run lint/tests after the wiring move and set the spec to final state.
   - Output:
     - Done spec and validated refactor.
   - Linked requirements: FR-001, FR-002, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-08-tx-scope-builder-wiring" bash scripts/spec-lint.sh` and `GOCACHE=/tmp/go-build go list ./...`
     - [x] Expected result: spec lint and package listing pass.
     - [x] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002, T-003
- FR-002 -> T-001, T-002, T-003
- NFR-001 -> T-002
- NFR-002 -> T-002
- NFR-005 -> T-003
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore the old helper-based tx-scope builder arrangement if the refactor introduces regression.
