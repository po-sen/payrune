---
doc: 04_test_plan
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

# Blockchain Receipt Polling Service - Test Plan

## Scope

- Covered:
  - Domain lifecycle transitions.
  - Poller use case orchestration and UoW boundaries.
  - Chain/network contract decoupling in poller core.
  - Chain-routed observer selection behavior.
  - Issue-time-scoped Bitcoin observer behavior.
  - Network-scoped dual poller deployment config.
- Not covered:
  - Production observer implementations for non-Bitcoin chains.
  - Advanced mempool conflict enrichment.

## Tests

### Unit

- TC-001:

  - Linked requirements: FR-004, FR-005
  - Steps:
    - Domain tests for `watching`, `partially_paid`, `paid_unconfirmed`, `paid_confirmed`, `double_spend_suspected` transitions.
  - Expected:
    - Status and timestamp transitions match lifecycle rules.

- TC-002:

  - Linked requirements: FR-007, FR-009
  - Steps:
    - Use case tests verify UoW callback use and chain/network scope propagation.
  - Expected:
    - Register/claim/save calls are UoW-mediated and receive expected scope filters.

- TC-003:

  - Linked requirements: FR-008, FR-014
  - Steps:
    - Observer tests with mocked data source for success/error/missing-endpoint paths.
    - Verify pre-`issued_at` inbound transactions are excluded from totals.
  - Expected:
    - Satoshi totals are computed from issue-time-scoped inbound data only; deterministic error behavior on invalid config/responses.

- TC-004:

  - Linked requirements: FR-011, FR-012
  - Steps:
    - Use case/config tests for chain/network scope validation and routing behavior.
  - Expected:
    - Generic chain/network values are accepted, network-without-chain is rejected, and unsupported chain observer routing returns deterministic error.

- TC-005:
  - Linked requirements: FR-014, NFR-009
  - Steps:
    - Use case tests verify `issued_at` lower-bound value is propagated to observer input.
    - Run two consecutive polling cycles with unchanged inbound transactions.
  - Expected:
    - Lower-bound is always provided; repeated cycles remain idempotent and do not inflate totals.

### Integration

- TC-101:

  - Linked requirements: FR-001, FR-002, FR-006
  - Steps:
    - Validate migration and registration/claim SQL path with short test suite.
  - Expected:
    - Idempotent registration and lock-safe due-row claiming behavior preserved.

- TC-102:

  - Linked requirements: FR-010
  - Steps:
    - Render merged compose config with base + both bitcoin overrides.
  - Expected:
    - `poller-mainnet` and `poller-testnet4` both appear; legacy base `poller` absent.

- TC-103:

  - Linked requirements: FR-011
  - Steps:
    - Validate Postgres receipt row scan path maps chain/network with generic value object parser.
  - Expected:
    - Poller repository read path does not depend on Bitcoin-only network parser.

- TC-104:

  - Linked requirements: FR-013, NFR-008
  - Steps:
    - Run `make -n up`.
    - Render merged config with `compose.yaml + compose.bitcoin.mainnet.yaml + compose.bitcoin.testnet4.yaml + compose.test-env.yaml`.
  - Expected:
    - Default make startup includes all three overrides and resulting services include both network pollers.
    - Rendered poller env contains default Esplora-compatible endpoints for `BITCOIN_MAINNET_ESPLORA_URL` and `BITCOIN_TESTNET4_ESPLORA_URL`.

- TC-105:
  - Linked requirements: FR-002, FR-014
  - Steps:
    - Validate registration SQL copies allocation `issued_at` into receipt-tracking row.
    - Validate read/claim mapping exposes stored `issued_at` into application model.
  - Expected:
    - Every registered tracking row has deterministic issue-time lower-bound data.

### E2E (if applicable)

- Scenario 1:
  - Run both scoped pollers concurrently.
  - Expected:
    - No cross-network row processing when scope filters are set.

## Edge cases and failure modes

- Case:
  - RPC timeout/unreachable endpoint.
- Expected behavior:

  - Per-row error persisted and retried later.

- Case:
  - Missing RPC endpoint for requested network.
- Expected behavior:

  - Deterministic error path without process panic.

- Case:
  - Allocation is marked `issued` but `issued_at` is null/invalid.
- Expected behavior:

  - Registration or observation path fails deterministically and stores retryable polling error.

- Case:
  - `POLL_NETWORK` provided but `POLL_CHAIN` omitted.
- Expected behavior:
  - Poller startup/config validation fails fast.

## NFR verification

- Reliability:
  - Verify dual pollers do not cross-claim networks.
- Security:
  - Verify credential fields are not logged or persisted.
- Maintainability:
  - Verify full precommit pipeline passes after integration.
- Data correctness:
  - Verify pre-issue transaction history never contributes to receipt totals.
