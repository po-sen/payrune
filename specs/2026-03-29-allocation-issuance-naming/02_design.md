---
doc: 02_design
spec_date: 2026-03-29
slug: allocation-issuance-naming
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Allocation Issuance Naming - Technical Design

## High-level approach

- Summary:
  - Rename allocation persistence and internal model fields so they explicitly describe address
    space, slot reservation, and typed issuance references.
  - Add one typed issuance-reference column instead of overloading one free-form string field.
  - Normalize Ethereum CREATE2 issuance persistence so only the salt payload is stored as the
    issuance reference.
- Key decisions:
  - Keep the public API unchanged; this is an internal naming and persistence-semantics refactor.
  - Keep the existing `scheme` field unchanged in this iteration to avoid mixing this rename with a
    broader contract change.
  - Introduce `issuance_ref_kind` as a small typed discriminator rather than creating a JSON blob or
    generic key/value metadata column.
  - Update both SQL adapters and core domain/application types to the same vocabulary so the schema
    and code review language stay aligned.

## System context

- Components:
  - Domain:
    - `PaymentAddressAllocation`
    - `AddressIssuancePolicy`
    - issuance-reference kind value object
  - Application:
    - allocation use case
    - chain and issued-address deriver ports
  - Outbound adapters:
    - PostgreSQL and Cloudflare Postgres allocation stores
    - Bitcoin HD address derivers
    - Ethereum CREATE2 derivers
  - Bootstrap:
    - policy configuration assembly
- Interfaces:
  - SQL migrations under `deployments/postgresql/migrations/`
  - internal ports under `internal/application/ports/outbound/`
  - allocation persistence stores under `internal/adapters/outbound/persistence/`

## Key flows

- Flow 1:
  - Reserve a slot for one allocation request.
  - The store keys reservation uniqueness by `(address_policy_id, address_space_ref, slot_index)`.
  - The domain entity keeps `slot_index` rather than `derivation_index`.
- Flow 2:
  - Complete one issued allocation.
  - Bitcoin writes `issuance_ref_kind=hd_path_absolute` plus an absolute HD path payload.
  - Ethereum CREATE2 writes `issuance_ref_kind=create2_salt` plus the salt payload only.

## Diagrams (optional)

- Mermaid sequence / flow:

## Data model

- Entities:
  - `PaymentAddressAllocation`
    - `SlotIndex`
    - `IssuanceRefKind`
    - `IssuanceRef`
  - `AddressIssuanceConfig`
    - `AddressSpaceRef`
    - `IssuanceRefPrefix` for methods that compose a payload from a stable prefix plus a
      relative reference, such as Bitcoin HD
- Schema changes or migrations:
  - Rename `address_policy_allocations.address_source_ref` to `address_space_ref`.
  - Rename `address_policy_allocations.derivation_index` to `slot_index`.
  - Rename `address_policy_allocations.address_reference` to `issuance_ref`.
  - Add `address_policy_allocations.issuance_ref_kind`.
  - Rename `address_policy_cursors.address_source_ref` to `address_space_ref`.
  - Update indexes and cursor queries to use the renamed columns.
  - Backfill existing issued rows:
    - Bitcoin: `issuance_ref_kind = 'hd_path_absolute'`
    - Ethereum CREATE2: `issuance_ref = <salt only>` and `issuance_ref_kind = 'create2_salt'`
- Consistency and idempotency:
  - Allocation uniqueness remains `(address_policy_id, address_space_ref, slot_index)`.
  - Public address uniqueness remains `(chain, address)`.
  - Idempotent replay continues to resolve issued allocations by ID and does not depend on the old
    column names.

## API or contracts

- Endpoints or events:
  - No public HTTP contract changes.
  - Internal outbound port contracts are renamed to the new vocabulary.
- Request/response examples:
  - N/A; this change is internal.

## Backward compatibility (optional)

- API compatibility:
  - Existing allocation and status APIs remain unchanged.
- Data migration compatibility:
  - Existing rows are migrated in place.
  - Down migration restores the previous column names and reconstructs Ethereum CREATE2 reference
    strings by prefixing the policy ID to the stored salt payload when rolling back.

## Failure modes and resiliency

- Retries/timeouts:
  - No new runtime retry model is introduced.
- Backpressure/limits:
  - Slot range limits remain the same as the current allocation cursor limits.
- Degradation strategy:
  - If the migration fails, the schema should remain transactionally unchanged by the migration
    runner.

## Observability

- Logs:
  - No new logs required beyond existing migration and application error logging.
- Metrics:
  - No new metrics required.
- Traces:
  - No new traces required.
- Alerts:
  - Existing migration and application startup failure alerting remains sufficient.

## Security

- Authentication/authorization:
  - No change.
- Secrets:
  - Ethereum CREATE2 derivation keys remain runtime-managed; the new schema must not persist them.
- Abuse cases:
  - Avoid storing CREATE2 references in a way that looks like public sequential derivation data.

## Alternatives considered

- Option A:
  - Keep the current schema and rely on documentation.
- Option B:
  - Force all issuance methods into a derivation-path-like string format.
- Why chosen:
  - Documentation alone does not fix the schema-level ambiguity.
  - Path-like normalization would misrepresent CREATE2 semantics instead of clarifying them.

## Risks

- Risk:
  - Widespread renames can break tests or leave one adapter on the old vocabulary.
- Mitigation:
  - Change the SQL migration, core entities, ports, adapters, bootstrap, and tests in one patch and
    verify with focused Go tests.
