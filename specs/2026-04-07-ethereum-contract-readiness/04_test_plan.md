---
doc: 04_test_plan
spec_date: 2026-04-07
slug: ethereum-contract-readiness
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-05-ethereum-usdt-payment-receiving
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
  - Ethereum issuance readiness checks for factory and token contracts
  - API bootstrap enforcement before serving requests
  - API request/error logging for diagnosable issuance failures
  - Regression coverage for Bitcoin and native ETH policy handling
- Not covered:
  - Real mainnet broadcasts or long-lived readiness caching
  - Non-Ethereum chains

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-002, FR-003, FR-004 / NFR-006
  - Steps:
    - Add checker tests for missing factory code, mismatched factory runtime hash, missing token code, failed `balanceOf`, and decimals mismatch.
  - Expected:
    - The checker returns not-ready results only for the affected Ethereum policy/network.
- TC-002:
  - Linked requirements: FR-001, FR-001A, FR-004 / NFR-002, NFR-006
  - Steps:
    - Add use-case tests confirming generate/allocate no longer depend on a readiness checker and keep normal behavior after startup validation has already succeeded.
    - Add policy/bootstrap tests confirming disabled policies are skipped, while enabled policies with missing required static config fail startup before readiness.
  - Expected:
    - Generate/allocate behavior stays unchanged and has no request-time readiness dependency.
    - Explicit `enabled` intent stays separate from static-config validation and startup-time readiness.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-004 / NFR-003, NFR-006
  - Steps:
    - Exercise API bootstrap with valid and invalid Ethereum readiness inputs.
  - Expected:
    - API startup fails closed on invalid readiness and succeeds on valid readiness.
- TC-102:
  - Linked requirements: FR-002, FR-003 / NFR-001, NFR-005
  - Steps:
    - Build the API container with Ethereum RPC config and verify Ethereum readiness checks can be constructed from existing env names.
  - Expected:
    - API bootstrap reuses existing Ethereum RPC env config and does not require duplicate API-only env names.
- TC-103:
  - Linked requirements: FR-005 / NFR-005
  - Steps:
    - Exercise the public router and mapped controller failures, then inspect captured logs.
  - Expected:
    - Request logs include method/path/status/duration, and mapped failures include internal error details.

### E2E (if applicable)

- Scenario 1:
  - Start the API with a deliberately wrong Ethereum factory or token contract and verify the API process fails to start.
- Scenario 2:
  - Start the API with valid Ethereum config and verify Ethereum issuance still succeeds.

## Edge cases and failure modes

- Case:
  - Ethereum policy is enabled and configured but no RPC config exists for its network.
  - Expected behavior:
    - API startup fails closed.
- Case:
  - Ethereum USDT policy is explicitly enabled but `assetReference` is empty.
  - Expected behavior:
    - API startup fails during static-config validation before readiness begins.
- Case:
  - Native ETH policy uses a valid factory but empty `assetReference`.
  - Expected behavior:
    - Readiness succeeds without token-contract checks.
- Case:
  - Token policy factory is valid but token `decimals()` mismatches policy decimals.
  - Expected behavior:
    - Issuance fails before address generation/allocation.

## NFR verification

- Performance:
  - Bitcoin request tests confirm no Ethereum RPC dependency is introduced for Bitcoin paths.
- Reliability:
  - Startup validation tests confirm failed readiness does not create issued allocations or idempotency side effects because the API never begins serving requests.
- Security:
  - Error and log assertions confirm no sensitive recovery material is exposed.
