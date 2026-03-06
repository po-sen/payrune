---
doc: 02_design
spec_date: 2026-03-06
slug: receipt-webhook-delivery
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-status-change-notification
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Technical Design

## High-level approach

- Extend the existing outbox table with delivery lifecycle columns.
- Add a dedicated dispatch use case that claims pending rows, posts them to a fixed webhook endpoint, and persists delivery results.
- Implement a webhook notifier adapter that signs raw JSON with HMAC-SHA256 and treats only `2xx` as success.
- Run the dispatcher in its own command/bootstrap/DI path, separate from blockchain pollers.
- Make the dispatcher a default compose service, with required env interpolation in `compose.yaml`, committed fake values in `compose.test.env`, and test-only fake receiver wiring in `compose.test.yaml`.

## Key flows

- Flow 1: claim due notifications
  - Worker tick -> claim pending rows by `next_attempt_at` and expired/empty `lease_until` -> set new lease -> return claimed notifications.
- Flow 2: successful delivery
  - Build payload -> POST fixed webhook URL -> receive `2xx` -> persist `delivery_status = 'sent'`, `delivered_at = now`, clear `lease_until`, clear `last_error`.
- Flow 3: retryable failure
  - POST returns non-`2xx` or transport error -> increment attempts -> set `last_error` -> if under max attempts, keep `delivery_status = 'pending'`, schedule `next_attempt_at = now + retry_delay`, clear `lease_until`.
- Flow 4: terminal failure
  - Attempt count reaches `max_attempts` -> set `delivery_status = 'failed'`, clear `lease_until`, preserve `last_error`.

## Data model

- Migration updates `payment_receipt_status_notifications`:
  - add `delivery_attempts INTEGER NOT NULL DEFAULT 0 CHECK (delivery_attempts >= 0)`
  - add `next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
  - add `lease_until TIMESTAMPTZ`
  - add `last_error TEXT`
  - add `delivered_at TIMESTAMPTZ`
  - add `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes:
  - partial due index on pending rows by `next_attempt_at`
  - partial lease index on pending rows by `lease_until`

## Contracts

- New inbound port:
  - `RunReceiptWebhookDispatchCycleUseCase`
- New outbound port:
  - `PaymentReceiptStatusNotifier`
    - `NotifyStatusChanged(ctx, input) error`
- Extended outbound repository:
  - `ClaimPending`
  - `MarkSent`
  - `MarkRetryScheduled`
  - `MarkFailed`

## Payload contract

- Event metadata:
  - `event_type = "payment_receipt.status_changed"`
  - `event_version = 1`
  - `notification_id`
- Business fields:
  - `payment_address_id`
  - `customer_reference`
  - `previous_status`
  - `current_status`
  - `observed_total_minor`
  - `confirmed_total_minor`
  - `unconfirmed_total_minor`
  - `conflict_total_minor`
  - `status_changed_at`
- Headers:
  - `Content-Type: application/json`
  - `X-Payrune-Event: payment_receipt.status_changed`
  - `X-Payrune-Event-Version: 1`
  - `X-Payrune-Notification-ID: <id>`
  - `X-Payrune-Signature-256: sha256=<hex-hmac>`

## Failure modes and resiliency

- Active lease prevents the same row from being dispatched concurrently.
- Worker crash during in-flight delivery is recovered by lease expiry.
- Non-`2xx` HTTP responses and transport errors are handled as delivery failures.
- Secret and endpoint remain runtime config only and are never stored in DB rows.
- Local test environment may use a self-signed HTTPS receiver behind an explicit insecure-skip-verify test-only flag.

## Observability

- Worker logs cycle summary with claimed, sent, retried, and failed counts.
- Database retains attempt count and last error for operational inspection.

## Configuration contract

- `RECEIPT_WEBHOOK_DISPATCH_INTERVAL`
- `RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE`
- `RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL`
- `RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS`
- `RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY`
- `PAYMENT_RECEIPT_WEBHOOK_URL`
- `PAYMENT_RECEIPT_WEBHOOK_SECRET`
- `PAYMENT_RECEIPT_WEBHOOK_TIMEOUT`
- `PAYMENT_RECEIPT_WEBHOOK_INSECURE_SKIP_VERIFY` (test-only support for local fake receiver)
- Local compose test runs load the fake webhook env via `--env-file deployments/compose/compose.test.env`.
