---
doc: 01_requirements
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

# Requirements

## Out-of-scope behaviors

- OOS1:
  - Changing API, webhook, or deployment runtime behavior.
- OOS2:
  - Documenting every internal package or every legacy spec.

## Functional requirements

### FR-001 - Root README must explain the product in one screen

- Description:
  - The repo root must have a `README.md` that quickly explains what Payrune is and what it does.
- Acceptance criteria:
  - [ ] `README.md` exists at repo root.
  - [ ] It describes Payrune as a Bitcoin payment-address, status-tracking, and webhook product.
  - [ ] It lists the supported public API surface at a high level.

### FR-002 - Root README must document direct API integration

- Description:
  - The README must provide enough API detail for a human or AI client to wire the core payment
    address flow.
- Acceptance criteria:
  - [ ] It documents create payment address request fields.
  - [ ] It documents status lookup path and payment status values.
  - [ ] It includes at least one concrete request example.

### FR-003 - Root README must document webhook integration

- Description:
  - The README must provide the webhook headers, payload shape, and signature verification rule.
- Acceptance criteria:
  - [ ] It documents `X-Payrune-Event`, `X-Payrune-Event-Version`,
        `X-Payrune-Notification-ID`, and `X-Payrune-Signature-256`.
  - [ ] It describes the HMAC-SHA256 signing rule over the raw request body.
  - [ ] It includes the webhook payload fields needed to consume status updates.

### FR-004 - Root README must document key parameters and deployment entrypoints

- Description:
  - The README must tell operators which parameters matter and how to deploy locally and on
    Cloudflare.
- Acceptance criteria:
  - [ ] It lists the main API, poller, and webhook parameters.
  - [ ] It references `.env.cloudflare.example`.
  - [ ] It documents `make up/down` and `make cf-up/down`.

## Non-functional requirements

- Maintainability (NFR-001):
  - The document must stay concise and fit a quick first read.
- Accuracy (NFR-002):
  - The README must align with current OpenAPI, compose, Cloudflare worker docs, and webhook
    signing behavior.

## Dependencies and integrations

- Internal sources of truth:
  - `deployments/swagger/openapi.yaml`
  - `deployments/compose/compose.yaml`
  - `deployments/cloudflare/*/README.md`
  - `internal/adapters/outbound/webhook/payment_receipt_status_notifier.go`
