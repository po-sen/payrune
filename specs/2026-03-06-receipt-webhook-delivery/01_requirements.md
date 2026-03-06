---
doc: 01_requirements
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

# Requirements

## Functional requirements

### FR-001 - Add delivery lifecycle fields to receipt notification outbox

- Description:
  - The notification outbox schema must track delivery attempts, scheduling, claim lease, errors, and successful delivery time.
- Acceptance criteria:
  - [x] Schema includes `delivery_attempts`, `next_attempt_at`, `lease_until`, `last_error`, `delivered_at`, and `updated_at`.
  - [x] Existing pending rows are compatible with the new schema.

### FR-002 - Claim pending notifications for parallel dispatch workers

- Description:
  - Dispatcher must claim due pending notifications with lease semantics so multiple workers can run safely.
- Acceptance criteria:
  - [x] Repository claims only rows with `delivery_status = 'pending'`, `next_attempt_at <= now`, and no active lease.
  - [x] Claim operation sets `lease_until = claim_until`.
  - [x] Save paths clear `lease_until` after success or failure handling.

### FR-003 - Deliver to one fixed webhook endpoint

- Description:
  - Dispatcher must send every claimed notification to one fixed URL configured by environment variables.
- Acceptance criteria:
  - [x] Webhook URL is loaded from env and must be a valid HTTPS URL.
  - [x] Secret is loaded from env and required at startup.
  - [x] Successful `2xx` responses mark the notification `sent`.

### FR-004 - Sign webhook payloads

- Description:
  - Every request must carry a deterministic payload and HMAC signature so the receiver can verify authenticity.
- Acceptance criteria:
  - [x] Payload includes notification identity and status-change snapshot fields.
  - [x] Request includes a SHA-256 HMAC signature header computed from the raw JSON body.
  - [x] Payload includes a stable event type and version.

### FR-005 - Retry failed deliveries with bounded attempts

- Description:
  - Failed deliveries must be retried until a configured attempt limit is reached.
- Acceptance criteria:
  - [x] Retryable failures increase `delivery_attempts`, persist `last_error`, and schedule `next_attempt_at`.
  - [x] When attempts reach `max_attempts`, row transitions to `failed`.
  - [x] Worker output distinguishes sent, retried, and terminal failed rows.

### FR-006 - Provide independent runtime for webhook delivery

- Description:
  - Webhook delivery must have its own binary/bootstrap/DI path rather than reuse the blockchain poller runtime.
- Acceptance criteria:
  - [x] A dedicated command exists for webhook dispatch.
  - [x] Runtime reads batch size, interval, claim TTL, max attempts, retry delay, URL, secret, and timeout from env.
  - [x] Startup fails fast on invalid config.

### FR-007 - Default compose service and test-env mock receiver

- Description:
  - Webhook dispatcher must be part of the default compose topology, while local test runs use a committed env file for fake settings and a compose override for the fake receiver service.
- Acceptance criteria:
  - [x] `compose.yaml` defines the webhook dispatcher service directly.
  - [x] `compose.yaml` uses required Compose env interpolation for `PAYMENT_RECEIPT_WEBHOOK_URL` and `PAYMENT_RECEIPT_WEBHOOK_SECRET`.
  - [x] `deployments/compose/compose.test.env` provides committed fake webhook settings for local test runs.
  - [x] `compose.test.yaml` defines a fake webhook receiver service that the dispatcher can call in local test runs.

## Non-functional requirements

- Reliability (NFR-001):
  - Delivery state changes must be persisted transactionally per notification row, and active leases must prevent duplicate concurrent handling.
- Security (NFR-002):
  - Only a fixed HTTPS endpoint may receive webhooks, and the shared secret must not be persisted in the database.
- Maintainability (NFR-003):
  - Delivery orchestration depends on outbound ports, with HTTP-specific behavior isolated in a webhook adapter.
- Operability (NFR-004):
  - Operators can tune retry/interval settings entirely via environment variables and inspect failed deliveries in the database.
- Testability (NFR-005):
  - Local compose test environment must provide a deterministic webhook endpoint without changing application code.

## Dependencies and integrations

- External systems:
  - One fixed webhook receiver endpoint owned by the platform.
- Internal services:
  - `payment_receipt_status_notifications` outbox rows produced by the receipt polling flow.
