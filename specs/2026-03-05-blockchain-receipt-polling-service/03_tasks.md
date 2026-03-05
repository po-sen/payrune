---
doc: 03_tasks
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

# Blockchain Receipt Polling Service - Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - Includes schema change, async worker runtime, external RPC integration, and deployment split.
- Upstream dependencies (`depends_on`):
  - `2026-03-03-postgresql18-migration-runner-container`
  - `2026-03-04-policy-payment-address-allocation`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`.

## Milestones

- M1: Receipt-tracking schema and domain transitions.
- M2: Poller use case + UoW + persistence integration.
- M3: Node-backed observer and dual-network deployment model.
- M4: Poller core decoupled from Bitcoin-specific chain/network contracts.
- M5: Issue-time-scoped inbound receipt observation model.

## Tasks (ordered)

1. T-001 - Add receipt-tracking schema and domain state model

   - Scope:
     - Add migration and domain entity/value object for receipt lifecycle.
   - Linked requirements: FR-001, FR-004, FR-005
   - Validation:
     - [x] `go test ./internal/domain/... -count=1`
     - [x] `go test ./... -short -count=1`

2. T-002 - Implement poller use case, ports, and runtime wiring

   - Scope:
     - Implement cycle orchestration, DTO/ports, `cmd/poller`, bootstrap, and DI container.
   - Linked requirements: FR-002, FR-003, FR-006, FR-007, FR-009
   - Validation:
     - [x] `go test ./internal/application/use_cases -count=1`
     - [x] `go test ./... -short -count=1`

3. T-003 - Implement postgres receipt repository and lock-safe claiming

   - Scope:
     - Implement registration/claim/save SQL including `FOR UPDATE SKIP LOCKED`.
   - Linked requirements: FR-001, FR-002, FR-006, FR-009
   - Validation:
     - [x] `go test ./internal/adapters/outbound/persistence/postgres -count=1`

4. T-004 - Integrate Bitcoin Esplora observer

   - Scope:
     - Replace noop behavior with initial Bitcoin node-backed observer baseline and network-specific endpoint selection.
   - Linked requirements: FR-008
   - Validation:
     - [x] `go test ./internal/adapters/outbound/bitcoin -count=1`

5. T-005 - Refactor poller DB write path to explicit UnitOfWork

   - Scope:
     - Use UoW callbacks for register/claim/save observation/save error paths.
   - Linked requirements: FR-007, NFR-006
   - Validation:
     - [x] `go test ./internal/application/use_cases -count=1`

6. T-006 - Add network-scoped dual pollers and remove legacy base poller

   - Scope:
     - Add `poller-mainnet` / `poller-testnet4` in override compose files.
     - Remove base `poller` service from `compose.yaml`.
   - Linked requirements: FR-010, NFR-002
   - Validation:
     - [x] `docker compose -f deployments/compose/compose.yaml -f deployments/compose/compose.bitcoin.mainnet.yaml -f deployments/compose/compose.bitcoin.testnet4.yaml config`
     - [x] `go test ./... -short -count=1`
     - [x] `bash scripts/precommit-run.sh`

7. T-007 - Decouple poller domain/application contracts from Bitcoin-specific network typing

   - Scope:
     - Introduce chain-agnostic network value object for receipt tracking and observer port contracts.
     - Remove implicit Bitcoin fallback from poll scope parsing.
   - Linked requirements: FR-011, NFR-007
   - Validation:
     - [x] `go test ./internal/domain/... -count=1`
     - [x] `go test ./internal/application/use_cases -count=1`

8. T-008 - Add chain-routed observer composition and wire Bitcoin adapter behind router

   - Scope:
     - Implement routing observer adapter keyed by chain.
     - Keep Bitcoin adapter chain-specific and selected through router in poller DI.
   - Linked requirements: FR-008, FR-012, NFR-006, NFR-007
   - Validation:
     - [x] `go test ./internal/adapters/outbound/bitcoin -count=1`
     - [x] `go test ./internal/adapters/outbound/blockchain -count=1`

9. T-009 - Update poller config/runtime validation and regression tests

   - Scope:
     - Update env parsing and use case tests for chain/network generic validation.
     - Re-run full precommit verification.
   - Linked requirements: FR-003, FR-009, FR-011, FR-012, NFR-002
   - Validation:
     - [x] `go test ./cmd/poller ./internal/bootstrap ./internal/application/use_cases -count=1`
     - [x] `go test ./... -short -count=1`
     - [x] `SPEC_DIR=\"specs/2026-03-05-blockchain-receipt-polling-service\" bash scripts/spec-lint.sh`
     - [x] `bash scripts/precommit-run.sh`

