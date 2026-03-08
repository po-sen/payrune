---
doc: 02_design
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

# Payment Address Idempotency Key - Technical Design

## High-level approach

- Summary:
  - Read `Idempotency-Key` from the HTTP header and pass it into the allocation use case.
  - Persist payment-address idempotency in a dedicated PostgreSQL table instead of on allocation rows.
  - Lookup existing completed idempotency records by `chain + idempotency_key` before allocating a new one.
  - Claim the idempotency key inside the issuance transaction, then finalize it with the winning `payment_address_id`.
- Key decisions:
  - `customerReference` remains part of the business payload and is compared as part of replay equivalence, not used as the replay key itself.
  - Idempotency scope is `chain + Idempotency-Key`.
  - Same key + same payload returns the existing success payload.
  - Same key + different payload returns `409 Conflict`.
  - Requests without the header keep current semantics.
  - Failed/non-issued attempts release the claimed key before commit so the same key is still retryable.

## System context

- Components:
  - Inbound adapter: `ChainAddressController`
  - Application: `AllocatePaymentAddressUseCase`
  - Outbound persistence: `PaymentAddressAllocationStore`, `PaymentAddressIdempotencyStore`, `PaymentReceiptTrackingStore`, `UnitOfWork`
  - Policy metadata: `AddressPolicyReader`
- Interfaces:
  - `POST /v1/chains/{chain}/payment-addresses`
  - PostgreSQL `payment_address_idempotency_keys`
  - PostgreSQL `address_policy_allocations`

## Key flows

- Flow 1: first request with idempotency key
  - Controller validates body, reads `Idempotency-Key`, and forwards both.
  - Use case looks for a completed idempotency record by `chain + idempotency_key`.
  - No hit: transaction starts, claims the idempotency key row, then runs the existing reserve/derive/issue flow.
  - After receipt tracking is created, the use case finalizes the idempotency row with the issued `payment_address_id`.
- Flow 2: sequential replay with same payload
  - Use case finds the completed idempotency record by key.
  - Stored payload matches request payload.
  - Use case loads the issued allocation by `payment_address_id`, marks the output as replayed, and returns the existing allocation response without opening a write transaction.
- Flow 3: sequential replay with different payload
  - Use case finds the completed idempotency record by key.
  - Stored payload differs from request payload.
  - Use case returns `409 Conflict`.
- Flow 4: concurrent replay race
  - Multiple requests miss the pre-read before the winner commits.
  - One transaction claims the idempotency key row and commits first.
  - Losers hit the dedicated idempotency-key unique constraint on claim, then reload the idempotency record by key.
  - If payload matches, losers return the winner's issued allocation; otherwise they return conflict.
- Flow 5: non-issued failure with idempotency key
  - Transaction claims the key row first.
  - Issuance fails before an issued allocation exists, but the use case still commits a derivation-failed allocation row.
  - Before commit, the use case deletes the claimed idempotency row so the key is not stuck.
- Flow 6: header absent
  - Use case skips idempotency lookup and keeps current non-idempotent allocation behavior.

## Data model

- Entities:
  - No new domain entity is introduced for idempotency because this is technical process state, not business behavior.
- Technical records:
  - `payment_address_idempotency_keys` stores:
    - `chain`
    - `idempotency_key`
    - `address_policy_id`
    - `expected_amount_minor`
    - `customer_reference`
    - `payment_address_id`
- Schema changes or migrations:
  - Migration `000008` creates `payment_address_idempotency_keys`.
  - Primary key is (`chain`, `idempotency_key`).
  - `payment_address_id` is nullable during the in-transaction claim, then set before commit for successful issuance.
  - `payment_address_id` references `address_policy_allocations(id)`.
  - `address_policy_allocations` is unchanged by this feature.
- Consistency and idempotency:
  - DB uniqueness on the dedicated idempotency table is the concurrency safety net.
  - Application pre-read is the fast path for sequential replay.
  - Claim, allocation issuance, receipt tracking creation, and idempotency completion happen in one DB transaction.

## API or contracts

- Endpoint:
  - `POST /v1/chains/{chain}/payment-addresses`
- Request contract:
  - Existing JSON body remains unchanged.
  - New optional header: `Idempotency-Key: <string>`.
- Response contract:
  - Existing success body remains unchanged.
  - Replay success keeps `201 Created` and adds response header `Idempotency-Replayed: true`.
  - New conflict path returns `409` with `{ "error": "idempotency key conflicts with existing payment address" }`.
  - CORS for the Swagger origin exposes `Idempotency-Replayed` so browser-based tooling can read it.

## Backward compatibility (optional)

- API compatibility:
  - Body schema and success payload remain unchanged.
  - Requests that omit the header keep existing behavior.
- Data migration compatibility:
  - Existing allocation rows require no backfill.
  - New idempotency records only exist for requests that send the header after rollout.

## Failure modes and resiliency

- Retries/timeouts:
  - Same-key retries become safe when the header is present.
- Backpressure/limits:
  - Replay path avoids unnecessary derivation and write load.
- Degradation strategy:
  - DB lookup or reload failures still return server error.
  - Failed issuance does not strand an idempotency key in a completed-looking state.
  - Derivation failures that are intentionally persisted are returned after commit as a distinct application-level failure path, not as transaction rollback errors.

## Observability

- Logs:
  - No new log contract required in this scope.
- Metrics:
  - Replay success/conflict counts can be inferred later from API outcomes.
- Traces:
  - Replay detection stays within the request trace boundary.
- Alerts:
  - No new alert requirement in this scope.

## Security

- Authentication/authorization:
  - No auth changes.
- Secrets:
  - No new secrets.
- Abuse cases:
  - Client misuse of one key for multiple payloads results in deterministic conflict instead of duplicate issuance.

## Alternatives considered

- Option A:
  - Reuse `customerReference` as the dedupe key.
- Option B:
  - Use `Idempotency-Key` header and persist it on the allocation row.
- Option C:
  - Use `Idempotency-Key` header with a dedicated payment-address idempotency table.
- Why chosen:
  - Option C matches standard API semantics, keeps business payload separate from transport replay control, and avoids coupling technical uniqueness to allocation-row storage.

## Risks

- Risk:
  - Future payload fields could be added without updating replay-equivalence checks.
- Mitigation:
  - Keep replay comparison centralized in the use case and update it with any future request-shape changes.
- Risk:
  - A committed transaction could accidentally leave an idempotency row without `payment_address_id` if the failure path is implemented incorrectly.
- Mitigation:
  - Test the business-failure path explicitly and release the key inside the same transaction before commit.
