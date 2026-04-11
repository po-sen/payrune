---
doc: 01_requirements
spec_date: 2026-04-09
slug: compose-env-example
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-07-ethereum-contract-readiness
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Base compose env example:
  - A checked-in example file for compose envs used by the formal/mainnet-style path.

## Out-of-scope behaviors

- OOS1:
  - Changing compose runtime behavior or env names.
- OOS2:
  - Removing the committed fake-value development env file.

## Functional requirements

### FR-001 - Keep compose env example as the main checked-in compose reference

- Description:
  - The repo must include `deployments/compose/compose.env.example` as the main checked-in compose env reference.
- Acceptance criteria:
  - [ ] The file exists under `deployments/compose/`.
  - [ ] The file covers the formal/mainnet-style env surface consumed by `deployments/compose/compose.yaml`.
  - [ ] The file may include extra local development-chain blocks when they are part of the same compose env contract.
  - [ ] Local development policy flags in the example stay disabled by default.
- Notes:
  - The file is an example/reference, not a secret-bearing checked-in runtime env.

### FR-002 - Include local development-chain env blocks in the example

- Description:
  - The main checked-in example must include the local development-chain env blocks needed to understand the full compose env surface.
- Acceptance criteria:
  - [ ] `deployments/compose/compose.env.example` includes Bitcoin testnet4 policy envs.
  - [ ] `deployments/compose/compose.env.example` includes Ethereum Sepolia policy envs.
  - [ ] `deployments/compose/compose.env.example` includes the development-profile webhook/dev-helper envs needed to understand the local development path.
- Notes:
  - Public fake values and public test-chain references are acceptable in the example.

### FR-002A - Keep the checked-in development env minimal

- Description:
  - The checked-in development env file must keep only the overrides that differ from `compose.yaml` defaults or that are required to satisfy development-only services.
- Acceptance criteria:
  - [ ] `deployments/compose/compose.dev.env` exists.
  - [ ] `deployments/compose/compose.dev.env` keeps the development policy `*_ENABLED=true` flags.
  - [ ] `deployments/compose/compose.dev.env` keeps the required Bitcoin testnet4 xpub values.
  - [ ] `deployments/compose/compose.dev.env` keeps the required Ethereum Sepolia CREATE2 config and USDT asset reference.
  - [ ] `deployments/compose/compose.dev.env` keeps the required webhook development settings.
  - [ ] `deployments/compose/compose.dev.env` does not restate values that already match `compose.yaml` defaults unless they are required to make the file understandable.
- Notes:
  - The goal is to keep the checked-in development file ready to run without duplicating default noise.

### FR-003 - Keep grouping readable and consistent

- Description:
  - The example file must follow the current repo convention of grouping the same network/kind together.
- Acceptance criteria:
  - [ ] Bitcoin mainnet policy entries are grouped as paired `ENABLED + XPUB` values per policy.
  - [ ] Bitcoin testnet4 policy entries are grouped as paired `ENABLED + XPUB` values per policy.
  - [ ] Ethereum mainnet entries keep shared CREATE2 config together instead of pretending each policy has a separate collector/derivation key.
  - [ ] Ethereum Sepolia entries in the example keep shared CREATE2 config together.
  - [ ] Comments clearly indicate required, conditional, or optional settings where that improves readability.
- Notes:
  - This should align with the recent env formatting cleanup in the compose/test env files.

### FR-004 - Example must validate against formal and development rendering

- Description:
  - The example file must be usable as a valid env source for formal/mainnet-style config rendering and for local development config rendering.
- Acceptance criteria:
  - [ ] `docker compose --env-file deployments/compose/compose.env.example -f deployments/compose/compose.yaml config` succeeds.
  - [ ] `docker compose --env-file deployments/compose/compose.env.example --profile development -f deployments/compose/compose.yaml config` succeeds.
  - [ ] `docker compose --env-file deployments/compose/compose.dev.env --profile development -f deployments/compose/compose.yaml config` succeeds.
  - [ ] Required values in the rendered compose paths are satisfied by placeholder/example values in the file.
- Notes:
  - Placeholder values may be fake but must be structurally valid enough for config rendering.

### FR-005 - Makefile must use explicit local compose targets

- Description:
  - The main local compose Make targets must expose explicit local development and formal/mainnet-style entrypoints instead of switching behavior based on `deployments/compose/compose.env` presence.
- Acceptance criteria:
  - [ ] `make up`, `make down`, and `make config` use `deployments/compose/compose.dev.env` with the `development` profile.
  - [ ] `make up-mainnet`, `make down-mainnet`, and `make config-mainnet` use `deployments/compose/compose.env` without a `mainnet` profile.
  - [ ] The Makefile logic stays simple and readable; it must not introduce opaque dynamic make metaprogramming.
  - [ ] Compose commands stay direct in the target recipes instead of being hidden behind multiple command-assembly variables.
  - [ ] The Makefile uses native Make target/prerequisite behavior to fail fast when the required env file is missing.
  - [ ] `Makefile` exposes a `help` target that lists the main local and Cloudflare entrypoints.
  - [ ] Existing Cloudflare Make targets remain available.
- Notes:
  - The operator should not need hidden env-file-presence behavior to understand which target does what.

### FR-006 - Keep one compose entrypoint and one development-only profile

- Description:
  - The repo must use a single `compose.yaml` as the local compose entrypoint, with one dedicated development profile for dev-only services.
- Acceptance criteria:
  - [ ] `deployments/compose/compose.yaml` contains both formal/mainnet-style and local development service definitions.
  - [ ] Formal/mainnet-style services are part of the base stack and do not require a `mainnet` profile.
  - [ ] Local test pollers and test-only helpers are guarded by a `development` profile.
  - [ ] `deployments/compose/compose.test.yaml` is no longer required for local compose usage.
  - [ ] The local development path still renders successfully with `deployments/compose/compose.dev.env`.
- Notes:
  - Profiles are used to decide which services start; env files still carry the actual settings.

### FR-007 - Cloudflare migration stays in `cf-up` but loses its standalone target

- Description:
  - The Cloudflare-facing Make targets should keep migration in the deploy flow, but the standalone migration Make target should be removed.
- Acceptance criteria:
  - [ ] `Makefile` no longer defines `cf-migrate`.
  - [ ] `make cf-up` runs `scripts/cf-cloudflare-migrate.sh` before deploying workers.
  - [ ] `make cf-up` still deploys the fake webhook worker and the main payrune worker.
  - [ ] `make cf-down` still deletes those workers.
- Notes:
  - This does not remove the migration script itself; it only removes the standalone Make wrapper.

## Non-functional requirements

- Performance (NFR-001):
  - No runtime performance impact; this is a docs/config-only change.
- Availability/Reliability (NFR-002):
  - The example file must not break compose config rendering.
- Security/Privacy (NFR-003):
  - No real secrets may be committed; secrets must stay placeholders.
- Compliance (NFR-004):
  - No additional compliance requirements.
- Observability (NFR-005):
  - No new observability behavior required.
- Maintainability (NFR-006):
  - Comments and grouping must make it obvious which blocks belong to the formal/mainnet-style path, the local development path, or both.

## Dependencies and integrations

- External systems:
  - Docker Compose config rendering.
- Internal services:
  - [`deployments/compose/compose.yaml`](/Users/posen/Desktop/payrune/deployments/compose/compose.yaml)
  - [`deployments/compose/compose.env.example`](/Users/posen/Desktop/payrune/deployments/compose/compose.env.example)
  - [`deployments/compose/compose.dev.env`](/Users/posen/Desktop/payrune/deployments/compose/compose.dev.env)
