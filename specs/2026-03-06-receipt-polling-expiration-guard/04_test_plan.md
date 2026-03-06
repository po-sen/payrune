---
doc: 04_test_plan
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

# Test Plan

## Unit

- TC-001:

  - Linked requirements: FR-002
  - Steps:
    - Validate status parser accepts `failed_expired`.
  - Expected:
    - Status parsing succeeds.

- TC-002:

  - Linked requirements: FR-002
  - Steps:
    - Polling cycle receives expired tracking row.
  - Expected:
    - Observer not called; row saved as `failed_expired`.

- TC-003:

  - Linked requirements: FR-003
  - Steps:
    - Polling cycle transitions into `paid_unconfirmed`, then runs again with unchanged `paid_unconfirmed`.
  - Expected:
    - `expires_at` extends on transition only and remains unchanged when status does not change.

- TC-006:

  - Linked requirements: FR-003, NFR-002
  - Steps:
    - Unit test entity transition helper with different prior statuses.
  - Expected:
    - Extension occurs only for transition to `paid_unconfirmed`, with unchanged behavior for non-transition cases.

- TC-004:

  - Linked requirements: FR-004
  - Steps:
    - Set invalid duration env for app/poller expiry configs.
  - Expected:
    - DI/env loader returns explicit validation error.

- TC-005:
  - Linked requirements: FR-004
  - Steps:
    - Set valid custom duration env values.
  - Expected:
    - Use cases apply configured values rather than defaults.

## Integration

- TC-101:

  - Linked requirements: FR-001, FR-005
  - Steps:
    - Verify repository scan/claim paths include `expires_at` mapping and `lease_until` filtering.
  - Expected:
    - Entities include persisted expiry timestamps and claim excludes active leases.

- TC-102:

  - Linked requirements: FR-001
  - Steps:
    - Run migration checks.
  - Expected:
    - Existing rows are backfilled; new constraint/index is valid.

- TC-103:

  - Linked requirements: FR-004, NFR-003
  - Steps:
    - Verify compose mainnet and testnet4 overlays define expiry env values for app and poller services.
  - Expected:
    - Both overlays expose explicit env entries and can be tuned independently.

- TC-104:

  - Linked requirements: FR-005, NFR-004
  - Steps:
    - Confirm `ClaimDue` sets `lease_until` only, and save paths clear `lease_until` while writing final `next_poll_at`.
  - Expected:
    - Claim lock and schedule semantics are independent and consistent.

- TC-105:
  - Linked requirements: NFR-002
  - Steps:
    - Build and run tests against `RunReceiptPollingCycleUseCase` constructor call sites after removing `WithConfig` constructor.
  - Expected:
    - All callers use one constructor API and behavior remains unchanged.
