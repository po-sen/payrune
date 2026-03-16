---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Account public key / xpub:
  - The configured public derivation key currently stored in `AddressDerivationConfig.AccountPublicKey`.
- Legacy fingerprint rows:
  - Historical cursor or allocation rows created before this change that only carry
    `xpub_fingerprint_*` columns.

## Out-of-scope behaviors

- OOS1:
  - Reconstructing the original xpub for every legacy fingerprint row.
- OOS2:
  - Adding a new admin or policy-management API.

## Functional requirements

### FR-001 - New allocation state keys by account public key directly

- Description:
  - Reservation, cursor, and reopen-failed logic must use the configured account public key/xpub as
    the active persistence key instead of fingerprint fields.
- Acceptance criteria:
  - [ ] New cursor rows are keyed by (`address_policy_id`, `account_public_key`).
  - [ ] New allocation rows persist `account_public_key`.
  - [ ] `address_policy_allocations.account_public_key` is `NOT NULL`.
  - [ ] Reopen-failed and reserve-fresh queries match rows by `address_policy_id` and
        `account_public_key`.
  - [ ] No active runtime SQL path requires `xpub_fingerprint_algo` or `xpub_fingerprint`.
- [ ] `address_policy_allocations` no longer exposes `xpub_fingerprint_algo` or
      `xpub_fingerprint` columns after the migration is applied.
- Notes:
  - Surviving allocation rows must all be fully xpub-backed; legacy null-key rows are not allowed.

### FR-002 - Domain and policy-reader code no longer model fingerprints

- Description:
  - The core domain and in-memory policy reader must treat the account public key as the only key
    material needed for allocation issuance.
- Acceptance criteria:
  - [ ] `AddressDerivationConfig` no longer exposes fingerprint fields.
  - [ ] `AddressIssuancePolicy.ValidateForAllocationIssuance` no longer requires fingerprint
        configuration.
  - [ ] `AddressPolicyReader` no longer computes or injects fingerprint values.
  - [ ] List/generate/allocate public API responses remain unchanged.
- Notes:
  - Existing account-public-key and derivation-path behavior must stay intact.

### FR-003 - Migration clears legacy process data and leaves a strict xpub-backed schema

- Description:
  - The migration may discard pre-production allocation process data so the final schema can enforce
    non-null xpub-backed rows without legacy compatibility logic.
- Acceptance criteria:
  - [ ] The migration clears `address_policy_allocations` and dependent payment-address process data
        before enforcing the final schema.
  - [ ] Schema changes add `account_public_key` as a required persistence field for active rows.
  - [ ] The first post-migration allocation for any (`address_policy_id`, `account_public_key`)
        starts from derivation index `0`.
  - [ ] Once a policy has xpub-backed rows, the same xpub continues from its own max index.
  - [ ] Once a policy has xpub-backed rows, a different xpub under the same policy starts from
        index `0`.
- Notes:
  - This intentionally prefers strict final invariants over preserving pre-production process state.

### FR-004 - Postgres adapter behavior stays aligned across runtimes

- Description:
  - The regular PostgreSQL adapter and the Cloudflare PostgreSQL adapter must implement the same
    xpub-backed keying semantics.
- Acceptance criteria:
  - [ ] `internal/adapters/outbound/persistence/postgres` uses the xpub-backed key path.
  - [ ] `internal/adapters/outbound/persistence/cloudflarepostgres` uses the same xpub-backed key
        path.
  - [ ] Targeted unit/integration tests cover policy-reader cleanup, domain cleanup, the new
        clean-state cursor seeding behavior, and the reserve-fresh insert shape.
  - Notes:
  - The repository should continue to compile and test cleanly across both runtimes.

## Non-functional requirements

- Reliability (NFR-001):
  - `go test ./...` must pass after the change.
- Security/Privacy (NFR-002):
  - The implementation must continue to use xpub/account-public-key data only and must not
    introduce private-key material.
- Observability (NFR-003):
  - Existing API error contracts and payment-address-status lookups by ID must remain unchanged.
- Maintainability (NFR-004):
  - Runtime code under `internal/` must not contain fingerprint helper functions or fingerprint
    field plumbing after the change.

## Dependencies and integrations

- External systems:
  - PostgreSQL schema and migration runner.
- Internal services:
  - `internal/domain/valueobjects`
  - `internal/domain/entities`
  - `internal/adapters/outbound/policy`
  - `internal/adapters/outbound/persistence/postgres`
  - `internal/adapters/outbound/persistence/cloudflarepostgres`
