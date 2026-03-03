---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Chain-scoped policy list endpoint.
  - Chain-scoped policy-based address generation endpoint for `legacy`, `segwit`, `nativeSegwit`, `taproot`.
  - Use-case policy resolution and error mapping.
  - Compose override behavior including multi-file `COMPOSE_OVERRIDE`.
  - OpenAPI contract updates for new paths.
- Not covered:
  - Production security controls and abuse prevention.
  - Non-BTC chain derivation runtime.

## Tests

### Unit

- TC-001: List policies for supported chain

  - Linked requirements: FR-003, NFR-002
  - Steps:
    - Execute list use case with bitcoin chain and mixed enabled/disabled policy config.
  - Expected:
    - Returns bitcoin policies with correct `enabled` values, includes `minorUnit`/`decimals`, and exposes four scheme values.

- TC-002: Generate address by policy success path

  - Linked requirements: FR-004, FR-005, NFR-001
  - Steps:
    - Execute generate use case with enabled `legacy`, `segwit`, `nativeSegwit`, and `taproot` policies and valid index.
  - Expected:
    - Returns deterministic response with policy metadata (`minorUnit`/`decimals`) and scheme-correct address type.

- TC-003: Generate address error paths

  - Linked requirements: FR-005, NFR-005
  - Steps:
    - Trigger unsupported chain, unknown policy, disabled policy, and deriver failure.
  - Expected:
    - Returns expected domain errors for controller mapping.

- TC-004: Outbound adapter scheme/type validation

  - Linked requirements: FR-007, NFR-006
  - Steps:
    - Run adapter unit tests for mainnet/testnet4 with `legacy`, `segwit`, `nativeSegwit`, and `taproot`.
    - Keep encoder tests split into four files:
      - `address_encoder_legacy_test.go`
      - `address_encoder_segwit_test.go`
      - `address_encoder_native_segwit_test.go`
      - `address_encoder_taproot_test.go`
    - Validate deterministic output for repeated input.
    - Validate unsupported network/scheme errors.
  - Expected:
    - Address types and encodings match scheme definition and all assertions pass.

- TC-005: Value-object parser normalization

  - Linked requirements: FR-004, NFR-006
  - Steps:
    - Test `ParseBitcoinNetwork` and `ParseBitcoinAddressScheme` with mixed-case and whitespace inputs.
    - Test unsupported values.
  - Expected:
    - Supported values are normalized and accepted; unsupported values are rejected.

- TC-006: Deriver configuration guard

  - Linked requirements: FR-007, NFR-006
  - Steps:
    - Construct deriver with empty encoder injection and try deriving with supported scheme.
  - Expected:
    - Returns explicit unsupported scheme error (detects DI misconfiguration early).

- TC-007: Use-case scheme routing table test

  - Linked requirements: FR-004, FR-007, NFR-006
  - Steps:
    - Execute generate use case with policy table containing all four schemes.
    - Assert deriver receives matching scheme for each policy.
  - Expected:
    - Every scheme policy routes to corresponding scheme value without mismatch.

- TC-008: Xpub depth path semantics
  - Linked requirements: FR-004, NFR-006
  - Steps:
    - Use account-level xpub (depth <= 3), derive address for index `n`, and compare result with expected `/0/n` derivation.
    - Use change-level xpub (depth >= 4), derive address for index `n`, and compare result with expected direct `n` derivation.
  - Expected:
    - Account-level xpub derivation matches `/0/index`.
    - Change-level xpub derivation matches direct `index`.

### Integration

- TC-101: Chain controller list endpoint

  - Linked requirements: FR-003, NFR-006
  - Steps:
    - Register routes and call `GET /v1/chains/bitcoin/address-policies`.
  - Expected:
    - Returns `200` with JSON list response.

- TC-102: Chain controller generate endpoint

  - Linked requirements: FR-004, FR-005, NFR-005
  - Steps:
    - Call `GET /v1/chains/bitcoin/addresses?addressPolicyId=<id>&index=<n>` in success and failure cases.
    - Open swagger UI and confirm no resolver errors for `400/404/405/500/501` responses on this endpoint.
    - Open swagger UI and confirm no resolver errors for `200` response schema on this endpoint.
  - Expected:
    - `200` on success; `400/404/405/501/500` mapped correctly; swagger schema references resolve cleanly.

- TC-103: Compose override rendering and make expansion

  - Linked requirements: FR-001, FR-002, NFR-002
  - Steps:
    - Render base compose and override combinations.
    - Validate `make -n up` with multi-file `COMPOSE_OVERRIDE` (space and comma forms).
  - Expected:
    - Base has no bitcoin xpub env; overrides inject expected env vars; make command expands all override files.

- TC-104: OpenAPI pre-commit guard

  - Linked requirements: FR-006, NFR-006
  - Steps:
    - Run `pre-commit run swagger-validation --all-files`.
  - Expected:
    - Command passes for valid OpenAPI and fails on broken `$ref`, including commits where `openapi.yaml` itself is not part of staged files.

- TC-105: Spec lint pre-commit guard

  - Linked requirements: FR-006, NFR-004
  - Steps:
    - Run `pre-commit run spec-lint --all-files`.
  - Expected:
    - Command passes for valid spec folders and fails when any changed spec folder violates lint rules.

- TC-106: Swagger bind-mount freshness
  - Linked requirements: FR-006, NFR-002
  - Steps:
    - Recreate swagger service: `docker compose -f deployments/compose/compose.yaml up -d --force-recreate swagger`.
    - Compare checksums for host `deployments/swagger/openapi.yaml` and container `/usr/share/nginx/html/specs/openapi.yaml`.
    - Request `GET http://localhost:8081/specs/openapi.yaml` and verify response contains `components.schemas.HealthResponse` and `components.schemas.ListAddressPoliciesResponse`.
  - Expected:
    - Host/container checksums match and served spec includes components required by `$ref`.

### E2E (if applicable)

- Scenario 1:
  - Start stack with mainnet override and call generation endpoint with `bitcoin-mainnet-legacy`.
  - Expected: `200` with derived address.
- Scenario 2:
  - Start stack with testnet4 override and call generation endpoint with `bitcoin-testnet4-native-segwit`.
  - Expected: `200` with derived address.
- Scenario 3:
  - Start stack without bitcoin override and call generation endpoint for bitcoin taproot policy.
  - Expected: `501` policy not enabled.

## Edge cases and failure modes

- Case: Unsupported chain in path.
- Expected behavior:

  - Return `404`.

- Case: Missing `addressPolicyId` query.
- Expected behavior:

  - Return `400`.

- Case: Invalid `index` (non-numeric/out-of-range).
- Expected behavior:

  - Return `400`.

- Case: Invalid xpub configured for enabled policy.
- Expected behavior:
  - Derivation fails and endpoint returns `500` while process stays alive.
- Case: Unsupported scheme value in internal policy config.
- Expected behavior:
  - Derivation fails with explicit scheme error and endpoint returns `500`.

## NFR verification

- Performance:
  - Verify local derivation request remains within 300ms target.
- Reliability:
  - Verify service starts without bitcoin overrides and list endpoint still reports policy states.
- Security:
  - Verify no private key material is introduced in code or compose settings.
