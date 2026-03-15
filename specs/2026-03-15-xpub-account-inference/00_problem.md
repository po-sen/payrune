---
doc: 00_problem
spec_date: 2026-03-15
slug: xpub-account-inference
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background:
  - Payment address issuance persists a Bitcoin `derivation_path` for each allocated address.
  - The current path builder combines a configured `DerivationPathPrefix` such as `m/84'/0'/0'`
    with a relative path derived from the xpub such as `0/11`.
- Users or stakeholders:
  - payrune maintainers and wallet operators who need persisted derivation paths to match the real
    xpub account that produced the address.
- Why now:
  - The current implementation hard-codes account `0'` in config and does not infer the actual
    account index from account-level xpubs.

## Constraints (optional)

- Technical constraints:
  - The change must stay within the existing Bitcoin address derivation flow and current schema.
- Timeline/cost constraints:
  - None.
- Compliance/security constraints:
  - No private keys may be introduced; only xpub metadata may be inspected.

## Problem statement

- Current pain:
  - When the configured prefix says `.../0'` but the supplied account-level xpub actually belongs to
    another account, the issued address can still be correct while the persisted `derivation_path`
    is wrong.
- Evidence or examples:
  - The current DI configuration hard-codes prefixes like `m/84'/0'/0'`, and the Bitcoin deriver
    currently emits `0/{index}` for account-level xpubs instead of reflecting the xpub's hardened
    account child index.

## Goals

- G1:
  - Persist derivation paths that reflect the actual account index encoded in account-level xpubs.
- G2:
  - Preserve current address derivation behavior for account-level and branch-level xpub inputs.
- G3:
  - Keep the fix localized to the existing Bitcoin outbound deriver and allocation flow.

## Non-goals (out of scope)

- NG1:
  - Changing the HTTP/API contract for address generation or allocation.
- NG2:
  - Adding a new wallet identifier model or changing database schema.

## Assumptions

- A1:
  - Configured `DerivationPathPrefix` values remain the source for purpose and coin-type path
    segments.
- A2:
  - Account-level xpubs expose their child index through BIP32 metadata and branch-level xpubs
    continue to rely on configured account prefixes.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Persisted Bitcoin derivation paths match the xpub account metadata for account-level xpubs.
- Target:
  - Targeted and full Go tests pass, and new tests cover non-zero account-level xpub derivation
    paths without requiring a hard-coded `0'` account.
