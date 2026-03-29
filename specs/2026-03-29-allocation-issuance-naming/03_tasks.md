---
doc: 03_tasks
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

# Allocation Issuance Naming - Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - This change adds a schema migration, renames core persistence fields, and updates issued-address
    semantics across adapters, so Quick mode would under-specify the data-model impact.
- Upstream dependencies (`depends_on`):
  - 2026-03-20-create2-eth-payment-receiving
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`

## Milestones

- M1:
  - Land the spec and migration design.
- M2:
  - Implement the rename across schema, code, and tests with no behavior regression.

## Tasks (ordered)

1. T-001 - Update the spec package and lock the rename scope
   - Scope:
     - Define the address-space / slot / issuance-reference vocabulary and record what this change
       intentionally does not redesign.
   - Output:
     - Completed Full-mode spec package for this refactor.
   - Linked requirements: FR-001, FR-002, FR-003, NFR-001
   - Validation:
     - [ ] How to verify (manual steps or command):
           `SPEC_DIR="specs/2026-03-29-allocation-issuance-naming" bash scripts/spec-lint.sh`
     - [ ] Expected result:
           Spec lint passes with consistent frontmatter, links, and dependency state.
     - [ ] Logs/metrics to check (if applicable):
           N/A
2. T-002 - Add the allocation schema migration and adapter SQL updates
   - Scope:
     - Rename persistence columns, add `issuance_ref_kind`, rewrite Ethereum CREATE2 issued rows to
       salt-only payloads, and update SQL queries/indexes.
   - Output:
     - New migration pair plus updated PostgreSQL and Cloudflare Postgres allocation stores/tests.
   - Linked requirements: FR-001, FR-002, FR-003, NFR-002, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command):
           `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
     - [ ] Expected result:
           Store tests pass and assert the renamed columns plus issuance-reference kind behavior.
     - [ ] Logs/metrics to check (if applicable):
           N/A
3. T-003 - Rename core models and deriver outputs to the new vocabulary
   - Scope:
     - Update entities, value objects, ports, use cases, bootstrap wiring, and Bitcoin/Ethereum
       derivers to use address-space / slot / issuance-reference names.
   - Output:
     - Consistent internal terminology and typed issuance-reference outputs.
   - Linked requirements: FR-001, FR-002, FR-003, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command):
           `go test ./internal/domain/... ./internal/application/... ./internal/adapters/outbound/bitcoin ./internal/adapters/outbound/ethereum ./internal/bootstrap`
     - [ ] Expected result:
           The refactored packages compile and pass focused tests with the new field names.
     - [ ] Logs/metrics to check (if applicable):
           N/A

## Traceability (optional)

- FR-001 -> T-001, T-002, T-003
- FR-002 -> T-001, T-002, T-003
- FR-003 -> T-001, T-002, T-003
- NFR-001 -> T-001, T-003
- NFR-002 -> T-002, T-003
- NFR-003 -> T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Apply the new migration before deploying code that expects renamed columns.
- Rollback steps:
  - Revert the deploy and run the migration down step to restore the previous column names and
    Ethereum CREATE2 prefixed reference format.

## Validation evidence

- 2026-03-29:
  - `go test ./...`
  - `SPEC_DIR="specs/2026-03-29-allocation-issuance-naming" bash scripts/spec-lint.sh`
  - `bash scripts/precommit-run.sh`
  - Result: passed
- 2026-03-30:
  - `go test ./cmd/ethereum-create2-tool ./...`
  - `bash -n scripts/ethereum_create2_verify_chain.sh`
  - Result: passed
