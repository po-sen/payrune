---
doc: 00_problem
spec_date: 2026-03-09
slug: shared-tip-height-polling
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

# Shared Tip Height Polling - Problem & Goals

## Context

- Background:
  - The Bitcoin receipt observer currently fetches Esplora tip height inside every address observation.
  - Receipt polling processes addresses in batches, so the same network tip height is repeatedly requested within one poll cycle.
- Users or stakeholders:
  - Backend maintainers preparing the Bitcoin payment poller for production cost control.
- Why now:
  - The user wants to reduce third-party API calls before enabling the system in production.

## Constraints (optional)

- Technical constraints:
  - Preserve current receipt status behavior and confirmation counting.
  - Keep clean architecture boundaries explicit; the poller use case should own per-cycle orchestration.
  - Do not change persistent models or API contracts.
- Timeline/cost constraints:
  - Prefer a focused optimization over a broader observer redesign.
- Compliance/security constraints:
  - None.

## Problem statement

- Current pain:
  - Each address observation fetches the same latest block height even when multiple tracked addresses share the same network in the same poll cycle.
  - This creates avoidable external API calls and provider cost.
- Evidence or examples:
  - `run_receipt_polling_cycle_use_case.go` loops through claimed receipt trackings one by one.
  - `esplora_receipt_observer.go` calls `/blocks/tip/height` inside every `ObserveAddress` call.

## Goals

- G1:
  - Fetch latest block height once per chain/network per poll cycle instead of once per address.
- G2:
  - Keep observer confirmation logic correct by passing the shared tip height into address observation.
- G3:
  - Preserve existing receipt lifecycle behavior, error mapping, and status transitions.

## Non-goals (out of scope)

- NG1:
  - Replacing Esplora with another Bitcoin backend.
- NG2:
  - Redesigning address discovery to use txid-centric tracking.
- NG3:
  - Removing `LastObservedBlockHeight` from the domain or API models.

## Assumptions

- A1:
  - A poll cycle may claim addresses from multiple networks, so the optimization should cache tip height per chain/network pair.
- A2:
  - One additional capability on the observer port for fetching latest block height is acceptable because it remains part of receipt observation infrastructure.

## Open questions

- Q1:
  - None for this scope.

## Success metrics

- Metric:
  - Within one poll cycle, the poller fetches tip height at most once per claimed chain/network pair.
- Target:
  - Repeated address observations on the same network reuse one shared latest block height while producing the same receipt updates as before.