10. T-010 - Align default compose startup and remove `up-test-env`

    - Scope:
      - Make `COMPOSE_OVERRIDE` default include test env + bitcoin mainnet/testnet4 overrides.
      - Remove `up-test-env` target from Makefile.
    - Linked requirements: FR-013, NFR-008
    - Validation:
      - [x] `make -n up`
      - [x] `docker compose -f deployments/compose/compose.yaml -f deployments/compose/compose.bitcoin.mainnet.yaml -f deployments/compose/compose.bitcoin.testnet4.yaml -f deployments/compose/compose.test-env.yaml config`
      - [x] `SPEC_DIR=\"specs/2026-03-05-blockchain-receipt-polling-service\" bash scripts/spec-lint.sh`

11. T-011 - Extract shared Postgres executor contract for repositories

    - Scope:
      - Move repository executor interface to a shared adapter-level contract in Postgres persistence package.
      - Reuse shared executor type across allocation and receipt-tracking repositories.
    - Linked requirements: NFR-006
    - Validation:
      - [x] `go test ./internal/adapters/outbound/persistence/postgres -count=1`
      - [x] `go test ./... -short -count=1`

12. T-012 - Move `PaymentReceiptObservation` to value objects

    - Scope:
      - Move `PaymentReceiptObservation` type from `entities` to `value_objects`.
      - Keep observation validation logic in the value object and adapt entity/use-case call sites.
    - Linked requirements: FR-004, FR-005, NFR-006
    - Validation:
      - [x] `go test ./internal/domain/value_objects ./internal/domain/entities -count=1`
      - [x] `go test ./internal/application/use_cases -count=1`

13. T-013 - Persist allocation `issued_at` into receipt tracking rows

    - Scope:
      - Extend receipt-tracking schema and registration SQL to store source allocation `issued_at`.
      - Ensure claimed tracking entities expose this value for observer lower-bound logic.
    - Linked requirements: FR-002, FR-014, NFR-009
    - Validation:
      - [x] `go test ./internal/adapters/outbound/persistence/postgres -count=1`
      - [x] `go test ./... -short -count=1`

14. T-014 - Extend observer contract for issue-time lower-bound

    - Scope:
      - Add issue-time lower-bound field to observer input contract and propagate through use case flow.
      - Keep chain-routed observer abstraction unchanged at core boundary.
    - Linked requirements: FR-014, FR-012, NFR-006, NFR-009
    - Validation:
      - [x] `go test ./internal/application/use_cases ./internal/adapters/outbound/blockchain -count=1`

15. T-015 - Replace UTXO snapshot aggregation with issue-time inbound aggregation

    - Scope:
      - Refactor Bitcoin adapter to aggregate inbound receipts from issue-time-scoped transaction data.
      - Remove dependency on global current-UTXO snapshot query path.
    - Linked requirements: FR-008, FR-014, NFR-009
    - Validation:
      - [x] `go test ./internal/adapters/outbound/bitcoin -count=1`
      - [x] `go test ./... -short -count=1`

16. T-016 - Add regression tests and spec/precommit verification for issue-time behavior

    - Scope:
      - Add tests covering pre-issue history exclusion and repeated-cycle idempotency.
      - Run full spec lint and precommit checks after implementation.
    - Linked requirements: FR-014, NFR-009
    - Validation:
      - [x] `go test ./internal/domain/... ./internal/application/use_cases ./internal/adapters/outbound/bitcoin -count=1`
      - [x] `SPEC_DIR=\"specs/2026-03-05-blockchain-receipt-polling-service\" bash scripts/spec-lint.sh`
      - [x] `bash scripts/precommit-run.sh`

17. T-017 - Add default Esplora endpoints in bitcoin compose overrides

    - Scope:
      - Set default `BITCOIN_MAINNET_ESPLORA_URL` and `BITCOIN_TESTNET4_ESPLORA_URL` to Esplora-compatible public endpoints for local startup convenience.
    - Linked requirements: FR-013, NFR-008
    - Validation:
      - [x] `docker compose -f deployments/compose/compose.yaml -f deployments/compose/compose.bitcoin.mainnet.yaml -f deployments/compose/compose.bitcoin.testnet4.yaml -f deployments/compose/compose.test-env.yaml config`

