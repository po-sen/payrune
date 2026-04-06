---
doc: 03_tasks
spec_date: 2026-04-05
slug: ethereum-usdt-payment-receiving
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
  - 2026-03-30-eth-balance-receipt-observer
  - 2026-04-03-ethereum-ledger-batch-sweep
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - This feature adds new persistence fields, application/API contract changes, Ethereum observer
    behavior, and CREATE2 contract/sweep changes. Quick mode would hide too many design decisions.
- Upstream dependencies (`depends_on`): `2026-03-20-create2-eth-payment-receiving`,
  `2026-03-30-eth-balance-receipt-observer`, `2026-04-03-ethereum-ledger-batch-sweep`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`

## Milestones

- M1:
  - Land the `asset_reference`-based payment model and remove redundant asset-shape fields from
    runtime contracts.
- M2:
  - Land ERC-20 observation and token-capable recovery, then verify the full flow.
- M3:
  - Land one-signature ERC-20 recovery, cut over to one current receiver/factory artifact set,
    and keep the Ledger USDT payment helper.

## Tasks (ordered)

1. T-001 - Finalize the Ethereum USDT feature spec and readiness gate
   - Scope:
     - Complete the Full-mode spec package for explicit Ethereum USDT receiving, including data
       model, observer, and recovery decisions.
   - Output:
     - Lintable spec folder `specs/2026-04-05-ethereum-usdt-payment-receiving/`.
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-04-05-ethereum-usdt-payment-receiving" bash scripts/spec-lint.sh`
     - [x] Expected result: spec lint passes and all five docs agree on mode, links, dependencies, and traceability.
     - [x] Logs/metrics to check (if applicable): none
1. T-002 - Implement asset-aware persistence, domain/application contracts, and HTTP/OpenAPI responses

   - Scope:
     - Rewrite migration `000017`, replace the issued-row asset metadata draft with nullable
       `asset_reference`, copy that snapshot into `payment_receipt_trackings`, wire the new model
       through domain/application contracts, remove `assetKind` / `tokenStandard` from the formal
       runtime/API/webhook contracts, remove `minorUnit` from policy and public contracts, remove
       `AssetSymbol` from core and public contracts, drop the thin `PaymentAssetReference` wrapper
       in favor of a plain nullable string, remove the thin `ConfiguredAssetReference()` helper,
       simplify bootstrap policy validation in place, tidy shared Ethereum address helpers into a
       dedicated file, and keep Bitcoin/native ETH compatibility intact.
   - Output:
     - Updated entities, ports, persistence adapters, migrations, controllers, OpenAPI, and
       bootstrap policy wiring for USDT policies without `assetKind`, `tokenStandard`,
       `minorUnit`, `AssetSymbol`, or a standalone `PaymentAssetReference` value object.
   - Linked requirements: FR-001, FR-002, FR-005, NFR-001, NFR-002, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/application/... ./internal/adapters/inbound/http/... ./internal/adapters/outbound/persistence/... ./internal/bootstrap/...`
     - [x] Expected result: payment allocation/status tests pass with nullable `asset_reference`
           persistence changes and tracking no longer needs allocation joins for token identity.
     - [x] Logs/metrics to check (if applicable): none

1. T-003 - Implement ERC-20 balance observation for Ethereum USDT payment rows

   - Scope:
     - Extend the Ethereum observer and polling wiring so ERC-20 USDT rows are observed via
       `balanceOf(address)` snapshots at latest and confirmed block heights using parsed
       `asset_reference` input where `NULL` means native asset and non-`NULL` means ERC-20 token.
   - Output:
     - Observer adapter changes, poller contract updates, and regression tests for ETH/native rows
       plus new USDT rows.
   - Linked requirements: FR-002, FR-003, NFR-002, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/outbound/ethereum ./internal/adapters/outbound/blockchain ./internal/application/usecases`
     - [x] Expected result: observer tests cover native ETH and ERC-20 branches and poller tests pass without cross-chain regressions.
     - [x] Logs/metrics to check (if applicable): row-level polling failures still use existing failure reasons.

