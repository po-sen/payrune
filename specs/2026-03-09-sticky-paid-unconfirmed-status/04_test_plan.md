---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Domain status parsing and sticky paid transitions.
  - Poller expiry behavior after full payment observation.
  - PostgreSQL receipt status parsing/constraint compatibility.
  - Status API/OpenAPI enum handling and webhook-related status propagation.
  - Removal of obsolete poller env config.
  - Removal of the unused `double_spend_suspected` status from all outward-facing contracts.
  - Removal of `ConflictTotalMinor` from contracts and persisted schema.
- Not covered:
  - Manual integration against a live Bitcoin network.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002, NFR-004
  - Steps:
    - Apply a full unconfirmed observation to a receipt, then apply a later sub-threshold observation.
  - Expected:
    - Status becomes `paid_unconfirmed_reverted`, not `watching` or `partially_paid`.
- TC-002:
  - Linked requirements: FR-002, FR-003, NFR-001
  - Steps:
    - Create a receipt with `PaidAt != nil` and expired `expires_at`; call lifecycle expiry logic.
  - Expected:
    - Expiry logic does not mark the receipt as `failed_expired`.
- TC-003:
  - Linked requirements: FR-004, NFR-005
  - Steps:
    - Construct poller lifecycle policy/config without `PAYMENT_RECEIPT_PAID_UNCONFIRMED_EXPIRY_EXTENSION`, using `POLL_RESCHEDULE_INTERVAL`.
  - Expected:
    - Poller config loads successfully, `POLL_RESCHEDULE_INTERVAL` is honored, and no obsolete env is required.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, FR-003, FR-005, FR-006, NFR-002, NFR-006, NFR-007
  - Steps:
    - Run domain/use-case/postgres tests covering receipt polling and status persistence.
  - Expected:
    - New status persists and polls cleanly, and fully-paid rows no longer expire.
- TC-102:
  - Linked requirements: FR-004, FR-005, FR-006, NFR-003, NFR-005, NFR-006, NFR-007
  - Steps:
    - Run controller, DI, and webhook-related tests plus package compilation.
  - Expected:
    - Status API and webhook-related code accept the new status, and removed config paths do not break startup.

### E2E (if applicable)

- Scenario 1:
  - Simulate `watching -> paid_unconfirmed -> paid_unconfirmed_reverted -> paid_unconfirmed -> paid_confirmed` through polling fixtures.
- Scenario 2:
  - Simulate a fully-paid observation crossing the original expiry deadline and confirm it does not become `failed_expired`.

## Edge cases and failure modes

- Case:
  - A receipt never reaches the expected amount.
- Expected behavior:
  - Existing `watching` / `partially_paid` expiry behavior remains unchanged.

## NFR verification

- Performance:
  - Confirm the change does not add extra observer calls or DB round-trips in one poll cycle.
- Reliability:
  - Confirm fully-paid receipts cannot silently fall back into unpaid statuses.
- Security:
  - Confirm no new secret/config path is introduced and the obsolete env path is removed cleanly.
