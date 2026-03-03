---
doc: 03_tasks
spec_date: 2026-03-03
slug: btc-xpub-address-api
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-swagger-ui-container-api-testing
  - 2026-03-03-cmd-app-compose-prefix
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
  - Change introduces new API contract shape, policy listing behavior, and chain-scoped path design.
  - Requires clear design for route parsing, error semantics, and policy enablement behavior.
- Upstream dependencies (`depends_on`):
  - `2026-03-03-swagger-ui-container-api-testing`
  - `2026-03-03-cmd-app-compose-prefix`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: not applicable
  - What would trigger switching to Full mode: not applicable
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable

## Milestones

- M1: Spec package updated to address-policy and chain-scoped contract.
- M2: Application/controller refactor completed with tests and four bitcoin schemes.
- M3: OpenAPI and compose-override behavior validated with bitcoin naming, scheme coverage, and amount metadata fields.
- M4: Outbound adapter scheme tests maintained as regression guard.

## Tasks (ordered)

1. T-001 - Update and lint Full-mode spec

   - Scope:
     - Update all five docs to reflect `addressPolicyId` + `/v1/chains/{chain}` API with `chain=bitcoin`, four scheme values, and `minorUnit`/`decimals` response fields.
   - Output:
     - `specs/2026-03-03-btc-xpub-address-api/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, NFR-004
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-03-btc-xpub-address-api" bash scripts/spec-lint.sh`
     - [x] Expected result: lint exits with code 0.
     - [x] Logs/metrics to check (if applicable): no lint failure output.

