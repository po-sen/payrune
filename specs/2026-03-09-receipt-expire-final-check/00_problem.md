---
doc: 00_problem
spec_date: 2026-03-09
slug: receipt-expire-final-check
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-05-blockchain-receipt-polling-service
  - 2026-03-09-sticky-paid-unconfirmed-status
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
  - Receipt polling currently treats `expires_at` as an immediate due condition and can mark a
    tracking as `failed_expired` before running the final blockchain observation for that poll
    cycle.
- Users or stakeholders:
  - Payrune operators and developers who rely on receipt polling to make the last payment-window
    decision correctly.
- Why now:
  - The desired behavior is to decide expiry only when the scheduled poll actually runs, after that
    poll has finished checking whether a late payment record exists.

## Constraints (optional)

- Technical constraints:
  - Keep the existing receipt tracking schema and polling architecture.
  - Preserve sticky paid semantics introduced for `paid_unconfirmed` and
    `paid_unconfirmed_reverted`.
- Timeline/cost constraints:
  - Prefer a targeted flow change over introducing a new scheduler or migration.
- Compliance/security constraints:
  - Not applicable.

## Problem statement

- Current pain:
  - `ClaimDue` can bypass `next_poll_at` when `expires_at <= now`, so expiry handling is tied to
    the claim query rather than the scheduled poll cadence.
  - The poll use case can call `ExpireIfDue` before fetching the latest observation, so the system
    may mark `failed_expired` without a final payment check.
- Evidence or examples:
  - A tracking with `next_poll_at` still in the future but `expires_at` already passed can be
    claimed immediately.
  - A tracking whose payment arrives near the deadline should be checked one final time before it is
    marked expired.

## Goals

- G1:
  - Only evaluate payment-window expiry when a tracking becomes due by `next_poll_at`.
- G2:
  - Run the normal observation flow first, then decide whether the tracking should become
    `failed_expired`.
- G3:
  - Keep observation failures as retryable processing errors rather than terminal expiry.

## Non-goals (out of scope)

- NG1:
  - Changing payment status names or sticky paid semantics.
- NG2:
  - Adding new tables, columns, or migrations.
- NG3:
  - Changing poll cadence defaults.

## Assumptions

- A1:
  - If the final observation succeeds and still shows an unpaid or underpaid tracking past
    `expires_at`, that cycle may mark it `failed_expired`.
- A2:
  - If the final observation fails, the system should not expire the tracking because it did not
    complete the last payment check.
- A3:
  - `failed_expired` remains terminal and non-pollable.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Final expiry decision timing.
- Target:
  - Expiry is evaluated only inside a due poll cycle, after the last observation attempt finishes.
- Metric:
  - False expiry risk near deadline.
- Target:
  - A successful final observation that finds payment keeps the tracking out of `failed_expired`.
