---
doc: 01_requirements
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

# Issuance Space Key Cleanup - Requirements

## Glossary (optional)

- Term:
  - Source-scoped cursor
- Definition:
  - The cursor identified by `(address_policy_id, address_space_ref)`; this preserves historical
    continuity when a policy rotates between multiple issuance sources.
- Term:
  - Source ref
- Definition:
  - The persisted `address_space_ref` value used internally for BTC xpub/tpub or ETH CREATE2
    issuance continuity.

## Out-of-scope behaviors

- OOS1:
  - Any change to `sweep_material_json` content or operator usage.
- OOS2:
  - Any new operator-visible JSON describing internal partitioning.
- OOS3:
  - Any change to public allocation or status API shapes.

## Functional requirements

### FR-001 - Keep continuity keyed by policy plus source ref

- Description:
  - Final runtime partitioning and cursor continuity must rely on
    `(address_policy_id, address_space_ref)`, not `address_policy_id` alone.
- Acceptance criteria:
  - [ ] Allocation stores reserve fresh slots using `address_policy_id` plus
        `IssuancePolicy.IssuanceConfig.AddressSpaceRef`.
  - [ ] Allocation stores reopen failed rows using the same policy-plus-source identity.
  - [ ] `address_policy_cursors` final schema keys cursor state by
        `(address_policy_id, address_space_ref)`.
  - [ ] A new source under the same policy starts from a fresh cursor.
  - [ ] Switching back to a previously used source under the same policy resumes that source's
        previous cursor.

### FR-002 - Keep the final schema simple

- Description:
  - Final phase 2 must use one allocation table and one cursor table, without an extra registry or
    allocation side table.
- Acceptance criteria:
  - [ ] `address_policy_allocations` retains `sweep_material_json`.
  - [ ] `address_policy_allocations` retains `address_space_ref`, `slot_index`, and
        `failure_reason`.
  - [ ] `address_policy_allocations` no longer stores `issuance_ref_kind` or `issuance_ref`.
  - [ ] `address_policy_cursors` retains `address_space_ref`.
  - [ ] The final schema does not add a second readable JSON column for internal partitioning.
  - [ ] The final schema does not add an extra registry table only to track policy/source
        continuity.

### FR-003 - Preserve allocation and cursor continuity

- Description:
  - Fresh DBs and legacy DBs must preserve source-aware slot allocation continuity and uniqueness
    after the cleanup.
- Acceptance criteria:
  - [ ] Existing cursor values remain continuous per `(address_policy_id, address_space_ref)`.
  - [ ] Existing slot indexes remain continuous per `(address_policy_id, address_space_ref)`.
  - [ ] New allocations do not reissue old addresses after upgrade.
  - [ ] `(chain, address)` uniqueness remains enforced after upgrade.

### FR-004 - Keep migration rollout additive until final cleanup

- Description:
  - Phase 2 must not break old continuity semantics while the migration chain is applied.
- Acceptance criteria:
  - [ ] The compatibility migration is additive and does not remove source-aware continuity.
  - [ ] The cleanup migration only removes the redundant issuance-ref persistence columns.
  - [ ] Migration file names and rollback messaging describe the actual final cleanup being applied.
  - [ ] The final binary works after both the compatibility step and the cleanup step without
        runtime schema-mode branching.
  - [ ] Cleanup rollback semantics are explicit when dropped issuance-ref columns cannot be
        losslessly reconstructed.

### FR-005 - Keep sweep and operator behavior unchanged

- Description:
  - Phase-1 recoverability must remain the stable baseline during and after the cleanup.
- Acceptance criteria:
  - [ ] `sweep_material_json` content for equivalent BTC / ETH issued rows is unchanged by phase 2.
  - [ ] ETH DB-driven sweep helper behavior is unchanged for dry-run and broadcast decision logic.
  - [ ] Operator docs and workflows continue to point only at `sweep_material_json`.
  - [ ] The DB rejects `address_policy_allocations` rows with `allocation_status='issued'` when
        `sweep_material_json` is `NULL`.

### FR-006 - Keep public allocation flows stable

- Description:
  - Public allocation APIs must continue to work after the internal cleanup.
- Acceptance criteria:
  - [ ] `POST /v1/chains/bitcoin/payment-addresses` still allocates successfully.
  - [ ] `POST /v1/chains/ethereum/payment-addresses` still allocates successfully.
  - [ ] Focused smoke coverage shows no new 500-class failures in allocation flows.

### FR-007 - Align issued-allocation runtime shape with final persistence

- Description:
  - The final runtime model must not claim that issued allocations persist issuance-ref columns
    after phase 2 removes them from the schema.
- Acceptance criteria:
  - [ ] `PaymentAddressAllocation` no longer models persisted `IssuanceRefKind` /
        `IssuanceRef` as part of the issued allocation row state.
  - [ ] `FindIssuedByID` returns the final persisted issued state without leaving zero-valued
        legacy fields behind.
  - [ ] Allocation issuance still succeeds because sweep material remains the persisted recovery
        payload.

### FR-008 - Fail fast on invalid BTC source config

- Description:
  - Enabled BTC issuance policies with invalid xpub / tpub config must be rejected during bootstrap
    instead of reaching public runtime handlers as generic 500s.
- Acceptance criteria:
  - [ ] API bootstrap validates enabled BTC issuance policies before serving traffic.
  - [ ] Cloudflare API bootstrap performs the same validation.
  - [ ] Validation errors identify the offending policy or env-backed source config.
  - [ ] Valid BTC policies and ETH CREATE2 policies keep their existing behavior.

## Non-functional requirements

- Reliability (NFR-001):
  - Focused fresh-DB and legacy-upgrade validation must prove unchanged source-aware slot
    continuity, cursor continuity, and address uniqueness.
- Maintainability (NFR-002):
  - Final schema should be understandable without extra continuity tables.
- Maintainability (NFR-006):
  - Final migration names, rollback messages, and runtime entity contracts should match the actual
    final model without stale phase-2 leftovers.
- Operability (NFR-003):
  - The rollout sequence must be explicit enough that maintainers can apply it without guessing
    deployment ordering.
- Compatibility (NFR-004):
  - Legacy DBs that already use multiple sources under one policy must continue to upgrade and
    behave correctly.
- Performance (NFR-005):
  - Allocation hot paths must not perform runtime schema-mode detection or startup-only guard
    probes per request.

## Dependencies and integrations

- External systems:
  - PostgreSQL / Cloudflare Postgres.
- Internal services:
  - Allocation stores and cursor writes.
  - Bootstrap policy definitions.
  - BTC / ETH issued-address derivation.
  - Existing phase-1 ETH sweep helper.
