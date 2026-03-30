---
doc: 01_requirements
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

# Requirements

## Glossary

- Balance snapshot:
  - The address balance returned by `eth_getBalance` at a specific block tag.

## Out-of-scope behaviors

- OOS1:
  - ERC-20 balances or internal ETH transfers.
- OOS2:
  - Address-history indexing or archive-node-only tracing.
- OOS3:
  - Outgoing transfer accounting for future deploy-and-sweep workflows.

## Functional requirements

### FR-001 - Observe ETH receipt rows from block-tagged balance snapshots

- Description:
  - Ethereum receipt observation must use bounded `eth_getBalance` calls instead of block scanning.
- Acceptance criteria:
  - [ ] Ethereum `ObserveAddress` uses `eth_getBalance` at `latestBlockHeight` for the observed
        snapshot.
  - [ ] Ethereum `ObserveAddress` uses `eth_getBalance` at
        `latestBlockHeight - requiredConfirmations + 1` when positive for the confirmed snapshot.
  - [ ] `unconfirmed_total_minor` is computed as
        `observed_total_minor - confirmed_total_minor`.
  - [ ] Ethereum `ObserveAddress` no longer calls `eth_getBlockByNumber`.

### FR-002 - Preserve current receipt lifecycle and output contract

- Description:
  - The refactor must keep the current receipt status model and observer output shape.
- Acceptance criteria:
  - [ ] `ObservePaymentAddressOutput` still returns observed, confirmed, unconfirmed totals and the
        latest block height.
  - [ ] Existing lifecycle transitions among `watching`, `partially_paid`, `paid_unconfirmed`,
        `paid_unconfirmed_reverted`, and `paid_confirmed` remain valid.
  - [ ] No public API or webhook payload changes are introduced.

### FR-003 - Fail safely on inconsistent snapshots and provider errors

- Description:
  - The ETH observer must reject impossible snapshot states rather than silently guessing.
- Acceptance criteria:
  - [ ] If any required balance query fails, the observer returns
        `ErrBlockchainReceiptObserverFailed`.
  - [ ] If `confirmed_total_minor > observed_total_minor`, the observer returns
        `ErrBlockchainReceiptObserverFailed`.
  - [ ] If `latestBlockHeight < requiredConfirmations`, the observer returns
        `confirmed_total_minor = 0` and `unconfirmed_total_minor = observed_total_minor`.

### FR-004 - Keep ETH-specific semantics explicit

- Description:
  - The design must document that ETH snapshot totals are not the same contract as Bitcoin’s
    post-issuance transaction filtering.
- Acceptance criteria:
  - [ ] The design states that current ETH totals reflect balance snapshots rather than strict
        post-issuance inbound totals.
  - [ ] The implementation keeps this limitation inside the observer path and does not add ETH
        baseline capture to allocation.
  - [ ] Tests cover the snapshot behavior explicitly.

## Non-functional requirements

- Performance (NFR-001):
  - For one ETH receipt row, observer-side work must remain bounded to at most two balance queries,
    independent of chain age or receipt age.
- Availability/Reliability (NFR-002):
  - One provider failure or inconsistent ETH row must only fail that row and must not break the
    rest of the cycle.
- Security/Privacy (NFR-003):
  - No new public fields or logs may expose raw CREATE2 salts or full source references.
- Compliance (NFR-004):
  - None beyond existing project requirements.
- Observability (NFR-005):
  - Existing poll-cycle start and completion/failure logs remain available.
- Maintainability (NFR-006):
  - The ETH observer should stay materially simpler than the prior block-scan implementation, with
    focused tests covering snapshot observation and failure handling only.

## Dependencies and integrations

- External systems:
  - Ethereum JSON-RPC supporting `eth_blockNumber` and `eth_getBalance`.
- Internal services:
  - Receipt polling use case
