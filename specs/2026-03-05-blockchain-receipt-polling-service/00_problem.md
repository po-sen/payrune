---
doc: 00_problem
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

# Blockchain Receipt Polling Service - Problem & Goals

## Context

- Background:
  - Address allocation was delivered, but on-chain collection state tracking needed a dedicated polling capability.
  - Receipt flow requires handling partial payments, multi-transaction top-up, confirmation progression, and conflict-risk signaling.
  - Allocated payment addresses are treated as clean at issuance, so receipt observation should focus on inbound activity after `issued_at` rather than global UTXO snapshot state.
  - Poller runtime now needs clear deployment isolation between `mainnet` and `testnet4`, while keeping core logic extensible for other chains.
- Users or stakeholders:
  - Merchant backend teams consuming payment status.
  - Reconciliation/risk workflows requiring auditable receipt lifecycle data.
  - Payrune operators maintaining blockchain poller services.
- Why now:
  - Production readiness required node-backed observation, explicit transaction boundaries, and network-scoped poller deployment.

## Constraints (optional)

- Technical constraints:
  - Keep Go clean architecture boundaries.
  - Isolate complex receipt lifecycle in dedicated table(s), not allocation table.
  - Avoid holding DB transactions across outbound RPC calls.
- Compliance/security constraints:
  - No private key handling.
  - RPC credentials must be environment-driven and not logged.

## Problem statement

- Current pain:
  - Without poller, issued addresses cannot progress into reliable collection states automatically.
  - A single unscoped poller deployment risks mixing network workloads.
- Evidence or examples:
  - Customer can pay in multiple transactions before reaching target amount.
  - Node/RPC failure needs retryable error state instead of hard-failing all work.

## Goals

- G1:
  - Introduce dedicated receipt-tracking persistence model and state machine.
- G2:
  - Provide independent poller microservice (`cmd/poller`) with periodic processing loop.
- G3:
  - Aggregate multi-transaction receipts and support partial/unconfirmed/confirmed transitions.
- G4:
  - Add conflict-risk status (`double_spend_suspected`) in domain lifecycle.
- G5:
  - Use explicit Unit of Work orchestration for poller DB operations.
- G6:
  - Integrate real Bitcoin observation using issue-time-scoped inbound receipt data (not current unspent snapshot scan).
- G7:
  - Support network-scoped dual pollers (`poller-mainnet`, `poller-testnet4`) running concurrently.
- G8:
  - Remove Bitcoin-specific type coupling from poller domain/application contracts so future chains can plug in via adapters.
- G9:
  - Keep local compose startup simple: `make up` should bring up both bitcoin network pollers and test env overrides by default.

## Non-goals (out of scope)

- NG1:
  - Implement ETH/TRON observer adapters in this iteration.
- NG2:
  - Treasury sweep/refund automation.
- NG3:
  - Full mempool-conflict model beyond current baseline.

## Assumptions

- A1:
  - Bitcoin endpoint URLs (Esplora-compatible API) are available in deployment environments.
- A2:
  - `issued_at` is always populated when an allocation transitions to `issued`.
- A3:
  - Receipt semantics only care about inbound value observed at/after allocation `issued_at`.

## Open questions

- Q1:
  - Which Bitcoin data source should be preferred for issue-time-scoped inbound transaction queries in production (bitcoind wallet/indexer/esplora)?
- Q2:
  - Should we expose read API by `paymentAddressId` in a follow-up spec?

## Success metrics

- Metric:
  - Polling freshness latency.
- Target:
  - `p95` status-update latency <= `poll_interval + 15s` in local environment.
- Metric:
  - Parallel processing correctness.
- Target:
  - `0` duplicate claim overlap across concurrent poller workers for same scope.
- Metric:
  - Network isolation correctness.
- Target:
  - mainnet poller only processes `network=mainnet`; testnet4 poller only processes `network=testnet4`.
- Metric:
  - Core decoupling quality.
- Target:
  - Adding one new chain observer requires no changes in receipt domain entities or polling use case orchestration.
- Metric:
  - Local startup ergonomics.
- Target:
  - `make up` renders compose config with `poller-mainnet`, `poller-testnet4`, and test-env xpub fixtures without extra override flags.
- Metric:
  - Issue-time scoping correctness.
- Target:
  - `100%` of validated receipt cases exclude inbound transactions earlier than allocation `issued_at`.
