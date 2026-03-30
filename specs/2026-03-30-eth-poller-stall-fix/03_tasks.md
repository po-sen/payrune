---
doc: 03_tasks
spec_date: 2026-03-30
slug: eth-poller-stall-fix
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale: This is a bounded bug fix to an existing Ethereum observer and poller log path with no schema changes.
- Upstream dependencies (`depends_on`): `2026-03-20-create2-eth-payment-receiving`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: The change stays inside an existing observer and poller path without altering the architecture or data model.
  - What would trigger switching to Full mode: A generalized Ethereum incremental reconciliation design with new persisted state.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): Each task lists concrete go test / spec lint / precommit checks.

## Milestones

- M1: Add the safe zero-total Ethereum incremental scan path.
- M2: Add poller start logging and regression validation.

## Tasks (ordered)

1. T-001 - Add safe zero-total incremental scanning for Ethereum
   - Scope: Extend observer input with current cumulative totals, teach the Ethereum observer to use `SinceBlockHeight + 1` only for zero-total rows, and keep the full-rescan path for non-zero rows.
   - Output: Sepolia Ethereum rows that already scanned to a prior block while still at zero totals can progress without repeated full-history rescans.
   - Linked requirements: FR-001 / FR-002 / NFR-001 / NFR-002 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/outbound/ethereum ./internal/application/usecases ./internal/adapters/outbound/blockchain`
     - [x] Expected result: Ethereum observer regression tests pass for zero-total incremental scanning and no-op latest-height reuse; existing use case tests still pass with the expanded observer input.
     - [x] Logs/metrics to check (if applicable): N/A
2. T-002 - Emit poll-cycle start logs and run repo validation
   - Scope: Add a start log before each poll cycle and run spec/repo validation for the bug fix.
   - Output: Operators can see that a poll cycle started even before a slow Ethereum observer finishes.
   - Linked requirements: FR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/bootstrap`, `SPEC_DIR="specs/2026-03-30-eth-poller-stall-fix" bash scripts/spec-lint.sh`, `bash scripts/precommit-run.sh`
     - [x] Expected result: Bootstrap tests, spec lint, and repo validation pass with the new observer path and poller start log.
     - [x] Logs/metrics to check (if applicable): N/A

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001
- FR-003 -> T-002
- NFR-001 -> T-001
- NFR-002 -> T-001
- NFR-005 -> T-002
- NFR-006 -> T-001, T-002

## Rollout and rollback

- Feature flag: None
- Migration sequencing: Land the observer and logging fix together; no schema or config migration is required.
- Rollback steps: Revert the observer input expansion, Ethereum zero-total optimization, and poller start log if the behavior regresses.

## Validation evidence

- `go test ./internal/adapters/outbound/ethereum ./internal/application/usecases ./internal/adapters/outbound/blockchain` passed.
- `go test ./internal/bootstrap ./internal/adapters/outbound/ethereum ./internal/application/usecases ./internal/adapters/outbound/blockchain` passed.
- `SPEC_DIR="specs/2026-03-30-eth-poller-stall-fix" bash scripts/spec-lint.sh` passed.
- `bash scripts/precommit-run.sh` passed.
