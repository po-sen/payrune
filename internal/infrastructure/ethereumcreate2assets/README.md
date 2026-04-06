# Ethereum CREATE2 Assets

This directory holds checked-in runtime assets for Ethereum CREATE2 issuance and recovery.

- `contracts/`: Solidity source for the fixed-collector receiver variants and the CREATE2 factory
- `artifacts/`: generated contract artifacts consumed by runtime config and operator tooling
- `metadata/`: checked-in network metadata that resolves the active factory address and artifact name

`mainnet` and `sepolia` metadata currently point new issuance at the checked-in
`FixedCollectorReceiver` artifact. Rebuild checked-in artifacts with
[`scripts/ethereum_create2_build_artifacts.sh`](/Users/posen/Desktop/payrune/scripts/ethereum_create2_build_artifacts.sh).

For operator recovery and CREATE2 sweep workflows, use
`address_policy_allocations.sweep_material_json` as the only operator-facing recovery payload.

Use [`scripts/ethereum_create2_factory_deploy.sh`](/Users/posen/Desktop/payrune/scripts/ethereum_create2_factory_deploy.sh)
for the one-time deployment step. It deploys the checked-in `Create2ReceiverFactory` artifact and
updates checked-in metadata with the deployed factory address for the selected network.

Use [`scripts/ethereum_create2_sweep.sh`](/Users/posen/Desktop/payrune/scripts/ethereum_create2_sweep.sh)
for Ledger-only recovery. The same helper accepts one or many explicit selections and sends a
single batch recovery call through the factory recorded in the selected rows' recovery material.
The helper still loads checked-in metadata so operators can compare that row-owned factory against
the current issuance factory for the selected network. The factory derives each receiver from
CREATE2 recovery payload, deploys the receiver if code is still missing, and then sweeps it. The
helper also queries each selected receiver's current on-chain balance and fails closed when any
selected receiver is already empty.
