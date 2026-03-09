---
doc: 00_problem
spec_date: 2026-03-09
slug: mempool-compose-defaults
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

# Mempool Compose Defaults - Problem & Goals

## Context

- Background:
  - The compose defaults for Bitcoin Esplora endpoints are inconsistent today.
  - `testnet4` defaults to `mempool.space`, while `mainnet` defaults to `blockstream.info`.
- Users or stakeholders:
  - Backend maintainers preparing a clearer production/dev configuration baseline.
- Why now:
  - The user wants the public Esplora defaults to be consistent.

## Constraints (optional)

- Technical constraints:
  - Keep the current Esplora-compatible adapter behavior unchanged.
  - Limit the change to default configuration and related tests.
- Timeline/cost constraints:
  - Prefer a small refactor with no behavior redesign.
- Compliance/security constraints:
  - None.

## Problem statement

- Current pain:
  - The compose defaults suggest two different public providers without a technical reason visible in the code.
- Evidence or examples:
  - `compose.bitcoin.mainnet.yaml` uses `https://blockstream.info/api`.
  - `compose.bitcoin.testnet4.yaml` uses `https://mempool.space/testnet4/api`.

## Goals

- G1:
  - Use one consistent public Esplora default provider across mainnet and testnet4 compose files.
- G2:
  - Align poller config tests with the new default.

## Non-goals (out of scope)

- NG1:
  - Changing production recommendations to require public endpoints.
- NG2:
  - Modifying the Bitcoin observer implementation.

## Assumptions

- A1:
  - `https://mempool.space/api` is an acceptable mainnet default because it is Esplora-compatible and publicly reachable.

## Open questions

- Q1:
  - None for this scope.

## Success metrics

- Metric:
  - Compose defaults for Bitcoin public Esplora endpoints are consistent across supported networks.
- Target:
  - Mainnet and testnet4 defaults both use `mempool.space`, and related config tests pass.
