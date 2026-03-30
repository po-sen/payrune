---
doc: 00_problem
spec_date: 2026-03-30
slug: eth-balance-receipt-observer
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
  - 2026-03-30-eth-poller-stall-fix
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
  - Ethereum receipt polling for CREATE2 payment addresses needs to avoid long-running block scans.
- Users or stakeholders:
  - Operators running Ethereum pollers and developers maintaining receipt tracking correctness.
- Why now:
  - The previous block-scan model was too slow and stall-prone.

## Constraints

- Technical constraints:
  - Keep the existing public API and payment receipt lifecycle.
  - Stay on standard Ethereum JSON-RPC; do not require a third-party indexer.
  - Support the current v1 scope only: native ETH to CREATE2 payment addresses.
- Timeline/cost constraints:
  - The change should remain a bounded refactor inside the observer path.
- Compliance/security constraints:
  - Do not expose raw CREATE2 salts or full source references.

## Problem statement

- Current pain:
  - Block scanning is operationally expensive.
  - Standard Ethereum RPC does not offer Bitcoin/Esplora-style address transaction history.
- Evidence or examples:
  - Bitcoin can filter post-issuance inbound transactions by address history.
  - Ethereum can cheaply read balance snapshots at exact block tags, but not native-ETH inbound
    transaction history for an address.

## Goals

- G1:
  - Make ETH receipt observation O(1) per row without block scanning.
- G2:
  - Keep ETH receipt observation responsibility inside the poller/observer path.
- G3:
  - Keep provider requirements limited to `eth_blockNumber` and `eth_getBalance`.

## Non-goals

- NG1:
  - ERC-20 or internal ETH transfer observation.
- NG2:
  - General-purpose EVM indexing infrastructure.
- NG3:
  - Automatic receiver deployment or sweep orchestration.

## Assumptions

- A1:
  - ETH payment addresses are allocation-specific and intended for one receiving flow.
- A2:
  - For this v1 flow, current balance snapshots are an acceptable approximation for ETH receipt
    observation.
- A3:
  - Matching Bitcoin’s strict post-issuance semantics would require additional infrastructure such
    as an archive-capable provider or address-history indexer.

## Open questions

- Q1:
  - When deploy-and-sweep is added, should ETH receipt observation split into pre-collection and
    post-collection modes?
- Q2:
  - If stricter post-issuance totals become required later, should the project adopt an archive
    provider or a dedicated indexer?

## Success metrics

- Metric:
  - RPC work per observed ETH receipt row.
- Target:
  - One latest-height fetch per cycle scope and at most two balance queries per row.
- Metric:
  - ETH observer dependence on block scans.
- Target:
  - Zero `eth_getBlockByNumber` calls inside Ethereum `ObserveAddress`.
- Metric:
  - Allocation-flow coupling.
- Target:
  - No ETH-specific receipt-observation baseline capture in allocation use cases.
