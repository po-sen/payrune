---
doc: 04_test_plan
spec_date: 2026-03-13
slug: readme-product-api-webhook
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-10-cloudflare-workers-postgres
  - 2026-03-11-cloudflare-poller-workers
  - 2026-03-13-cloudflare-webhook-dispatcher-worker
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Test Plan

## Scope

- Covered:
  - Top-level README accuracy for product summary, API integration, webhook integration, and
    deployment commands.
- Not covered:
  - Runtime behavior changes or live deployment checks.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-004, NFR-001, NFR-002
  - Steps:
    - Compare README product/deploy sections against current compose, Cloudflare worker docs, and
      `.env.cloudflare.example`.
  - Expected:
    - README reflects current local and Cloudflare deployment entrypoints and key parameters.

### Integration

- TC-101:
  - Linked requirements: FR-002, FR-003, NFR-002
  - Steps:
    - Compare README API and webhook examples against `deployments/swagger/openapi.yaml` and
      `internal/adapters/outbound/webhook/payment_receipt_status_notifier.go`.
  - Expected:
    - Field names, headers, and signing rule match the current implementation.

## Edge cases and failure modes

- Case:
  - README grows too long and stops being a quick-start document.
- Expected behavior:
  - Keep sections short and focused on first-use tasks.

## NFR verification

- Maintainability:
  - Confirm README stays concise and avoids duplicating large implementation detail.
- Accuracy:
  - Confirm examples and parameter names align with current runtime sources of truth.
