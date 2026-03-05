---
doc: 01_requirements
spec_date: 2026-03-05
slug: blockchain-receipt-polling-service
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-postgresql18-migration-runner-container
  - 2026-03-04-policy-payment-address-allocation
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Blockchain Receipt Polling Service - Requirements

## Glossary (optional)

- Receipt tracking:
  - Persistent lifecycle state for one allocated payment address.
- Due row:
  - Tracking row where `next_poll_at <= now` and active status allows claiming.

## Out-of-scope behaviors

- OOS1:
  - Implementing production ETH/TRON observers in this iteration.
- OOS2:
  - Treasury sweep/refund workflows.

## Functional requirements

### FR-001 - Dedicated receipt tracking persistence

- Description:
  - Maintain receipt state in dedicated table independent from allocation lifecycle table.
- Acceptance criteria:
  - [x] `payment_receipt_trackings` table exists via migration.
  - [x] Table contains lifecycle totals/status/timestamps and due-poll metadata.
  - [x] Uniqueness and due-query indexes exist.

### FR-002 - Idempotent tracking registration from issued allocations

- Description:
  - Poller registers missing receipt rows from `address_policy_allocations` with `issued` status.
- Acceptance criteria:
  - [x] Registration uses insert-on-conflict semantics.
  - [x] Repeated cycles do not duplicate tracking rows.
  - [x] Tracking row preserves source allocation `issued_at` for later observation scoping.

### FR-003 - Polling microservice runtime

- Description:
  - Provide standalone poller process with interval and batch controls.
- Acceptance criteria:
  - [x] `cmd/poller` binary exists.
  - [x] Poll interval/batch/claim TTL/required confirmations are env-configurable.
  - [x] Graceful shutdown on signal/context cancellation.

### FR-004 - Receipt amount aggregation lifecycle

- Description:
  - Aggregate totals across multiple observations and map to lifecycle states.
- Acceptance criteria:
  - [x] Supports `watching`, `partially_paid`, `paid_unconfirmed`, `paid_confirmed`.
  - [x] Over-target amounts remain valid and preserved in totals.

### FR-005 - Conflict-risk status

- Description:
  - Domain supports conflict-risk lifecycle status.
- Acceptance criteria:
  - [x] `double_spend_suspected` status is modeled and persisted.
  - [x] Observer output supports `conflict_total_minor` input for state transition.

### FR-006 - Concurrency-safe due claiming

- Description:
  - Multiple workers can process due rows without duplicate claim overlap.
- Acceptance criteria:
  - [x] Claim SQL uses `FOR UPDATE SKIP LOCKED`.
  - [x] Claim step updates `next_poll_at` claim window atomically.

### FR-007 - Poller Unit of Work orchestration

- Description:
  - Poller DB operations run through explicit UoW callbacks.
- Acceptance criteria:
  - [x] Register + claim path executes inside UoW callback.
  - [x] Save observation path executes inside UoW callback.
  - [x] Save polling error path executes inside UoW callback.

### FR-008 - Bitcoin Esplora observer integration

- Description:
  - Poller observer queries Bitcoin transaction data and maps inbound amounts to satoshi through a chain-agnostic observer contract.
- Acceptance criteria:
  - [x] Node-backed observer adapter exists and is wired in poller DI.
  - [x] `mainnet` and `testnet4` endpoints are selected by tracking network.
  - [x] Missing endpoint for target network returns deterministic error.
  - [x] Adapter implementation does not depend on global current-UTXO snapshot queries.
  - [x] Runtime env naming for Bitcoin observer endpoints is explicitly Esplora-scoped (`BITCOIN_*_ESPLORA_*`).

### FR-009 - Network-scoped poller filtering

- Description:
  - Poller can constrain work to configured chain/network scope.
- Acceptance criteria:
  - [x] `POLL_CHAIN` and `POLL_NETWORK` are parsed and validated.
  - [x] Register/claim SQL apply optional chain/network filters.

### FR-010 - Dual network poller deployment

- Description:
  - Deployment uses separate poller services for `mainnet` and `testnet4`.
- Acceptance criteria:
  - [x] `compose.bitcoin.mainnet.yaml` defines `poller-mainnet`.
  - [x] `compose.bitcoin.testnet4.yaml` defines `poller-testnet4`.
  - [x] Base `compose.yaml` no longer defines legacy shared `poller` service.
  - [x] Both pollers can be rendered together by compose config merge.

### FR-011 - Chain/network contract decoupling in poller core

- Description:
  - Poller domain and application contracts must not depend on Bitcoin-specific network types or implicit Bitcoin defaults.
- Acceptance criteria:
  - [x] Receipt tracking entity uses chain-agnostic `chain` + `network` value objects.
  - [x] Observer port input uses chain-agnostic `network` type.
  - [x] Poll scope validation does not auto-infer Bitcoin when only `network` is supplied.

### FR-012 - Chain-routed observer composition

- Description:
  - Poller runtime composes chain-specific observer adapters behind a routing adapter, so core use case does not branch on chain implementation details.
- Acceptance criteria:
  - [x] A routing observer selects downstream observer by `chain`.
  - [x] Unsupported chain returns deterministic error.
  - [x] Bitcoin observer remains a chain-specific adapter behind router and does not receive `chain` routing input directly.

### FR-013 - Default compose startup profile for dual pollers

- Description:
  - `make up` should default to a compose override set that starts both bitcoin pollers and test env fixtures, and `up-test-env` should not be required.
- Acceptance criteria:
  - [x] `COMPOSE_OVERRIDE` default in `Makefile` includes:
    - `deployments/compose/compose.bitcoin.mainnet.yaml`
    - `deployments/compose/compose.bitcoin.testnet4.yaml`
    - `deployments/compose/compose.test-env.yaml`
  - [x] `up-test-env` target is removed from Makefile.
  - [x] `make -n up` renders compose command containing all three override files.
  - [x] Bitcoin poller compose overrides provide default Esplora-compatible endpoint URLs for `mainnet` and `testnet4`.

### FR-014 - Issue-time-scoped receipt observation

- Description:
  - Poller only counts inbound value observed at/after address `issued_at`, and ignores pre-issue history.
- Acceptance criteria:
  - [x] Observer input contract includes issue-time lower-bound data (`issued_at` or equivalent cursor).
  - [x] Receipt aggregation excludes inbound transfers earlier than allocation `issued_at`.
  - [x] Repeated polling remains idempotent while advancing observation lower-bound state.

## Non-functional requirements

- Performance (NFR-001):
  - Observer call path uses bounded timeout from config.
- Availability/Reliability (NFR-002):
  - Failed per-row RPC observations are persisted as retryable polling errors.
- Security/Privacy (NFR-003):
  - RPC credentials are env-driven and not emitted in logs.
- Observability (NFR-005):
  - Poll cycle summary logs include claimed/updated/failed counts.
- Maintainability (NFR-006):
  - All integration remains behind application ports and UoW boundaries.
- Extensibility (NFR-007):
  - Adding a new chain observer should only require a new adapter and DI registration, without modifying receipt domain entity/use case code.
- Operability (NFR-008):
  - Local startup command for full bitcoin polling profile remains a single default command (`make up`).
- Data correctness (NFR-009):
  - Receipt totals must be computed from issue-time-scoped inbound transfers only; pre-issue chain history must have zero impact on state.

## Dependencies and integrations

- External systems:
  - PostgreSQL.
  - Bitcoin endpoint(s) exposing Esplora-compatible APIs.
- Internal services:
  - Existing allocation persistence (`address_policy_allocations`).
