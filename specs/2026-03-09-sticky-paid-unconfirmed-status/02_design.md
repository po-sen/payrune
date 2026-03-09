---
doc: 02_design
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

# Technical Design

## High-level approach

- Summary:
  - Add a new externally visible receipt status `paid_unconfirmed_reverted`.
  - Make the "full payment observed" condition sticky by deriving unpaid-state eligibility from `PaidAt == nil`.
  - Replace transition-based expiry extension with a simpler invariant: payment-window expiry applies only before full payment is ever observed.
  - Remove `double_spend_suspected` entirely from the status model.
  - Remove `ConflictTotalMinor` entirely from contracts and persistence.
- Key decisions:
  - Use `PaidAt` as the sticky invariant instead of adding another persistence flag.
  - Keep regression as an explicit status instead of overloading `watching`/`partially_paid`.
  - Remove the `PAYMENT_RECEIPT_PAID_UNCONFIRMED_EXPIRY_EXTENSION` config path entirely.
  - Rename the poller reschedule env to `POLL_RESCHEDULE_INTERVAL` and align internal naming to `RescheduleInterval`.
  - Remove `ConflictTotalMinor` instead of keeping an always-zero compatibility field.
  - Keep historical migrations immutable and express schema evolution only in `000009_sticky_paid_unconfirmed_status`.

## System context

- Components:
  - Domain:
    - `PaymentReceiptStatus`
    - `PaymentReceiptTracking`
    - `PaymentReceiptTrackingLifecyclePolicy`
  - Application:
    - `RunReceiptPollingCycleUseCase`
    - `GetPaymentAddressStatusUseCase`
    - Webhook dispatch use case consumes status events without structural change
  - Adapters:
    - Postgres receipt tracking store / status finder
    - HTTP controller + OpenAPI
    - Webhook notifier
- Interfaces:
  - Status strings remain plain strings at API/webhook boundaries.
  - No new inbound endpoint or outbound port type is introduced.
  - Poller config distinguishes worker tick cadence (`POLL_TICK_INTERVAL`) from per-row rescheduling (`POLL_RESCHEDULE_INTERVAL`).

## Key flows

- Flow 1:
  - `watching` or `partially_paid` receives an observation with `ObservedTotalMinor >= ExpectedAmountMinor` and insufficient confirmations.
  - `PaidAt` is set if missing.
  - Status becomes `paid_unconfirmed`.
  - Expiry is no longer relevant once `PaidAt` exists.
- Flow 2:
  - A later observation for a receipt with `PaidAt != nil` drops below `ExpectedAmountMinor`.
  - Status becomes `paid_unconfirmed_reverted`.
  - Polling continues; the row does not expire by payment window.
  - A later sufficient unconfirmed observation returns to `paid_unconfirmed`; a confirmed observation becomes `paid_confirmed`.
- Flow 3:
  - A receipt with `PaidAt == nil` still uses existing expiry rules and may become `failed_expired`.

## Diagrams (optional)

- Mermaid sequence / flow:
  - `watching|partially_paid -> paid_unconfirmed -> paid_unconfirmed_reverted -> paid_unconfirmed|paid_confirmed`

## Data model

- Entities:
  - `PaymentReceiptStatus` gains `paid_unconfirmed_reverted`.
  - `PaymentReceiptTracking.ApplyObservation` uses `PaidAt` and current observation together to decide whether unpaid states are still allowed, without a special conflict-only status.
  - `PaymentReceiptTrackingLifecyclePolicy.ExpireIfDue` ignores payment-window expiry when `PaidAt != nil`.
- Schema changes or migrations:
  - Leave `000003`, `000005`, and `000006` unchanged as historical records.
  - In `000009_sticky_paid_unconfirmed_status.up.sql`, update receipt-tracking constraints/indexes to remove `double_spend_suspected`, add `paid_unconfirmed_reverted`, and drop `conflict_total_minor` from `payment_receipt_trackings` and `payment_receipt_status_notifications`.
  - In `000009_sticky_paid_unconfirmed_status.down.sql`, re-add `conflict_total_minor` and restore the pre-`000009` status constraints.
  - No row backfill is required.
- Consistency and idempotency:
  - Status remains a deterministic function of current observation plus sticky historical marker `PaidAt`.
  - Repeated observations with the same inputs keep the same status and do not re-trigger expiry behavior.

## API or contracts

- Endpoints or events:
  - `GET /v1/chains/{chain}/payment-addresses/{paymentAddressId}` may return `paid_unconfirmed_reverted`.
  - Existing webhook status-changed payload may emit `currentStatus` / `previousStatus = paid_unconfirmed_reverted`.
- Request/response examples:
  - Status API example: `paymentStatus: paid_unconfirmed_reverted`
  - Webhook payload example: `"currentStatus": "paid_unconfirmed_reverted"`
  - Neither contract includes `conflictTotalMinor`.

## Backward compatibility (optional)

- API compatibility:
  - Breaking change for consumers that assume the old finite status set.
- Data migration compatibility:
  - Existing rows remain valid; only future observations may produce the new status.

## Failure modes and resiliency

- Retries/timeouts:
  - Observer failures and polling errors keep existing behavior.
- Backpressure/limits:
  - Polling cadence is unchanged.
- Degradation strategy:
  - If a previously paid unconfirmed transaction disappears temporarily, the receipt remains in an explicit pollable regression state instead of silently looking unpaid again.

## Observability

- Logs:
  - Existing poll cycle logs remain sufficient; status transition counts naturally include the new state.
- Metrics:
  - No new metrics are required, but the new status should appear in any status-based dashboards.
- Traces:
  - No tracing changes required.
- Alerts:
  - None in scope.

## Security

- Authentication/authorization:
  - No auth changes.
- Secrets:
  - One poller env is removed; no new secret is introduced.
- Abuse cases:
  - A malicious or unstable payment that oscillates below/above threshold will now oscillate between explicit paid states rather than disguising itself as unpaid.

## Alternatives considered

- Option A:
  - Keep `paid_unconfirmed` sticky forever even if current observation regresses.
- Option B:
  - Revert to `watching`/`partially_paid` and re-arm expiry.
- Option C:
  - Keep `double_spend_suspected` as a reserved status despite the current observer never producing actionable conflict signals.
- Why chosen:
  - A dedicated regression status preserves current-observation truth while keeping the sticky history that full payment was once observed, and removing the unused double-spend status keeps the model aligned with real observer capability.

## Risks

- Risk:
  - External consumers may not recognize the new status.
- Mitigation:
  - Update OpenAPI, parser enums, and webhook-related tests in the same change.
- Risk:
  - Existing expiry-related config becomes dead and misleading.
- Mitigation:
  - Remove env parsing and Compose exposure in the same change.
