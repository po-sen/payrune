---
doc: 03_tasks
spec_date: 2026-03-08
slug: container-store-injection-refactor
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-08-payment-address-idempotency-key
  - 2026-03-08-payment-address-status-api
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Container Store Injection Refactor - Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a behavior-preserving dependency refactor with no schema change, no new integration, and no API contract expansion.
- Upstream dependencies (`depends_on`):
  - `2026-03-08-payment-address-idempotency-key`
  - `2026-03-08-payment-address-status-api`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The change is limited to use-case dependency shape, replay read orchestration, and container wiring cleanup.
  - What would trigger switching to Full mode:
    - A persistence-model redesign or new transaction abstraction.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable; `04_test_plan.md` is included.

## Milestones

- M1:
  - Finalize the refactor scope and dependency targets in spec.
- M2:
  - Remove unnecessary store injection from allocate flow and simplify container wiring.

## Tasks (ordered)

1. T-001 - Finalize the dependency-refactor spec
   - Scope:
     - Capture which injected stores are unnecessary and what behavior must remain unchanged.
   - Output:
     - `specs/2026-03-08-container-store-injection-refactor/*.md`
   - Linked requirements: FR-001, FR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-08-container-store-injection-refactor" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec lint passes and documents consistently describe the dependency refactor.
     - [ ] Logs/metrics to check (if applicable): none
2. T-002 - Remove direct replay-store injection from allocate use case
   - Scope:
     - Rework replay lookup and duplicate-claim recovery to use `UnitOfWork` read transactions instead of DB-scoped injected stores.
     - Update constructor shape and tests accordingly.
   - Output:
     - Allocate use case depends on `UnitOfWork` rather than direct allocation/idempotency store injection.
   - Linked requirements: FR-001, NFR-001, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/application/use_cases -count=1`
     - [ ] Expected result: use-case tests pass for success, replay, duplicate-claim recovery, and conflict behavior.
     - [ ] Logs/metrics to check (if applicable): none
3. T-003 - Simplify API container payment-flow wiring
   - Scope:
     - Remove no-longer-needed named store assignments from the API container.
     - Inline the status finder at its single use site if that leaves a clearer constructor.
   - Output:
     - Cleaner container payment-flow wiring with only meaningful named dependencies.
   - Linked requirements: FR-002, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/infrastructure/di ./internal/adapters/inbound/http/controllers ./internal/adapters/outbound/persistence/postgres -count=1` and `GOCACHE=/tmp/go-build go list ./...`
     - [ ] Expected result: DI, controller, and persistence suites pass and packages compile cleanly.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-003
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
  - Restore the previous allocate use-case constructor and container wiring if replay behavior regresses.
