---
doc: 03_tasks
spec_date: 2026-03-20
slug: create2-eth-payment-receiving
mode: Full
status: READY
owners:
  - payrune-team
depends_on:
  - 2026-03-16-remove-xpub-fingerprint
  - 2026-03-05-blockchain-receipt-polling-service
  - 2026-03-08-payment-address-status-api
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# CREATE2 ETH Payment Receiving - Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - The change introduces a new external integration, contract deployment flow, new technical
    process state, and non-trivial failure and security behavior.
- Upstream dependencies (`depends_on`):
  - `2026-03-16-remove-xpub-fingerprint`
  - `2026-03-05-blockchain-receipt-polling-service`
  - `2026-03-08-payment-address-status-api`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - Not applicable.
  - What would trigger switching to Full mode:
    - Already Full because the feature includes contract behavior, migration changes, async worker
      design, and operational risks.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Make the issuance model and persistence schema capable of expressing Ethereum CREATE2 without
    breaking Bitcoin.
- M2:
  - Issue deterministic ETH payment addresses and observe ETH payments through the existing poller
    and status pipeline.
- M3:
  - Collect received ETH safely through a dedicated deploy-and-sweep flow and verify the full local
    end-to-end path.

## Tasks (ordered)

1. T-001 - Generalize issuance configuration and allocation metadata for CREATE2
   - Scope:
     - Refactor the issuance model so Bitcoin HD and Ethereum CREATE2 can coexist with explicit
       config instead of Bitcoin-biased placeholder fields.
     - Add the migration that replaces `account_public_key` and `derivation_path` semantics with
       neutral equivalents such as `address_source_ref` and `address_reference`.
     - Keep current Bitcoin allocation and status reads compatible with the updated schema.
   - Output:
     - Domain, application ports, policy reader, and persistence model can represent Ethereum
       CREATE2 issuance cleanly.
   - Linked requirements: FR-001, FR-002, FR-006, FR-007, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): run
           `go test ./internal/domain/... ./internal/adapters/outbound/policy ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
           and apply the migration on a disposable database.
     - [ ] Expected result: Bitcoin tests stay green, migration succeeds, and the updated schema no
           longer encodes Ethereum issuance behind xpub-specific names.
     - [ ] Logs/metrics to check (if applicable): none
2. T-002 - Add Ethereum CREATE2 address derivation and API wiring
   - Scope:
     - Add `ethereum` support to chain parsing and policy configuration.
     - Expose both `ethereum/mainnet` and `ethereum/sepolia` CREATE2 policies through explicit
       collector runtime config, while keeping deployment-derived factory and init-code inputs out
       of env wiring.
     - Keep local issuance testable before T-003 by using checked-in deterministic fixture metadata
       instead of operator-supplied deployment env vars.
     - Implement an Ethereum CREATE2 deriver adapter and wire it through allocation flow using
       non-public per-allocation salt material derived from runtime-managed secret input rather
       than a publicly enumerable index-only salt rule.
     - Keep `GET /v1/chains/{chain}/addresses` available for Bitcoin, but reject or disable
       Ethereum CREATE2 policies so the public route cannot enumerate future payment addresses.
     - Add or update controller and use-case tests so Ethereum policy listing and address issuance
       work through the existing chain-scoped HTTP routes.
   - Output:
     - `payrune` can issue deterministic ETH payment addresses for configured `mainnet` and
       `sepolia` policies through the current API contract.
   - Linked requirements: FR-001, FR-002, FR-005, FR-008, NFR-001, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): run
           `go test ./internal/application/usecases ./internal/adapters/inbound/http/controllers ./internal/adapters/outbound/blockchain`
           plus CREATE2 prediction vector tests.
     - [ ] Expected result: Ethereum policy list and allocation tests pass, and the same input
           produces the same predicted ETH address in Go and contract vectors without leaving a
           public index-based enumeration path enabled for Ethereum.
     - [ ] Logs/metrics to check (if applicable): none
3. T-003 - Add factory and receiver contract artifacts plus local Ethereum tooling
   - Scope:
     - Add the contract source or generated artifacts required for CREATE2 prediction and
       deployment.
     - Add checked-in deployment metadata so runtime policy wiring can resolve factory addresses
       without operator-supplied env vars.
     - Define the contract caller model explicitly so factory identity stays fixed per address
       space while operator-signer credentials remain rotatable runtime configuration.
     - Update prediction tooling to work from explicit stored salt material or operator-supplied
       verification inputs, not from a public sequential index that would reveal future addresses.
     - Add Go CLI tooling under `cmd/` for contract build, prediction, and explicit chain
       verification, with only thin wrappers left under `scripts/`.
     - Keep verification network-driven so CREATE2 tooling can exercise a configured Ethereum RPC
       endpoint without repo-managed devnet infrastructure.
   - Output:
     - Local tooling can deploy the factory, compute expected addresses, and fund predicted payment
       addresses against a configured verification chain, and the resulting metadata clearly
       separates fixed factory inputs from rotatable operator-signer inputs.
   - Linked requirements: FR-002, FR-006, FR-007, FR-008, NFR-003
   - Validation:
     - [ ] How to verify (manual steps or command): run the contract deploy and prediction
           verification script against the configured Ethereum RPC network.
     - [ ] Expected result: deployed factory metadata matches the Go-side predictor and one funded
           predicted address can later be deployed at the expected address.
     - [ ] Logs/metrics to check (if applicable): capture factory address, init code hash, and
           predicted receiver address
4. T-004 - Implement Ethereum native ETH receipt observation in the poller
   - Scope:
     - Add an Ethereum observer adapter that scans bounded block ranges for native ETH transfers to
       issued payment addresses.
     - Wire the observer into the existing chain-routed receipt poller and DI setup.
     - Preserve current payment status transitions and row-level retry behavior.
   - Output:
     - Receipt polling updates Ethereum payment status through the existing receipt-tracking model.
   - Linked requirements: FR-004, FR-005, FR-006, NFR-002, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): run
           `go test ./internal/application/usecases ./internal/adapters/outbound/blockchain`
           and a verification-network smoke that funds one issued ETH payment address, then runs
           one poller cycle.
     - [ ] Expected result: the Ethereum receipt row moves from `watching` to the expected paid
           state with correct ETH totals in `wei`.
     - [ ] Logs/metrics to check (if applicable): poller summary logs include Ethereum chain,
           address, scan range, and observed totals without raw salt leakage
5. T-005 - Implement deploy-and-sweep technical state and worker
   - Scope:
     - Add explicit Ethereum CREATE2 technical persistence for deployment and sweep progress.
     - Implement a dedicated use case and worker or scheduler path that deploys funded receivers and
       sweeps ETH to the configured collector idempotently.
     - Handle already-deployed and already-swept cases safely.
   - Output:
     - A funded deterministic ETH payment address can be collected into the operator collector
       address without duplicate collection on retry.
   - Linked requirements: FR-003, FR-006, FR-007, NFR-002, NFR-003, NFR-005
   - Validation:
     - [ ] How to verify (manual steps or command): execute an end-to-end smoke on the configured
           verification network:
           allocate ETH address -> fund predicted address -> poll payment -> run sweeper twice.
     - [ ] Expected result: first sweep deploys and collects ETH to the collector; second sweep is
           a deterministic no-op or already-complete outcome; persisted technical state includes tx
           hashes.
     - [ ] Logs/metrics to check (if applicable): sweeper logs include payment address id,
           receiver address, deploy tx hash, and sweep tx hash
6. T-006 - Update docs, contracts, and verification evidence for rollout
   - Scope:
     - Update OpenAPI or API docs, operator env examples, and Ethereum runtime documentation.
     - Document the privacy boundary explicitly: public Ethereum preview-by-index is disabled, and
       v1 privacy covers future-address enumeration resistance rather than post-sweep anonymity.
     - Capture spec, migration, and local smoke verification evidence before implementation is
       marked complete.
   - Output:
     - Documentation and validation evidence reflect the final ETH CREATE2 flow and rollout
       constraints.
   - Linked requirements: FR-005, FR-006, FR-008, NFR-001, NFR-004, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): run
           `SPEC_DIR="specs/2026-03-20-create2-eth-payment-receiving" bash scripts/spec-lint.sh`,
           `go list ./...`, `go test ./...`, and `bash scripts/precommit-run.sh`.
     - [ ] Expected result: spec lint passes, package graph is clean, repository tests are green,
           and the ETH CREATE2 flow is documented.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002, T-003
- FR-003 -> T-005
- FR-004 -> T-004
- FR-005 -> T-002, T-004, T-006
- FR-006 -> T-001, T-003, T-004, T-005, T-006
- FR-007 -> T-001, T-003, T-005
- FR-008 -> T-002, T-003, T-006
- NFR-001 -> T-002, T-006
- NFR-002 -> T-004, T-005
- NFR-003 -> T-002, T-003, T-005
- NFR-004 -> T-006
- NFR-005 -> T-004, T-005
- NFR-006 -> T-001, T-002, T-004, T-006

## Rollout and rollback

- Feature flag:
  - Keep Ethereum CREATE2 policies disabled by default until factory, observer, and sweeper
    verification is complete.
- Migration sequencing:
  - Apply the allocation-schema cleanup migration and Ethereum CREATE2 technical table migration
    before enabling the Ethereum policy.
  - Deploy binaries or workers that understand the new schema before issuing any Ethereum payment
    addresses.
- Rollback steps:
  - Disable Ethereum policies first.
  - Stop the sweeper runtime before reverting binaries.
  - If no Ethereum payment addresses have been issued yet, the new migrations can be rolled back in
    the usual way.
  - If live Ethereum CREATE2 rows or deployment state already exist, prefer a forward fix or
    database snapshot restore over destructive partial rollback.
