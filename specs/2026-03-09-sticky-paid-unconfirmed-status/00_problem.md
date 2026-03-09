---
doc: 00_problem
spec_date: 2026-03-09
slug: sticky-paid-unconfirmed-status
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-polling-expiration-guard
  - 2026-03-08-payment-address-status-api
  - 2026-03-06-receipt-webhook-delivery
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Sticky Paid-Unconfirmed Status - Problem & Goals

## Context

- Background:
  - Receipt polling currently derives the next status only from the latest observation totals.
  - `PaidAt` is sticky once the expected amount has ever been observed, but the status is not sticky.
- Users or stakeholders:
  - Backend maintainers preparing Bitcoin receipt tracking for production-grade semantics.
- Why now:
  - The current state machine allows a receipt that was already seen as fully paid to fall back into unpaid statuses and to expire later, which does not match the user's desired semantics.

## Constraints (optional)

- Technical constraints:
  - Preserve clean architecture boundaries: sticky payment rules belong in domain/policy logic, not in SQL or HTTP adapters.
  - Update persistence constraints, APIs, and webhook-compatible status values consistently.
  - Because the project has not reached `0.1.0`, breaking status/config changes are acceptable.
- Timeline/cost constraints:
  - Prefer a focused state-machine change over a larger receipt-tracking redesign.
- Compliance/security constraints:
  - None.

## Problem statement

- Current pain:
  - A receipt can move from `paid_unconfirmed` back to `watching` or `partially_paid` if the latest observation drops below the expected amount.
  - The same receipt can later become `failed_expired` even though a full payment was already observed once.
- Evidence or examples:
  - `PaymentReceiptTracking.ApplyObservation` sets `PaidAt` once and never clears it.
  - `decidePaymentReceiptStatus(...)` still returns `watching` for `ObservedTotalMinor == 0` and `partially_paid` for sub-threshold observations.
  - `ExpireIfDue(...)` currently only checks `expires_at`, regardless of whether full payment was already observed.

## Goals

- G1:
  - Make fully-paid observation sticky so a receipt never returns to unpaid statuses after `PaidAt` has been set.
- G2:
  - Introduce an explicit regression status for receipts that were fully paid unconfirmed but later lose that observation.
- G3:
  - Ensure receipts that have ever been fully paid do not fail by payment-window expiry.
- G4:
  - Update API/webhook-visible status contracts and persistence constraints consistently.
- G5:
  - Remove obsolete `paid_unconfirmed` expiry-extension configuration once the payment-window-expiry rule no longer applies after full payment is seen.
- G6:
  - Remove the unused `double_spend_suspected` status and simplify the receipt status model.
- G7:
  - Rename the poller reschedule env so the scheduling intent is explicit and consistent with poller terminology.
- G8:
  - Remove `ConflictTotalMinor` entirely so the receipt model only exposes data the current observer actually produces and uses.

## Non-goals (out of scope)

- NG1:
  - Introducing provider-specific mempool heuristics or tx-level reconciliation.
- NG2:
  - Redesigning webhook payload structure beyond the new status value.
- NG3:
  - Changing how confirmed payments become terminal.

## Assumptions

- A1:
  - The new regression status will be named `paid_unconfirmed_reverted` for explicitness.
- A2:
  - `PaidAt != nil` is the correct invariant for "a full payment was observed at least once".
- A3:
  - `ConflictTotalMinor` can be removed safely because the current observer always produces zero and no business rule depends on it anymore.

## Open questions

- Q1:
  - None for this scope.

## Success metrics

- Metric:
  - Receipts that were once fully paid never return to `watching` or `partially_paid`.
- Target:
  - Regressed observations after `PaidAt` transition to the new explicit regression status instead.
- Metric:
  - Payment-window expiry only affects receipts that have never been fully paid.
- Target:
  - Rows with `PaidAt != nil` no longer transition to `failed_expired` via the poller expiry path.
- Metric:
  - External status consumers understand the new state.
- Target:
  - Status API/OpenAPI and webhook-related status parsing support the final simplified status set without `double_spend_suspected`.
- Metric:
  - Poller scheduling config is self-descriptive.
- Target:
  - `POLL_RESCHEDULE_INTERVAL` replaces `RECEIPT_POLL_INTERVAL` in runtime config and Compose defaults.
- Metric:
  - Receipt polling and webhook contracts expose no dead conflict field.
- Target:
  - Status API, webhook payloads, events, ports, and persisted schema no longer include `ConflictTotalMinor`.
