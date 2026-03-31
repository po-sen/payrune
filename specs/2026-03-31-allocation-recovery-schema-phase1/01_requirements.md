---
doc: 01_requirements
spec_date: 2026-03-31
slug: allocation-recovery-schema-phase1
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-29-allocation-issuance-naming
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Allocation Recovery Schema Phase 1 - Requirements

## Glossary (optional)

- Term:
  - Sweep material JSON
- Definition:
  - The only operator-facing recovery payload persisted on `address_policy_allocations` for one
    issued payment address.
- Term:
  - Internal compatibility fields
- Definition:
  - Existing fields such as `address_space_ref`, `slot_index`, `issuance_ref`, and
    `issuance_ref_kind` that remain for allocation / issuance internals during phase 1.

## Out-of-scope behaviors

- OOS1:
  - Any change to allocation or cursor partitioning semantics.
- OOS2:
  - Any new readable operator JSON besides `sweep_material_json`.
- OOS3:
  - Any migration that rewrites slot numbering or existing cursor continuity.

## Functional requirements

### FR-001 - Add one operator-facing recover-material column without changing internal partitioning

- Description:
  - `address_policy_allocations` must gain exactly one operator-facing recovery JSON column named
    `sweep_material_json`, while current allocation / cursor continuity fields remain unchanged.
- Acceptance criteria:
  - [ ] `address_policy_allocations` adds a nullable `sweep_material_json` JSON column in phase 1.
  - [ ] No second readable operator JSON column is added in this change.
  - [ ] `address_space_ref` remains present after the migration.
  - [ ] No unique index, cursor key, or allocation-partitioning column is changed by this phase-1
        work.
- Notes:
  - `issuance_ref` and `issuance_ref_kind` may remain as internal compatibility fields.

### FR-002 - Persist sweep material for all newly issued rows

- Description:
  - Successful issuance must persist `sweep_material_json` for every newly issued Bitcoin and
    Ethereum allocation row.
- Acceptance criteria:
  - [ ] Every newly issued Bitcoin row persists `sweep_material_json` with at least:
        `material_type`, `material_version`, `chain`, `network`, `address`,
        `hd_derivation_path`, `account_xpub`, `script_type`.
  - [ ] Every newly issued Ethereum CREATE2 row persists `sweep_material_json` with at least:
        `material_type`, `material_version`, `chain`, `network`, `address`,
        `predicted_address`, `factory_address`, `collector_address`, `create2_salt`,
        `init_code_hex`, `init_code_hash`.
  - [ ] Reserved and derivation-failed rows may keep `sweep_material_json` null.
  - [ ] Issuance writes to `sweep_material_json` do not require changes to allocation reservation,
        cursor lookup, or slot assignment behavior.
- Notes:
  - `predicted_address` may equal the persisted `address` for issued CREATE2 rows.

### FR-003 - Backfill existing issued rows without changing allocation continuity

- Description:
  - Existing issued rows must be backfilled from current persisted issuance data so operator
    recovery works for both old and new allocations.
- Acceptance criteria:
  - [ ] A migration backfills `sweep_material_json` for existing issued Bitcoin rows.
  - [ ] A migration backfills `sweep_material_json` for existing issued Ethereum CREATE2 rows.
  - [ ] The backfill does not modify `address_space_ref`, `slot_index`, `issuance_ref`,
        `issuance_ref_kind`, cursor rows, or allocation-status transitions.
  - [ ] The backfill only targets rows that are already `allocation_status='issued'`.
- Notes:
  - Any broader cleanup of internal issuance / partitioning data is deferred to phase 2.

### FR-004 - Make ETH sweep helper DB-driven and operator-safe

- Description:
  - The ETH sweep helper must load one issued allocation row from the database, read
    `sweep_material_json`, validate operator inputs, and use that data to assemble or broadcast the
    sweep transaction.
- Acceptance criteria:
  - [ ] The helper selects one allocation by env-provided selector, at minimum supporting either
        `payment_address_id` or `address`.
  - [ ] Operators are not required to manually provide factory, collector, or salt.
  - [ ] Supported env inputs are limited to `DATABASE_URL`, one allocation selector,
        `ETHEREUM_SWEEP_RPC_URL`, `ETHEREUM_SWEEP_FROM_ADDRESS`, and optional
        `ETHEREUM_SWEEP_DERIVATION_PATH`.
  - [ ] The helper validates env completeness before attempting a broadcast.
  - [ ] The helper validates that the connected Ledger sender matches
        `ETHEREUM_SWEEP_FROM_ADDRESS` before attempting a broadcast.
  - [ ] The helper rejects non-Ethereum or non-CREATE2 rows instead of trying to infer recovery
        material from internal fields.
- Notes:
  - Default behavior may be dry-run; broadcast can require an explicit opt-in flag.

### FR-005 - Keep operator docs and workflows focused on one JSON source of truth

- Description:
  - Operator-facing docs and helper usage must point to `sweep_material_json` as the only recovery
    payload for this workflow.
- Acceptance criteria:
  - [ ] README and relevant docs explain operator recovery in terms of `sweep_material_json`.
  - [ ] ETH helper usage docs reference only `sweep_material_json` for factory / collector / salt /
        init-code recovery.
  - [ ] New docs do not instruct operators to read `issuance_ref`, `issuance_ref_kind`, or
        `address_space_ref` directly for the phase-1 workflow.

### FR-006 - Preserve allocation API behavior during phase 1

- Description:
  - The phase-1 refactor must not break existing BTC / ETH payment-address allocation endpoints.
- Acceptance criteria:
  - [ ] `POST /v1/chains/bitcoin/payment-addresses` continues to allocate successfully.
  - [ ] `POST /v1/chains/ethereum/payment-addresses` continues to allocate successfully.
  - [ ] The change does not introduce new 500-class failures in focused allocation-path tests.

## Non-functional requirements

- Reliability (NFR-001):
  - Focused tests must show unchanged allocation / cursor behavior, and the phase-1 migrations must
    not alter `address_policy_cursors`, `slot_index`, or allocation uniqueness definitions.
- Security/Privacy (NFR-002):
  - The ETH helper must abort before broadcast when required env is missing, when the selected row
    is not an Ethereum CREATE2 issuance, or when the derived Ledger sender does not match
    `ETHEREUM_SWEEP_FROM_ADDRESS`.
- Maintainability (NFR-003):
  - After this change, only one operator-facing recover-material JSON column exists on
    `address_policy_allocations`, and operator docs no longer depend on internal compatibility
    fields.
- Operability (NFR-004):
  - A dry-run operator flow must succeed with only DB access, the allocation selector, RPC URL,
    sender address, and optional Ledger derivation path.

## Dependencies and integrations

- External systems:
  - PostgreSQL / Cloudflare Postgres.
  - Foundry `cast`, `psql`, and `jq` for the shell-based ETH sweep helper.
- Internal services:
  - Allocation issuance use case and issued-address derivers.
  - PostgreSQL and Cloudflare Postgres allocation stores.
  - Embedded Ethereum CREATE2 deployment metadata and receiver artifacts.
