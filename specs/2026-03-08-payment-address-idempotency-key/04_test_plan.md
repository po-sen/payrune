---
doc: 04_test_plan
spec_date: 2026-03-08
slug: payment-address-idempotency-key
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-04-policy-payment-address-allocation
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Payment Address Idempotency Key - Test Plan

## Scope

- Covered:
  - Header-based replay success path.
  - Conflicting key-reuse behavior.
  - Dedicated idempotency-store lookup, claim, release, and duplicate-claim translation.
  - OpenAPI/controller wiring for `Idempotency-Key`.
  - CORS allow-header behavior for Swagger browser requests using `Idempotency-Key`.
- Not covered:
  - Generic idempotency support for unrelated endpoints.
  - Full end-to-end external load testing.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002, FR-006, NFR-006
  - Steps:
    - Add use-case test where a completed idempotency record already exists for the same `Idempotency-Key` and the payload matches.
  - Expected:
    - Use case returns the existing allocation, marks the output as replayed, and does not open a write transaction.
- TC-002:
  - Linked requirements: FR-003, FR-006, NFR-005
  - Steps:
    - Add use-case test where the key exists but amount or other compared payload fields differ.
  - Expected:
    - Use case returns deterministic idempotency conflict.
- TC-003:
  - Linked requirements: FR-001, FR-006
  - Steps:
    - Add controller test confirming `Idempotency-Key` header is trimmed and passed into application input.
  - Expected:
    - Application input contains the normalized header value.
- TC-007:
  - Linked requirements: FR-002, FR-006, NFR-005
  - Steps:
    - Add controller test where the use case returns a replayed success result.
  - Expected:
    - HTTP status remains `201` and response header `Idempotency-Replayed` is `true`.
- TC-004:
  - Linked requirements: FR-001
  - Steps:
    - Keep use-case tests without `Idempotency-Key`.
  - Expected:
    - Existing non-idempotent behavior still works.
- TC-005:
  - Linked requirements: FR-005, FR-006, NFR-006
  - Steps:
    - Add use-case test where derivation failure is persisted as a business outcome after claiming the idempotency key.
  - Expected:
    - Use case releases the idempotency claim before commit, and a later retry can claim the same key again.
- TC-006:
  - Linked requirements: FR-001
  - Steps:
    - Add CORS middleware test for preflight request from Swagger origin with requested header `Idempotency-Key`.
  - Expected:
    - `Access-Control-Allow-Headers` includes `Idempotency-Key`.
- TC-008:
  - Linked requirements: FR-001, NFR-005
  - Steps:
    - Add CORS middleware test for browser request from Swagger origin.
  - Expected:
    - `Access-Control-Expose-Headers` includes `Idempotency-Replayed`.

### Integration

- TC-101:
  - Linked requirements: FR-002, FR-004, FR-006, NFR-002
  - Steps:
    - Add adapter test for idempotency-record lookup by `chain + idempotency_key`.
  - Expected:
    - Stored row is mapped back correctly, including request snapshot fields and `payment_address_id`.
- TC-102:
  - Linked requirements: FR-004, FR-006, NFR-002
  - Steps:
    - Add adapter test where `Claim` hits the dedicated idempotency primary-key conflict.
  - Expected:
    - Adapter maps DB duplicate key error to the dedicated idempotency-key collision error.
- TC-103:
  - Linked requirements: FR-001, FR-004
  - Steps:
    - Validate migration SQL shape.
  - Expected:
    - `payment_address_idempotency_keys` is present and `address_policy_allocations` does not gain an `idempotency_key` column.
- TC-104:
  - Linked requirements: FR-004, FR-006
  - Steps:
    - Add adapter test for issued allocation lookup by `payment_address_id`.
  - Expected:
    - Allocation store returns the issued row needed for replay response building.

### E2E (if applicable)

- Scenario 1:
  - Call `POST /v1/chains/bitcoin/payment-addresses` twice with the same body and same `Idempotency-Key`.
  - Expected:
    - Both responses return the same `paymentAddressId` and `address`.
- Scenario 2:
  - Call the same endpoint twice with the same `Idempotency-Key` but different payload.
  - Expected:
    - The second response is `409 Conflict` and no new address is issued.

## Edge cases and failure modes

- Case:
  - `Idempotency-Key` absent or blank.
- Expected behavior:
  - Replay protection is skipped and current allocation behavior remains unchanged.
- Case:
  - Unique violation occurs after a concurrent winning request commits.
- Expected behavior:
  - Losing request reloads the committed idempotency record and returns success or conflict based on payload equivalence.
- Case:
  - Derivation failure is committed after the key was claimed.
- Expected behavior:
  - The key is released before commit, so the next retry is not blocked.

## NFR verification

- Performance:
  - Targeted replay-path tests remain within local warm-DB expectations.
- Reliability:
  - Same-key retries do not create duplicate issued allocations.
- Security:
  - No new secret/private-key handling and no derivation-index exposure.
