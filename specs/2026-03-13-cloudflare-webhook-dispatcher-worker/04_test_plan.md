---
doc: 04_test_plan
spec_date: 2026-03-13
slug: cloudflare-webhook-dispatcher-worker
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-webhook-delivery
  - 2026-03-10-cloudflare-workers-postgres
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
  - Cloudflare webhook dispatcher runtime, notifier bridge, mock worker, deployment shell, and
    automation.
- Not covered:
  - Live end-to-end delivery against a real merchant webhook endpoint in production.

## Tests

### Unit

- TC-001:

  - Linked requirements: FR-002, NFR-004
  - Steps:
    - Test the Cloudflare dispatcher handler against a fake
      `RunReceiptWebhookDispatchCycleUseCase`.
  - Expected:
    - Handler maps scheduled input to the use case and returns the expected summary.

- TC-002:

  - Linked requirements: FR-003, NFR-001, NFR-002
  - Steps:
    - Test the Cloudflare notifier adapter for success, retryable failure, terminal failure
      conditions, explicit Cloudflare binding transport mode, and Cloudflare PostgreSQL scan
      support for webhook outbox time columns.
  - Expected:
    - Delivery behavior matches the existing webhook dispatch policy expectations and outbox scan
      logic handles `time.Time`, `*time.Time`, and `sql.NullTime` destinations.

- TC-003:
  - Linked requirements: FR-004, NFR-003
  - Steps:
    - Test the mock worker JS handler for `POST /receipt-status`.
  - Expected:
    - Valid requests return `204`, invalid signatures return `401`, and invocation logs are
      emitted.

### Integration

- TC-101:

  - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-006
  - Steps:
    - Run `go test` on worker runtime packages and `npm test` in both Cloudflare shells.
  - Expected:
    - Go/Wasm entrypoint, JS shell, notifier bridge, and mock worker all pass.

- TC-102:
  - Linked requirements: FR-005, FR-006, NFR-003
  - Steps:
    - Run `wrangler deploy --dry-run` for both workers and `make -n` for orchestration.
  - Expected:
    - Deploy/delete wiring is valid, mock deploys before dispatcher, and observability config is
      present.

### E2E (if applicable)

- Scenario 1:
  - Deploy `receipt-webhook-mock` and `payrune-webhook-dispatcher`, then trigger one scheduled run.
- Scenario 2:
  - Verify the outbox item transitions to sent or retry/failed as expected.

## Edge cases and failure modes

- Case:
  - Webhook target returns non-2xx.
- Expected behavior:

  - Use case records retry or failed state according to existing delivery policy.

- Case:
  - Dispatcher worker has no pending notifications.
- Expected behavior:

  - Scheduled run completes successfully with zero claimed notifications.

- Case:
  - Mock worker receives invalid signature.
- Expected behavior:
  - Return `401` and log the failure.

## NFR verification

- Reliability:
  - Confirm sent / retry / failed transitions match existing delivery policy and binding path
    avoids a public network hop.
- Security:
  - Confirm shared secret remains secret-synced and no public webhook URL is required.
- Observability:
  - Confirm dispatcher and mock `wrangler.toml` files enable Cloudflare invocation logs.
