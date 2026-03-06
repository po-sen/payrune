---
doc: 00_problem
spec_date: 2026-03-06
slug: receipt-polling-expiration-guard
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-write-through-receipt-tracking
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
  - Receipt polling currently needs an explicit stop rule to avoid indefinite retries.
- Users or stakeholders:
  - Platform/backend team operating poller at scale.
- Why now:
  - Without expiry control, stale addresses can keep consuming claim/observe capacity.

## Problem statement

- Current pain:
  - A tracking row may remain active forever if no payment ever completes.
- Evidence or examples:
  - Active statuses (`watching`, `partially_paid`, `paid_unconfirmed`, `double_spend_suspected`) are repeatedly eligible for polling.
  - Current claim flow reuses `next_poll_at` for temporary claim lock and final schedule, which mixes two different concerns.

## Goals

- G1:
  - Introduce `expires_at` as lifecycle deadline for receipt tracking rows.
- G2:
  - Mark expired rows to terminal `failed_expired` and stop observing them.
- G3:
  - Extend expiry only when status transitions into `paid_unconfirmed`, preventing indefinite extension on unchanged states.
- G4:
  - Make expiry windows configurable via environment variables for app and poller deployments.
- G5:
  - Separate scheduling (`next_poll_at`) from claim lease (`lease_until`) to improve poller semantics and multi-worker safety.

## Non-goals (out of scope)

- NG1:
  - Introducing poll-attempt-based hard fail policy.
- NG2:
  - Refactoring blockchain observer provider integration.

## Assumptions

- A1:
  - Initial issue-time expiry defaults to `issued + 7 days`.
- A2:
  - Existing rows can be safely backfilled using `COALESCE(issued_at, created_at) + interval '7 days'`.

## Success metrics

- Metric:
  - Expired active rows no longer trigger observer calls.
- Target:
  - Poller transitions expired rows once to `failed_expired`.
- Metric:
  - Lifecycle remains resilient for late/partial payments.
- Target:
  - Transition to `paid_unconfirmed` receives one extension window without repeated extension when status stays unchanged.
- Metric:
  - Operators can tune expiry windows per deployment without code changes.
- Target:
  - Compose mainnet/testnet4 files expose dedicated env values for initial expiry and status-based extensions.
- Metric:
  - Claim lock and scheduling are independently controlled.
- Target:
  - `ClaimDue` sets `lease_until` only, while save paths clear `lease_until` and persist final `next_poll_at`.
