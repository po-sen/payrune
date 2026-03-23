---
doc: 04_test_plan
spec_date: 2026-03-20
slug: create2-eth-payment-receiving
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-16-remove-xpub-fingerprint
  - 2026-03-05-blockchain-receipt-polling-service
  - 2026-03-08-payment-address-status-api
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# CREATE2 ETH Payment Receiving - Test Plan

## Scope

- Covered:
  - Ethereum CREATE2 address prediction correctness
  - Ethereum payment-address allocation flow
  - Ethereum native ETH receipt polling
  - Schema migration compatibility with existing Bitcoin behavior
- Not covered:
  - ERC-20 transfer support
  - Public mainnet canary rollout procedures
  - Future collection-specific signer infrastructure

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-006, NFR-006
  - Steps:
    - Add parsing and policy-normalization tests for `ethereum`, `mainnet`, `sepolia`, and
      `create2` policy config.
  - Expected:
    - Chain parsing, policy enablement, and config validation behave deterministically and do not
      break existing Bitcoin cases.
- TC-002:
  - Linked requirements: FR-002, FR-007, NFR-003
  - Steps:
    - Feed known factory address, collector address, salt, and init-code values into the Go-side
      CREATE2 predictor and compare them against expected vectors from the contract side.
  - Expected:
    - Predicted address matches the known vector exactly for every case.
- TC-003:
  - Linked requirements: FR-001, FR-002, FR-005, FR-007, NFR-001
  - Steps:
    - Unit-test `AllocatePaymentAddressUseCase` and `GenerateAddressUseCase` with an Ethereum
      CREATE2 policy and a deterministic fake or real predictor.
  - Expected:
    - Allocation persists the correct Ethereum-specific internal source and reference values, while
      public preview-by-index is unavailable or rejected for Ethereum CREATE2 policies, and one
      issued address remains reconstructible from allocation metadata plus runtime secret input.
- TC-004:
  - Linked requirements: FR-003, NFR-002, NFR-005
  - Steps:
    - Unit-test the Ethereum receipt observer with synthetic block or transaction inputs that cover
      no payment, partial payment, paid-unconfirmed, and paid-confirmed transitions.
  - Expected:
    - ETH totals and confirmation math are correct, and row-level errors are surfaced
      deterministically.
- TC-007:
  - Linked requirements: FR-001, FR-002, FR-007, NFR-003
  - Steps:
    - Attempt to derive or preview Ethereum payment addresses using only public policy metadata
      plus sequential `index` guesses through the public address-preview route.
  - Expected:
    - The public route rejects Ethereum CREATE2 policies, and no public response reveals raw salt,
      full source-ref, or enough data to enumerate future Ethereum payment addresses.

### Integration

- TC-101:
  - Linked requirements: FR-002, FR-005, NFR-006
  - Steps:
    - Apply the migration set to a disposable database, verify the neutralized allocation schema,
      ensure current Bitcoin persistence adapters still read and write correctly, and validate that
      compose/cloudflare deployment examples expose
      `ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS` and
      `ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS` plus per-network
      `ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY` and
      `ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY` runtime config, plus overrideable default
      Ethereum RPC endpoints for `mainnet` and `sepolia`, with scope-explicit poller service
      names per `(chain, network)` and `sepolia` living in the local/test Compose override.
  - Expected:
    - Migration succeeds, schema is in the expected shape, and Bitcoin adapter tests remain green.
- TC-102:
  - Linked requirements: FR-001, FR-002, FR-005, FR-006, FR-007
  - Steps:
    - Deploy the CREATE2 factory to a configured Ethereum verification network, compute one
      predicted address in Go from explicit stored salt inputs, then deploy the receiver with the
      same inputs.
  - Expected:
    - The deployed receiver address matches the predicted payment address exactly.
- TC-103:
  - Linked requirements: FR-003, FR-004, NFR-002, NFR-005
  - Steps:
    - Allocate one Ethereum payment address against a local database, fund it on the configured
      Ethereum verification network, then run one or more poller cycles from the matching
      chain-network poller scope.
  - Expected:
    - The payment receipt row updates to the expected status and the payment-status API returns the
      new totals.
- TC-105:
  - Linked requirements: FR-002, FR-007, NFR-003
  - Steps:
    - Issue multiple Ethereum payment addresses under one policy, then verify that checked-in
      factory metadata plus public API outputs remain insufficient to derive the next issued
      address without access to internal allocation salt material.
  - Expected:
    - Future Ethereum payment addresses are not reproducible from public metadata and sequential
      guesses alone.

### E2E (if applicable)

- Scenario 1:
  - Create one Ethereum payment address, send ETH to the predicted address before deployment, and
    poll until the payment becomes paid.
- Scenario 2:
  - Repeat one or more poller cycles after the payment is already paid and confirm the persisted
    status remains stable and idempotent.
- Scenario 3:
  - Issue a payment address with an idempotency key, replay the same request, and confirm the same
    Ethereum payment address and status resource are returned.

## Edge cases and failure modes

- Case:
  - Ethereum policy is configured with a factory address or init-code hash that does not match the
    contract artifacts used by the runtime.
  - Expected behavior:
    - Startup or preflight validation fails closed before payment addresses are issued.
- Case:
  - A payment address receives less ETH than `expectedAmountMinor`.
  - Expected behavior:
    - Receipt status remains partial or watching according to observed totals.
- Case:
  - A payment address receives more ETH than `expectedAmountMinor`.
  - Expected behavior:
    - The receipt status is still marked paid; any overpayment handling remains outside the first
      rollout scope.
- Case:
  - A chain reorg changes block confirmation depth near the payment threshold.
  - Expected behavior:
    - The observer recomputes totals from bounded history and the receipt status remains consistent
      with the final chain state.

## NFR verification

- Performance:
  - Measure allocation latency for repeated Ethereum issuance requests in a warm local environment.
    Verify p95 <= 250 ms.
- Reliability:
  - Re-run allocation and poller cycles multiple times on the same Ethereum payment address and
    confirm no duplicate receipt rows or inconsistent state transitions appear.
- Security:
  - Review that no per-payment private key material is stored, CREATE2 derivation remains
    deterministic from checked-in artifacts plus runtime-managed derivation key material, and
    default logs and public APIs do not expose raw CREATE2 salts or full source references.
