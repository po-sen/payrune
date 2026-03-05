---
doc: 00_problem
spec_date: 2026-03-04
slug: policy-payment-address-allocation
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-btc-xpub-address-api
  - 2026-03-03-postgresql18-migration-runner-container
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Policy-Based Payment Address Allocation - Problem & Goals

## Context

- Background:
  - Customer allocation API must allocate by `addressPolicyId` without exposing derivation index.
  - Address sequence must reset when xpub rotates under the same policy.
  - Allocation records must carry enough lifecycle data for reconciliation and backtracking.
  - During implementation, several architecture micro-refactors (UoW/repository naming, adapter placement, directory shape) were split into multiple small specs.
- Users or stakeholders:
  - Merchant backend teams requesting payment addresses.
  - Payrune backend maintainers and reconciliation workflows.
- Why now:
  - Feature delivery is complete, but spec artifacts should be consolidated to one feature-level source of truth.

## Constraints (optional)

- Technical constraints:
  - Keep Go project layout and clean architecture boundaries.
  - Keep xpub-only derivation (no private keys).
  - Use non-hardened BIP32 index range (`0..2147483647`).
- Timeline/cost constraints:
  - Consolidation is documentation-focused; no new product scope.
- Compliance/security constraints:
  - No new key-handling scope.

## Problem statement

- Current pain:
  - Feature details are fragmented across multiple micro-spec folders, reducing readability and traceability.
- Evidence or examples:
  - Separate specs were created for naming cleanup, UoW contract shape, adapter rollback, and config-adapter removal.

## Goals

- G1:
  - Keep customer API index-free while preserving one-time unique allocation.
- G2:
  - Partition allocation sequence by (`addressPolicyId`, `xpubFingerprintAlgo`, `xpubFingerprint`) so xpub rotation starts at index `0`.
- G3:
  - Persist allocation lifecycle (`reserved`, `issued`, `derivation_failed`) with rich trace fields.
- G4:
  - Return stable `paymentAddressId` and required amount metadata for reconciliation.
- G5:
  - Ensure derivation failure does not permanently consume index.
- G6:
  - Keep architecture boundaries clean (use case -> ports only; UoW owns transaction lifecycle; repository outputs entities/aggregates only).
- G7:
  - Consolidate this entire feature and all architecture micro-adjustments into one spec package.

## Non-goals (out of scope)

- NG1:
  - On-chain spend detection and automatic address recycling.
- NG2:
  - Multi-chain expansion beyond current Bitcoin scope.
- NG3:
  - Quote/rate model introduction.

## Assumptions

- A1:
  - Once issued by API, address allocation is considered consumed and not re-issued.
- A2:
  - `addressPolicyId` is stable while xpub can rotate over time.
- A3:
  - Policy source remains DI-provided configuration for this phase.

## Open questions

- Q1:
  - Should we add idempotency key support for allocation retries in next phase?
- Q2:
  - Should we expose retrieval endpoint by `paymentAddressId` in follow-up spec?

## Success metrics

- Metric:
  - Duplicate allocation rate per (`addressPolicyId`, `xpubFingerprintAlgo`, `xpubFingerprint`).
- Target:
  - `0` duplicates across sequential and concurrent test runs.
- Metric:
  - Xpub rotation sequence behavior.
- Target:
  - First allocation after xpub change uses index `0`.
- Metric:
  - Traceability coverage.
- Target:
  - Each issued record stores policy/fingerprint/address/path/status/timestamps/amount/reference.
- Metric:
  - Spec consolidation.
- Target:
  - Single canonical spec folder documents feature + architecture refinements.
