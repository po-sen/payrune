---
doc: 04_test_plan
spec_date: 2026-03-04
slug: policy-payment-address-allocation
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-btc-xpub-address-api
  - 2026-03-03-postgresql18-migration-runner-container
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Policy-Based Payment Address Allocation - Test Plan

## Scope

- Covered:
  - Allocation reserve/derive/finalize/failure lifecycle.
  - Xpub-rotation sequence reset behavior.
  - API validation and response fields.
  - Repository/UoW boundary behavior and policy reader path.
- Not covered:
  - On-chain payment confirmation and settlement.
  - Production-scale performance benchmarking.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-004, NFR-006
  - Steps:
    - Controller tests for `POST /v1/chains/bitcoin/payment-addresses` with required `expectedAmountMinor` and optional `customerReference`.
  - Expected:
    - `201` response includes `paymentAddressId` and `expectedAmountMinor`, without derivation index.
- TC-002:
  - Linked requirements: FR-002, FR-003
  - Steps:
    - Use case tests for reserve success, finalize success, derivation failure marking, and exhaustion mapping.
  - Expected:
    - Correct transitions and error mapping.
- TC-005:
  - Linked requirements: FR-006
  - Steps:
    - Domain tests verify `PaymentAddressAllocation` transitions from `reserved` to `issued` and `derivation_failed`.
  - Expected:
    - No separate completion object is required; aggregate carries finalized/failure state.
    - Status type comes from `internal/domain/value_objects`.
- TC-003:
  - Linked requirements: FR-002, FR-006
  - Steps:
    - Assert reserve input contains `xpubFingerprintAlgo` and `xpubFingerprint` for policy.
  - Expected:
    - Fingerprint-partition key is always present.
- TC-004:
  - Linked requirements: FR-005, FR-006
  - Steps:
    - List/generate use case tests through `AddressPolicyReader`.
  - Expected:
    - Deterministic behavior unchanged.

### Integration

- TC-101:
  - Linked requirements: FR-002, FR-003, NFR-002
  - Steps:
    - Run migration and execute sequential allocations for same policy/fingerprint key.
  - Expected:
    - Unique, monotonic indexes and issued records persisted.
- TC-102:
  - Linked requirements: FR-002, NFR-002
  - Steps:
    - Simulate xpub rotation (fingerprint change), run first allocation.
  - Expected:
    - New key starts at index `0`.
- TC-103:
  - Linked requirements: FR-003, NFR-005
  - Steps:
    - Force derivation failure and inspect persisted row.
  - Expected:
    - `derivation_failed` status + reason persisted; retry can reopen same index.
- TC-104:
  - Linked requirements: FR-006, NFR-006
  - Steps:
    - Run static check: `rg -n "AddressPolicyQueryService|adapters/outbound/config" internal`.
  - Expected:
    - No runtime dependency on removed query-service/config-adapter path.

### E2E (if applicable)

- Scenario 1:
  - Allocate twice with same policy/fingerprint key.
  - Expected: different addresses and different `paymentAddressId`.
- Scenario 2:
  - Rotate xpub then allocate.
  - Expected: new key sequence starts cleanly.

## Edge cases and failure modes

- Case:
  - Missing/blank `addressPolicyId`.
- Expected behavior:
  - `400` with structured error payload.
- Case:
  - Non-positive `expectedAmountMinor`.
- Expected behavior:
  - `400`.
- Case:
  - Unsupported chain.
- Expected behavior:
  - `404`.
- Case:
  - Disabled policy (`xpub` missing).
- Expected behavior:
  - deterministic not-enabled error.
- Case:
  - Exhausted index range.
- Expected behavior:
  - deterministic pool exhausted error.

## NFR verification

- Performance:
  - Local p95 target check for allocation endpoint.
- Reliability:
  - Verify `0` duplicates under concurrent same-key load.
- Security:
  - Confirm no private-key handling and no derivation-index leak in customer API.
