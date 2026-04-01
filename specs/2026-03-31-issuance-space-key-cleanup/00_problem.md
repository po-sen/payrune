---
doc: 00_problem
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

# Issuance Space Key Cleanup - Problem & Goals

## Context

- Background:
  - Phase 1 is complete and stable: `sweep_material_json` exists on
    `address_policy_allocations`, operator recovery and ETH sweep depend only on that JSON, and the
    public allocation APIs still work.
  - Historical allocator behavior is source-aware: if one `address_policy_id` rotates from one BTC
    xpub / tpub or ETH CREATE2 source to another, the new source starts from a fresh cursor; if
    maintainers later switch back to the old source, allocation resumes from that old source's
    stored cursor.
  - The previous phase-2 draft simplified continuity to `address_policy_id` only and added an
    extra source-registry table. That made the model stricter but no longer matched the allocator's
    long-standing operational behavior.
- Users or stakeholders:
  - Payrune maintainers responsible for allocation correctness and migration safety.
  - Operators who want the DB to stay readable and only care about `sweep_material_json`.
- Why now:
  - Phase 2 should finish the cleanup without breaking source-aware cursor continuity and without
    introducing extra tables that only exist to recover state already implicit in the existing
    cursor/allocation schema.

## Constraints (optional)

- Technical constraints:
  - `sweep_material_json` must remain the only operator-facing JSON and its shape must not change.
  - Source-aware allocation continuity, slot uniqueness, and `(chain, address)` uniqueness must
    remain stable across fresh DBs and legacy upgrades.
  - Public allocation APIs must keep working.
- Timeline/cost constraints:
  - Prefer the smallest schema change that preserves historical behavior.
- Compliance/security constraints:
  - Operators must continue using the phase-1 DB-driven sweep workflow; no new manual recovery
    inputs may be introduced.

## Problem statement

- Current pain:
  - `issuance_ref_kind` / `issuance_ref` are still persisted even though phase 1 already made
    `sweep_material_json` the operator-facing recovery payload.
  - The recent draft added `address_policy_sources` plus startup guards to compensate for removing
    source-aware continuity from the main schema.
  - That extra registry table is unnecessary if cursor continuity simply stays keyed by the same
    source ref that allocations already used historically.
- Evidence or examples:
  - `address_policy_cursors` originally tracked one cursor per
    `(address_policy_id, account_public_key/address_source_ref/address_space_ref)`.
  - `address_policy_allocations` originally tracked the same source ref so failed reservations
    could reopen on the correct source.
  - Removing source ref from the main schema breaks the historical "rotate source, later switch
    back, resume old cursor" behavior unless a second registry or side table is introduced.
  - Operator recovery already proved that only `sweep_material_json` matters for operator
    workflows; internal cleanup should not change that.

## Goals

- G1:
  - Keep phase-1 recoverability exactly as-is: `sweep_material_json` remains the only
    operator-facing recovery payload, and ETH DB-driven sweep remains unchanged.
- G2:
  - Keep `address_policy_cursors` as the source-aware sequence-pointer table:
    `(address_policy_id, address_space_ref) -> next_index`.
- G3:
  - Keep `address_policy_allocations` as the only allocation table and do not introduce an extra
    registry / side table for source continuity.
- G4:
  - Keep only the minimum necessary internal allocation fields on `address_policy_allocations`:
    `address_space_ref`, `slot_index`, and `failure_reason`.
- G5:
  - Remove `issuance_ref_kind` and `issuance_ref` from the final schema.
- G6:
  - Preserve historical source-rotation behavior: a new source under the same policy starts from a
    fresh cursor, and switching back to an old source resumes that old source's cursor.
- G7:
  - Keep public BTC / ETH allocation APIs stable with no new 500-class regression.
- G8:
  - Enforce at the DB layer that every `issued` allocation row has non-null
    `sweep_material_json`.
- G9:
  - Keep the issued allocation entity and persistence contract aligned so replay / lookup does not
    return partially populated issued allocations.
- G10:
  - Fail fast on invalid bootstrap-time BTC xpub / tpub config instead of surfacing request-time
    500 responses.

## Non-goals (out of scope)

- NG1:
  - Changing `sweep_material_json` shape or operator recovery workflow.
- NG2:
  - Introducing a second readable JSON such as `issuance_space_json`.
- NG3:
  - Redesigning issuance methods, allocation semantics, or sweep behavior.
- NG4:
  - Changing public HTTP request or response shapes.
- NG5:
  - Forcing one `address_policy_id` to map to only one source for its whole lifetime.

## Assumptions

- A1:
  - A single `address_policy_id` may legitimately be used with multiple source refs over time.
- A2:
  - `address_space_ref` is the minimum row-level state needed to preserve source-aware cursor
    continuity without an extra table.
- A3:
  - `issuance_ref_kind` / `issuance_ref` are redundant as persisted columns once
    `sweep_material_json` exists.
- A4:
  - BTC xpub / tpub config is still loaded from bootstrap env and can be validated before serving
    traffic.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Final schema readability.
- Target:
  - `address_policy_allocations` keeps operator-facing data plus only the minimum necessary
    allocation-level internal fields; `address_policy_cursors` stays source-scoped and simple.
- Metric:
  - Extra internal-table count.
- Target:
  - `0` extra allocator tables beyond `address_policy_allocations` and `address_policy_cursors`.
- Metric:
  - Source-rotation continuity regressions.
- Target:
  - `0`; switching source starts a fresh cursor and switching back resumes the previous cursor for
    that source.
