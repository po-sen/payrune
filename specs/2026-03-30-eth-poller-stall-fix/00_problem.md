---
doc: 00_problem
spec_date: 2026-03-30
slug: eth-poller-stall-fix
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background: The local Sepolia Ethereum poller process starts but can appear completely silent, while `ethereum/sepolia` receipt rows remain in `watching` without updated observed totals.
- Users or stakeholders: Operators validating Ethereum CREATE2 receipt tracking locally before rollout.
- Why now: The user has deployed a real Sepolia factory and needs the local `ethereum-sepolia-create2` flow to actually advance receipt tracking instead of stalling on old rows.

## Constraints (optional)

- Technical constraints: Keep the current receipt lifecycle and storage model; do not introduce new persistence tables in this fix.
- Timeline/cost constraints: Quick bug fix scoped to observer behavior, poller observability, and regression tests.
- Compliance/security constraints: Do not change public API shapes or require new operator secrets.

## Problem statement

- Current pain: The Sepolia poller can claim old `watching` rows and then spend so long rescanning Ethereum blocks that the cycle never visibly completes; while stuck, the container emits no progress logs and newer Sepolia rows do not get processed.
- Evidence or examples:
  - `docker logs payrune-poller-ethereum-sepolia-1` remains empty while the container is up.
  - Existing `ethereum/sepolia` rows have `lease_until` set but `observed_total_minor` and related fields remain unchanged.
  - The Ethereum RPC observer receives `SinceBlockHeight` but currently ignores it, rescanning from `issued_at` every time.

## Goals

- G1: Make Sepolia receipt polling advance again for rows that already have zero cumulative totals and a previously scanned block height.
- G2: Ensure the poller prints at least a cycle-start log before doing potentially slow work so operators can distinguish "stuck" from "not started".

## Non-goals (out of scope)

- NG1: Redesign the entire Ethereum receipt observer into a fully general incremental scanner for all historical non-zero states.
- NG2: Introduce new deploy-and-sweep automation or change CREATE2 issuance semantics.

## Assumptions

- A1: The immediate stall is driven by old Sepolia rows whose stored cumulative totals are still zero, so a safe zero-total incremental optimization is enough to unblock local validation.
- A2: For non-zero cumulative Ethereum rows, correctness is more important than optimization, so a full rescan fallback remains acceptable in this patch.

## Open questions

- Q1: None.
- Q2: None.

## Success metrics

- Metric: Whether the local Sepolia poller resumes emitting logs and updating due `ethereum/sepolia` receipt rows.
- Target: After the fix, the poller emits a cycle-start log immediately, and zero-total Sepolia rows with a prior `last_observed_block_height` no longer stall the whole cycle on repeated full-history rescans.
