---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Fake receiver signature verification, verification logging inputs, and local compose env wiring.
- Not covered:
  - Production webhook retry policy and production receiver implementation.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002, NFR-001, NFR-002, NFR-003
  - Steps:
    - Add tests around the fake receiver handler for valid and invalid signatures.
  - Expected:
    - Valid requests return `204`, invalid signatures return non-2xx, and the handler uses HMAC-SHA256 verification.

### Integration

- TC-101:
  - Linked requirements: FR-003, NFR-002
  - Steps:
    - Render compose config with `compose.test.env` and `compose.test.yaml`.
  - Expected:
    - The fake receiver service receives the webhook secret in the rendered config.

## Edge cases and failure modes

- Case:
  - Missing signature header.
- Expected behavior:

  - The receiver logs verification failure and returns non-2xx.

- Case:
  - Secret is not configured.
- Expected behavior:
  - The receiver logs that verification is skipped or unavailable in an explicit way.

## NFR verification

- Observability:
  - Logs show one readable request block with full headers and raw body, without duplicate success-only sections.
- Security:
  - Signature comparison uses constant-time HMAC verification.
