---
doc: 03_tasks
spec_date: 2026-03-06
slug: receipt-polling-expiration-guard
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-write-through-receipt-tracking
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
  - Includes schema migration, status model update, and polling lifecycle behavior change.

## Tasks (ordered)

1. T-001 - Add expiration schema and status support

   - Scope:
     - Add migration `000005_receipt_tracking_expiration` with `expires_at` and `failed_expired` status constraint.
   - Linked requirements: FR-001, FR-002
   - Validation:
     - [x] `go test ./internal/domain/value_objects ./internal/domain/entities -count=1`

2. T-002 - Write initial expiry at issue time

   - Scope:
     - Extend repository contract and allocation issue flow to persist `expires_at` when registering tracking.
   - Linked requirements: FR-001
   - Validation:
     - [x] `go test ./internal/application/use_cases ./internal/adapters/outbound/persistence/postgres -count=1`

3. T-003 - Enforce expiry in polling cycle and dynamic extension

   - Scope:
     - Mark expired rows as `failed_expired` before observer call.
     - Extend expiry only when status transitions to `paid_unconfirmed`.
   - Linked requirements: FR-002, FR-003, NFR-001
   - Validation:
     - [x] `go test ./internal/application/use_cases -count=1`

4. T-004 - Add env-driven expiry configuration

   - Scope:
     - Load configurable expiry durations from env in app/poller DI.
     - Inject settings into allocation and polling use cases.
     - Add compose env entries for mainnet and testnet4 overlays.
   - Linked requirements: FR-004, NFR-003
   - Validation:
     - [x] `go test ./internal/infrastructure/di ./cmd/poller -count=1`

5. T-005 - Verify full stack and finalize spec

   - Scope:
     - Run short tests, precommit, and spec-lint.
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, NFR-001, NFR-002, NFR-003, NFR-004
   - Validation:
     - [x] `go test ./... -short -count=1`
     - [x] `bash scripts/precommit-run.sh`
     - [x] `SPEC_DIR="specs/2026-03-06-receipt-polling-expiration-guard" bash scripts/spec-lint.sh`

6. T-006 - Move transition extension rule into domain entity

   - Scope:
     - Move `paid_unconfirmed` transition expiry extension rule from polling use case helper into `PaymentReceiptTracking` entity.
     - Keep use case focused on orchestration and persistence calls.
   - Linked requirements: FR-003, NFR-002
   - Validation:
     - [x] `go test ./internal/domain/entities ./internal/application/use_cases -count=1`

7. T-007 - Split claim lease from poll schedule using `lease_until`

   - Scope:
     - Add `lease_until` column and active lease index in migration.
     - Change `ClaimDue` to claim by `lease_until` and keep `next_poll_at` as schedule.
     - Clear `lease_until` in save paths.
   - Linked requirements: FR-005, NFR-004
   - Validation:
     - [x] `go test ./internal/adapters/outbound/persistence/postgres ./internal/application/use_cases -count=1`

8. T-008 - Consolidate polling use case constructor into one API

   - Scope:
     - Remove `NewRunReceiptPollingCycleUseCaseWithConfig`.
     - Keep `NewRunReceiptPollingCycleUseCase` with explicit config argument and update callers.
   - Linked requirements: NFR-002
   - Validation:
     - [x] `go test ./internal/application/use_cases ./internal/infrastructure/di ./cmd/poller -count=1`

9. T-009 - Remove redundant terminal next-poll scheduling branches

   - Scope:
     - Remove `paid_confirmed` 24h next-poll special-case.
     - Remove `failed_expired` 24h next-poll special-case and use common poll interval scheduling in use-case save paths.
   - Linked requirements: NFR-002
   - Validation:
     - [x] `go test ./internal/application/use_cases -count=1`

10. T-010 - Inline redundant private save helpers in polling use case

- Scope:
  - Remove `savePollingError` and `saveObservation` private methods.
  - Inline save flow into `Execute` while keeping behavior unchanged.
- Linked requirements: NFR-002
- Validation:
  - [x] `go test ./internal/application/use_cases -count=1`

## Traceability

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-003
- FR-003 -> T-003, T-006
- FR-004 -> T-004, T-005
- FR-005 -> T-007, T-005
- NFR-001 -> T-003, T-005
- NFR-002 -> T-005, T-006, T-008, T-009
- NFR-003 -> T-004, T-005
- NFR-004 -> T-007, T-005
