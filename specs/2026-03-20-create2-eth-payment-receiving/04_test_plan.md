---
doc: 04_test_plan
spec_date: 2026-03-20
slug: create2-eth-payment-receiving
mode: Full
status: READY
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
  - Ethereum deploy-and-sweep retry safety
  - Schema migration compatibility with existing Bitcoin behavior
- Not covered:
  - ERC-20 transfer support
  - Public mainnet canary rollout procedures
  - Production KMS or signer infrastructure beyond adapter contract boundaries

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
  - Linked requirements: FR-001, FR-002, FR-005, FR-008, NFR-001
  - Steps:
    - Unit-test `AllocatePaymentAddressUseCase` and `GenerateAddressUseCase` with an Ethereum
      CREATE2 policy and a deterministic fake or real predictor.
  - Expected:
    - Allocation persists the correct Ethereum-specific internal source and reference values, while
      public preview-by-index is unavailable or rejected for Ethereum CREATE2 policies, and one
      issued address remains reconstructible from allocation metadata plus runtime secret input.
- TC-004:
  - Linked requirements: FR-004, NFR-002, NFR-005
  - Steps:
    - Unit-test the Ethereum receipt observer with synthetic block or transaction inputs that cover
      no payment, partial payment, paid-unconfirmed, and paid-confirmed transitions.
  - Expected:
    - ETH totals and confirmation math are correct, and row-level errors are surfaced
      deterministically.
- TC-005:
  - Linked requirements: FR-003, FR-007, NFR-002, NFR-003
  - Steps:
    - Unit-test the deploy-and-sweep use case for fresh deploy, already deployed, already swept,
      and failed retry paths.
  - Expected:
    - Retries are idempotent, and collector routing cannot be changed by unexpected inputs.
- TC-006:
  - Linked requirements: FR-002, FR-003, FR-006, FR-007
  - Steps:
    - Keep factory, collector, salt rule, and receiver artifact fixed while changing only the
      configured operator signer.
  - Expected:
    - Predicted payment addresses remain unchanged, and only the transaction sender for
      deploy-and-sweep changes.
- TC-007:
  - Linked requirements: FR-001, FR-002, FR-008, NFR-003
  - Steps:
    - Attempt to derive or preview Ethereum payment addresses using only public policy metadata
      plus sequential `index` guesses through the public address-preview route.
  - Expected:
    - The public route rejects Ethereum CREATE2 policies, and no public response reveals raw salt,
      full source-ref, or enough data to enumerate future Ethereum payment addresses.

### Integration

- TC-101:
  - Linked requirements: FR-002, FR-006, NFR-006
  - Steps:
    - Apply the migration set to a disposable database, verify the neutralized allocation schema,
      ensure current Bitcoin persistence adapters still read and write correctly, and validate that
      compose/cloudflare deployment examples expose
      `ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS` and
      `ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS` plus per-network
      `ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY` and
      `ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY` runtime config.
  - Expected:
    - Migration succeeds, schema is in the expected shape, and Bitcoin adapter tests remain green.
- TC-102:
  - Linked requirements: FR-001, FR-002, FR-006, FR-007, FR-008
  - Steps:
    - Deploy the CREATE2 factory to a configured Ethereum verification network, compute one
      predicted address in Go from explicit stored salt inputs, then deploy the receiver with the
      same inputs.
  - Expected:
    - The deployed receiver address matches the predicted payment address exactly.
- TC-103:
  - Linked requirements: FR-004, FR-005, NFR-002, NFR-005
  - Steps:
    - Allocate one Ethereum payment address against a local database, fund it on the configured
      Ethereum verification network, then run one or more poller cycles.
  - Expected:
    - The payment receipt row updates to the expected status and the payment-status API returns the
      new totals.
- TC-104:
  - Linked requirements: FR-003, FR-006, FR-007, FR-008, NFR-002, NFR-003
  - Steps:
    - Run the sweeper against a funded, not-yet-deployed ETH payment address, then rerun it,
      including one retry after rotating the operator signer.
  - Expected:
    - The first run deploys and sweeps funds to the collector; the second run does not duplicate
      collection and reports deterministic persisted state without changing the predicted address.
- TC-105:
  - Linked requirements: FR-002, FR-008, NFR-003
  - Steps:
    - Issue multiple Ethereum payment addresses under one policy, then verify that checked-in
      factory metadata plus public API outputs remain insufficient to derive the next issued
      address without access to internal allocation salt material.
  - Expected:
    - Future Ethereum payment addresses are not reproducible from public metadata and sequential
      guesses alone.

### E2E (if applicable)

- Scenario 1:
  - Create one Ethereum payment address, send ETH to the predicted address before deployment, poll
    until the payment becomes paid, then sweep it into the collector address.
- Scenario 2:
  - Repeat the sweep workflow after the first successful collection and confirm the result is
    idempotent and observable through persisted technical state.
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
    - Receipt status remains partial or watching according to observed totals, and sweeping does not
      run unless the configured eligibility rule allows it.
- Case:
  - A payment address receives more ETH than `expectedAmountMinor`.
  - Expected behavior:
    - The receipt status is still marked paid; any overpayment handling follows the configured sweep
      policy or is surfaced for manual review.
- Case:
  - A chain reorg changes block confirmation depth near the payment threshold.
  - Expected behavior:
    - The observer recomputes totals from bounded history and the receipt status remains consistent
      with the final chain state.
- Case:
  - ETH was funded, but deployment or sweep transaction fails because of gas, signer, or RPC
    issues.
  - Expected behavior:
    - Payment status remains queryable; deploy-and-sweep technical state records the failure and can
      retry later.
- Case:
  - The operator signer is rotated after payment-address issuance but before deploy-and-sweep.
  - Expected behavior:
    - Existing predicted addresses stay valid, and collection continues with the new signer as long
      as factory metadata and receiver artifacts are unchanged.

## NFR verification

- Performance:
  - Measure allocation latency for repeated Ethereum issuance requests in a warm local environment.
    Verify p95 <= 250 ms.
- Reliability:
  - Re-run poller and sweeper cycles multiple times on the same funded payment address and confirm
    no duplicate collection or duplicate state records appear.
- Security:
  - Review that no per-payment private key material is stored, receiver sweep target is fixed, and
    operator-signer secrets are consumed only through runtime config without becoming part of
    address derivation. Also verify default logs and public APIs do not expose raw CREATE2 salts or
    full source references.
