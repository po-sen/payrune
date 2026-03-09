---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Due poll:
  - A receipt tracking claimed because `next_poll_at <= now`.
- Final observation:
  - The observation attempt performed by the due poll cycle immediately before deciding whether to
    expire the tracking.

## Out-of-scope behaviors

- OOS1:
  - Schema or migration changes.
- OOS2:
  - Changes to sticky paid status semantics after a tracking has already been paid.

## Functional requirements

### FR-001 - Expiry claims must respect `next_poll_at`

- Description:
  - Receipt tracking claims for polling must not bypass the scheduled poll cadence just because the
    payment window has already expired.
- Acceptance criteria:
  - [ ] `ClaimDue` only claims rows whose `next_poll_at <= now`, alongside existing lease,
        status, chain, and network filters.
  - [ ] A tracking with `expires_at <= now` but `next_poll_at > now` is not claimed early.
- Notes:
  - Payment-window expiry remains a domain rule, not a query-side due shortcut.

### FR-002 - Expiry must happen after the final observation

- Description:
  - A due poll cycle must complete the normal observation path before deciding whether the tracking
    should become `failed_expired`.
- Acceptance criteria:
  - [ ] For a due tracking with a successful observation result, status calculation runs before any
        expiry decision.
  - [ ] If that same cycle ends with an unpaid or underpaid tracking, `expires_at <= now`, and the
        tracking can still expire by payment window, the tracking becomes `failed_expired`.
  - [ ] If the successful final observation finds enough payment to move the tracking into a paid
        status, the tracking does not become `failed_expired`.
- Notes:
  - This applies to both fully unpaid and partially paid trackings that are still eligible to expire.

### FR-003 - Observation failures remain retryable after expiry time

- Description:
  - A tracking whose payment window has passed must still avoid terminal expiry if the final
    observation could not be completed.
- Acceptance criteria:
  - [ ] If fetching latest block height fails after `expires_at`, the cycle records a processing
        error and reschedules instead of marking `failed_expired`.
  - [ ] If observing the address fails after `expires_at`, the cycle records a processing error and
        reschedules instead of marking `failed_expired`.
  - [ ] If applying the observation result fails after `expires_at`, the cycle records a processing
        error and reschedules instead of marking `failed_expired`.
- Notes:
  - This preserves the rule that expiry requires a completed final payment check.

### FR-004 - Sticky paid protections must still win

- Description:
  - A tracking that has already reached sticky paid semantics must continue to avoid payment-window
    expiry.
- Acceptance criteria:
  - [ ] `paid_unconfirmed` and `paid_unconfirmed_reverted` trackings do not become `failed_expired`
        through this new ordering.
  - [ ] Existing `CanExpireByPaymentWindow()` behavior remains the guard for payment-window expiry.
- Notes:
  - This requirement keeps the new expiry timing aligned with the current paid-state model.

## Non-functional requirements

- Performance (NFR-001):
  - The change must not add an extra blockchain observation beyond the normal due poll cycle.
- Availability/Reliability (NFR-002):
  - Expiry decisions must be deterministic for a given successful final observation and poll time.
- Security/Privacy (NFR-003):
  - Not applicable.
- Compliance (NFR-004):
  - Not applicable.
- Observability (NFR-005):
  - Existing polling counters (`UpdatedCount`, `ProcessingErrorCount`, `TerminalFailedCount`) must
    remain meaningful under the new ordering.
- Maintainability (NFR-006):
  - Expiry timing should stay explicit in application flow and domain policy, without moving
    business rules into the SQL claim query.

## Dependencies and integrations

- External systems:
  - Existing blockchain observer implementations.
- Internal services:
  - `runReceiptPollingCycleUseCase`
  - `PaymentReceiptTrackingStore`
  - `PaymentReceiptTrackingLifecyclePolicy`
