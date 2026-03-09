---
doc: 03_tasks
spec_date: 2026-03-09
slug: sticky-paid-unconfirmed-status
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-polling-expiration-guard
  - 2026-03-08-payment-address-status-api
  - 2026-03-06-receipt-webhook-delivery
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
  - This change alters the domain state machine, adds a new persisted status value, changes API-visible enums, removes an env/config path, and affects async webhook status delivery semantics.
- Upstream dependencies (`depends_on`):
  - `2026-03-06-receipt-polling-expiration-guard`
  - `2026-03-08-payment-address-status-api`
  - `2026-03-06-receipt-webhook-delivery`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - Not applicable.
  - What would trigger switching to Full mode:
    - Already Full.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not applicable; `04_test_plan.md` is included.

## Milestones

- M1:
  - Finalize sticky paid semantics and the new regression status design.
- M2:
  - Implement domain, migration, API contract, and config changes, including removal of the unused double-spend status and conflict field.
- M3:
  - Verify polling, status API, webhook compatibility, and schema constraints.

## Tasks (ordered)

1. T-001 - Finalize sticky paid-unconfirmed spec
   - Scope:
     - Capture the new status, sticky-paid invariant, expiry rule change, and config removal.
   - Output:
     - `specs/2026-03-09-sticky-paid-unconfirmed-status/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-002, NFR-004, NFR-005
   - Validation:
   - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-09-sticky-paid-unconfirmed-status" bash scripts/spec-lint.sh`
   - [x] Expected result: spec lint passes and all docs describe the same new status and sticky expiry semantics.
   - [x] Logs/metrics to check (if applicable): none
2. T-002 - Implement sticky paid status model and persistence updates
   - Scope:
     - Add `paid_unconfirmed_reverted`, update state transitions, remove paid-unconfirmed expiry extension logic, remove `double_spend_suspected`, remove `ConflictTotalMinor`, and add the required PostgreSQL migration/constraint updates without rewriting historical migrations.
   - Output:
     - Domain and persistence layers enforce sticky paid semantics and the new status.
   - Linked requirements: FR-001, FR-002, FR-003, FR-005, FR-006, NFR-001, NFR-002, NFR-004, NFR-006, NFR-007, NFR-008
   - Validation:
   - [x] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/domain/... ./internal/application/use_cases ./internal/adapters/outbound/persistence/postgres -count=1`
   - [x] Expected result: domain, polling, and postgres tests pass with the new status and non-expiring fully-paid behavior, and historical migrations stay untouched.
   - [x] Logs/metrics to check (if applicable): none
3. T-003 - Update API/webhook-visible contracts and poller config
   - Scope:
     - Update OpenAPI/status API enum handling, webhook-related status parsing/tests, remove obsolete `PAYMENT_RECEIPT_PAID_UNCONFIRMED_EXPIRY_EXTENSION` env exposure from DI/Compose, rename `RECEIPT_POLL_INTERVAL` to `POLL_RESCHEDULE_INTERVAL`, and remove `double_spend_suspected` plus `ConflictTotalMinor` from outward-facing contracts.
   - Output:
     - External status consumers can see the new status, and obsolete config paths are gone.
   - Linked requirements: FR-004, FR-005, FR-006, NFR-003, NFR-005, NFR-006, NFR-007
   - Validation:
   - [x] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/adapters/inbound/http/controllers ./internal/infrastructure/di ./internal/adapters/outbound/webhook -count=1`, `SPEC_DIR="specs/2026-03-09-sticky-paid-unconfirmed-status" bash scripts/spec-lint.sh`, and `GOCACHE=/tmp/go-build go list ./...`
   - [x] Expected result: controller/DI/webhook tests pass, spec stays valid, and all packages compile.
   - [x] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002
- FR-003 -> T-001, T-002
- FR-004 -> T-001, T-003
- FR-005 -> T-001, T-002, T-003
- FR-006 -> T-001, T-002, T-003
- NFR-001 -> T-002
- NFR-002 -> T-001, T-002
- NFR-003 -> T-003
- NFR-004 -> T-001, T-002
- NFR-005 -> T-003
- NFR-006 -> T-002, T-003
- NFR-007 -> T-002, T-003
- NFR-008 -> T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Apply the new receipt-status migration before running poller/application binaries that may persist or read `paid_unconfirmed_reverted`.
- Rollback steps:
  - Revert binaries and migration together; do not leave rows with the new status value or the simplified status set while older binaries are active.
