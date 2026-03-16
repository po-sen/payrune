---
doc: 03_tasks
spec_date: 2026-03-16
slug: remove-xpub-fingerprint
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-04-policy-payment-address-allocation
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - The change modifies database schema, persistence keys, and migration/rollback behavior.
- Upstream dependencies (`depends_on`):
  - `2026-03-04-policy-payment-address-allocation`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - Not applicable.
  - What would trigger switching to Full mode:
    - Already Full because schema and migration behavior change.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Remove fingerprint modeling from domain and policy-reader code.
- M2:
  - Migrate persistence and allocation stores to use `account_public_key`, then drop the legacy
    fingerprint columns.
- M3:
  - Verify strict non-null schema, clean-state cursor seeding, adapter parity, and full repository
    health.

## Tasks (ordered)

1. T-001 - Remove fingerprint plumbing from core code
   - Scope:
     - Update domain value objects/entities, policy reader, and affected tests so the active model
       uses only account public key plus derivation path prefix.
   - Output:
     - Runtime code no longer carries fingerprint fields or helper functions.
   - Linked requirements: FR-002, NFR-002, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command): run
           `go test ./internal/domain/... ./internal/adapters/outbound/policy ./internal/application/usecases`
     - [ ] Expected result: domain, policy-reader, and use-case tests pass without fingerprint
           fields.
     - [ ] Logs/metrics to check (if applicable): none
2. T-002 - Migrate active persistence keys to account public key
   - Scope:
     - Add the PostgreSQL migration and update both PostgreSQL allocation-store adapters to reserve,
       reopen, and seed cursors by `account_public_key`, clear pre-production allocation process
       data, and remove legacy fingerprint columns from `address_policy_allocations`.
   - Output:
     - New cursor/allocation writes use xpub-backed keys only, `address_policy_allocations` enforces
       non-null `account_public_key`, and legacy process rows are gone from the live schema.
   - Linked requirements: FR-001, FR-003, FR-004, NFR-001, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command): run
           `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
           and execute a real allocation request against PostgreSQL-backed API runtime.
     - [ ] Expected result: store tests pass and cursor seeding behavior matches the migration
           design, reserve-fresh SQL executes successfully against a real database, and migrated
           allocations cannot have null `account_public_key`.
     - [ ] Logs/metrics to check (if applicable): none
3. T-003 - Run full verification and capture rollout constraints
   - Scope:
     - Run spec lint, repository-wide Go tests, and record the migration/rollback caveats in the
       spec before marking the work complete.
   - Output:
     - Final spec and code are aligned, and verification evidence is captured.
   - Linked requirements: FR-004, NFR-001, NFR-003, NFR-004
   - Validation:
     - [ ] How to verify (manual steps or command): run
           `SPEC_DIR="specs/2026-03-16-remove-xpub-fingerprint" bash scripts/spec-lint.sh`,
           `go list ./...`, and `go test ./...`
     - [ ] Expected result: spec lint passes, package graph is clean, and the repository test suite
           is green.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-002
- FR-002 -> T-001
- FR-003 -> T-002
- FR-004 -> T-002, T-003
- NFR-001 -> T-002, T-003
- NFR-002 -> T-001
- NFR-003 -> T-002, T-003
- NFR-004 -> T-001, T-003

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Apply the single merged migration `000010` before deploying binaries that expect the final
    xpub-backed schema.
- Rollback steps:
  - If no new allocations were written after deployment, run the down migration and redeploy the
    previous binaries.
  - If new xpub-backed allocations were written, use a database snapshot restore for a fully
    lossless rollback because old fingerprint values cannot be reconstructed automatically.
  - The migration is destructive to pre-production allocation-related data by design.

## Validation evidence

- 2026-03-17:
  - `SPEC_DIR="specs/2026-03-16-remove-xpub-fingerprint" bash scripts/spec-lint.sh`
  - `go list ./...`
  - `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
  - `go test ./...`
  - `bash scripts/precommit-run.sh`
  - Disposable PostgreSQL smoke:
    - Migrate a disposable database to v9 using only `000001` to `000009`
    - Insert legacy rows into `address_policy_allocations`, `payment_receipt_trackings`,
      `payment_address_idempotency_keys`, `payment_receipt_status_notifications`, and
      `address_policy_cursors`
    - Apply `000010` and verify row counts become `0|0|0|0|0`
    - Verify `address_policy_allocations.account_public_key` has `is_nullable = NO`
    - Verify `schema_migrations.version = 10`
    - Verify `down -> up` still returns the schema to v10 cleanly
