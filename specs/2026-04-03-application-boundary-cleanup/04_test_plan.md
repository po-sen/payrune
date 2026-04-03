---
doc: 04_test_plan
spec_date: 2026-04-03
slug: application-boundary-cleanup
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-domain-model-boundary-cleanup
  - 2026-04-02-sweep-material-redesign
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
  - Application use-case regression for address allocation and health check.
  - HTTP controller response compatibility for current JSON payloads.
  - Outbound contract cleanup regression.
- Not covered:
  - DB payload content changes, because this spec must preserve existing behavior.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001 / FR-003 / NFR-001
  - Steps:
    - Run `go test ./internal/application/usecases -run TestAllocatePaymentAddress`
  - Expected:
    - Allocation use-case tests pass with opaque sweep-material contract names and unchanged
      behavior.
- TC-002:
  - Linked requirements: FR-002 / FR-003 / NFR-002
  - Steps:
    - Run `go test ./internal/adapters/inbound/http/controllers/...`
  - Expected:
    - Controller tests pass without payload shape changes.

### Integration

- TC-101:
  - Linked requirements: FR-001 / FR-002 / FR-003 / NFR-001 / NFR-002 / NFR-006
  - Steps:
    - Run `go test ./internal/application/... ./internal/adapters/inbound/http/controllers/...`
  - Expected:
    - Application-layer and inbound HTTP adapter regression suites pass together.

### E2E (if applicable)

- Scenario 1:
  - Not applicable for this refactor.

## Edge cases and failure modes

- Case:
  - Health response still serializes timestamp in the current HTTP format after moving formatting
    responsibility to controllers.
- Expected behavior:
  - Controller tests lock the output shape.
- Case:
  - Idempotency replay responses still omit replay-only internal fields from HTTP output.
- Expected behavior:
  - Controller tests lock the omission behavior.

## NFR verification

- Performance:
  - No additional DB or network calls introduced.
- Reliability:
  - Existing tests for allocation, health, and status retrieval remain green.
- Security:
  - No change.
