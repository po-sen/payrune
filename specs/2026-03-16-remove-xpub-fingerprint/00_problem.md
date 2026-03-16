---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - Payment address allocation currently partitions cursor and reservation state by
    `xpub_fingerprint_algo` and `xpub_fingerprint`.
  - The in-memory policy reader computes the fingerprint from the configured account public key
    before the allocation flow can run.
  - The original account public key already exists in policy configuration and is the real value
    operators reason about when debugging address issuance.
- Users or stakeholders:
  - payrune maintainers operating the allocation service and inspecting PostgreSQL state.
  - Wallet operators rotating account xpubs for the same `addressPolicyId`.
- Why now:
  - The current fingerprint model adds indirection without enough product value, and the user wants
    the database to store the xpub directly instead.

## Constraints (optional)

- Technical constraints:
  - Keep the existing Go layout and current Clean Architecture boundaries.
  - Keep public HTTP/API contracts stable.
  - Favor a strict final schema over preserving pre-production allocation history.
- Timeline/cost constraints:
  - No additional product scope beyond removing the active fingerprint feature.
- Compliance/security constraints:
  - Continue using xpub/account-public-key material only; no private keys may be introduced.

## Problem statement

- Current pain:
  - The system computes and stores a derived fingerprint even though the configured xpub already
    identifies the active derivation source.
  - Database records are harder to inspect because operators must map fingerprints back to xpubs.
  - Domain and adapter code carry fingerprint-specific validation, normalization, and SQL keying
    that are not needed for the current product shape.
- Evidence or examples:
  - `AddressDerivationConfig` and `AddressIssuancePolicy` both carry fingerprint fields.
  - `address_policy_cursors` and `address_policy_allocations` key reservation state by fingerprint
    columns rather than the stored account public key.
  - The policy reader computes `sha256-trunc64-hex-v1` fingerprints at runtime before allocation.

## Goals

- G1:
  - Remove `xpub_fingerprint` as an active runtime concept and key new allocation state directly by
    account public key/xpub.
- G2:
  - Simplify domain and policy-reader behavior by removing fingerprint-specific fields and
    validation.
- G3:
  - Preserve current external API behavior for listing policies, generating addresses, allocating
    payment addresses, and fetching payment-address status.
- G4:
  - Make the post-migration allocation schema strict so every surviving allocation row carries a
    non-null `account_public_key`.
- G5:
  - Remove the legacy fingerprint columns from `address_policy_allocations` once the xpub-backed
    model is in place.
- G6:
  - Clear pre-production allocation process data during migration so the final schema does not need
    legacy compatibility paths.

## Non-goals (out of scope)

- NG1:
  - Adding policy CRUD or a new metadata table to backfill historical fingerprint rows to xpubs.
- NG2:
  - Changing public request/response DTO shapes or route structure.
- NG3:
  - Retroactively recovering the original xpub for legacy rows that only store a fingerprint.
- NG4:
  - Preserving pre-production allocation, receipt-tracking, notification, or idempotency data across
    this migration.

## Assumptions

- A1:
  - `AddressPolicyId` remains stable while the configured account public key may rotate over time.
- A2:
  - Historical allocation rows may exist without a reversible mapping from fingerprint back to
    xpub.
- A3:
  - Storing account public keys in PostgreSQL text columns and indexes is acceptable for current
    Bitcoin-only usage.
- A4:
  - This environment may discard existing allocation-related PostgreSQL rows because the service is
    not formally in production yet.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Active runtime references to fingerprint-specific fields and helpers.
- Target:
  - `internal/` runtime code no longer depends on `PublicKeyFingerprint*` fields or fingerprint
    helper functions.
- Metric:
  - Strictness of surviving allocation rows after migration.
- Target:
  - Every row in `address_policy_allocations` has non-null `account_public_key`.
- Metric:
  - Persistence clarity for new rows.
- Target:
  - New allocation rows persist `account_public_key`, and new cursor rows key by
    (`address_policy_id`, `account_public_key`).
- Metric:
  - Legacy process data compatibility paths in active persistence logic.
- Target:
  - Allocation stores no longer need transitional seeding for legacy fingerprint-backed rows.
- Metric:
  - Legacy fingerprint columns in active schema.
- Target:
  - `address_policy_allocations` no longer contains `xpub_fingerprint_algo` or `xpub_fingerprint`.
