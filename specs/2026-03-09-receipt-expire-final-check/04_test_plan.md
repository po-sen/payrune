---
doc: 04_test_plan
spec_date: 2026-03-09
slug: receipt-expire-final-check
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-05-blockchain-receipt-polling-service
  - 2026-03-09-sticky-paid-unconfirmed-status
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
  - Domain lifecycle expiry behavior.
  - Polling use case ordering for final observation vs expiry.
  - Postgres due-claim query behavior.
- Not covered:
  - Real Esplora network calls.
  - Full end-to-end compose polling runs.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-002, FR-004
  - Steps:
    - Start from a tracking that has `PaidAt != nil` and `expires_at <= now`.
    - Call `ExpireIfDue`.
  - Expected:
    - The tracking does not become `failed_expired`.
- TC-002:
  - Linked requirements: FR-002
  - Steps:
    - Start from a watching or partially paid tracking with `expires_at <= now`.
    - Apply a successful final observation that still leaves it underpaid, then call `ExpireIfDue`.
  - Expected:
    - The tracking becomes `failed_expired`.

### Integration

- TC-101:
  - Linked requirements: FR-001
  - Steps:
    - Exercise `ClaimDue` with a tracking whose `expires_at <= now` but `next_poll_at > now`.
  - Expected:
    - The store does not claim the row.
- TC-102:
  - Linked requirements: FR-002, FR-003, NFR-005
  - Steps:
    - Exercise the polling use case with an expired tracking and a successful observation that
      finds payment.
    - Exercise the same flow with a tip-height or observer failure.
  - Expected:
    - Successful final observation prevents expiry when payment is found.
    - Observer-stage failures increment processing errors and do not produce terminal expiry.

### E2E (if applicable)

- Scenario 1:
  - Not applicable.
- Scenario 2:
  - Not applicable.

## Edge cases and failure modes

- Case:
  - Observation succeeds with zero amount exactly at the expiry boundary.
- Expected behavior:
  - The same due cycle may mark the tracking `failed_expired`.
- Case:
  - Observation fails after the expiry boundary.
- Expected behavior:
  - The tracking records a retryable processing error and remains non-terminal.

## NFR verification

- Performance:
  - Confirm no extra observer call is introduced for a due poll.
- Reliability:
  - Confirm expiry requires a completed final observation path.
- Security:
  - Not applicable.
