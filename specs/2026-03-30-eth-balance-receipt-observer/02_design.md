---
doc: 02_design
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

# Technical Design

## High-level approach

- Summary:
  - Replace Ethereum block scanning with exact block-tagged balance snapshots.
- Key decisions:
  - Keep the change inside the Ethereum observer path.
  - Do not add ETH-specific baseline capture or schema changes.
  - Accept documented ETH snapshot semantics for the current v1 flow.

## System context

- Components:
  - ETH receipt observer
  - Receipt polling use case
- Interfaces:
  - ETH observe: observed / confirmed / unconfirmed totals at exact heights

## Key flows

- Flow 1: ETH polling from balance snapshots

  - Poller fetches latest block height once per scope.
  - ETH observer queries current balance at `latestBlockHeight`.
  - ETH observer queries confirmed balance at `latestBlockHeight - requiredConfirmations + 1` when
    positive.
  - ETH observer returns observed / confirmed / unconfirmed totals.

## Data model

- Entities:
  - No entity changes.
- Schema changes or migrations:
  - None.
- Consistency and idempotency:
  - Polling remains stateless with respect to new ETH-only baseline metadata.

## API or contracts

- Endpoints or events:
  - No public API changes.
- Request/response examples:
  - Observer output stays unchanged.

## Backward compatibility

- API compatibility:
  - Fully backward compatible.
- Data migration compatibility:
  - No migration required.

## Failure modes and resiliency

- Retries/timeouts:
  - Reuse existing HTTP timeout handling.
- Backpressure/limits:
  - ETH polling remains O(1) per row.
- Degradation strategy:
  - Polling fails row-level if the balance snapshots are inconsistent.

## Observability

- Logs:
  - Reuse existing poll-cycle start and completion/failure logs.
- Metrics:
  - None added in this change.
- Traces:
  - None added.
- Alerts:
  - Existing row-level failure reasons remain the main signal.

## Security

- Authentication/authorization:
  - No change.
- Secrets:
  - No new secret surfaces.
- Abuse cases:
  - The design assumes current address balance is an acceptable proxy for the current ETH receiving
    flow.

## Alternatives considered

- Option A:
  - Keep block scanning with bounded incremental logic.
- Option B:
  - Use allocation-time baseline capture plus snapshot subtraction.
- Option C:
  - Add a third-party indexer.
- Why chosen:
  - Raw balance snapshots remove the stall-prone scan path without violating the desired
    architecture boundary around allocation.

## Risks

- Risk:
  - ETH totals are not equivalent to Bitcoin’s strict post-issuance inbound totals.
- Mitigation:
  - Keep the limitation explicit in requirements and tests.
- Risk:
  - Future deploy-and-sweep behavior could invalidate pure balance semantics.
- Mitigation:
  - Revisit the observer design when collection is introduced.
