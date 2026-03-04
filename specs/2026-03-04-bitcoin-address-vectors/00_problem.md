---
doc: 00_problem
spec_date: 2026-03-04
slug: bitcoin-address-vectors
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-btc-xpub-address-api
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: We need deterministic verification of bitcoin address derivation against known xpub vectors.
- Users or stakeholders: Backend developers maintaining derivation logic and release safety.
- Why now: A recent derivation-path bug showed we need fixed vectors to catch regressions early.

## Constraints (optional)

- Technical constraints:
  - Keep tests as unit tests in `internal/adapters/outbound/bitcoin`.
  - Fixture vectors are hardcoded in test code and not read from env.
- Timeline/cost constraints:
  - Implement as a small follow-up without changing API contracts.
- Compliance/security constraints:
  - Use xpub only (no private key material).

## Problem statement

- Current pain:
  - Existing tests verify types and deterministic behavior but not known external vectors from wallet outputs.
- Evidence or examples:
  - Mainnet/testnet4 index-0 addresses were provided and need exact-match validation.

## Goals

- G1: Keep all provided vector fixtures hardcoded in test code.
- G2: Add a make target to run vector verification unit tests consistently.
- G3: Validate provided mainnet/testnet4 vectors for `legacy`, `segwit`, `nativeSegwit`, and `taproot` at `index=0`.
- G4: Keep vector assertions split by encoder test file for easier maintenance per address scheme.

## Non-goals (out of scope)

- NG1: Add production runtime xpub validation policy.
- NG2: Add DB storage for vectors or runtime address snapshots.

## Assumptions

- A1: Provided vectors are account-level or compatible with current derivation semantics.
- A2: Local test runs should work without extra env setup.

## Open questions

- Q1: Should future vectors include non-zero indices?
- Q2: Should we add additional cross-wallet vectors in future revisions?

## Success metrics

- Metric: Deterministic vector verification coverage.
- Target: Unit test validates 8 provided vectors with exact-match results.
- Metric: Developer ergonomics.
- Target: `make test-address-vectors` runs without additional shell setup.
