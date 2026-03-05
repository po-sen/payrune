---
doc: 02_design
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

# Policy-Based Payment Address Allocation - Technical Design

## High-level approach

- Summary:
  - Keep `POST /v1/chains/{chain}/payment-addresses` as customer allocation entrypoint.
  - Partition cursor by (`addressPolicyId`, `xpubFingerprintAlgo`, `xpubFingerprint`).
  - Persist lifecycle states and reconciliation fields in allocation table.
  - Return `paymentAddressId` and `expectedAmountMinor` in response.
- Key decisions:
  - `expectedAmountMinor` is integer minor unit in API request/response.
  - Cursor key includes fingerprint algorithm + fingerprint to support xpub rotation reset.
  - Reservation fallback is deterministic: `reopen derivation_failed -> reserve fresh`.
  - Shared `UnitOfWork` owns transaction lifecycle; repositories never begin/commit/rollback transactions.
  - `UnitOfWork` contract is repository-agnostic and accepts tx-bound repository bundle in callback.
  - Tx-bound repository bundle uses application port struct (`out.TxRepositories`) to keep DI wiring explicit and remove DI-local adapter wrapper types.
  - DI wiring uses adapter factory method reference (`postgres.NewTxRepositories`) to avoid inline closure noise in container composition.
  - Transaction-scoped repository execution is composed in DI/runtime via tx-repository builder.
  - Postgres command repository stays pure `database/sql`; repository methods execute SQL through injected executor (`*sql.DB`/`*sql.Tx`) without owning transaction boundaries.
  - Command-side persistence is repository pattern and returns domain entities/aggregates only.
  - Allocation domain model is a single aggregate (`PaymentAddressAllocation`) with explicit state transition methods; no separate completion aggregate/object.
  - Allocation status enum (`PaymentAddressAllocationStatus`) is a value object in `internal/domain/value_objects`.
  - Policy read path uses one `AddressPolicyReader` port (`ListByChain`, `FindByID`).
  - Postgres allocation persistence is a single repository implementation; tx-scoped use reuses the same repository type instead of building a nested reservation-repository implementation.
  - Policy source is DI-provided in-memory provider under infrastructure, not config outbound adapter.
  - Fingerprint algorithm is deterministic (`sha256-trunc64-hex-v1`) and computed in provider layer.

## System context

- Components:
  - Inbound adapter: `ChainAddressController`.
  - Application use cases:
    - `ListAddressPoliciesUseCase`
    - `GenerateAddressUseCase`
    - `AllocatePaymentAddressUseCase`
  - Domain:
    - `AddressPolicy` entity.
    - `PaymentAddressAllocation` aggregate.
  - Outbound ports:
    - `AddressPolicyReader`
    - `BitcoinAddressDeriver`
    - `PaymentAddressAllocationRepository`
    - `UnitOfWork`
  - Outbound adapters:
    - PostgreSQL allocation persistence.
    - Bitcoin xpub address derivation.
  - Infrastructure:
    - DI policy reader (`address_policy_reader.go`) builds normalized policies from env config.
- Interfaces:
  - `POST /v1/chains/{chain}/payment-addresses`
  - `GET /v1/chains/{chain}/address-policies`
  - `GET /v1/chains/{chain}/addresses`

## Key flows

- Flow 1: Successful allocation
  - Controller validates body (`addressPolicyId`, `expectedAmountMinor`, optional `customerReference`).
  - Use case loads policy via `AddressPolicyReader.FindByID`.
  - Use case executes UoW transaction and reserves index by fallback order.
  - Deriver builds address and relative path from policy xpub + index.
  - Domain aggregate composes absolute derivation path.
  - Repository finalizes row to `issued` with chain/network/scheme/address/path.
  - Controller returns `201` with `paymentAddressId` and metadata.
- Flow 2: Derivation failure
  - Reservation succeeds.
  - Derivation fails.
  - Repository marks row `derivation_failed` with reason.
  - API returns server error.
- Flow 3: Retry after failure
  - Reservation first attempts reopening a `derivation_failed` row/index.
  - If none available, reserves fresh cursor index.
  - Both decisions run in one UoW transaction.
- Flow 4: Xpub rotation reset
  - New xpub changes fingerprint.
  - New cursor row starts `next_index = 0` for new key.

## Data model

- Entities:
  - `address_policy_cursors`
    - key: `address_policy_id`, `xpub_fingerprint_algo`, `xpub_fingerprint`
    - payload: `next_index`, timestamps
  - `address_policy_allocations`
    - identity/keying: `id`, `address_policy_id`, `xpub_fingerprint_algo`, `xpub_fingerprint`, `derivation_index`
    - amount/reference: `expected_amount_minor`, `customer_reference`
    - issued fields: `chain`, `network`, `scheme`, `address`, `derivation_path`
    - lifecycle: `allocation_status`, `failure_reason`, `reserved_at`, `issued_at`
- Schema changes or migrations:
  - Migration set `000002` contains cursor/allocation schema for this feature.
- Consistency and idempotency:
  - Consistency via transaction + row lock + constraints.
  - Endpoint remains intentionally non-idempotent.

## API or contracts

- Endpoint:
  - `POST /v1/chains/{chain}/payment-addresses`
- Request body:
  - `addressPolicyId` (required)
  - `expectedAmountMinor` (required, integer > 0)
  - `customerReference` (optional)
- Response body:
  - `paymentAddressId`, `expectedAmountMinor`, `addressPolicyId`, `chain`, `network`, `scheme`, `minorUnit`, `decimals`, `address`, optional `customerReference`

## Backward compatibility (optional)

- Existing deterministic derivation endpoint behavior is preserved.
- Existing policy list endpoint is preserved.
- New allocation response fields are additive.

## Failure modes and resiliency

- DB unavailable: service startup fails fast.
- Derivation failure after reserve: row moves to `derivation_failed`, index remains reusable.
- Index exhaustion: mapped to deterministic business error.

## Observability

- Error payload remains stable.
- Failure lifecycle is persisted in DB for troubleshooting/reconciliation.

## Security

- xpub-only handling; no private keys.
- No new auth model changes in this scope.

## Alternatives considered

- A: sequence keyed only by `addressPolicyId`.
- B: sequence keyed by (`addressPolicyId`, `xpubFingerprintAlgo`, `xpubFingerprint`).
- Chosen: B, because it guarantees xpub rotation reset and prevents key-space collision across fingerprint algorithms.

## Risks

- Risk:
  - Architecture drift when many micro-refactors are tracked independently.
- Mitigation:
  - Consolidate into one canonical feature spec (this folder).
- Risk:
  - Future need for policy DB source.
- Mitigation:
  - Keep use case dependency on `AddressPolicyReader` port so source can swap later.

## Consolidation notes

- This spec supersedes and merges feature-level micro-specs:
  - `2026-03-04-uow-tx-repositories-bundle`
  - `2026-03-04-unit-of-work-naming-cleanup`
  - `2026-03-04-address-policy-repository-layering`
  - `2026-03-04-repository-naming-consistency`
  - `2026-03-04-address-policy-repository-name-fix`
  - `2026-03-04-xpub-fingerprint-di-strategy`
  - `2026-03-04-payment-allocation-repository-file-split`
  - `2026-03-04-adapter-directory-rollback`
  - `2026-03-05-remove-config-address-policy-adapter`
