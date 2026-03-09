---
doc: 03_tasks
spec_date: 2026-03-09
slug: shared-tip-height-polling
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Shared Tip Height Polling - Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a focused orchestration and adapter refactor with no schema change, no new integration, and no public API contract change.
- Upstream dependencies (`depends_on`):
  - None.
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The change is limited to receipt polling orchestration and observer port shape.
  - What would trigger switching to Full mode:
    - A new persistence model, async pipeline redesign, or Bitcoin backend replacement.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable; `04_test_plan.md` is included.

## Milestones

- M1:
  - Finalize the shared tip-height optimization spec.
- M2:
  - Refactor polling and observer code to reuse one tip height per network per cycle.

## Tasks (ordered)

1. T-001 - Finalize shared tip-height optimization spec
   - Scope:
     - Capture the intended polling optimization and preserved behavior.
   - Output:
     - `specs/2026-03-09-shared-tip-height-polling/*.md`
   - Linked requirements: FR-001, FR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-09-shared-tip-height-polling" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec lint passes and documents consistently describe the optimization.
     - [ ] Logs/metrics to check (if applicable): none
2. T-002 - Share latest block height across a poll cycle
   - Scope:
     - Extend receipt observer ports so latest block height can be fetched separately and passed into address observation.
     - Update the poller use case to cache latest block height per claimed chain/network pair within one execution.
   - Output:
     - Latest block height is fetched once per claimed chain/network pair and reused for matching address observations.
   - Linked requirements: FR-001, FR-002, NFR-001, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/application/use_cases ./internal/adapters/outbound/blockchain ./internal/adapters/outbound/bitcoin -count=1`
     - [ ] Expected result: polling, routing, and Bitcoin observer tests pass with the shared tip-height flow.
     - [ ] Logs/metrics to check (if applicable): none
3. T-003 - Verify DI and package compile health
   - Scope:
     - Ensure the updated observer interface still wires cleanly through the poller container.
   - Output:
     - Poller DI compiles and tests cleanly with the shared tip-height flow.
   - Linked requirements: FR-002, NFR-002, NFR-005
   - Validation:
     - [ ] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/infrastructure/di -count=1` and `GOCACHE=/tmp/go-build go list ./...`
     - [ ] Expected result: DI tests pass and packages compile cleanly.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002, T-003
- NFR-001 -> T-002
- NFR-002 -> T-002, T-003
- NFR-005 -> T-003
- NFR-006 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore latest block height fetching inside the Bitcoin observer if the shared flow regresses receipt polling.
