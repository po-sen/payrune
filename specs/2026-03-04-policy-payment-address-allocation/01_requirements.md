---
doc: 01_requirements
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

# Policy-Based Payment Address Allocation - Requirements

## Glossary (optional)

- Xpub fingerprint:
  - Deterministic hash identifier of configured xpub, tagged by algorithm name.
- Payment address id:
  - Stable server-generated identifier for one allocation record (`paymentAddressId`).
- Allocation lifecycle:
  - State transitions `reserved -> issued` or `reserved -> derivation_failed`.

## Out-of-scope behaviors

- OOS1:
  - Address reuse prevention based on on-chain payment detection.
- OOS2:
  - DB-backed policy CRUD and admin management.

## Functional requirements

### FR-001 - Customer allocation API without index input

- Description:
  - Provide allocation endpoint by `addressPolicyId`, without client-provided derivation index.
- Acceptance criteria:
  - [x] Endpoint remains `POST /v1/chains/{chain}/payment-addresses`.
  - [x] Request requires `addressPolicyId`.
  - [x] Request requires positive integer `expectedAmountMinor`.
  - [x] Request accepts optional `customerReference`.
  - [x] Missing/blank `addressPolicyId` returns `400`.
  - [x] Missing or non-positive `expectedAmountMinor` returns `400`.
- Notes:
  - Customer never receives derivation index.

### FR-002 - Sequence partition by policy and xpub fingerprint

- Description:
  - Allocate index by key (`addressPolicyId`, `xpubFingerprintAlgo`, `xpubFingerprint`) instead of policy-only.
- Acceptance criteria:
  - [x] Same key never re-issues same index.
  - [x] Same fingerprint value under different algorithms is isolated by key.
  - [x] Xpub change under same policy starts sequence at index `0`.
  - [x] Failed derivation index can be reused and is not permanently burned.
  - [x] If index exceeds `2147483647`, API returns deterministic business error.
- Notes:
  - Fingerprint is computed server-side from configured xpub.

### FR-003 - Persist rich allocation lifecycle records

- Description:
  - Persist reservation/issued/failure records for future reconciliation and tracing.
- Acceptance criteria:
  - [x] Reservation persists `addressPolicyId`, `xpubFingerprintAlgo`, `xpubFingerprint`, `derivationIndex`, `expectedAmountMinor`, `customerReference`, `status=reserved`, `reservedAt`.
  - [x] Issued state persists `chain`, `network`, `scheme`, `address`, `derivationPath`, `issuedAt`, `status=issued`.
  - [x] `derivationPath` is absolute from root (example `m/84'/0'/0'/0/42`).
  - [x] Failure persists `status=derivation_failed` and failure reason.
  - [x] Reopen failed reservation is attempted before fresh cursor allocation, within one transaction boundary.
  - [x] Data model supports reverse lookup by (`chain`, `address`) and by `paymentAddressId`.
- Notes:
  - Internal persistence fields need not all appear in API response.

### FR-004 - Return reconciliation-friendly response fields

- Description:
  - Allocation response includes stable ID and amount metadata.
- Acceptance criteria:
  - [x] Success response includes `paymentAddressId`.
  - [x] Success response includes `expectedAmountMinor`.
  - [x] Success response includes `addressPolicyId`, `chain`, `network`, `scheme`, `minorUnit`, `decimals`, `address`.
  - [x] Success response may echo `customerReference` when provided.
  - [x] Response does not expose derivation index.

### FR-005 - Preserve deterministic listing/derivation behavior

- Description:
  - Existing listing and index-based derivation APIs keep deterministic behavior.
- Acceptance criteria:
  - [x] `GET /v1/chains/{chain}/address-policies` remains available.
  - [x] `GET /v1/chains/{chain}/addresses?addressPolicyId=...&index=...` behavior remains unchanged.
  - [x] Existing tests for listing and index-based derivation remain passing.

### FR-006 - Keep architecture boundaries clean with one policy reader port

- Description:
  - Use cases depend only on ports; policy read path is unified under one repository abstraction.
- Acceptance criteria:
  - [x] Policy read abstraction uses unified naming `AddressPolicyReader` and provides both `ListByChain` and `FindByID`.
  - [x] Use cases (`list`, `generate`, `allocate`) use `AddressPolicyReader` only.
  - [x] Application use case tests are split into dedicated files per use case (`list`, `generate`, `allocate`) for maintainability.
  - [x] Transaction lifecycle (`BeginTx/Commit/Rollback`) is owned by shared Unit of Work adapter, not repository.
  - [x] Shared UoW abstraction is repository-agnostic (UoW does not depend on concrete repositories).
  - [x] Transaction and repository composition is wired in DI/composition root via tx-repository builder.
  - [x] Use case transaction callback receives tx-bound repositories (not tx/context directly).
  - [x] Postgres command repository uses pure `database/sql` and never owns `Begin/Commit/Rollback`.
  - [x] Command-side persistence returns domain entities/aggregates only.
  - [x] Runtime source no longer depends on `internal/adapters/outbound/config/address_policy_repository.go`.
  - [x] `PaymentAddressAllocation` uses a single aggregate state model for `reserved -> issued/derivation_failed` transitions.
  - [x] Completion data is represented as aggregate state, not as a separate domain object type.
  - [x] Allocation status type (`PaymentAddressAllocationStatus`) is modeled as a value object under `internal/domain/value_objects`.
  - [x] Postgres allocation adapter uses one repository implementation only; no nested reservation-repository constructor/factory.

## Non-functional requirements

- Performance (NFR-001):
  - Allocation endpoint p95 latency <= 200ms in local warm-DB environment.
- Availability/Reliability (NFR-002):
  - For 100 parallel requests on same allocation key, returned addresses are unique (`0` duplicates).
- Security/Privacy (NFR-003):
  - xpub-only derivation; no private-key handling introduced.
- Compliance (NFR-004):
  - No additional compliance controls in this iteration.
- Observability (NFR-005):
  - Error payload format remains `{ "error": "..." }`; failure status is persisted.
- Maintainability (NFR-006):
  - Single canonical spec package and clean port-driven architecture; no adapter business logic leakage.

## Dependencies and integrations

- External systems:
  - PostgreSQL for cursor and allocation lifecycle persistence.
- Internal services:
  - `BitcoinAddressDeriver` outbound port.
  - `AddressPolicyReader` outbound port implemented by DI-provided in-memory policy reader.
