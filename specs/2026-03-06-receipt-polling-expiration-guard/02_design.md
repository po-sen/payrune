---
doc: 02_design
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

# Technical Design

## High-level approach

- Add `expires_at` lifecycle field and `failed_expired` status.
- Set initial expiry at issue-time registration.
- In polling cycle: if row is expired, mark terminal and skip observer.
- Extend expiry only when status transitions to `paid_unconfirmed`.
- Replace hardcoded expiry constants with env-driven configuration loaded in DI and injected into use cases.
- Keep transition-based expiry rule in domain entity to avoid business rule leakage in use-case orchestration.
- Separate poll scheduling (`next_poll_at`) from claim lease (`lease_until`).
- Keep a single constructor API for polling use case (`NewRunReceiptPollingCycleUseCase`) with explicit config argument.
- Remove terminal-status (`paid_confirmed`/`failed_expired`) special next-poll scheduling branches since these states are not claimed again.
- Inline persistence save helpers (`savePollingError` / `saveObservation`) into `Execute` to reduce private method surface and keep flow explicit.

## Key flows

- Flow 1 (issue):
  - Allocation issue -> register tracking with required confirmations + initial `expires_at` from configured default duration.
- Flow 2 (poll):
  - Claim due rows with expired/empty lease (including deadline-crossing rows) -> set `lease_until` claim window -> check expiration -> mark expired or observe -> save, clear lease, then extend expiry only on transition to `paid_unconfirmed`.

## Data model

- `payment_receipt_trackings`:
  - add `expires_at TIMESTAMPTZ NOT NULL`
  - add `lease_until TIMESTAMPTZ NULL`
  - status check adds `failed_expired`
  - index for active-expiry rows
  - index for active lease lookup
- Backfill:
  - `expires_at = COALESCE(issued_at, created_at) + interval '7 days'`

## Failure modes and resiliency

- Expired rows are terminal (`failed_expired`) and excluded from active polling statuses.
- Observer failures still use existing retry/error path.

## Observability

- Poll cycle `failed` counter includes expired transitions.
- No new external API surfaces required.

## Configuration contract

- App container:
  - `BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER`: duration for mainnet issued tracking expiry.
  - `BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER`: duration for testnet4 issued tracking expiry.
- Poller container:
  - `PAYMENT_RECEIPT_PAID_UNCONFIRMED_EXPIRY_EXTENSION`: duration used when status becomes `paid_unconfirmed`.
- Validation:
  - Empty values use defaults.
  - Non-positive or invalid duration values fail startup.
