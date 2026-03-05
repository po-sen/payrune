---
doc: 02_design
spec_date: 2026-03-05
slug: blockchain-receipt-polling-service
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-postgresql18-migration-runner-container
  - 2026-03-04-policy-payment-address-allocation
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Blockchain Receipt Polling Service - Technical Design

## High-level approach

- Summary:
  - Use dedicated receipt table + stateful domain transitions.
  - Run poller as independent binary with periodic cycle.
  - Integrate Bitcoin Esplora observer behind chain-routed observer composition.
  - Scope receipt observation to inbound activity at/after allocation `issued_at` (not global UTXO snapshot).
  - Deploy network-scoped dual pollers through compose overrides.
- Key decisions:
  - Poller DB interactions are transaction-scoped through UoW callbacks.
  - RPC observation executes outside DB transactions to avoid long lock windows.
  - Scope filters (`POLL_CHAIN`, `POLL_NETWORK`) are validated as generic chain/network identifiers and applied at register/claim query layer.
  - Receipt rows carry source `issued_at`; observer calls use it as lower-bound for amount aggregation.
  - Legacy base poller service removed; network pollers are explicit (`poller-mainnet`, `poller-testnet4`).
  - Poller domain/application contracts use chain-agnostic value objects (`ChainID`, `NetworkID`); Bitcoin-specific parsing is isolated in Bitcoin adapter.
  - Outbound observer contracts are split into multi-chain routing (`BlockchainReceiptObserver`) and chain-specific execution (`ChainReceiptObserver`) to avoid role ambiguity and redundant chain validation.
  - Makefile default compose overrides include test env + both bitcoin poller overrides so operators can use `make up` directly.

## System context

- Components:
  - Domain: `PaymentReceiptTracking`, `PaymentReceiptStatus`.
  - Application: `RunReceiptPollingCycleUseCase` (UoW + observer orchestration).
- Outbound adapters:
  - Postgres receipt repository.
  - Chain-routed observer (dispatches by chain).
  - Bitcoin inbound-transaction observer (Esplora-compatible endpoint) as one routed implementation.
  - Inbound runtime:
    - `cmd/poller` + `bootstrap.RunPoller`.
- Interfaces:
  - Internal worker process only (no new HTTP endpoint).

## Key flows

- Flow 1: Register + claim

  - Start UoW transaction.
  - Register missing issued rows with optional scope filters and carry allocation `issued_at` into tracking row.
  - Claim due rows using `FOR UPDATE SKIP LOCKED` and move `next_poll_at` to claim window.
  - Commit transaction.

- Flow 2: Observe + save success

  - Router selects chain observer by tracking `chain`.
  - Selected observer calls underlying node/RPC or index source for inbound transfers at/after row `issued_at` (and optional incremental cursor).
  - Domain applies observation and computes next status.
  - Save observation in separate UoW transaction.

- Flow 3: Observe failure

  - Mark row error and retry schedule in separate UoW transaction.

- Flow 4: Dual poller runtime
  - `poller-mainnet` and `poller-testnet4` run concurrently.
  - Scope filters prevent cross-network claiming.

## Data model

- Entities:
  - `payment_receipt_trackings` with lifecycle totals/status, timestamps, polling metadata, and issue-time observation floor (`issued_at`).
- Schema changes or migrations:
  - `000003_payment_receipt_trackings.{up,down}.sql` includes `issued_at` in receipt-tracking schema.
- Consistency and idempotency:
  - Unique `payment_address_id` prevents duplicate tracking row creation.
  - Claim path lock strategy prevents same-row concurrent claim overlap.

## API or contracts

- Observer input:
  - `chain`, `network`, `address`, `required_confirmations`, `issued_at`, `since_block_height`.
- Observer output:
  - `observed_total_minor`, `confirmed_total_minor`, `unconfirmed_total_minor`, `conflict_total_minor`, `latest_block_height`.

## Backward compatibility (optional)

- Existing HTTP API contracts remain unchanged.
- Existing allocation persistence remains source for tracking registration.

## Failure modes and resiliency

- Missing network endpoint:
  - Deterministic row-level error persisted, retry scheduled.
- Missing chain observer route:
  - Deterministic row-level error persisted, retry scheduled.
- RPC timeout/unreachable:
  - Row marked with `last_error`, cycle continues.
- Invalid poll scope env:
  - Poller startup/config validation fails fast (`network` cannot be provided without `chain`).

## Observability

- Poll-cycle summary logs: registered/claimed/updated/failed.
- Row-level failure logs via error propagation to save-error path.

## Security

- RPC credentials sourced from env; no secret persistence in DB.
- No private key handling introduced.

## Alternatives considered

- Option A:
  - Continue current UTXO snapshot scan baseline.
- Option B:
  - Chain-routed observer + network-scoped dual pollers with issue-time-scoped inbound transaction aggregation.
- Why chosen:
  - Better data correctness for clean addresses and avoids coupling state to current unspent snapshot.

## Risks

- Risk:
  - Inbound transaction source may differ across providers (RPC/indexer) and affect parsing consistency.
- Mitigation:
  - Keep adapter boundary strict and add contract tests for issue-time lower-bound behavior.

## Consolidation notes

- This spec supersedes and merges:
  - `2026-03-05-bitcoin-node-poller-uow`
  - `2026-03-05-network-scoped-dual-poller-compose`
