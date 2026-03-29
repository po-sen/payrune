---
doc: 00_problem
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

# Allocation Issuance Naming - Problem & Goals

## Context

- Background:
  - `payrune` currently persists payment-address allocation rows for both Bitcoin HD issuance and
    Ethereum CREATE2 issuance in the same tables.
  - Recent multi-chain work renamed Bitcoin-specific columns such as `account_public_key` and
    `derivation_path`, but the remaining names still mix address-space, slot, and issuance
    reference concepts in ways that are hard to read from the database alone.
  - Ethereum CREATE2 rows currently persist an `address_reference` value shaped like
    `ethereum-mainnet-create2/0x...`, which mixes reference kind labeling with the actual
    reconstructible payload.
- Users or stakeholders:
  - payrune maintainers reviewing allocation rows and schema changes.
  - wallet and payment operators debugging issued address metadata.
  - future feature contributors adding new issuance methods without reverse-engineering ambiguous
    column names.
- Why now:
  - The repo now supports both Bitcoin HD and Ethereum CREATE2 issuance, so naming that was merely
    awkward in a Bitcoin-only world now actively obscures the model.
  - The current discussion exposed that even a careful reader can misread `address_reference` as
    “should always be a derivation path,” which means the schema is still not carrying its own
    meaning clearly enough.

## Constraints (optional)

- Technical constraints:
  - Preserve existing allocation behavior, deterministic address generation, and public API shapes.
  - Stay within the current Clean Architecture boundaries under `internal/`.
  - Keep PostgreSQL and Cloudflare Postgres adapters aligned.
- Timeline/cost constraints:
  - Prefer a focused refactor that improves naming and persisted semantics without bundling broader
    product changes.
- Compliance/security constraints:
  - Do not weaken Ethereum CREATE2 privacy assumptions by converting internal secret-derived salt
    material into public-index-like semantics.

## Problem statement

- Current pain:
  - `derivation_index` implies every allocation is fundamentally a derivation-path concept, even
    though Ethereum CREATE2 uses it only as an internal reserved slot.
  - `address_source_ref` and `address_reference` are still too generic to distinguish address-space
    identity from per-address reconstruction metadata when reading the database.
  - `address_reference` currently stores mixed-shape values: Bitcoin uses an HD path payload while
    Ethereum stores a prefixed CREATE2 reference string.
- Evidence or examples:
  - Bitcoin-issued rows persist absolute HD paths such as `m/84'/1'/0'/0/0`.
  - Ethereum CREATE2 rows persist strings such as
    `ethereum-mainnet-create2/0x6056ff5174c9fdf0e6798b5b6f7602dd9f5ce3c45f5df575be2f61f20451692e`.
  - The current code still keys uniqueness on `(address_policy_id, address_source_ref,
derivation_index)`, which is technically correct but semantically uneven across issuance
    methods.

## Goals

- G1:
  - Make allocation persistence self-explanatory by separating address-space identity, internal slot
    identity, issuance-reference kind, and issuance-reference payload.
- G2:
  - Keep Bitcoin HD and Ethereum CREATE2 issuance under one shared model without pretending both
    are derivation-path-based.
- G3:
  - Persist Ethereum issuance reference payloads as the actual CREATE2 salt rather than a
    prefix-plus-payload string.
- G4:
  - Limit this refactor to naming, persistence semantics, and internal wiring so behavior remains
    stable for existing flows.

## Non-goals (out of scope)

- NG1:
  - Changing public HTTP request or response shapes.
- NG2:
  - Replacing Ethereum CREATE2 issuance with HD-wallet issuance.
- NG3:
  - Splitting the existing `scheme` field into separate “issuance method” and “address format”
    columns in the same change.

## Assumptions

- A1:
  - `address_policy_id` continues to identify the configured issuance policy; this refactor does not
    redesign policy IDs.
- A2:
  - Existing rows can be migrated in place without data loss by renaming columns, adding one
    reference-kind column, and rewriting Ethereum CREATE2 issuance references to salt-only payloads.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Allocation rows are readable without chain-specific mental translation.
- Target:
  - Schema, internal types, and tests all use explicit address-space / slot / issuance-reference
    terminology, and focused Go verification plus migration/spec lint checks pass.
