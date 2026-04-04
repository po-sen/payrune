---
doc: 02_design
spec_date: 2026-04-03
slug: ethereum-ledger-batch-sweep
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-sweep-material-redesign
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
  - Keep one `Create2ReceiverFactory` per network, one active metadata entry per network, and one
    canonical CREATE2 batch recovery call.
  - Make factory-driven recovery CREATE2-aware so the same batch transaction can deploy missing
    receivers and sweep them immediately.
  - Remove legacy recovery entry points and old private-key-based helper paths.
- Key decisions:
  - Ethereum-only, explicit, non-generic design.
  - Ledger-only signer path.
  - Explicit selector lists instead of hidden DB discovery.
  - One singleton contract per network: `Create2ReceiverFactory`.

## System context

- Components:
  - `internal/infrastructure/ethereumcreate2assets/contracts/Create2ReceiverFactory.sol`
  - generated artifact in `internal/infrastructure/ethereumcreate2assets/artifacts/`
  - `scripts/ethereum_create2_build_artifacts.sh`
  - `scripts/ethereum_create2_factory_deploy.sh`
  - `scripts/ethereum_create2_sweep.sh`
- Interfaces:
  - PostgreSQL reads via `psql`
  - ABI encoding/broadcast via `cast`
  - Ledger sender resolution via `cast wallet address --ledger`

## Key flows

- Flow 1:
  - Operator runs the factory deploy script once per network.
  - The script resolves chain ID, maps it to `mainnet` or `sepolia`, validates the Ledger sender,
    prints the deploy command in dry-run mode, and on broadcast writes the deployed factory address
    back into checked-in metadata.
- Flow 2:
  - Operator passes one or more explicit payment address IDs or ETH addresses to the sweep script.
  - The script loads matching issued rows, validates all rows, checks each selected receiver's
    current on-chain balance, validates the CREATE2 payload fields needed for recovery, recomputes
    `keccak(init_code_hex)` and the predicted CREATE2 address, and if a receiver is already
    deployed verifies its `collector()` matches the recorded recovery payload.
  - The script resolves the active factory address from checked-in metadata for the selected
    network, rejects rows whose recorded `factory_address` does not match that active factory,
    validates the Ledger sender, and prints a single `cast send` command in dry-run mode.
- Flow 3:
  - Operator reruns the sweep script with `--broadcast`.
  - The script performs the same validations and then sends one transaction to the selected active
    factory contract using the same CREATE2 batch recovery call for both deployed and undeployed
    receivers.

## Diagrams (optional)

- Mermaid sequence / flow:

```mermaid
flowchart TD
  A[Deploy factory once] --> B[Write factory address into metadata]
  B --> C[Operator explicit selection]
  C --> D[Load issued create2 rows]
  D --> E[Validate rows and sweep_material_json]
  E --> F[Validate create2 salt and init code]
  F --> G[Query receiver balances]
  G --> H[Resolve factory from recovery payload]
  H --> H2[Validate payload factory equals active metadata factory]
  H2 --> I[Resolve Ledger sender]
  I --> J[Render cast send command]
  J --> K{Broadcast?}
  K -- No --> L[Dry-run output only]
  K -- Yes --> M[Ledger signs one tx]
  M --> N[Create2ReceiverFactory.sweep(salts, initCodes)]
  N --> O[Deploy missing receiver if code absent]
  O --> P[receiver.sweep()]
```

## Data model

- Entities:
  - No domain/application model changes.
- Schema changes or migrations:
  - None.
- Consistency and idempotency:
  - This remains an operator-initiated recovery flow. No write-side DB changes are introduced.

## API or contracts

- Endpoints or events:
  - None.
- Request/response examples:
  - Factory batch recovery function:
    - `function sweep(bytes32[] calldata salts, bytes[] calldata initCodes) external`
  - Required env:
    - `ETHEREUM_SWEEP_RPC_URL`
    - `ETHEREUM_SWEEP_FROM_ADDRESS`
  - Sweep-only required env:
    - `DATABASE_URL`
  - Selector env:
    - exactly one of:
      - `ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS`
      - `ETHEREUM_SWEEP_ADDRESSES`
  - Optional env:
    - `ETHEREUM_SWEEP_DERIVATION_PATH`

## Backward compatibility (optional)

- API compatibility:
  - Existing APIs unchanged.
- Data migration compatibility:
  - No migration.
  - Existing `sweep_material_json` remains valid only when its recorded `factory_address` matches
    the currently active metadata factory for that network.

## Failure modes and resiliency

- Retries/timeouts:
  - No automated retry loop in the batch helper.
- Backpressure/limits:
  - The script should reject oversized batches above a conservative safety limit.
- Degradation strategy:
  - Operators can run the same sweep helper with one selected row.

## Observability

- Logs:
  - Deploy flow prints network, metadata path, factory address (when known), Ledger sender, and
    final deploy command.
  - Sweep flow prints selector, network, receiver count, payment address IDs, receiver addresses,
    receiver balances in wei, receiver deployment states, factory address, call signature, Ledger
    sender, and final command.
- Metrics:
  - None.
- Traces:
  - None.
- Alerts:
  - None.

## Security

- Authentication/authorization:
  - Ledger sender must match `ETHEREUM_SWEEP_FROM_ADDRESS`.
  - The factory batch sweep path stays permissionless because each receiver still hardcodes its own
    collector.
- Secrets:
  - No private key env vars in operator helper flows.
- Abuse cases:
  - Mixed-network, malformed, or duplicate selections must fail before broadcast.
  - Zero-balance receiver selections must fail before broadcast to avoid wasting gas on empty sweeps.
  - Receiver rows whose CREATE2 recovery payload is internally inconsistent must fail before
    broadcast.
  - A deployed receiver whose `collector()` does not match the recorded `collector_address` must
    fail before broadcast.
  - Selected rows whose recorded `factory_address` does not match the active metadata factory must
    fail before broadcast.
  - Receiver call failure reverts the whole batch.
  - Legacy private-key helper paths are removed so operators are not nudged toward unsafe shortcuts.

## Alternatives considered

- Option A:
  - Keep a separate `SweepBatchCaller` singleton contract and batch-only scripts.
- Option B:
  - Loop a single-address helper over many rows.
- Option C:
  - Use a hot wallet or private key for automation.
- Why chosen:
  - Option A creates a second singleton contract and a second operator address to remember.
  - Option B keeps the Ledger bottleneck unchanged.
  - Option C violates the repository's required security model.
  - Extending the existing factory keeps initialization, metadata, and sweep recovery under one
    contract without weakening signer safety.

## Risks

- Risk:
  - One bad receiver call blocks the whole batch.
- Mitigation:
  - Fail closed and keep the same sweep helper usable with one selected row.
- Risk:
  - Operators may keep stale local issued rows after redeploying a new active factory.
- Mitigation:
  - Fail closed when recorded `factory_address` no longer matches the active metadata factory and
    require regenerating local development data.
