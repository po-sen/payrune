---
doc: 03_tasks
spec_date: 2026-03-09
slug: mempool-compose-defaults
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

# Mempool Compose Defaults - Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a small configuration cleanup with no schema change, no new integration type, and no API behavior change.
- Upstream dependencies (`depends_on`):
  - None.
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The change only touches default config values and related tests.
  - What would trigger switching to Full mode:
    - A broader provider-selection architecture or failover design.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable; `04_test_plan.md` is included.

## Milestones

- M1:
  - Finalize the provider-default unification spec.
- M2:
  - Update compose defaults and matching tests.

## Tasks (ordered)

1. T-001 - Finalize the mempool default unification spec
   - Scope:
     - Capture the intended default-provider change and validation scope.
   - Output:
     - `specs/2026-03-09-mempool-compose-defaults/*.md`
   - Linked requirements: FR-001, FR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-09-mempool-compose-defaults" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec lint passes and documents consistently describe the config change.
     - [ ] Logs/metrics to check (if applicable): none
2. T-002 - Unify compose defaults on mempool.space
   - Scope:
     - Change the mainnet compose fallback endpoint to `https://mempool.space/api`.
   - Output:
     - Mainnet and testnet4 compose defaults both point to mempool.space.
   - Linked requirements: FR-001, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): inspect the compose yaml defaults and run `GOCACHE=/tmp/go-build go test ./internal/infrastructure/di -count=1`
     - [ ] Expected result: compose defaults are consistent and DI tests pass.
     - [ ] Logs/metrics to check (if applicable): none
3. T-003 - Align tests with the unified defaults
   - Scope:
     - Update poller container tests to use mempool mainnet examples.
   - Output:
     - Tests no longer assume blockstream.info as the default public mainnet endpoint.
   - Linked requirements: FR-002, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/infrastructure/di -count=1` and `GOCACHE=/tmp/go-build go list ./...`
     - [ ] Expected result: DI tests and package compilation pass.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-003
- NFR-002 -> T-002, T-003
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore the previous `blockstream.info` mainnet compose default if the team prefers mixed public defaults.
