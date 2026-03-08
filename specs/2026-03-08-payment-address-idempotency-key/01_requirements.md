---
doc: 01_requirements
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

# Payment Address Idempotency Key - Requirements

## Glossary (optional)

- Idempotency key:
  - Client-provided `Idempotency-Key` header used to make create-request retries safe.
- Replay:
  - A repeated `POST /v1/chains/{chain}/payment-addresses` request using the same idempotency key.

## Out-of-scope behaviors

- OOS1:
  - Dedupe driven by `customerReference`.
- OOS2:
  - Idempotency support for requests that do not provide the header.

## Functional requirements

### FR-001 - Accept optional `Idempotency-Key` header

- Description:
  - The allocation endpoint must accept an optional idempotency header for replay-safe request handling.
- Acceptance criteria:
  - [x] `POST /v1/chains/{chain}/payment-addresses` accepts `Idempotency-Key` header.
  - [x] Header value is trimmed before use.
  - [x] Missing or blank header keeps the endpoint's current non-idempotent behavior.
  - [x] CORS responses for allowed Swagger origin include `Idempotency-Key` in `Access-Control-Allow-Headers`.
  - [x] CORS responses for allowed Swagger origin expose `Idempotency-Replayed` through `Access-Control-Expose-Headers`.
- Notes:
  - `customerReference` remains in the JSON body and is not repurposed as the replay key.

### FR-002 - Reuse existing allocation for same-key same-payload replay

- Description:
  - Same idempotency key with the same request payload must return the existing issued allocation.
- Acceptance criteria:
  - [x] When a completed idempotency record already exists for the same `chain + Idempotency-Key`, the API loads it before allocating a fresh index.
  - [x] If stored payload matches `addressPolicyId`, `expectedAmountMinor`, and `customerReference`, the API returns the existing success payload.
  - [x] Replay success keeps the endpoint response status at `201 Created` and includes response header `Idempotency-Replayed: true`.
  - [x] Replay path does not create a new allocation row.
  - [x] Replay path does not create a second receipt-tracking row.

### FR-003 - Reject conflicting key reuse

- Description:
  - Same idempotency key with a different request payload must be rejected.
- Acceptance criteria:
  - [x] If a completed idempotency record exists for the same `chain + Idempotency-Key` but request payload differs, the API returns `409 Conflict`.
  - [x] Conflict response uses `{ "error": "..." }` payload shape.
  - [x] Conflict response does not allocate a new derivation index.

### FR-004 - Use a dedicated payment-address idempotency store

- Description:
  - Duplicate protection must persist independently from `address_policy_allocations`.
- Acceptance criteria:
  - [x] Persistence creates a dedicated table for payment-address idempotency keyed by `chain + Idempotency-Key`.
  - [x] The idempotency record stores request snapshot fields needed to detect conflicting key reuse and the completed `payment_address_id`.
  - [x] `address_policy_allocations` no longer stores `idempotency_key` and no longer owns uniqueness for it.
  - [x] If concurrent same-key same-payload requests race, the losing request returns the committed allocation after reloading the idempotency record.
  - [x] If concurrent same-key different-payload requests race, the losing request returns deterministic conflict behavior.

### FR-005 - Keep failed issuance retryable

- Description:
  - Idempotency should not permanently reserve a key when the create flow does not produce an issued allocation.
- Acceptance criteria:
  - [x] If the create flow rolls back, no idempotency record remains visible after the transaction ends.
  - [x] If derivation failure is persisted as a business outcome, the claimed idempotency record is released before commit.
  - [x] A later retry with the same key after a non-issued outcome can attempt allocation again.

### FR-006 - Keep architecture boundaries clean

- Description:
  - Idempotency behavior must be split cleanly across inbound, application, and persistence layers.
- Acceptance criteria:
  - [x] Controller reads `Idempotency-Key` header and maps it into application input only.
  - [x] Application use case decides whether a replay is equivalent or conflicting and coordinates claim/release/complete of the idempotency key.
  - [x] A dedicated idempotency store owns key lookup and duplicate-claim translation only.
  - [x] Allocation persistence keeps ownership of allocation rows and issued-allocation lookup by ID.
  - [x] Existing `GET` address APIs remain unchanged.

## Non-functional requirements

- Performance (NFR-001):
  - Idempotent replay path keeps local warm-DB p95 latency <= 200 ms.
- Availability/Reliability (NFR-002):
  - For 100 parallel same-key retries, issued allocation duplicates remain `0`.
- Security/Privacy (NFR-003):
  - No new secret or private-key handling is introduced.
- Compliance (NFR-004):
  - No additional compliance controls are introduced in this scope.
- Observability (NFR-005):
  - Error payload format remains stable, conflict text is deterministic, and replay responses expose a deterministic replay header.
- Maintainability (NFR-006):
  - Spec and tests clearly distinguish `Idempotency-Key` transport behavior from `customerReference` business data and from allocation persistence.

## Dependencies and integrations

- External systems:
  - PostgreSQL idempotency and allocation tables.
- Internal services:
  - `AllocatePaymentAddressUseCase`
  - `PaymentAddressAllocationStore`
  - `PaymentAddressIdempotencyStore`
  - `AddressPolicyReader`
