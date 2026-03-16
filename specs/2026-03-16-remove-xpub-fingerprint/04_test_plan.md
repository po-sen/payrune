---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Domain and policy-reader cleanup after removing fingerprint fields.
  - PostgreSQL and Cloudflare PostgreSQL allocation-store behavior with xpub-backed keys.
  - Destructive migration behavior that clears pre-production allocation process data.
  - Live schema verification that the legacy fingerprint columns are dropped and
    `account_public_key` is required.
- Not covered:
  - A full SQL backfill from historical fingerprints to original xpub values.
  - Production rollback rehearsal with real historical data snapshots.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-002, NFR-002, NFR-004
  - Steps:
    - Run domain tests covering `AddressDerivationConfig` normalization and
      `AddressIssuancePolicy.ValidateForAllocationIssuance`.
  - Expected:
    - Domain types normalize and validate without fingerprint fields or fingerprint-specific errors.
- TC-002:
  - Linked requirements: FR-002, NFR-004
  - Steps:
    - Run `internal/adapters/outbound/policy` tests.
  - Expected:
    - The policy reader returns issuance policies without computing or comparing fingerprints.
- TC-003:
  - Linked requirements: FR-001, FR-003, FR-004, NFR-001
  - Steps:
    - Run allocation-store tests for both PostgreSQL adapters, including clean-state seed behavior
      and same-xpub continuation after rows exist.
  - Expected:
    - Fresh reservation uses `account_public_key`, and cursor seeding no longer carries legacy
      compatibility branches.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-004, NFR-001, NFR-003
  - Steps:
    - Run:
      - `go test ./internal/application/usecases`
      - `go test ./internal/adapters/outbound/persistence/postgres`
      - `go test ./internal/adapters/outbound/persistence/cloudflarepostgres`
      - `go list ./...`
      - Execute `POST /v1/chains/bitcoin/payment-addresses` against a PostgreSQL-backed API runtime.
  - Expected:
    - Allocation use cases and both persistence adapters compile and behave consistently with the
      new xpub-backed model, and a real reserve-fresh insert succeeds without SQL shape errors.
- TC-102:
  - Linked requirements: FR-001, FR-003, FR-004, NFR-001
  - Steps:
    - Run `go test ./...`
    - Inspect `address_policy_allocations` in a migrated database.
  - Expected:
    - The repository remains green after the fingerprint removal, and
      `address_policy_allocations` no longer contains fingerprint columns and does not allow null
      `account_public_key`.

### E2E (if applicable)

- Scenario 1:
  - Apply the new migration to a disposable database with pre-existing legacy rows and verify the
    allocation-related tables are cleared before new xpub-backed rows are written.
- Scenario 2:
  - Allocate with one xpub after migration, rotate to a different xpub under the same policy, and
    verify the new xpub starts at index `0`.

## Edge cases and failure modes

- Case:
  - The migration runs against a pre-production database that already has allocation rows and
    dependent receipt/idempotency data.
- Expected behavior:
  - The migration clears that process data, and the next allocation starts from a clean xpub-backed
    state.
- Case:
  - A policy already has xpub-backed rows for xpub A and then rotates to xpub B.
- Expected behavior:
  - Xpub B gets its own cursor and starts at index `0`.
- Case:
  - A historical `derivation_failed` row existed before the migration.
- Expected behavior:
  - The row is removed by the destructive migration and cannot be reopened afterward.

## NFR verification

- Reliability:
  - `go test ./...` passes.
- Security:
  - No private-key material is introduced.
- Observability:
  - Existing status lookup and API error contracts remain unchanged.
- Maintainability:
  - Runtime fingerprint plumbing is removed from `internal/`.
