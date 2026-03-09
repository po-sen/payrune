---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Sticky paid invariant:
  - Once a receipt has observed the full expected amount and `PaidAt` is set, it must never re-enter unpaid statuses.
- Regression status:
  - A status representing that a previously fully-paid unconfirmed observation is no longer currently observed as fully paid.

## Out-of-scope behaviors

- OOS1:
  - Auto-refund or manual-review workflow for reverted payments.
- OOS2:
  - Introducing new webhook event types or versions.

## Functional requirements

### FR-001 - Add explicit regression status

- Description:
  - The domain and persistence layer must support a new status for receipts that previously reached `paid_unconfirmed` semantics but later lose full-payment observation.
- Acceptance criteria:
  - [ ] `PaymentReceiptStatus` includes `paid_unconfirmed_reverted`.
  - [ ] PostgreSQL receipt status constraints accept `paid_unconfirmed_reverted`.
  - [ ] Active polling status sets include `paid_unconfirmed_reverted`.
  - [ ] Status parsing in read/query adapters accepts `paid_unconfirmed_reverted`.
- Notes:
  - This status is expected to be externally visible through status APIs and webhook payloads as a normal status string.

### FR-002 - Make full-payment observation sticky

- Description:
  - After `PaidAt` is set, receipt status may no longer return to `watching` or `partially_paid`.
- Acceptance criteria:
  - [ ] If a receipt has `PaidAt != nil` and the latest observation is below the expected amount, status becomes `paid_unconfirmed_reverted`.
  - [ ] If a receipt has `PaidAt != nil` and a later observation again reaches the expected amount unconfirmed, status becomes `paid_unconfirmed`.
  - [ ] If a receipt has `PaidAt != nil` and a later observation reaches confirmed amount, status becomes `paid_confirmed`.
- Notes:
  - Sticky behavior applies only after the full expected amount has been observed at least once.

### FR-003 - Stop expiring once fully paid was observed

- Description:
  - Payment-window expiry must no longer transition receipts to `failed_expired` after a full payment has ever been observed.
- Acceptance criteria:
  - [ ] Poller expiry logic ignores receipts with `PaidAt != nil`.
  - [ ] `paid_unconfirmed` does not rely on expiry extension windows.
  - [ ] `paid_unconfirmed_reverted` also remains non-expiring through the payment-window expiry path.
  - [ ] Receipts that never reached full observed payment continue to use the existing expiry behavior.
- Notes:
  - This replaces the previous "extend expiry on transition to `paid_unconfirmed`" rule.

### FR-004 - Remove obsolete config and update contracts

- Description:
  - With sticky fully-paid semantics, obsolete config and public status contracts must be updated consistently.
- Acceptance criteria:
  - [ ] Poller DI no longer reads `PAYMENT_RECEIPT_PAID_UNCONFIRMED_EXPIRY_EXTENSION`.
  - [ ] Compose poller env blocks no longer expose `PAYMENT_RECEIPT_PAID_UNCONFIRMED_EXPIRY_EXTENSION`.
  - [ ] Poller config reads `POLL_RESCHEDULE_INTERVAL` instead of `RECEIPT_POLL_INTERVAL`.
  - [ ] Compose poller env blocks expose `POLL_RESCHEDULE_INTERVAL`.
  - [ ] OpenAPI payment status enum includes `paid_unconfirmed_reverted`.
  - [ ] Existing webhook delivery flow can emit the new status value without requiring a new event shape.
- Notes:
  - This is a pre-`0.1.0` breaking config/status change.

### FR-005 - Remove unused double-spend status

- Description:
  - The receipt status model must no longer contain `double_spend_suspected`; conflict totals remain informational only.
- Acceptance criteria:
  - [ ] `PaymentReceiptStatus` no longer includes `double_spend_suspected`.
  - [ ] Domain status decisions ignore `ConflictTotalMinor` when choosing the receipt status.
  - [ ] PostgreSQL receipt-status constraints and notification status constraints no longer include `double_spend_suspected`.
  - [ ] OpenAPI and webhook/status parsers no longer advertise `double_spend_suspected`.
- Notes:
  - This intentionally simplifies the model because the current observer never produces actionable conflict signals.

### FR-006 - Remove conflict total field

- Description:
  - The receipt model and outward-facing contracts must no longer carry `ConflictTotalMinor`.
- Acceptance criteria:
  - [ ] `PaymentReceiptObservation`, `PaymentReceiptTracking`, and status-changed events no longer contain `ConflictTotalMinor`.
  - [ ] Status API DTOs and read models no longer expose `conflictTotalMinor`.
  - [ ] Webhook notifier input/payload and outbox message models no longer contain `ConflictTotalMinor`.
  - [ ] PostgreSQL receipt tracking and notification schemas no longer store `conflict_total_minor`.
- Notes:
  - This is a pre-`0.1.0` schema and contract cleanup aligned with the current observer capability.

## Non-functional requirements

- Reliability (NFR-001):
  - Status transitions after full payment observation must be deterministic and must not depend on adapter-specific behavior.
- Correctness (NFR-002):
  - Sticky paid logic must live in domain or lifecycle policy code, not in SQL queries or controller branches.
- Compatibility (NFR-003):
  - All outward-facing status enums and parsers must remain consistent with the stored receipt status set.
- Maintainability (NFR-004):
  - The new state machine should be explainable by one invariant: "`PaidAt != nil` receipts never re-enter unpaid states."
- Operability (NFR-005):
  - Removing the obsolete expiry-extension env must not leave dead config paths in Compose or DI.
- Simplicity (NFR-006):
  - No status value should remain in the model if the current production observation pipeline cannot meaningfully produce it.
- Minimality (NFR-007):
  - No field should remain in runtime contracts or schema if the current production observation pipeline cannot meaningfully produce it.
- Migration hygiene (NFR-008):
  - Historical migrations must remain immutable; schema evolution for this spec must be expressed only through the new forward migration added by this change.

## Dependencies and integrations

- External systems:
  - PostgreSQL migration for receipt status constraint/index updates.
  - Swagger/OpenAPI documentation consumers.
  - Webhook consumers receiving status strings.
- Internal services:
  - Receipt polling lifecycle policy and entity state transitions.
  - Payment address status API read model.
  - Receipt webhook notification pipeline.
