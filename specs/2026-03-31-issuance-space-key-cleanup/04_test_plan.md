---
doc: 04_test_plan
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

# Issuance Space Key Cleanup - Test Plan

## Scope

- Covered:
  - Compatibility migration and final cleanup migration.
  - Final schema shape without an extra source-registry table.
  - Fresh DB migration path.
  - Legacy DB upgrade path with BTC and ETH issued rows.
  - Source-rotation continuity under one policy.
  - `issued`-row `sweep_material_json` DB enforcement.
  - Issued-allocation entity / persistence contract alignment.
  - Bootstrap fail-fast validation for invalid BTC xpub / tpub config.
  - Cursor continuity and slot continuity.
  - `(chain, address)` uniqueness preservation.
  - BTC / ETH allocation API smoke.
  - `sweep_material_json` stability and ETH sweep helper behavior stability.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002 / NFR-002
  - Steps:
    - Test allocation stores with focused coverage proving fresh reserve and reopen paths use
      `address_policy_id` plus `address_space_ref`.
  - Expected:
    - Runtime code uses source-aware continuity without extra tables.
- TC-002:
  - Linked requirements: FR-005 / NFR-001
  - Steps:
    - Re-run BTC and ETH sweep-material generation coverage after the cleanup.
  - Expected:
    - `sweep_material_json` content is unchanged for equivalent inputs.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004 / NFR-001, NFR-004
  - Steps:
    - Apply the full migration chain to a fresh DB.
  - Expected:
    - Final schema keeps `address_space_ref` on allocations/cursors, removes issuance-ref columns,
      and enforces the issued-row sweep-material invariant.
- TC-102:
  - Linked requirements: FR-001, FR-003, FR-004 / NFR-001, NFR-004
  - Steps:
    - Upgrade a legacy DB with existing BTC issued rows and ETH issued rows.
  - Expected:
    - Issued rows keep continuity, `sweep_material_json` is unchanged, and no old address is
      reissued.
- TC-103:
  - Linked requirements: FR-001, FR-003 / NFR-001, NFR-004
  - Steps:
    - Allocate with source A, rotate to source B under the same policy, then switch back to source
      A.
  - Expected:
    - Source B starts from a fresh cursor and source A resumes its previous cursor.
- TC-104:
  - Linked requirements: FR-005 / NFR-001
  - Steps:
    - Attempt to mark an allocation row `issued` while leaving `sweep_material_json` as `NULL`.
  - Expected:
    - PostgreSQL rejects the write.
- TC-105:
  - Linked requirements: FR-006 / NFR-001, NFR-005
  - Steps:
    - Run BTC and ETH allocation smoke tests against the upgraded schema.
  - Expected:
    - Allocation succeeds without 500-class regressions.
- TC-106:
  - Linked requirements: FR-007 / NFR-006
  - Steps:
    - Load issued allocations through Postgres and Cloudflare stores after the issuance-ref columns
      are removed.
  - Expected:
    - Returned issued allocations match the final persisted shape and do not rely on zero-valued
      legacy issuance-ref fields.
- TC-107:
  - Linked requirements: FR-008 / NFR-001, NFR-006
  - Steps:
    - Bootstrap API/container setup with an invalid BTC xpub / tpub for an enabled policy.
  - Expected:
    - Startup fails with a descriptive validation error before any request is served.

## Edge cases and failure modes

- Case:
  - A failed reservation exists for source A while the policy is currently configured to source B.
- Expected behavior:
  - Reopen filtering stays on source A; the failed row is not silently reopened under source B.
- Case:
  - ETH helper runs after the cleanup.
- Expected behavior:
  - Helper still succeeds because it depends only on `sweep_material_json`.
- Case:
  - Cleanup down migration is attempted after issuance-ref columns were dropped.
- Expected behavior:
  - Migration fails fast with an explicit irreversible-cleanup error instead of reconstructing fake
    payloads.
