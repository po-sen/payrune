---
doc: 04_test_plan
spec_date: 2026-03-08
slug: payment-address-status-api
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-04-policy-payment-address-allocation
  - 2026-03-05-blockchain-receipt-polling-service
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Payment Address Status API - Test Plan

## Scope

- Covered:
  - Use case success, not-found, and invariant-failure behavior.
  - Controller routing, input validation, status-code mapping, and response shape.
  - Postgres finder mapping for one issued payment status view.
- Not covered:
  - Full end-to-end polling against a live blockchain node.
  - Authentication or tenant scoping, which do not exist in this change.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-002, FR-003, NFR-006
  - Steps:
    - Execute the payment-status read use case with a fake finder returning a complete view.
    - Execute again with not-found and invariant-failure cases.
  - Expected:
    - Success returns the DTO unchanged, not-found maps to the application not-found error, and missing receipt state maps to an internal error.

### Integration

- TC-101:

  - Linked requirements: FR-002, FR-003, NFR-001, NFR-003, NFR-006
  - Steps:
    - Run Postgres finder tests that seed an issued allocation joined to receipt tracking and read by `paymentAddressId`.
    - Add cases for unknown ID and non-issued allocations.
  - Expected:
    - Finder returns the complete payment status view only for issued rows with matching receipt tracking and does not expose derivation path data.

- TC-102:
  - Linked requirements: FR-001, FR-002, FR-003, NFR-002
  - Steps:
    - Run controller tests for `GET /v1/chains/bitcoin/payment-addresses/{paymentAddressId}` with success, invalid id, not found, wrong method, and unsupported chain cases.
  - Expected:
    - HTTP responses match the contract: `200`, `400`, `404`, and `405` as appropriate.

### E2E (if applicable)

- Scenario 1:
  - Create a payment address, then call the new `GET` endpoint before and after receipt polling updates the row.
- Scenario 2:
  - Confirm the same payment status can be observed through the read API even if webhook delivery is unavailable.

## Edge cases and failure modes

- Case:
  - `paymentAddressId` is missing, non-numeric, zero, or negative.
  - Expected behavior:
    - Controller returns `400`.
- Case:
  - Allocation exists but is not `issued`.
  - Expected behavior:
    - Finder/use case reports not found and controller returns `404`.
- Case:
  - Issued allocation exists but receipt tracking row is missing.
  - Expected behavior:
    - Use case treats it as an internal invariant failure and controller returns `500`.

## NFR verification

- Performance:
  - Confirm the finder reads the status view without any external blockchain calls.
- Reliability:
  - Confirm the endpoint reads persisted state directly and does not depend on webhook runtime availability.
- Security:
  - Confirm response payload omits xpubs, derivation path, and webhook-delivery internals.
