---
doc: 00_problem
spec_date: 2026-03-06
slug: write-through-receipt-tracking
mode: Full
status: DONE
owners:
  - payrune-team
depends_on: []
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
  - Current poller cycle executes `RegisterMissingIssued` every run, using `INSERT ... SELECT` from `address_policy_allocations` before claiming due rows.
- Users or stakeholders:
  - Platform/backend team maintaining polling performance and correctness.
- Why now:
  - Per-cycle registration scan is unnecessary after allocation rows are issued and increases table-scan pressure with growth.

## Constraints (optional)

- Technical constraints:
  - Keep Clean Architecture boundaries: use case through ports, adapter SQL in outbound repo.
  - Keep existing API and poller runtime behavior compatible.
- Timeline/cost constraints:
  - Implement in current feature cycle without introducing new services.

## Problem statement

- Current pain:
  - Poller path does extra registration work each cycle and mixes registration concern into polling loop.
- Evidence or examples:
  - `RunReceiptPollingCycleUseCase` currently calls `RegisterMissingIssued` before `ClaimDue` on each run.

## Goals

- G1:
  - Move receipt-tracking registration to allocation issue transaction (write-through).
- G2:
  - Remove poller per-cycle registration scan and keep claiming/observation behavior unchanged.
- G3:
  - Backfill existing issued allocations without tracking rows through migration.

## Non-goals (out of scope)

- NG1:
  - Changing observer chain logic (Esplora/BTC adapter behavior).
- NG2:
  - Introducing policy-level configurable confirmations in this iteration.

## Assumptions

- A1:
  - Required confirmations defaults to `1` when network-specific env is not configured.
- A2:
  - Existing `payment_receipt_trackings` schema is kept; only data backfill SQL is added.

## Open questions

- Q1:
  - None for this iteration.

## Success metrics

- Metric:
  - Poller cycle no longer executes allocation-to-tracking registration query.
- Target:
  - `RunReceiptPollingCycleUseCase` transaction path only claims due rows (no register step).
- Metric:
  - Allocation issue always has corresponding tracking row.
- Target:
  - New issued allocation can be claimed by poller without relying on periodic backfill.
