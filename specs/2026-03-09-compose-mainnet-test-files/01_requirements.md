---
doc: 01_requirements
spec_date: 2026-03-09
slug: compose-mainnet-test-files
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-09-bitcoin-compose-defaults
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Production-like stack:
  - The mainnet-oriented Compose stack defined by `deployments/compose/compose.yaml`.
- Test stack:
  - The local/testnet4 Compose stack defined by `deployments/compose/compose.yaml` plus `deployments/compose/compose.test.yaml`.
- Fake test env file:
  - The committed `deployments/compose/compose.test.env` file used to satisfy required base env during local/test rendering.

## Out-of-scope behaviors

- OOS1:
  - Cloudflare or public ingress changes.
- OOS2:
  - Runtime business logic changes.

## Functional requirements

### FR-001 - Keep only two Compose entrypoint files

- Description:
  - Deployment entrypoints must be reduced to `compose.yaml` and `compose.test.yaml`.
- Acceptance criteria:
  - [ ] `deployments/compose/compose.bitcoin.mainnet.yaml` is removed.
  - [ ] `deployments/compose/compose.bitcoin.testnet4.yaml` is removed.
  - [ ] `deployments/compose/compose.swagger.yaml` is removed.
  - [ ] `deployments/compose/compose.webhook.yaml` is removed.
- Notes:
  - Historical specs may still mention the old files, but the active deployment shape must not.

### FR-002 - `compose.yaml` must be a full mainnet stack

- Description:
  - `compose.yaml` must directly define the production-like mainnet deployment shape.
- Acceptance criteria:
  - [ ] `compose.yaml` includes `postgres`, `migrate`, `app`, `poller-mainnet`, and `receipt-webhook-dispatcher`.
  - [ ] `compose.yaml` explicitly lists the Bitcoin mainnet env keys for `app`.
  - [ ] `compose.yaml` leaves individual mainnet xpub env keys optional so unused address schemes can stay empty.
  - [ ] `compose.yaml` explicitly lists the webhook dispatcher env keys.
  - [ ] `compose.yaml` does not include Swagger or fake webhook receiver.
- Notes:
  - Production-like env may remain required in this file.

### FR-003 - `compose.test.yaml` must be a local/test override

- Description:
  - `compose.test.yaml` must override the production-like base into a local/test deployment shape.
- Acceptance criteria:
  - [ ] `compose.test.yaml` disables `poller-mainnet` by default for local/test usage.
  - [ ] `compose.test.yaml` adds `poller-testnet4`, `receipt-webhook-mock`, and `swagger`.
  - [ ] `compose.test.yaml` overrides local/test service settings on top of `compose.yaml`.
  - [ ] `compose.test.yaml` keeps `receipt-webhook-dispatcher` override minimal and does not restate inherited webhook env keys.
  - [ ] Testnet4 xpub values and fake webhook values come from `deployments/compose/compose.test.env`, not duplicated inline in `compose.test.yaml`.
- Notes:
  - The test stack should favor convenience, but remain an override.

### FR-004 - Mainnet/testnet defaults must be preserved

- Description:
  - The existing mainnet/testnet4 operational defaults must survive the file consolidation.
- Acceptance criteria:
  - [ ] Mainnet confirmations remain `2`.
  - [ ] Testnet4 confirmations remain `2`.
  - [ ] Mainnet and testnet4 unpaid receipt expiry remain `24h`.
  - [ ] Mainnet and testnet4 polling defaults remain `POLL_TICK_INTERVAL=5s`, `POLL_RESCHEDULE_INTERVAL=10m`, `POLL_BATCH_SIZE=2`, and `POLL_CLAIM_TTL=30s`.
- Notes:
  - This is a file-layout change, not a default-value change.

### FR-005 - Local tooling entrypoint must stay obvious

- Description:
  - Local Swagger and fake webhook receiver must remain easy to run through the one supported override path.
- Acceptance criteria:
  - [ ] `docker compose --env-file deployments/compose/compose.test.env -f deployments/compose/compose.yaml -f deployments/compose/compose.test.yaml config` succeeds without extra overlay files.
  - [ ] `docker compose --env-file deployments/compose/compose.test.env -f deployments/compose/compose.yaml -f deployments/compose/compose.test.yaml up` would include Swagger and fake webhook receiver.
  - [ ] `Makefile` hardcodes the single supported local/test Compose path instead of rebuilding it from multiple override variables.
- Notes:
  - This keeps local operator instructions short and concrete while preserving a single production base.

## Non-functional requirements

- Performance (NFR-001):
  - No runtime polling or webhook behavior may change as part of this refactor.
- Availability/Reliability (NFR-002):
  - `compose.test.yaml` must remain a clean override that composes directly on top of `compose.yaml`.
- Availability/Reliability (NFR-003):
  - The committed fake env file must remain sufficient to render the local/test override path.
- Security/Privacy (NFR-004):
  - `compose.yaml` must keep production-like required env explicit rather than hiding them in removed overlay files.
- Security/Privacy (NFR-005):
  - The committed test env file must contain only fake non-secret values.
- Compliance (NFR-006):
  - Not applicable.
- Observability (NFR-007):
  - Service names must remain stable enough that existing logs and operator expectations still make sense.
- Maintainability (NFR-008):
  - Deployment usage should optimize for clarity over reuse.

## Dependencies and integrations

- External systems:
  - Docker Compose.
- Internal services:
  - Existing `app`, `poller`, webhook dispatcher, Swagger UI, and fake webhook receiver services.
  - `deployments/compose/compose.test.env`
