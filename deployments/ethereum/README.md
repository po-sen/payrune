# Ethereum Deployment

This directory contains the Solidity contracts plus the contract build/test toolchain used by
Payrune. The operator deployment path is now `cmd/evm-factory-deploy`, which deploys the factory
and registers it into `evm_factories` in one command.

## Contracts

- `DepositVaultFactory`
  - Owns deterministic CREATE2 deployment of payment vaults.
  - Restricts deploy/sweep operations to the factory owner.
- `DepositVault`
  - Receives ETH and ERC20 balances at a counterfactual address.
  - Allows the factory to sweep balances to the configured collector.
- `TestUSDTMock`
  - Local test token with 6 decimals and a USDT-like `transfer` that does not return a boolean.

## Environment

`cmd/evm-factory-deploy` reads one config set per network:

- `ETHEREUM_MAINNET_RPC_URL`
- `ETHEREUM_MAINNET_DEPLOY_PRIVATE_KEY`
- `ETHEREUM_MAINNET_DEPLOY_CONFIRMATIONS` (optional, default `1`)
- `ETHEREUM_MAINNET_COLLECTOR_ADDRESS`
- `ETHEREUM_MAINNET_FACTORY_DEPLOYMENT_MANIFEST` (optional output path)
- `ETHEREUM_SEPOLIA_RPC_URL`
- `ETHEREUM_SEPOLIA_DEPLOY_PRIVATE_KEY`
- `ETHEREUM_SEPOLIA_DEPLOY_CONFIRMATIONS` (optional, default `1`)
- `ETHEREUM_SEPOLIA_COLLECTOR_ADDRESS`
- `ETHEREUM_SEPOLIA_FACTORY_DEPLOYMENT_MANIFEST` (optional output path)

The runtime wiring for `cmd/evm-sweeper` reads one config set per network:

- `ETHEREUM_MAINNET_RPC_URL`
- `ETHEREUM_MAINNET_SWEEPER_PRIVATE_KEY`
- `ETHEREUM_SEPOLIA_RPC_URL`
- `ETHEREUM_SEPOLIA_SWEEPER_PRIVATE_KEY`
- Sweeper now loads active `factory_address` and `collector_address` from the database-backed
  `evm_factories` registry.

In compose, public RPC defaults live in `deployments/compose/compose.yaml`. The mainnet collector
default also lives in `compose.yaml`, while Sepolia-specific local test values stay in
`deployments/compose/compose.test.yaml` plus `deployments/compose/compose.test.env`. Private keys
remain empty by default and are only required when you actually run deploy or sweep execution.

Local `make up` does not auto-deploy or auto-register a factory. A DB viewer is exposed on
`http://localhost:8082` so `evm_factories` can be inspected after you run
`cmd/evm-factory-deploy`. Use `PostgreSQL` / `postgres` / `payrune` / `payrune` / `payrune` for
system/server/user/password/database in the local viewer.

## Commands

Install dependencies and run tests:

```bash
bash scripts/ethereum-contract-test.sh
```

The local test suite covers:

- deterministic vault deployment
- native ETH sweep from a deployed vault
- ERC20 sweep from a counterfactual address
- deployment manifest generation against a local JSON-RPC node

Ganache does not retain ETH balance sent to an undeployed CREATE2 address in this test setup, so the
native test funds the vault after deployment. The ERC20 counterfactual path is still covered
end-to-end.

Compile contracts only:

```bash
npm --prefix deployments/ethereum run compile
```

Deploy and register a factory:

```bash
ETHEREUM_SEPOLIA_RPC_URL=https://sepolia.example \
ETHEREUM_SEPOLIA_DEPLOY_PRIVATE_KEY=0x... \
ETHEREUM_SEPOLIA_COLLECTOR_ADDRESS=0xCollector \
bash scripts/evm-factory-deploy.sh --network=sepolia
```

Rotate an existing active factory explicitly:

```bash
bash scripts/evm-factory-deploy.sh --network=mainnet --replace-active
```

If deployment already succeeded but DB registration needs to be replayed, reuse the emitted
manifest instead of deploying again:

```bash
DATABASE_URL=postgres://payrune:payrune@localhost:5432/payrune?sslmode=disable \
bash scripts/evm-factory-deploy.sh \
  --deployment-manifest=deployments/ethereum/build/deployments/11155111-deposit-vault-factory.json
```

Or use compose defaults from `deployments/compose/compose.test.env`:

```bash
docker compose \
  -f deployments/compose/compose.yaml \
  --env-file deployments/compose/compose.test.env \
  run --rm --profile ops evm-factory-deploy --network=sepolia
```

Run the sweeper container in compose dry-run mode:

```bash
docker compose \
  -f deployments/compose/compose.yaml \
  --env-file deployments/compose/compose.test.env \
  run --rm --profile ops evm-sweeper --network=sepolia --asset-code=usdt --dry-run
```
