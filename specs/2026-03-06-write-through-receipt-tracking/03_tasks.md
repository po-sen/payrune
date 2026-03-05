---
doc: 03_tasks
spec_date: 2026-03-06
slug: write-through-receipt-tracking
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

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - Includes migration update and async poller flow change.
- Upstream dependencies (`depends_on`): []
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`

## Milestones

- M1:
  - Port + repository write-through registration.
- M2:
  - Allocation/poller use case refactor and tests.
- M3:
  - Migration backfill + verification + precommit.
- M4:
  - Network-specific confirmations env wiring and validation.

## Tasks (ordered)

1. T-001 - Add write-through registration repository contract and Postgres implementation

   - Scope:
     - Replace bulk `RegisterMissingIssued` with single-row `RegisterIssuedAllocation`.
   - Output:
     - Updated application port and postgres adapter SQL method.
   - Linked requirements: FR-001, FR-004, NFR-006
   - Validation:
     - [x] `go test ./internal/adapters/outbound/persistence/postgres -count=1`
     - [x] Build passes without interface mismatch errors.

2. T-002 - Move registration into allocation issue transaction and remove poller pre-register step

   - Scope:
     - Update allocation use case to register tracking in same UoW.
     - Update poller cycle to claim directly and keep output compatibility.
     - Remove dead poller confirmation/register output fields from DTO/config/runtime.
   - Output:
     - Refactored use cases and updated unit tests.
   - Linked requirements: FR-001, FR-002, FR-005, NFR-001, NFR-002, NFR-005
   - Validation:
     - [x] `go test ./internal/application/use_cases -count=1`
     - [x] Poller tests prove claim/observe flow still works with cleaned output/config fields.

3. T-003 - Add one-time backfill migration for pre-existing issued allocations

   - Scope:
     - Add missing-row backfill SQL in a new migration file (`000004_backfill_payment_receipt_trackings.up.sql`).
   - Output:
     - Migration script updated with idempotent insert-select backfill.
   - Linked requirements: FR-003, NFR-002
   - Validation:
     - [x] Review SQL conditions: issued-only + non-null network/address + conflict-safe.

4. T-004 - Run full verification and spec completion

   - Scope:
     - Run short/full checks and mark spec state based on results.
   - Output:
     - Validation evidence and finalized spec status.
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-005, NFR-006
   - Validation:
     - [x] `go test ./... -short -count=1`
     - [x] `bash scripts/precommit-run.sh`
     - [x] `SPEC_DIR="specs/2026-03-06-write-through-receipt-tracking" bash scripts/spec-lint.sh`

5. T-005 - Add per-network confirmations env for issue-time registration

   - Scope:
     - Add DI config parsing for `BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS` and `BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS`.
     - Wire network-specific defaults into allocation issue use case.
     - Add separate env declarations in mainnet/testnet4 compose overrides.
   - Output:
     - Allocation issue path writes `required_confirmations` by target network env.
   - Linked requirements: FR-006, NFR-002, NFR-006
   - Validation:
     - [x] `go test ./internal/application/use_cases ./internal/infrastructure/di ./cmd/poller -count=1`
     - [x] `go test ./... -short -count=1`
     - [x] `bash scripts/precommit-run.sh`

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-002
- FR-003 -> T-003
- FR-004 -> T-001
- FR-005 -> T-002
- FR-006 -> T-005
- NFR-001 -> T-002
- NFR-002 -> T-002, T-003
- NFR-005 -> T-004
- NFR-006 -> T-001, T-004, T-005

## Rollout and rollback

- Migration sequencing:
  - Run migration before deploying new poller/application binaries.
- Rollback steps:
  - Revert binaries first; data backfill is additive and conflict-safe.
