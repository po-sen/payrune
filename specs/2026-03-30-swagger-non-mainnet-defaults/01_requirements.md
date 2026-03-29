---
doc: 01_requirements
spec_date: 2026-03-30
slug: swagger-non-mainnet-defaults
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-03-swagger-ui-container-api-testing
  - 2026-03-20-create2-eth-payment-receiving
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Swagger default:
- The example/default payload or schema value shown to a user in local Swagger UI before they edit it.

## Out-of-scope behaviors

- OOS1: No new endpoints or schema fields.
- OOS2: No change to payment business semantics beyond null-tolerant optional `customerReference` decoding.

## Functional requirements

### FR-001 - Payment-allocation Swagger examples must default to non-mainnet values

- Description: OpenAPI examples/defaults used for local payment-allocation testing must prefer non-mainnet address policies and network examples.
- Acceptance criteria:
  - [ ] The Bitcoin allocate example uses a testnet4 policy and `expectedAmountMinor: 2000`.
  - [ ] The Ethereum allocate example uses a Sepolia policy and `expectedAmountMinor: 100000000000000`.
  - [ ] Related success/status examples that Swagger surfaces by default no longer point to mainnet policies.
- Notes: The Ethereum amount is expressed in wei even though the user-facing intent is `0.0001 ETH`.

### FR-002 - customerReference must be documented and accepted as null

- Description: The allocate request contract must support `customerReference: null` as an optional input and document that default clearly.
- Acceptance criteria:
  - [ ] `deployments/swagger/openapi.yaml` marks `customerReference` as nullable and shows a default/example of `null`.
  - [ ] The allocate controller accepts JSON `null` for `customerReference` and passes an empty string into the use case.
  - [ ] Controller tests cover a request body with `customerReference: null`.
- Notes: Omitted and null customer references should have the same runtime meaning.

### FR-003 - Swagger contract must remain valid under repo validation

- Description: The updated OpenAPI document and controller behavior must pass the repo's standard validation workflow.
- Acceptance criteria:
  - [ ] `SPEC_DIR="specs/2026-03-30-swagger-non-mainnet-defaults" bash scripts/spec-lint.sh` passes.
  - [ ] `bash scripts/precommit-run.sh` passes with the updated OpenAPI spec and controller tests.
- Notes: This guards against invalid OpenAPI nullable/default syntax.

## Non-functional requirements

- Performance (NFR-001): No measurable runtime performance impact; only request decoding for one optional field changes.
- Availability/Reliability (NFR-002): Existing allocate requests with a string `customerReference` or omitted field continue to behave the same.
- Security/Privacy (NFR-003): Local Swagger documentation should avoid presenting mainnet-looking defaults as the first suggested values.
- Compliance (NFR-004):
- Observability (NFR-005): Validation must include both OpenAPI lint/validation and controller test coverage for null handling.
- Maintainability (NFR-006): Swagger examples should make local testing intent obvious without requiring verbal explanation.

## Dependencies and integrations

- External systems: Swagger UI / OpenAPI validation through repo tooling
- Internal services: `deployments/swagger/openapi.yaml` and `internal/adapters/inbound/http/controllers`
