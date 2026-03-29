---
doc: 01_requirements
spec_date: 2026-03-30
slug: compose-eth-sepolia-split
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
  - 2026-03-28-eth-create2-config-update
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Base compose:
- `deployments/compose/compose.yaml`, the default stack shared by non-test local usage.
- Test overlay:
- `deployments/compose/compose.test.yaml`, the Compose override loaded together with `compose.test.env` for local test/dev workflows.

## Out-of-scope behaviors

- OOS1: No env-value rotation or secret changes.
- OOS2: No new services or Compose profiles.

## Functional requirements

### FR-001 - Base compose must stop owning Sepolia API issuance envs

- Description: `deployments/compose/compose.yaml` must no longer define Sepolia API env vars for Ethereum issuance.
- Acceptance criteria:
  - [ ] `deployments/compose/compose.yaml` removes API env entries for `ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS`, `ETHEREUM_SEPOLIA_RECEIPT_EXPIRES_AFTER`, `ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS`, and `ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY`.
  - [ ] Existing mainnet Ethereum API env entries in `deployments/compose/compose.yaml` remain intact.
- Notes: This keeps the base stack focused on default/mainnet behavior.

### FR-002 - Test overlay must own Sepolia API issuance envs and blank mainnet issuance inputs

- Description: `deployments/compose/compose.test.yaml` must carry Sepolia API issuance envs and explicitly disable Ethereum mainnet CREATE2 issuance by blanking the mainnet issuance vars.
- Acceptance criteria:
  - [ ] `deployments/compose/compose.test.yaml` adds the four `ETHEREUM_SEPOLIA_*` API env entries under `services.api.environment`.
  - [ ] `deployments/compose/compose.test.yaml` sets `ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS` and `ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY` to `""`.
  - [ ] Existing Bitcoin test overlay behavior remains unchanged.
- Notes: Receipt-term defaults for Ethereum mainnet do not need to be blanked to satisfy the Bitcoin-like disablement pattern.

### FR-003 - The merged test stack must stay valid under the existing Compose workflow

- Description: The standard local Compose invocation used by the repo must still render a valid merged config after the split.
- Acceptance criteria:
  - [ ] `docker compose --env-file deployments/compose/compose.test.env -f deployments/compose/compose.yaml -f deployments/compose/compose.test.yaml config` succeeds.
  - [ ] The rendered merged config shows blank `ETHEREUM_MAINNET_CREATE2_*` values for `services.api.environment`.
  - [ ] The rendered merged config still includes the Sepolia API env entries under `services.api.environment`.
- Notes: This requirement validates the layering rather than the Go runtime.

## Non-functional requirements

- Performance (NFR-001): No impact on service runtime performance; change is config-only.
- Availability/Reliability (NFR-002): Existing mainnet/local Compose services remain structurally unchanged aside from the intended env split.
- Security/Privacy (NFR-003): The local test overlay must not accidentally source Ethereum mainnet CREATE2 issuance credentials from the host environment.
- Compliance (NFR-004):
- Observability (NFR-005): Validation must include a rendered Compose config check, not just static grep.
- Maintainability (NFR-006): Compose files should express the same network split pattern for Ethereum test usage that they already express for Bitcoin test usage.

## Dependencies and integrations

- External systems: Docker Compose CLI
- Internal services: `deployments/compose/compose.yaml`, `deployments/compose/compose.test.yaml`, `deployments/compose/compose.test.env`, and `Makefile`
