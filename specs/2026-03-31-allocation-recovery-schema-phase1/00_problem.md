---
doc: 00_problem
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

# Allocation Recovery Schema Phase 1 - Problem & Goals

## Context

- Background:
  - `payrune` now allocates both Bitcoin HD and Ethereum CREATE2 payment addresses from the same
    `address_policy_allocations` table.
  - After the recent address-space / issuance-ref rename, allocation and cursor continuity now
    depend on the current `address_space_ref` + `slot_index` schema and must not be disturbed by
    this change.
  - Operators need a recoverable, DB-readable material view so they can understand one issued row
    and perform ETH sweep operations without reconstructing internal partitioning semantics.
- Users or stakeholders:
  - Payment operators recovering or sweeping issued payment addresses.
  - Payrune maintainers responsible for migrations, allocation safety, and API stability.
- Why now:
  - Current operator recovery still requires reading internal fields such as `issuance_ref`,
    `issuance_ref_kind`, and CREATE2 config details separately.
  - The ETH sweep workflow should become DB-driven before broader schema cleanup is attempted.

## Constraints (optional)

- Technical constraints:
  - This is phase 1 only: do not change allocation continuity, cursor partitioning, slot reuse
    behavior, or `address_space_ref` semantics.
  - Keep the existing chain-scoped allocation APIs and idempotency / receipt-tracking flow intact.
  - Use the existing spec workflow, Go project layout, and Clean Architecture boundaries.
- Timeline/cost constraints:
  - Prefer the smallest safe refactor that improves operator recoverability now and defers
    high-risk internal cleanup.
- Compliance/security constraints:
  - ETH sweep tooling must validate operator env completeness and Ledger sender identity before any
    broadcast decision.

## Problem statement

- Current pain:
  - The database does not expose one canonical operator-facing recovery material for issued payment
    addresses.
  - ETH CREATE2 sweep helpers still require operators to manually provide factory / collector /
    salt inputs that already exist implicitly in the issued allocation model.
  - If operator recovery keeps depending on internal partitioning fields, future cleanup becomes
    more dangerous and more confusing.
- Evidence or examples:
  - Bitcoin recovery requires mentally combining `address_space_ref`, `issuance_ref`, `scheme`, and
    chain/network columns.
  - Ethereum CREATE2 recovery requires combining DB fields with checked-in asset metadata and
    manual script inputs for factory / collector / salt.
  - The user explicitly requires phase-1 safety: any change that touches allocation/cursor
    partitioning or slot continuity must be deferred.

## Goals

- G1:
  - Add one DB-readable, operator-facing `sweep_material_json` to `address_policy_allocations`.
- G2:
  - Ensure every newly issued Bitcoin HD and Ethereum CREATE2 row stores recoverable sweep
    material in that JSON.
- G3:
  - Backfill existing issued rows so operator recovery does not depend on legacy rows being reissued.
- G4:
  - Make ETH sweep DB-driven so operators only need DB access plus a small env set and a Ledger.
- G5:
  - Preserve `POST /v1/chains/bitcoin/payment-addresses` and
    `POST /v1/chains/ethereum/payment-addresses` behavior.

## Non-goals (out of scope)

- NG1:
  - Removing `address_space_ref` or redesigning internal issuance / partitioning storage.
- NG2:
  - Changing allocation or cursor partitioning, slot continuity, or legacy space-key compatibility.
- NG3:
  - Adding a second operator-facing JSON column such as `issuance_space_json`.
- NG4:
  - Beautifying or normalizing the rest of the internal allocation schema in the same change.

## Assumptions

- A1:
  - The current `address_space_ref`, `slot_index`, `issuance_ref`, and `issuance_ref_kind` fields
    remain the internal compatibility model during phase 1.
- A2:
  - All CREATE2 policies continue to use the checked-in receiver artifact and deployment metadata
    already embedded in `internal/infrastructure/ethereumcreate2assets/`.
- A3:
  - It is acceptable in phase 1 to backfill `sweep_material_json` from existing row data and
    checked-in CREATE2 artifact knowledge without redesigning the internal schema.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - New issued row coverage for `sweep_material_json`.
- Target:
  - `100%` of newly issued Bitcoin and Ethereum allocation rows persist non-null
    `sweep_material_json`.
- Metric:
  - Existing issued row backfill coverage.
- Target:
  - `100%` of previously issued Bitcoin and Ethereum rows receive non-null
    `sweep_material_json` after migration.
- Metric:
  - ETH sweep helper recoverability.
- Target:
  - A dry-run can assemble the `cast send ... "sweep()"` command using only DB-selected
    `sweep_material_json` plus the allowed env inputs.
- Metric:
  - Allocation API stability.
- Target:
  - Focused BTC and ETH allocation tests continue to pass with no new 500-class regression in the
    allocation path.