2. T-002 - Refactor application/domain/controller to policy-driven chain-scoped API

   - Scope:
     - Add chain/scheme value objects, address policy DTOs/use cases/controller updates for `chain=bitcoin`, and response metadata fields.
     - Fix xpub derivation depth semantics so account-level xpub derives via external branch `/0/index`.
   - Output:
     - New/updated files under `internal/domain`, `internal/application`, `internal/adapters/inbound`, `internal/infrastructure/di`, `internal/bootstrap`.
     - `internal/adapters/outbound/bitcoin/hd_xpub_address_deriver.go`
     - `internal/adapters/outbound/bitcoin/hd_xpub_address_deriver_test.go`
   - Linked requirements: FR-003, FR-004, FR-005, NFR-001, NFR-002, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`
     - [x] Expected result: all tests pass including controller/use-case coverage for success and error paths.
     - [x] Expected result: account-level xpub uses `/0/index`; change-level xpub uses direct `index`.
     - [x] Logs/metrics to check (if applicable): deterministic derivation for same policy/index input.

3. T-003 - Update OpenAPI to chain-scoped endpoints

   - Scope:
     - Replace BTC-specific derivation contract with chain-scoped policy list/generate contracts using bitcoin naming, four scheme values, and metadata fields.
     - Ensure Swagger UI resolves error response schemas without unresolved `$ref` errors.
     - Add pre-commit OpenAPI validation for `deployments/swagger/openapi.yaml` and spec validation for changed `specs/*` directories.
     - Use directory bind mount for swagger spec files to prevent stale single-file mount inode after atomic file rewrites.
   - Output:
     - `deployments/swagger/openapi.yaml`
     - `deployments/compose/compose.yaml`
     - `.pre-commit-config.yaml`
   - Linked requirements: FR-006, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): inspect OpenAPI paths and schemas; run swagger UI in compose stack if needed.
     - [x] How to verify (manual steps or command): `pre-commit run swagger-validation --all-files`
     - [x] How to verify (manual steps or command): `pre-commit run spec-lint --all-files`
     - [x] How to verify (manual steps or command): `docker compose -f deployments/compose/compose.yaml up -d --force-recreate swagger`
     - [x] How to verify (manual steps or command): compare host/container checksums for `deployments/swagger/openapi.yaml` and `/usr/share/nginx/html/specs/openapi.yaml`.
     - [x] Expected result: OpenAPI includes `/v1/chains/{chain}/address-policies` and `/v1/chains/{chain}/addresses`, and no unresolved schema reference errors.
     - [x] Expected result: pre-commit fails when OpenAPI schema references are broken, even when `openapi.yaml` is unchanged in the staged diff.
     - [x] Expected result: pre-commit fails when any changed spec folder fails `scripts/spec-lint.sh`.
     - [x] Expected result: swagger container serves the same openapi content as host file after edits.
     - [x] Logs/metrics to check (if applicable): none.

4. T-004 - Validate compose override and multi-file `COMPOSE_OVERRIDE`

   - Scope:
     - Verify mainnet/testnet4 scheme-specific overrides and multi-file compose override ergonomics.
   - Output:
     - `deployments/compose/compose.bitcoin.mainnet.yaml`
     - `deployments/compose/compose.bitcoin.testnet4.yaml`
     - `Makefile`
   - Linked requirements: FR-001, FR-002, NFR-002
   - Validation:
     - [x] How to verify (manual steps or command):
       - `docker compose -f deployments/compose/compose.yaml config`
       - `BITCOIN_MAINNET_LEGACY_XPUB=xpubDummy BITCOIN_MAINNET_SEGWIT_XPUB=xpubDummy BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB=xpubDummy BITCOIN_MAINNET_TAPROOT_XPUB=xpubDummy docker compose -f deployments/compose/compose.yaml -f deployments/compose/compose.bitcoin.mainnet.yaml config`
       - `BITCOIN_TESTNET4_LEGACY_XPUB=tpubDummy BITCOIN_TESTNET4_SEGWIT_XPUB=tpubDummy BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB=tpubDummy BITCOIN_TESTNET4_TAPROOT_XPUB=tpubDummy docker compose -f deployments/compose/compose.yaml -f deployments/compose/compose.bitcoin.testnet4.yaml config`
       - `BITCOIN_MAINNET_LEGACY_XPUB=xpubDummy BITCOIN_MAINNET_SEGWIT_XPUB=xpubDummy BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB=xpubDummy BITCOIN_MAINNET_TAPROOT_XPUB=xpubDummy BITCOIN_TESTNET4_LEGACY_XPUB=tpubDummy BITCOIN_TESTNET4_SEGWIT_XPUB=tpubDummy BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB=tpubDummy BITCOIN_TESTNET4_TAPROOT_XPUB=tpubDummy docker compose -f deployments/compose/compose.yaml -f deployments/compose/compose.bitcoin.mainnet.yaml -f deployments/compose/compose.bitcoin.testnet4.yaml config`
       - `BITCOIN_MAINNET_LEGACY_XPUB=xpubDummy BITCOIN_MAINNET_SEGWIT_XPUB=xpubDummy BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB=xpubDummy BITCOIN_MAINNET_TAPROOT_XPUB=xpubDummy BITCOIN_TESTNET4_LEGACY_XPUB=tpubDummy BITCOIN_TESTNET4_SEGWIT_XPUB=tpubDummy BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB=tpubDummy BITCOIN_TESTNET4_TAPROOT_XPUB=tpubDummy COMPOSE_OVERRIDE="deployments/compose/compose.bitcoin.mainnet.yaml deployments/compose/compose.bitcoin.testnet4.yaml" make -n up`
       - `BITCOIN_MAINNET_LEGACY_XPUB=xpubDummy BITCOIN_MAINNET_SEGWIT_XPUB=xpubDummy BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB=xpubDummy BITCOIN_MAINNET_TAPROOT_XPUB=xpubDummy BITCOIN_TESTNET4_LEGACY_XPUB=tpubDummy BITCOIN_TESTNET4_SEGWIT_XPUB=tpubDummy BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB=tpubDummy BITCOIN_TESTNET4_TAPROOT_XPUB=tpubDummy COMPOSE_OVERRIDE="deployments/compose/compose.bitcoin.mainnet.yaml,deployments/compose/compose.bitcoin.testnet4.yaml" make -n up`
     - [x] Expected result: compose renders expected env blocks; `COMPOSE_OVERRIDE` expands all override files.
     - [x] Logs/metrics to check (if applicable): command output includes both `-f` override files.

5. T-005 - Maintain outbound bitcoin adapter tests
   - Scope:
     - Keep outbound derivation tests for all supported schemes/networks and deterministic behavior.
     - Keep address encoder tests split by scheme (one file per encoder implementation).
   - Output:
     - `internal/adapters/outbound/bitcoin/hd_xpub_address_deriver_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_legacy_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_segwit_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_native_segwit_test.go`
     - `internal/adapters/outbound/bitcoin/address_encoder_taproot_test.go`
     - `internal/domain/value_objects/bitcoin_network_test.go`
     - `internal/domain/value_objects/bitcoin_address_scheme_test.go`
     - `internal/application/use_cases/address_policy_use_cases_test.go`
   - Linked requirements: FR-007, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/outbound/bitcoin`
     - [x] Expected result: all adapter scheme tests pass.
     - [x] Logs/metrics to check (if applicable): address type assertions match expected scheme.

## Traceability (optional)

- FR-001 -> T-001, T-004
- FR-002 -> T-001, T-004
- FR-003 -> T-001, T-002
- FR-004 -> T-001, T-002
- FR-005 -> T-001, T-002
- FR-006 -> T-001, T-003
- FR-007 -> T-001, T-005
- NFR-001 -> T-002
- NFR-002 -> T-002, T-004
- NFR-004 -> T-001
- NFR-005 -> T-002
- NFR-006 -> T-002, T-003, T-005

## Rollout and rollback

- Feature flag:
  - Policy enablement is controlled by xpub env var presence.
- Migration sequencing:
  - Not applicable.
- Rollback steps:
  - Revert chain-scoped controller/use case changes and restore previous route contract if needed.

## Ready-to-code checklist

- [x] Full-mode docs are present (`00` through `04`).
- [x] Frontmatter values are consistent across docs.
- [x] `depends_on` references existing DONE specs.
- [x] Mode decision and rationale are documented.
- [x] Requirement/task/test traceability IDs are present.
