---
doc: 04_test_plan
spec_date: 2026-04-05
slug: ethereum-usdt-payment-receiving
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
  - 2026-03-30-eth-balance-receipt-observer
  - 2026-04-03-ethereum-ledger-batch-sweep
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Test Plan

## Scope

- Covered:
  - Asset-aware persistence and HTTP/OpenAPI contract changes using `assetReference` as the only
    explicit non-native asset identifier
  - Ethereum ERC-20 `balanceOf(address)` receipt observation
  - CREATE2 unified receiver recovery payloads, stable interfaces, and operator tooling
  - one-signature ERC-20 batch recovery behavior
  - Ledger-only USDT payment helper
  - Regression coverage for Bitcoin and native ETH paths touched by shared code
- Not covered:
  - Real mainnet broadcast or treasury reconciliation
  - Generic multi-token or multi-EVM validation outside the explicit USDT rollout

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002, FR-005 / NFR-006
  - Steps:
    - Add policy/bootstrap tests for enabled and disabled Ethereum USDT policy construction.
  - Expected:
    - USDT policies expose `assetReference` and `decimals` without `assetKind` / `tokenStandard` /
      `minorUnit`, and malformed config fails closed.
- TC-002:
  - Linked requirements: FR-002, FR-003 / NFR-002, NFR-006
  - Steps:
    - Add observer tests for ERC-20 `balanceOf(address)` snapshots at latest and confirmed blocks
      using nullable `asset_reference` input.
  - Expected:
    - Observed, confirmed, and unconfirmed totals are derived correctly for token rows.
- TC-003:
  - Linked requirements: FR-004 / NFR-003, NFR-006
  - Steps:
    - Add receiver helper and sweep-material tests covering unified receiver payloads, ERC-20
      token metadata, and non-standard token-call handling.
  - Expected:
    - Unified receiver logic validates and serializes ETH and ERC-20 recovery payloads
      correctly.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002 / NFR-001, NFR-002
  - Steps:
    - Allocate a USDT payment address through the HTTP/controller/use-case stack and read it back
      through the status API path.
  - Expected:
    - Allocation and status responses expose `assetReference` and `decimals` without `minorUnit`,
      and the issued rows plus receipt-tracking rows persist the same nullable `asset_reference`
      snapshot.
- TC-102:
  - Linked requirements: FR-003 / NFR-002, NFR-005
  - Steps:
    - Run poller-cycle tests that feed ERC-20 observer snapshots through the current status
      transition path.
  - Expected:
    - USDT rows transition through watching, partially paid, paid unconfirmed, and paid confirmed
      without allocation joins for token identity and without affecting native ETH or Bitcoin
      regression cases.
- TC-103:
  - Linked requirements: FR-004 / NFR-003, NFR-005
  - Steps:
    - Validate the ERC-20 sweep script/contract path in dry-run mode with explicit selected rows.
  - Expected:
    - Mixed malformed selections are rejected; valid USDT selections render one Ledger-based
      ERC-20 factory batch recovery command even for multiple compatible receivers.
- TC-105:
  - Linked requirements: FR-004, FR-005 / NFR-006
  - Steps:
    - Verify that native ETH and USDT CREATE2 policies now derive through the same receiver
      artifact while still producing different addresses because of policy-specific salt inputs.
  - Expected:
    - The checked-in metadata receiver artifact is the unified current artifact, sweep material uses
      the same init code for ETH and USDT issuance on one network, and ETH/USDT addresses remain
      distinct.
- TC-106:
  - Linked requirements: FR-005 / NFR-006
  - Steps:
    - Verify that the checked-in CREATE2 asset bundle no longer references or rebuilds the unused
      token-only receiver source/artifact after the unified receiver cutover.
  - Expected:
    - The build script, tests, and embedded asset set only require the current unversioned receiver
      and factory artifacts.
- TC-107:
  - Linked requirements: FR-004, FR-005 / NFR-005, NFR-006
  - Steps:
    - Validate the sweep helper against rows whose recorded factory address differs from the
      network's current metadata factory, while keeping the selection on one shared factory.
  - Expected:
    - Recovery uses row-owned factory material and only rejects selections that mix multiple
      factories in one batch.
- TC-104:
  - Linked requirements: FR-006 / NFR-003, NFR-005, NFR-006
  - Steps:
    - Validate the Ledger USDT payment helper in dry-run mode for Sepolia and, if configured,
      mainnet.
  - Expected:
    - The helper resolves the expected asset reference and emits one Ledger-signed `transfer`
      command.

### E2E (if applicable)

- Scenario 1:
  - Issue one Sepolia USDT-compatible payment address, transfer token balance to it, run the poller,
    and confirm the status reaches `paid_confirmed`.
- Scenario 2:
  - Dry-run one operator sweep for the same confirmed USDT address and verify token-aware recovery
    validation/output.
- Scenario 3:
  - Dry-run one Ledger-signed Sepolia USD₮ payment to a test payment address.

## Edge cases and failure modes

- Case:
  - Ethereum USDT policy is configured without an asset reference.
  - Expected behavior:
    - Startup or policy bootstrap fails closed and the policy is not issuable.
- Case:
  - ERC-20 observer call fails or returns malformed ABI output.
  - Expected behavior:
    - Poller saves row-level failure reason and reschedules retry.
- Case:
  - Sweep selection mixes native ETH rows and rows for one shared USDT asset reference.
  - Expected behavior:
    - Sweep helper fails before broadcast regardless of row order and renders the native side as
      `<native>` instead of an empty asset reference.
- Case:
  - Sweep selection mixes two different non-native asset references.
  - Expected behavior:
    - Sweep helper fails before broadcast.
- Case:
  - Sweep selection mixes two different factory addresses.
  - Expected behavior:
    - Sweep helper fails before broadcast and asks the operator to split the batch by factory.
- Case:
  - ERC-20 recovery is attempted against a factory that has not yet been redeployed with the batch
    token-sweep ABI.
  - Expected behavior:
    - Operator workflow documents the redeploy requirement; the network metadata must be updated
      before the one-signature batch recovery path is used.

## NFR verification

- Performance:
  - Allocation-path tests confirm no mandatory observer/RPC dependency is introduced into the
    request path.
- Reliability:
  - Repeat poll-cycle tests on the same USDT row do not create duplicate terminal updates.
- Security:
  - No test or docs path introduces a private-key sweep mode; ERC-20 recovery remains Ledger-only.