1. T-004 - Extend CREATE2 receiver artifacts and sweep tooling for ERC-20 recovery

   - Scope:
     - Update the CREATE2 receiver/factory path, sweep material, and operator scripts so selected
       USDT rows can be recovered safely with Ledger, including one-signature batch recovery for
       multiple compatible token receivers and a cutover from dual receiver families to one
       current receiver/factory artifact set, then remove unshipped token-only legacy receiver
       source and artifact files from the checked-in asset bundle.
   - Output:
     - Updated Solidity source/artifacts, sweep payload shape, and sweep tooling validations for
       ERC-20 token recovery.
   - Linked requirements: FR-004, FR-005, NFR-003, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `bash scripts/ethereum_create2_build_artifacts.sh`, `bash -n scripts/ethereum_create2_sweep.sh`, `go test ./internal/infrastructure/ethereumcreate2assets ./internal/adapters/outbound/ethereum`
     - [x] Expected result: artifacts rebuild cleanly, scripts remain shell-valid, and recovery payload/contract tests pass for the current receiver/factory artifacts plus one-signature batch ERC-20 recovery and row-owned factory handling.
     - [x] Logs/metrics to check (if applicable): dry-run output includes token-aware validation context.

1. T-005 - Add a small Ledger-signed USDT payment helper

   - Scope:
     - Add one operator helper under `scripts/` that sends USDT with a Ledger signer for Sepolia or
       mainnet testing.
   - Output:
     - Ledger-only USDT payment helper and concise operator docs.
   - Linked requirements: FR-006, NFR-003, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `bash -n scripts/ethereum_usdt_pay_with_ledger.sh`, `bash scripts/ethereum_usdt_pay_with_ledger.sh --help`
     - [x] Expected result: helper is shell-valid, documented, and dry-run capable.
     - [x] Logs/metrics to check (if applicable): helper output includes network, sender, recipient, amount, and asset reference.

1. T-006 - Verify integrated Ethereum USDT flow and update operator/public docs
   - Scope:
     - Update README/examples and run targeted end-to-end validation for allocation, status, and
       operator recovery paths.
   - Output:
     - Updated docs plus recorded validation commands for repo review.
   - Linked requirements: FR-001, FR-003, FR-004, FR-005, FR-006, NFR-003, NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`, `go list ./...`, `SPEC_DIR="specs/2026-04-05-ethereum-usdt-payment-receiving" bash scripts/spec-lint.sh`
     - [x] Expected result: targeted tests and repo listing succeed, and docs/examples match the shipped USDT flow.
     - [x] Logs/metrics to check (if applicable): none

## Validation evidence

- `GOCACHE=/tmp/go-build go test ./internal/...`
- `GOCACHE=/tmp/go-build go test ./...`
- `go test ./internal/bootstrap ./internal/application/usecases ./internal/adapters/outbound/persistence/...`
- `go list ./...`
- `bash scripts/ethereum_create2_build_artifacts.sh`
- `bash -n scripts/ethereum_create2_sweep.sh`
- `bash -n scripts/ethereum_usdt_pay_with_ledger.sh`
- `SPEC_DIR="specs/2026-04-05-ethereum-usdt-payment-receiving" bash scripts/spec-lint.sh`
- `bash scripts/precommit-run.sh`

## Traceability (optional)

- FR-001 -> T-001, T-002, T-006
- FR-002 -> T-001, T-002, T-003
- FR-003 -> T-001, T-003, T-006
- FR-004 -> T-001, T-004, T-006
- FR-005 -> T-001, T-002, T-004, T-006
- FR-006 -> T-001, T-005, T-006
- NFR-001 -> T-002
- NFR-002 -> T-002, T-003
- NFR-003 -> T-004, T-005, T-006
- NFR-005 -> T-003, T-004, T-005, T-006
- NFR-006 -> T-001, T-002, T-003, T-004, T-005

## Rollout and rollback

- Feature flag:
  - No dedicated feature flag; rollout is controlled by whether explicit Ethereum USDT policy
    configuration is present.
- Migration sequencing:
  - Apply DB migration before deploying binaries that persist or read USDT asset metadata.
  - Rebuild CREATE2 artifacts before using ERC-20 recovery tooling.
- Rollback steps:
  - Remove USDT policy env configuration to disable new issuance.
  - Roll back binaries and, if necessary, revert the migration only before production data depends
    on the new asset columns.
  - Keep existing Bitcoin/native ETH flows available during rollback.
