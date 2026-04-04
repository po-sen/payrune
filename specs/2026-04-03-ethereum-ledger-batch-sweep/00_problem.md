---
doc: 00_problem
spec_date: 2026-04-03
slug: ethereum-ledger-batch-sweep
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-sweep-material-redesign
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
  - The ETH CREATE2 flow should have one deployed singleton contract per network, one checked-in
    metadata record per network, and one sweep path regardless of whether a receiver is already
    deployed.
  - The current implementation still carries compatibility branches for older recovery paths and
    older factory payloads, which makes the contract and the sweep script harder to reason about
    than necessary.
- Users or stakeholders:
  - Operators recovering ETH from many CREATE2 receiver addresses while keeping Ledger as the only
    signer.
- Why now:
  - This repo is not live yet, so we can drop legacy compatibility instead of preserving every old
    local development path.
  - Operators want one-time initialization, one active factory per network, and one boring batch
    sweep command.

## Constraints (optional)

- Technical constraints:
  - Ledger interactive signing is mandatory.
  - No hot wallet, private-key-based broadcaster, or unattended signer may be introduced.
  - Keep current `sweep_material_json` as the operator-facing recovery payload.
- Timeline/cost constraints:
  - Prefer a small, explicit operator workflow over a general automation framework.
- Compliance/security constraints:
  - Safety and operator reviewability take precedence over maximum throughput.

## Problem statement

- Current pain:
  - The current implementation exposes more than one recovery path, which means the contract and
    script have legacy branches that are harder to review.
  - Operators do not want to think about old factory variants or old recovery entry points; they
    want one current factory per network and one supported sweep path.
- Evidence or examples:
  - The current contract exposes both general deployment primitives and more than one sweep-style
    recovery entry point.
  - The current script contains branching for deployed vs undeployed receivers and old vs current
    factories.

## Goals

- G1:
  - Let operators deploy one Ethereum CREATE2 singleton contract per network and use that same
    contract for batch recovery.
- G2:
  - Preserve the current security model: explicit operator selection, Ledger-only signing, dry-run
    first, and no hot wallet.
- G3:
  - Keep the flow boring and reviewable: one deploy script, one sweep script, explicit metadata,
    and no generic cross-chain abstraction.
- G4:
  - Remove CREATE2 operator helper paths that depend on non-Ledger signing so the repo has one
    clear security stance.
- G5:
  - Make the sweep helper safe for predicted CREATE2 addresses that have balance before deployment:
    one operator command should still recover them correctly through the same contract entry point.
- G6:
- Keep no legacy operator compatibility code for superseded local-development factories or old
  recovery paths.

## Non-goals (out of scope)

- NG1:
  - Removing Ledger interaction or adding any private-key signer path.
- NG2:
  - Auto-sweeping all eligible DB rows without explicit operator selection.
- NG3:
  - Redesigning `sweep_material_json`.
- NG4:
  - Supporting recovery through superseded local-development factories once a new active factory is
    deployed for the same network.

## Assumptions

- A1:
  - Operators are willing to update checked-in CREATE2 metadata after a real network deployment if
    the deployment script can do that automatically.
- A2:
  - Reverting the whole batch when one receiver call fails is safer than partial success for the
    first version.
- A3:
  - Because the repo is not live, simplifying to one active factory per network is preferable to
    preserving legacy local data.

## Open questions

- None. This spec intentionally excludes non-Ledger alternatives and removes the existing
  non-Ledger helper path.

## Success metrics

- Metric:
  - One operator invocation can sweep more than one issued Ethereum CREATE2 allocation with one
    Ledger-signed transaction through the deployed CREATE2 factory contract.
- Target:
  - Batch flow supports at least 2 receiver addresses in one transaction.
- Metric:
  - Operators only need one deployed CREATE2 singleton address per network for issuance metadata and
    batch sweep recovery.
- Target:
  - No separate batch-caller address or second singleton contract remains in the operator flow.
- Metric:
  - A selected Ethereum CREATE2 address with ETH balance but no deployed receiver code can still be
    recovered in one Ledger-signed transaction.
- Target:
  - The recovery flow deploys missing receivers and sweeps them through the singleton factory
    without requiring a manual rescue command.
- Metric:
  - The operator flow has one canonical recovery entry point and no legacy fallback branch.
- Target:
  - The contract and the sweep script expose one supported batch recovery path only.
- Metric:
  - No non-Ledger signer path exists in the CREATE2 operator workflow.
- Target:
  - Zero private-key or hot-wallet options remain in CREATE2 operator scripts/tooling.
