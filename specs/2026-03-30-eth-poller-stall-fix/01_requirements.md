---
doc: 01_requirements
spec_date: 2026-03-30
slug: eth-poller-stall-fix
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Zero-total incremental scan:
- A safe optimization path where the Ethereum observer starts from `SinceBlockHeight + 1` only when the stored cumulative observed, confirmed, and unconfirmed totals are all zero.

## Out-of-scope behaviors

- OOS1: No generalized incremental reconciliation for non-zero Ethereum cumulative totals.
- OOS2: No new public API fields or new database columns.

## Functional requirements

### FR-001 - Unblock zero-total Ethereum poller rows with safe incremental scanning

- Description: The Ethereum receipt observer must stop doing repeated full-history rescans for rows that have already scanned to a prior block height while still holding zero cumulative totals.
- Acceptance criteria:
  - [ ] The observer input can receive the tracking row's current cumulative observed, confirmed, and unconfirmed totals.
  - [ ] When all three cumulative totals are zero and `SinceBlockHeight > 0`, the Ethereum observer starts scanning from `SinceBlockHeight + 1` instead of the issuance-derived starting block.
  - [ ] When `SinceBlockHeight >= LatestBlockHeight`, the observer returns the existing cumulative totals unchanged without scanning blocks again.
- Notes: This keeps the optimization correctness-constrained to rows where the stored cumulative state is still zero.

### FR-002 - Preserve existing correctness for non-zero cumulative Ethereum rows

- Description: The bug fix must not silently corrupt receipt totals for rows that already have non-zero cumulative Ethereum observations.
- Acceptance criteria:
  - [ ] For Ethereum rows with any non-zero cumulative total, the observer continues to use the full-history cumulative path.
  - [ ] Existing Bitcoin observer behavior remains unchanged.
- Notes: This intentionally trades performance for correctness outside the safe zero-total path.

### FR-003 - Emit a poll-cycle start log before slow observer work

- Description: The poller must log the start of each cycle before invoking the receipt polling handler so operators can tell that the container is active even when processing is slow.
- Acceptance criteria:
  - [ ] Each cycle prints one start log containing the chain, network, and batch size.
  - [ ] Existing success/failure cycle logs remain in place.
- Notes: This is an observability fix, not a behavioral contract change.

## Non-functional requirements

- Performance (NFR-001): Zero-total Ethereum rows with a prior `SinceBlockHeight` should avoid repeated rescans of already-checked blocks.
- Availability/Reliability (NFR-002): One slow Ethereum row must no longer keep the entire local Sepolia poller effectively silent for multiple cycles when a safe zero-total incremental path is available.
- Security/Privacy (NFR-003):
- Compliance (NFR-004):
- Observability (NFR-005): The poller must emit a start log before cycle processing begins.
- Maintainability (NFR-006): Regression tests must cover the new Ethereum zero-total incremental path and poller log behavior should remain easy to reason about from code review.

## Dependencies and integrations

- External systems:
- Internal services: `internal/adapters/outbound/ethereum`, `internal/application/ports/outbound`, `internal/application/usecases`, `internal/bootstrap`
