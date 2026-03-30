---
doc: 03_tasks
spec_date: 2026-03-30
slug: eth-balance-receipt-observer
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
  - 2026-03-30-eth-poller-stall-fix
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - This change replaces ETH block scanning with balance-snapshot observation and documents the
    resulting semantics tradeoff.
- Upstream dependencies (`depends_on`):
  - `2026-03-20-create2-eth-payment-receiving`
  - `2026-03-30-eth-poller-stall-fix`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: N/A
  - What would trigger switching to Full mode: N/A
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): N/A

## Milestones

- M1:
  - Replace ETH block scanning with balance-snapshot observation.
- M2:
  - Align tests and spec with the observer-only design.
- M3:
  - Validate repo health.

## Tasks (ordered)

1. T-001 - Replace ETH block scan with balance snapshots
   - Scope:
     - Keep ETH polling in the observer path and replace block scanning with exact
       `eth_getBalance` snapshots.
   - Output:
     - ETH observer is O(1) per row and no longer scans blocks.
   - Linked requirements: FR-001 / FR-002 / FR-003 / NFR-001 / NFR-002 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command):
           `go test ./internal/adapters/outbound/ethereum ./internal/adapters/outbound/blockchain`
     - [x] Expected result:
           ETH observer tests pass for snapshot observation, insufficient confirmations, and
           inconsistent snapshots.
     - [x] Logs/metrics to check (if applicable):
           N/A
2. T-002 - Keep allocation and persistence unchanged
   - Scope:
     - Avoid schema and allocation-flow coupling for ETH-specific observation semantics.
   - Output:
     - Allocation and receipt tracking persistence keep their prior responsibilities.
   - Linked requirements: FR-002 / FR-004 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command):
           `go test ./internal/application/usecases ./internal/bootstrap`
     - [x] Expected result:
           Existing allocation/bootstrap tests still pass without ETH baseline wiring.
     - [x] Logs/metrics to check (if applicable):
           N/A
3. T-003 - Update spec and tests to reflect ETH snapshot semantics
   - Scope:
     - Document that ETH uses snapshot semantics under standard RPC and keep that limitation
       explicit.
   - Output:
     - Code and spec agree on the ETH observer contract.
   - Linked requirements: FR-004 / NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command):
           `SPEC_DIR="specs/2026-03-30-eth-balance-receipt-observer" bash scripts/spec-lint.sh`
     - [x] Expected result:
           Spec matches the implemented observer-only design.
     - [x] Logs/metrics to check (if applicable):
           N/A
4. T-004 - Run spec and repo validation
   - Scope:
     - Sync spec status and run repo validation.
   - Output:
     - Implementation and spec remain aligned and shippable.
   - Linked requirements: NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command):
           `SPEC_DIR="specs/2026-03-30-eth-balance-receipt-observer" bash scripts/spec-lint.sh`,
           `bash scripts/precommit-run.sh`
     - [x] Expected result:
           Spec lint and repo validation pass.
     - [x] Logs/metrics to check (if applicable):
           Existing poll-cycle logs remain available.

## Validation evidence

- `SPEC_DIR="specs/2026-03-30-eth-balance-receipt-observer" bash scripts/spec-lint.sh`
- `go test ./internal/application/usecases ./internal/adapters/outbound/ethereum ./internal/bootstrap`
- `bash scripts/precommit-run.sh`

## Traceability

- FR-001 -> T-001
- FR-002 -> T-001, T-002
- FR-003 -> T-001
- FR-004 -> T-002, T-003
- NFR-001 -> T-001
- NFR-002 -> T-001
- NFR-003 -> T-004
- NFR-005 -> T-003, T-004
- NFR-006 -> T-001, T-002, T-003, T-004

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Revert the observer/test changes together if snapshot semantics prove insufficient.
