---
doc: 03_tasks
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

# Issuance Space Key Cleanup - Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - This change rewrites late-stage migrations, changes persistence shape, and requires
    fresh/legacy validation across real PostgreSQL states.

## Tasks (ordered)

1. T-001 - Update and lint the phase-2 spec

   - Scope:
     - Record the final model: source-aware continuity stays on allocations/cursors, redundant
       issuance-ref columns are dropped, and `sweep_material_json` stays the only operator-facing
       JSON.
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, NFR-001, NFR-002, NFR-003, NFR-004, NFR-005
   - Validation:
     - [x] `SPEC_DIR="specs/2026-03-31-issuance-space-key-cleanup" bash scripts/spec-lint.sh`

1. T-002 - Simplify migrations back to the source-aware model

   - Scope:
     - Remove the extra policy-source registry design, keep `address_space_ref` on allocations and
       cursors, add the issued-row sweep-material invariant, drop only redundant issuance-ref
       columns in the cleanup step, and rename the phase-2 migration files/messages so they match
       the actual final cleanup.
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, NFR-001, NFR-003, NFR-004, NFR-006
   - Validation:
     - [x] Real PostgreSQL fresh DB and legacy upgrade checks pass.

1. T-003 - Restore source-aware allocation store behavior

   - Scope:
   - Make Postgres and Cloudflare Postgres reserve/reopen logic key by
     `(address_policy_id, address_space_ref)` again, remove the startup source-guard helpers,
     ensure bootstrap no longer references those removed helpers, and align the issued allocation
     entity/lookup contract with the final schema by dropping stale issuance-ref persistence from
     runtime allocation rows.
   - Linked requirements: FR-001, FR-003, FR-004, FR-006, FR-007, NFR-001, NFR-002, NFR-005, NFR-006
   - Validation:
     - [x] `go test ./internal/bootstrap ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
     - [x] `go test ./internal/application/usecases/...`
     - [x] `go build ./cmd/...`

1. T-004 - Add bootstrap fail-fast validation for invalid BTC xpub / tpub config

   - Scope:
     - Validate enabled BTC issuance policies during API/bootstrap setup for both native API and
       Cloudflare API worker so malformed source config fails before public traffic is served, and
       keep that validation co-located with the API policy-building code instead of a standalone
       bootstrap helper file.
   - Linked requirements: FR-006, FR-008, NFR-001, NFR-006
   - Validation:
     - [x] `go test ./internal/bootstrap/...`
     - [x] Invalid BTC source config fails during bootstrap with a descriptive error.

1. T-005 - Re-run sweep / API validation

   - Scope:
     - Prove that phase-1 recoverability and BTC / ETH allocation flows still behave correctly,
       including source-rotation continuity.
   - Linked requirements: FR-003, FR-005, FR-006, FR-008, NFR-001, NFR-004
   - Validation:
     - [x] `go test ./...`
     - [x] `bash -n scripts/ethereum_create2_sweep.sh`
     - [x] Real PostgreSQL fresh/legacy/API smoke validation

1. T-006 - Run repo validation and close the spec
   - Scope:
     - Run repo-wide validation and update the spec package to final `DONE` state.
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, NFR-001, NFR-002, NFR-003, NFR-004, NFR-005, NFR-006
   - Validation:
     - [x] `bash scripts/precommit-run.sh`

## Validation evidence

- `SPEC_DIR="specs/2026-03-31-issuance-space-key-cleanup" bash scripts/spec-lint.sh`
- `go test ./internal/bootstrap ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
- `go test ./internal/application/usecases/...`
- `go build ./cmd/...`
- `go test ./internal/bootstrap ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
- `go test ./...`
- `bash -n scripts/ethereum_create2_sweep.sh`
- `bash scripts/precommit-run.sh`
- Bootstrap config validation:
  - Invalid enabled BTC xpub / tpub now fails during API/bootstrap setup with a descriptive error
    that includes the policy ID, the env key, and the underlying parse failure.
- Fresh PostgreSQL migration:
  - Full `go run ./cmd/migrate up` succeeded on an empty DB.
  - Final schema keeps `address_space_ref` on `address_policy_allocations` and
    `address_policy_cursors`.
  - Final schema no longer keeps `issuance_ref_kind` or `issuance_ref` on
    `address_policy_allocations`.
  - PostgreSQL rejected `issued` allocation inserts with `NULL` `sweep_material_json` via
    `chk_address_policy_allocations_issued_sweep_material`.
- Legacy PostgreSQL upgrade:
  - Legacy BTC rows for `xpub-a` and `xpub-b` under the same policy upgraded in place.
  - Source-aware cursor continuity remained:
    `bitcoin-mainnet-native-segwit/xpub-a -> 7`,
    `bitcoin-mainnet-native-segwit/xpub-b -> 3`,
    `ethereum-sepolia-create2/create2.v1:... -> 5`.
  - A `derivation_failed` row kept `address_space_ref='xpub-a'` after upgrade.
  - `sweep_material_json` text for issued rows remained unchanged across the upgrade.

## Rollout and rollback

- Migration sequencing:
  - Step 1: apply compatibility migration only.
  - Step 2: deploy the new binary that stops persisting `issuance_ref_kind` / `issuance_ref`.
  - Step 3: apply cleanup migration that drops those redundant columns.
- Rollback:
  - Do not apply cleanup before the new binary is fully deployed.
  - After cleanup, rollback requires restoring from backup or a new forward repair migration.
