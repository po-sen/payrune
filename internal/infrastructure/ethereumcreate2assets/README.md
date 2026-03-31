# Ethereum CREATE2 Assets

This directory holds checked-in runtime assets for Ethereum CREATE2 issuance.

- `contracts/`: Solidity source for the fixed-collector receiver and the CREATE2 factory
- `artifacts/`: generated contract artifacts consumed by runtime config and local tooling
- `metadata/`: checked-in network metadata that resolves the active factory address and artifact name

`mainnet` and `sepolia` metadata currently carry deterministic fixture factory addresses until real
network deployments replace them. Local verification should use the Go CLI under
[`cmd/ethereum-create2-tool`](/Users/posen/Desktop/payrune/cmd/ethereum-create2-tool) or the thin
wrapper scripts under [`scripts/`](/Users/posen/Desktop/payrune/scripts).

For operator recovery and CREATE2 sweep workflows, use
`address_policy_allocations.sweep_material_json` as the only operator-facing recover payload. The
phase-1 ETH sweep helper at
[`scripts/ethereum_create2_sweep.sh`](/Users/posen/Desktop/payrune/scripts/ethereum_create2_sweep.sh)
loads that JSON directly from the DB instead of asking operators to re-enter factory / collector /
salt values.
