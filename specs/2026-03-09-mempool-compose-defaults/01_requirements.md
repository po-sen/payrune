---
doc: 01_requirements
spec_date: 2026-03-09
slug: mempool-compose-defaults
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Glossary (optional)

- Public Esplora default:
  - The fallback endpoint value used in compose or tests when no explicit Esplora URL env var is provided.

## Out-of-scope behaviors

- OOS1:
  - Choosing a production SLA-backed Bitcoin provider.
- OOS2:
  - Adding failover or multi-provider logic.

## Functional requirements

### FR-001 - Unify compose Esplora defaults on mempool.space

- Description:
  - Bitcoin compose defaults must use `mempool.space` for both mainnet and testnet4.
- Acceptance criteria:
  - [ ] `compose.bitcoin.mainnet.yaml` defaults `BITCOIN_MAINNET_ESPLORA_URL` to `https://mempool.space/api`.
  - [ ] `compose.bitcoin.testnet4.yaml` continues to default `BITCOIN_TESTNET4_ESPLORA_URL` to `https://mempool.space/testnet4/api`.
- Notes:
  - This only changes defaults; explicit env overrides still take precedence.

### FR-002 - Keep poller config tests aligned with the default provider choice

- Description:
  - Tests covering poller Esplora config loading must reflect the unified default provider.
- Acceptance criteria:
  - [ ] Poller config tests no longer hardcode `blockstream.info` as the mainnet default example.
  - [ ] DI tests continue to pass after the config update.
- Notes:
  - The test goal is consistency, not provider-specific behavior.

## Non-functional requirements

- Performance (NFR-001):
  - No runtime performance behavior changes are introduced.
- Availability/Reliability (NFR-002):
  - Existing env override behavior remains unchanged.
- Security/Privacy (NFR-003):
  - No new credentials or secret flows are introduced.
- Compliance (NFR-004):
  - None.
- Observability (NFR-005):
  - None.
- Maintainability (NFR-006):
  - Default provider choices should be easy to understand from the compose files and tests.

## Dependencies and integrations

- External systems:
  - `mempool.space` public Esplora API.
- Internal services:
  - `deployments/compose/compose.bitcoin.mainnet.yaml`
  - `deployments/compose/compose.bitcoin.testnet4.yaml`
  - `internal/infrastructure/di/poller_container_test.go`
