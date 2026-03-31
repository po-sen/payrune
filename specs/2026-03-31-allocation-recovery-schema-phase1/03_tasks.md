---
doc: 03_tasks
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

# Allocation Recovery Schema Phase 1 - Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - This change adds schema migrations, issued-row backfill, new chain-specific recovery payloads,
    and a new operator ETH sweep helper with explicit failure-mode handling.
- Upstream dependencies (`depends_on`):
  - 2026-03-29-allocation-issuance-naming
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`

## Milestones

- M1:
  - Lock phase-1 scope and land the additive schema changes.
- M2:
  - Persist / backfill sweep material, add the DB-driven ETH helper, and verify allocation APIs
    stay stable.

## Deferred to phase 2

- Any change to `address_space_ref`, `slot_index`, cursor partitioning, or allocation continuity.
- Any attempt to remove internal compatibility fields or normalize the whole partitioning schema.
- Any second operator-facing JSON or broader schema beautification.

## Tasks (ordered)

1. T-001 - Write and lint the phase-1 safe-refactor spec package
   - Scope:
     - Capture the exact phase-1 boundary, prohibited allocation/cursor changes, and the operator
       recovery acceptance criteria.
   - Output:
     - Completed Full-mode spec package under
       `specs/2026-03-31-allocation-recovery-schema-phase1/`.
   - Linked requirements: FR-001, FR-003, FR-004, FR-005, FR-006, NFR-001, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command):
           `SPEC_DIR="specs/2026-03-31-allocation-recovery-schema-phase1" bash scripts/spec-lint.sh`
     - [ ] Expected result:
           Spec lint passes with Full-mode links, dependency state, and consistent frontmatter.
     - [ ] Logs/metrics to check (if applicable):
           N/A
2. T-002 - Add the additive `sweep_material_json` schema migration
   - Scope:
     - Add the new JSON column only, without touching cursor keys, indexes, or allocation
       continuity semantics.
   - Output:
     - Migration pair that adds/removes `sweep_material_json`.
   - Linked requirements: FR-001, NFR-001, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command):
           `go test ./cmd/migrate ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
     - [ ] Expected result:
           Migration-aware persistence tests compile and pass with the new additive column.
     - [ ] Logs/metrics to check (if applicable):
           N/A
3. T-003 - Persist and read sweep material in the issuance path
   - Scope:
     - Update derivers, entities, use case wiring, and allocation stores so new issued rows write
       and read back `sweep_material_json`.
   - Output:
     - Updated BTC and ETH issued-address derivers plus allocation entities / stores carrying
       `sweep_material_json`.
   - Linked requirements: FR-002, FR-006, NFR-001, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command):
           `go test ./internal/domain/... ./internal/application/... ./internal/adapters/outbound/bitcoin ./internal/adapters/outbound/ethereum ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
     - [ ] Expected result:
           Both chains persist sweep material for new issued rows, and focused store/use-case tests
           pass.
     - [ ] Logs/metrics to check (if applicable):
           N/A
4. T-004 - Backfill existing issued rows
   - Scope:
     - Populate `sweep_material_json` for existing issued BTC and ETH rows using current persisted
       issuance data, without changing slot/cursor continuity.
   - Output:
     - Backfill migration pair for existing issued rows only.
   - Linked requirements: FR-003, NFR-001, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command):
           `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
     - [ ] Expected result:
           Focused persistence coverage passes and backfill SQL compiles with the existing schema
           shape.
     - [ ] Logs/metrics to check (if applicable):
           N/A
5. T-005 - Add a DB-driven ETH CREATE2 sweep helper
   - Scope:
     - Create a shell helper that selects one issued allocation row from the DB, loads
       `sweep_material_json`, validates env and Ledger sender identity, and assembles or broadcasts
       the `sweep()` call without manual factory / collector / salt inputs.
   - Output:
     - New script under `scripts/` and any supporting CLI adjustments needed for operator usage.
   - Linked requirements: FR-004, NFR-002, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command):
           `bash -n scripts/ethereum_create2_sweep.sh`
     - [ ] Expected result:
           Script syntax is valid and dry-run mode can build the sweep command from DB-selected
           `sweep_material_json`.
     - [ ] Logs/metrics to check (if applicable):
           Confirm the script prints selector, target receiver, and dry-run/broadcast mode.