18. T-018 - Rename Bitcoin observer env and adapter naming to explicit Esplora terms

    - Scope:
      - Rename compose env keys from `BITCOIN_*_RPC_*` to `BITCOIN_*_ESPLORA_*`.
      - Rename Bitcoin adapter implementation naming away from `rpc_receipt_observer` to `esplora_receipt_observer`.
      - Update DI wiring and tests to use the new naming consistently.
    - Linked requirements: FR-008, FR-013, NFR-008
    - Validation:
      - [x] `go test ./internal/adapters/outbound/bitcoin ./internal/infrastructure/di -count=1`
      - [x] `docker compose -f deployments/compose/compose.yaml -f deployments/compose/compose.bitcoin.mainnet.yaml -f deployments/compose/compose.bitcoin.testnet4.yaml -f deployments/compose/compose.test-env.yaml config`

19. T-019 - Remove unused noop bitcoin receipt observer

    - Scope:
      - Delete `noop_receipt_observer` implementation and its isolated tests because poller runtime always uses Esplora observer wiring.
    - Linked requirements: FR-008, NFR-006
    - Validation:
      - [x] `go test ./internal/adapters/outbound/bitcoin -count=1`

20. T-020 - Replace prefix-based Esplora env parsing with explicit key loaders

    - Scope:
      - Remove `prefix`-concatenation style env resolution in poller DI.
      - Load mainnet/testnet4 Esplora config through explicit key sets to improve readability and reduce typo-prone runtime coupling.
    - Linked requirements: FR-008, NFR-006
    - Validation:
      - [x] `go test ./internal/infrastructure/di -count=1`
      - [x] `go test ./... -short -count=1`

21. T-021 - Decouple Bitcoin Esplora adapter from fixed mainnet/testnet fields

    - Scope:
      - Replace fixed `mainnet/testnet4` client fields with a network-keyed client map inside Bitcoin adapter.
      - Ensure constructor supports boot with only one configured network endpoint.
      - Keep deterministic error for unsupported or missing target network endpoint.
    - Linked requirements: FR-008, NFR-006, NFR-007
    - Validation:
      - [x] `go test ./internal/adapters/outbound/bitcoin -count=1`
      - [x] `go test ./... -short -count=1`

22. T-022 - Split multi-chain router and chain-specific observer interfaces

    - Scope:
      - Introduce separate outbound contracts for multi-chain routing input and chain-specific observer input.
      - Keep chain routing responsibility in `chain_router_receipt_observer` and remove redundant chain validation from Bitcoin adapter.
      - Update use case, DI, and tests to use the split contracts consistently.
    - Linked requirements: FR-012, NFR-006, NFR-007
    - Validation:
      - [x] `go test ./internal/application/use_cases ./internal/adapters/outbound/blockchain ./internal/adapters/outbound/bitcoin -count=1`
      - [x] `go test ./... -short -count=1`

## Traceability (optional)

- FR-001 -> T-001, T-003
- FR-002 -> T-002, T-003, T-013
- FR-003 -> T-002
- FR-004 -> T-001, T-012
- FR-005 -> T-001, T-012
- FR-006 -> T-002, T-003
- FR-007 -> T-002, T-005
- FR-008 -> T-004, T-008, T-015, T-018, T-019, T-020, T-021
- FR-009 -> T-002, T-003
- FR-010 -> T-006
- FR-011 -> T-007, T-009
- FR-012 -> T-008, T-009, T-014, T-022
- FR-013 -> T-010, T-017, T-018
- FR-014 -> T-013, T-014, T-015, T-016
- NFR-002 -> T-006
- NFR-006 -> T-005, T-011, T-012, T-014, T-019, T-020, T-021, T-022
- NFR-007 -> T-007, T-008, T-021, T-022
- NFR-008 -> T-010, T-017, T-018
- NFR-009 -> T-013, T-014, T-015, T-016

## Rollout and rollback

- Rollout:
  - Deploy with network-specific override pollers and Esplora env vars.
- Rollback:
  - Roll back poller image/config to previous stable release and disable override poller services.
