---
doc: 02_design
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

# Technical Design

## High-level approach

- Summary:
  - Move payment-window expiry from a pre-observation branch to a post-observation branch inside
    the receipt polling cycle.
- Key decisions:
  - `ClaimDue` will stop treating `expires_at` as an alternate due condition.
  - `runReceiptPollingCycleUseCase` will always attempt the due observation path before checking
    whether the tracking should expire.
  - Existing sticky paid guard (`CanExpireByPaymentWindow`) remains the only paid-state expiry
    protection.

## System context

- Components:
  - `PaymentReceiptTrackingStore.ClaimDue`
  - `runReceiptPollingCycleUseCase.Execute`
  - `PaymentReceiptTrackingLifecyclePolicy.ExpireIfDue`
- Interfaces:
  - Existing store and observer ports remain unchanged.

## Key flows

- Flow 1:
  - Claim a due tracking only when `next_poll_at <= now`.
- Flow 2:
  - For each claimed tracking:
    1. Fetch latest block height.
    2. Observe the address.
    3. Apply observation to tracking state.
    4. Evaluate `ExpireIfDue` on the updated tracking.
    5. Save final state and enqueue any status-change notification.
- Flow 3:
  - If any observation-stage step fails, save a polling error and reschedule without expiry.

## Diagrams (optional)

- Mermaid sequence / flow:

  ```mermaid
  flowchart TD
    A[Claim due by next_poll_at] --> B[Fetch latest tip]
    B --> C[Observe address]
    C --> D[Apply observation]
    D --> E{Can expire and expires_at <= now?}
    E -- yes --> F[Mark failed_expired]
    E -- no --> G[Keep observed status]
    B --> H[Processing error + reschedule]
    C --> H
    D --> H
  ```

## Data model

- Entities:
  - `PaymentReceiptTracking` keeps `ExpiresAt`, `PaidAt`, and status fields unchanged.
- Schema changes or migrations:
  - None.
- Consistency and idempotency:
  - The final status for a due poll is still derived from a single observation result plus current
    poll time within one transaction.

## API or contracts

- Endpoints or events:
  - No API contract changes.
- Request/response examples:
  - Not applicable.

## Backward compatibility (optional)

- API compatibility:
  - Preserved.
- Data migration compatibility:
  - Preserved; no migration required.

## Failure modes and resiliency

- Retries/timeouts:
  - Observer errors continue to use the existing retry path via `MarkPollingError` and reschedule.
- Backpressure/limits:
  - No change.
- Degradation strategy:
  - If the final observation cannot complete, prefer retry over terminal expiry.

## Observability

- Logs:
  - No new logs required.
- Metrics:
  - `ProcessingErrorCount` should increase for post-expiry observation failures instead of
    `TerminalFailedCount`.
- Traces:
  - Not applicable.
- Alerts:
  - Existing polling error alerts remain the signal for repeated observer failures.

## Security

- Authentication/authorization:
  - Not applicable.
- Secrets:
  - Not applicable.
- Abuse cases:
  - Not applicable.

## Alternatives considered

- Option A:
  - Keep early expiry in `ClaimDue` and only reorder use-case branches.
- Option B:
  - Add a dedicated final-check scheduler separate from normal polling.
- Why chosen:
  - Removing query-side early expiry and reusing the normal due poll keeps the behavior explicit,
    minimal, and aligned with current architecture.

## Risks

- Risk:
  - Overdue unpaid trackings may wait until the next scheduled poll before becoming
    `failed_expired`.
- Mitigation:
  - This delay is intentional and matches the new rule that expiry only happens at the scheduled
    final check.
