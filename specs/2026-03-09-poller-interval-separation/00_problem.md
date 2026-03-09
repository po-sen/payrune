---
doc: 00_problem
spec_date: 2026-03-09
slug: poller-interval-separation
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

# Poller Interval Separation - Problem & Goals

## Context

- Background:
  - The poller currently uses one `POLL_INTERVAL` value for both worker wake-up cadence and receipt `next_poll_at` rescheduling.
  - These two behaviors have different meanings: worker tick controls how often due rows are claimed, while receipt reschedule controls how long one address waits before becoming due again.
- Users or stakeholders:
  - Backend maintainers preparing the Bitcoin payment poller for production rollout and cost control.
- Why now:
  - The current configuration hides two independent knobs behind one env var, which makes production tuning harder and more error-prone.
  - The related poller Compose env blocks have also become harder to scan as scheduling and provider settings grew.

## Constraints (optional)

- Technical constraints:
  - Keep the poller architecture simple; do not introduce status-aware scheduling in this change.
  - Preserve current receipt lifecycle and observer behavior.
  - Because the project has not reached `0.1.0`, backward compatibility with `POLL_INTERVAL` is not required.
- Timeline/cost constraints:
  - Prefer a focused config and orchestration refactor.
- Compliance/security constraints:
  - None.

## Problem statement

- Current pain:
  - `POLL_INTERVAL` is overloaded and changes two different runtime behaviors at once.
  - Operators cannot tune worker cadence separately from address polling cadence.
- Evidence or examples:
  - `internal/bootstrap/poller.go` uses `config.Interval` for the ticker.
  - `internal/application/use_cases/run_receipt_polling_cycle_use_case.go` uses the same value to calculate `next_poll_at`.

## Goals

- G1:
  - Separate worker tick interval from receipt reschedule interval in configuration and code.
- G2:
  - Remove the legacy `POLL_INTERVAL` input so only the explicit new env names remain.
- G3:
  - Make the poller wiring and naming reflect the distinct meanings of the two intervals.
- G4:
  - Keep the poller Compose env blocks readable by grouping related settings in a stable order.

## Non-goals (out of scope)

- NG1:
  - Introducing status-specific polling intervals.
- NG2:
  - Changing receipt status transition rules.
- NG3:
  - Redesigning the poller scheduler or persistence model.

## Assumptions

- A1:
  - Keeping current default durations is safer than silently changing production polling cadence in this refactor.

## Open questions

- Q1:
  - None for this scope.

## Success metrics

- Metric:
  - Poller runtime can configure worker wake-up cadence and receipt reschedule cadence independently.
- Target:
  - `POLL_TICK_INTERVAL` affects only the worker ticker, `RECEIPT_POLL_INTERVAL` affects only `next_poll_at`, and `POLL_INTERVAL` is no longer accepted by the poller config.
