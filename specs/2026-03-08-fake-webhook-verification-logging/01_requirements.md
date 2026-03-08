---
doc: 01_requirements
spec_date: 2026-03-08
slug: fake-webhook-verification-logging
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-webhook-delivery
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
  - The production webhook dispatcher retry logic is unchanged.
- OOS2:
  - The fake receiver does not need to persist webhook payloads.

## Functional requirements

### FR-001 - Fake receiver must verify webhook signatures with the shared secret

- Description:
  - The fake webhook receiver must validate `X-Payrune-Signature-256` against the raw request body using the shared webhook secret.
- Acceptance criteria:
  - [x] The fake receiver computes `HMAC-SHA256(secret, raw_body)` and compares it to the incoming signature header.
  - [x] The fake receiver returns a non-2xx status when signature verification fails.
  - [x] The verification path uses the same signature method as the production webhook notifier.

### FR-002 - Fake receiver logs must show the verification process clearly

- Description:
  - The fake webhook receiver must emit one readable request log that shows the full incoming webhook without extra duplicate sections.
- Acceptance criteria:
  - [x] Logs include the full received headers and full raw request body.
  - [x] Logs are formatted in readable multi-line sections rather than a single dense key-value line.
  - [x] The fake receiver no longer emits separate `verification` or `payload` log sections for successful requests.

### FR-003 - Test compose wiring must provide the receiver secret

- Description:
  - The local compose test environment must provide the fake receiver with the webhook secret so the verification path works end-to-end.
- Acceptance criteria:
  - [x] `compose.test.yaml` passes the webhook secret into the fake receiver container.
  - [x] The local fake receiver uses that secret without requiring changes to the production webhook dispatcher env contract.

## Non-functional requirements

- Maintainability (NFR-001):
  - Receiver-side verification logic should live behind small helper functions or a dedicated handler so it is directly testable.
- Observability (NFR-002):
  - Request logs must be concise enough to avoid duplicate sections while still showing the full incoming webhook.
- Security (NFR-003):
  - Verification must use constant-time HMAC comparison for signature matching.

## Dependencies and integrations

- External systems:
  - None.
- Internal services:
  - `cmd/fake_webhook_receiver`
  - `deployments/compose/compose.test.yaml`
  - `deployments/compose/compose.test.env`
