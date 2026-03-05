---
doc: 02_design
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

# Technical Design

## High-level approach

- Summary:
  - Move tracking registration from poller pre-claim phase to allocation issue transaction.
  - Keep poller focused on `claim -> observe -> save`.
- Key decisions:
  - Add a single-allocation registration repository method.
  - Remove dead poller fields tied to legacy register path (`DefaultRequiredConfirmations`, `RegisteredCount`).
  - Add one-time migration backfill in a dedicated `000004` migration.
  - Resolve required confirmations at allocation issue time via network-scoped config (`mainnet` vs `testnet4`).

## System context

- Components:
  - `AllocatePaymentAddressUseCase`
  - `RunReceiptPollingCycleUseCase`
  - Postgres `PaymentReceiptTrackingRepository`
- Interfaces:
  - `PaymentReceiptTrackingRepository.RegisterIssuedAllocation(...)`

## Key flows

- Flow 1:
  - Allocation reserve + derive + `Complete` -> resolve required confirmations by network -> `RegisterIssuedAllocation` in same UoW tx.
- Flow 2:
  - Poller cycle -> `ClaimDue` -> observer -> `SaveObservation` / `SavePollingError`.

## Data model

- Entities:
  - `payment_receipt_trackings` unchanged.
- Schema changes or migrations:
  - Add `000004_backfill_payment_receipt_trackings.up.sql` for one-time backfill `INSERT ... SELECT ... ON CONFLICT DO NOTHING`.
- Consistency and idempotency:
  - Write-through and backfill both use conflict-safe insert to avoid duplicate rows.

## API or contracts

- Endpoints or events:
  - No external API contract changes.
  - Add internal env contract for app DI:
    - `BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS`
    - `BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS`

## Backward compatibility (optional)

- API compatibility:
  - Poller output DTO drops obsolete `RegisteredCount` field.
- Data migration compatibility:
  - Existing rows are preserved; only missing rows are added.

## Failure modes and resiliency

- Retries/timeouts:
  - Poller retry behavior unchanged.
- Degradation strategy:
  - If allocation-time registration fails, transaction fails and no partial issued state leaks.

## Observability

- Logs:
  - Poll cycle summary log reports `claimed`, `updated`, `failed`.
- Metrics:
  - Remove obsolete `registered` counter from cycle output.

## Security

- Secrets:
  - No new secrets or credentials introduced.

## Alternatives considered

- Option A:
  - Keep per-cycle `RegisterMissingIssued` scan.
- Option B:
  - Write-through registration with one-time backfill.
- Why chosen:
  - Lower recurring query cost and cleaner use-case responsibility split.

## Risks

- Risk:
  - Missing tracking repo in tx repositories can break allocation issue flow.
- Mitigation:
  - Add explicit nil-guard and test coverage.
