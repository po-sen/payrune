---
doc: 02_design
spec_date: 2026-03-31
slug: issuance-space-key-cleanup
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-31-allocation-recovery-schema-phase1
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Issuance Space Key Cleanup - Technical Design

## High-level approach

- Summary:
  - Keep phase-1 recoverability untouched and clean only the redundant issuance-ref persistence.
  - Keep source-aware continuity exactly where it already belongs: on
    `address_policy_cursors(address_policy_id, address_space_ref)` and on the allocation rows that
    may need to reopen failed reservations for that same source.
  - Keep `address_policy_allocations` as the only allocation table and retain the minimum
    allocation-level technical fields needed by runtime behavior:
    `address_space_ref`, `slot_index`, and `failure_reason`.
  - Remove `issuance_ref_kind` and `issuance_ref` from the final schema because issued-row
    recovery already persists equivalent data inside `sweep_material_json`.
  - Remove `IssuanceRefKind` / `IssuanceRef` from the allocation entity's persisted-row contract as
    well so replay lookups do not load partial issued allocations.
  - Add bootstrap validation for enabled BTC xpub / tpub configs in both API runtimes so malformed
    source config fails fast before public requests.
  - Keep that validation co-located with API policy-construction code; do not keep a separate
    bootstrap-only helper file for it.
- Key decisions:
  - `address_space_ref` stays on the allocation row because failed rows and source-rotation
    continuity need a row-level source identity unless an extra table is introduced.
  - `address_policy_cursors` stays keyed by `(address_policy_id, address_space_ref)` so rotating
    to a new source starts a new cursor and switching back resumes the old cursor.
  - `issuance_ref_kind` and `issuance_ref` do not stay on the final row because they are redundant
    once `sweep_material_json` exists.
  - Do not add `address_policy_sources` or any startup source-guard table; continuity remains
    implicit in the cursor table.
  - Add one DB invariant: `issued` rows must have `sweep_material_json`.
  - Treat the cleanup migration as one-way after redundant issuance-ref columns are dropped; a
    lossless rollback requires restoring from backup or a new forward repair migration.

## System context

- Components:
  - Application:
    - Allocation use case remains unchanged in behavior.
  - Outbound adapters:
    - Postgres and Cloudflare Postgres allocation stores read/write
      `address_policy_allocations` directly and use source-aware `address_policy_cursors` for
      sequence allocation.
  - Migrations:
    - Compatibility migration only adds the issued-row `sweep_material_json` invariant.
    - Cleanup migration removes `issuance_ref_kind` / `issuance_ref`.
- Interfaces:
  - Public APIs remain unchanged.
  - ETH sweep helper remains unchanged because it reads `sweep_material_json`, not issuance-ref
    columns.

## Key flows

- Flow 1:
  - Fresh DB bootstraps with `address_policy_allocations` plus source-scoped cursors keyed by
    `(address_policy_id, address_space_ref)`.
  - New binary reserves fresh slots and reopens failed rows directly on the allocation table.
- Flow 2:
  - Source rotation under one policy.
  - If a policy starts using a new `address_space_ref`, `ReserveFresh` creates a fresh cursor row
    for that `(address_policy_id, address_space_ref)` pair and starts from slot `0`.
  - If maintainers later switch back to an old source, `ReserveFresh` finds that old cursor row and
    continues from its stored `next_index`.
- Flow 3:
  - Failed-row reopen.
  - `ReopenFailedReservation` filters by both `address_policy_id` and `address_space_ref`, so a
    failed reservation created under one source is not accidentally reopened under another.
- Flow 4:
  - Cleanup rollout.
  - Add the issued-row `sweep_material_json` invariant first.
  - After the new binary that no longer persists issuance refs is deployed, drop
    `issuance_ref_kind` / `issuance_ref`.

## Diagrams (optional)

- Mermaid sequence / flow:

```mermaid
flowchart LR
    A[Configured policy + source ref] --> B[Lookup cursor by policy + source]
    B --> C[Reserve or reopen allocation row with same source ref]
    C --> D[Issue address and persist sweep_material_json]
    D --> E[Drop redundant issuance_ref columns in final cleanup]
```

## Data model

- Schema changes or migrations:
  - Compatibility migration:
    - add `CHECK (allocation_status <> 'issued' OR sweep_material_json IS NOT NULL)` on
      `address_policy_allocations`
  - Cleanup migration:
    - drop `issuance_ref_kind` and `issuance_ref` from `address_policy_allocations`
    - keep `address_space_ref`, `slot_index`, and `failure_reason` on
      `address_policy_allocations`
    - keep `address_policy_cursors(address_policy_id, address_space_ref, next_index, ...)`
    - fail fast on down migration with an explicit irreversible-cleanup message instead of
      pretending to reconstruct dropped issuance-ref columns
- Consistency and idempotency:
  - Uniqueness remains enforced by
    `(address_policy_id, address_space_ref, slot_index)`.
  - Public address uniqueness remains enforced by `(chain, address)`.
  - Runtime stores use one final data model only; compatibility lives in migration SQL, not in
    runtime schema-mode branching.

## API or contracts

- Endpoints or events:
  - No public endpoint, webhook, or sweep-material contract changes.
- Request/response examples:
  - N/A

## Backward compatibility (optional)

- API compatibility:
  - Allocation and status APIs are unchanged.
  - Startup behavior changes for invalid BTC source config: runtimes now fail fast instead of
    serving a generic runtime 500 later.
- Data migration compatibility:
  - Existing BTC issued rows, ETH issued rows, cursor rows, and multi-source-per-policy histories
    are upgraded in place.
  - `sweep_material_json` remains unchanged through the migration chain.
  - Cleanup down migration is intentionally not lossless because dropped issuance-ref columns are
    not part of the final schema contract.

## Failure modes and resiliency

- Retries/timeouts:
  - Not changed.
- Backpressure/limits:
  - Not changed.
- Degradation strategy:
  - Source-aware cursor continuity remains available throughout the migration chain.
  - Cleanup migration is explicitly deferred until the new binary is already deployed.

## Observability

- Logs:
  - No new operator log contract is required.
- Metrics:
  - No new runtime metric is required in phase 2.

## Security

- Authentication/authorization:
  - Not changed.
- Secrets:
  - Not changed.

## Alternatives considered

- Option A:
  - Collapse continuity to `address_policy_id` only and add a startup source registry.
- Option B:
  - Remove `address_space_ref` from the main schema and add a side table for failed-row/source
    continuity.
- Why chosen:
  - Keeping `address_space_ref` on the allocation row and cursor row is simpler than either a
    startup registry or an extra side table, and it preserves the allocator's historical behavior.

## Risks

- Risk:
  - Removing `address_space_ref` from the allocation/cursor schema would break long-standing
    source-rotation continuity.
- Mitigation:
  - Keep source-aware continuity keyed directly by `(address_policy_id, address_space_ref)`.
- Risk:
  - Existing phase-1 recoverability could regress accidentally.
- Mitigation:
  - Include explicit validation that `sweep_material_json` content and ETH sweep helper behavior do
    not change, and make the DB reject issued rows without `sweep_material_json`.
- Risk:
  - Removing issuance-ref persistence from the schema but leaving issuance-ref fields on the
    runtime entity would create a partially populated issued allocation model.
- Mitigation:
  - Remove those fields from the persisted allocation entity contract and keep issuance refs only in
    derivation outputs / sweep material generation paths where they are still needed.
