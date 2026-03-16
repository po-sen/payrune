---
doc: 02_design
spec_date: 2026-03-16
slug: remove-xpub-fingerprint
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

# Technical Design

## High-level approach

- Summary:
  - Remove fingerprint fields from domain and policy-reader code.
  - Persist new reservation state by `account_public_key` directly.
  - Use a destructive pre-production migration so the final allocation schema can require
    non-null xpub-backed rows and simpler cursor logic.
- Key decisions:
  - `AddressDerivationConfig` keeps only `AccountPublicKey` and `DerivationPathPrefix`.
  - `address_policy_cursors` is rebuilt around (`address_policy_id`, `account_public_key`).
  - `address_policy_allocations.account_public_key` becomes required for all surviving rows.
  - `address_policy_allocations` drops the legacy fingerprint columns entirely.
  - The final repo shape keeps one merged migration file for this feature instead of splitting the
    xpub-backed transition and legacy-column drop across two adjacent migrations.
  - Pre-production allocation process data is intentionally cleared during migration so no legacy
    null-key rows survive.

## System context

- Components:
  - Domain:
    - `AddressDerivationConfig`
    - `AddressIssuancePolicy`
  - Outbound policy adapter:
    - `address_policy_reader`
  - Outbound persistence adapters:
    - PostgreSQL allocation store
    - Cloudflare PostgreSQL allocation store
- Interfaces:
  - No public API shape changes.
  - SQL migration `000010` updates the persistence model.

## Key flows

- Flow 1:
  - Policy reader loads a configured xpub and derivation path prefix.
  - Allocation issuance validates chain, enabled state, and amount without fingerprint checks.
  - Fresh reservation writes/locks cursor state by (`address_policy_id`, `account_public_key`).
- Flow 2:
  - The migration clears legacy allocation process rows, so the first post-migration xpub-backed
    cursor for any policy/xpub starts from `0`.
- Flow 3:
  - After xpub-backed rows exist for the policy, reserve-fresh seeds from the max index for the
    same xpub only.
  - A future xpub rotation under the same policy therefore starts at index `0`.
- Flow 4:
  - Payment-address-status lookup by `paymentAddressId` keeps reading the allocation record by ID
    and remains independent from the key migration.

## Data model

- Entities:
  - `address_policy_cursors`
    - key: `address_policy_id`, `account_public_key`
    - payload: `next_index`, timestamps
  - `address_policy_allocations`
    - active keying: `address_policy_id`, `account_public_key`, `derivation_index`
    - lifecycle/state: unchanged (`reserved`, `issued`, `derivation_failed`)
- Schema changes or migrations:
  - Clear `address_policy_allocations` with `TRUNCATE ... CASCADE` so dependent receipt/idempotency
    rows do not survive into the strict schema.
  - Add `account_public_key` to `address_policy_allocations` as `NOT NULL`.
  - Replace fingerprint-based cursor table with an xpub-backed cursor table.
  - Replace active unique/index paths to use `account_public_key`.
  - Let PostgreSQL automatically drop legacy fingerprint-dependent unique/index objects when the
    fingerprint columns are dropped.
  - Drop `xpub_fingerprint_algo` and `xpub_fingerprint` from `address_policy_allocations`.
- Consistency and idempotency:
  - Cursor row creation remains transactional.
  - After the destructive migration, cursor seed only needs to consider the same xpub's max index.

## API or contracts

- Endpoints or events:
  - No request/response or route changes.
- Request/response examples:
  - Existing address-policy, address-derivation, allocation, and status DTOs remain unchanged.

## Backward compatibility (optional)

- API compatibility:
  - Fully preserved.
- Data migration compatibility:
  - Allocation-related pre-production rows are intentionally discarded.
  - Rollback after new writes is still not lossless without a database snapshot because the old
    fingerprint values cannot be reconstructed from newly written rows.

## Failure modes and resiliency

- Retries/timeouts:
  - Existing retry behavior does not change.
- Backpressure/limits:
  - None beyond current transaction/lock behavior.
- Degradation strategy:
  - If migration is not applied, new binaries fail fast when SQL hits missing schema.
  - If the migration is applied to an environment where historical allocation data matters, data
    loss occurs by design.

## Observability

- Logs:
  - Existing error surfaces remain unchanged.
- Metrics:
  - None added.
- Traces:
  - None added.
- Alerts:
  - None added.

## Security

- Authentication/authorization:
  - No change.
- Secrets:
  - No private keys are introduced.
- Abuse cases:
  - Storing xpub directly increases operator visibility only; it does not expand signing capability.

## Alternatives considered

- Option A:
  - Keep fingerprint model and continue hashing xpubs.
- Option B:
  - Drop and recreate all allocation process data with xpub-backed keys.
- Option C:
  - Use xpub-backed keys for active rows and drop the fingerprint columns after the transition.
- Why chosen:
  - In a pre-production environment, clearing process data is the simplest way to enforce a strict
    final schema without null legacy rows or compatibility branches.

## Risks

- Risk:
  - The migration destroys existing allocation-related process data.
- Mitigation:
  - Apply it only before formal production rollout and document the destructive behavior.
- Risk:
  - Rollback after new writes is not trivially lossless.
- Mitigation:
  - Document migration sequencing and require a DB snapshot for full rollback after production
    traffic.