6. T-006 - Update operator-facing docs and usage
   - Scope:
     - Update README / relevant docs / script usage so operator recovery points only at
       `sweep_material_json`.
   - Output:
     - Documentation updates for DB inspection and ETH sweep usage.
   - Linked requirements: FR-005, NFR-003, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command):
           `rg -n "issuance_ref|issuance_ref_kind|address_space_ref" README.md internal/infrastructure/ethereumcreate2assets/README.md scripts/ethereum_create2_sweep.sh`
     - [ ] Expected result:
           Docs no longer instruct operators to use internal compatibility fields for this workflow.
     - [ ] Logs/metrics to check (if applicable):
           N/A
7. T-007 - Verify BTC and ETH allocation APIs still allocate correctly
   - Scope:
     - Run focused tests covering both payment-address allocation endpoints and their underlying use
       cases/controllers.
   - Output:
     - Passing focused API regression evidence for both chains.
   - Linked requirements: FR-006, NFR-001
   - Validation:
     - [ ] How to verify (manual steps or command):
           `go test ./internal/application/usecases ./internal/adapters/inbound/http/controllers ./internal/bootstrap`
     - [ ] Expected result:
           BTC and ETH allocation-path tests continue to pass after the sweep-material changes.
     - [ ] Logs/metrics to check (if applicable):
           N/A
8. T-008 - Run repo validation and close the spec
   - Scope:
     - Run the repo validation flow after implementation, then update the spec package to `DONE`
       with validation evidence.
   - Output:
     - Validation evidence in this task doc plus final spec status update.
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, NFR-001, NFR-002, NFR-003, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command):
           `bash scripts/precommit-run.sh`
     - [ ] Expected result:
           Repo validation passes, and all produced spec docs are updated consistently.
     - [ ] Logs/metrics to check (if applicable):
           Record the final command list and result under validation evidence.

## Traceability (optional)

- FR-001 -> T-001, T-002, T-008
- FR-002 -> T-003, T-008
- FR-003 -> T-001, T-004, T-008
- FR-004 -> T-001, T-005, T-008
- FR-005 -> T-001, T-006, T-008
- FR-006 -> T-003, T-007, T-008
- NFR-001 -> T-001, T-002, T-003, T-004, T-007, T-008
- NFR-002 -> T-005, T-008
- NFR-003 -> T-001, T-002, T-003, T-004, T-006, T-008
- NFR-004 -> T-005, T-006, T-008

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Apply the additive column migration first, deploy code that writes / reads `sweep_material_json`,
    then apply the backfill migration before operator adoption of the new sweep helper.
- Rollback steps:
  - Stop using the new sweep helper.
  - Roll back the backfill migration if needed.
  - Roll back the additive column migration only after reverting code that reads / writes
    `sweep_material_json`.

## Validation evidence

- 2026-03-31:
  - `SPEC_DIR="specs/2026-03-31-allocation-recovery-schema-phase1" bash scripts/spec-lint.sh`
  - `go test ./internal/domain/... ./internal/application/... ./internal/adapters/outbound/bitcoin ./internal/adapters/outbound/ethereum ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
  - `go test ./cmd/migrate ./internal/adapters/inbound/http/controllers ./internal/bootstrap`
  - `bash -n scripts/ethereum_create2_sweep.sh`
  - `PATH="<tmpdir>:$PATH" DATABASE_URL=postgres://example ETHEREUM_SWEEP_PAYMENT_ADDRESS_ID=145 ETHEREUM_SWEEP_RPC_URL=https://rpc.example ETHEREUM_SWEEP_FROM_ADDRESS=0x1111111111111111111111111111111111111111 ETHEREUM_SWEEP_DERIVATION_PATH="m/44'/60'/0'/0/0" bash scripts/ethereum_create2_sweep.sh`
  - `bash scripts/precommit-run.sh`
  - Result: passed
