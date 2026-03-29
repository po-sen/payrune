---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale: This change is limited to Swagger/OpenAPI defaults plus one small inbound request-decoding compatibility update; no new integration, schema migration, or architectural redesign is needed.
- Upstream dependencies (`depends_on`): `2026-03-03-swagger-ui-container-api-testing`, `2026-03-20-create2-eth-payment-receiving`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: OpenAPI example edits and nullable request parsing do not introduce new data flow or deployment concerns.
  - What would trigger switching to Full mode: If the API contract had to change beyond optional-null handling or if Swagger generation became code-generated instead of static.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): Each task includes concrete `go test`, grep, spec-lint, and repo validation commands.

## Milestones

- M1: Update Swagger defaults/examples to non-mainnet values.
- M2: Add null-tolerant customer reference handling and close validation.

## Tasks (ordered)

1. T-001 - Update Swagger OpenAPI defaults to non-mainnet examples
   - Scope: Adjust payment-address request/response/schema examples in `deployments/swagger/openapi.yaml` to use testnet4 / sepolia values, BTC `2000`, ETH `100000000000000`, and nullable/default-null customer reference documentation.
   - Output: Swagger UI shows non-mainnet defaults for local payment-address testing.
   - Linked requirements: FR-001 / FR-002 / NFR-003 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `rg -n "bitcoin-testnet4|ethereum-sepolia|100000000000000|customerReference|nullable|default: null" deployments/swagger/openapi.yaml`
     - [x] Expected result: OpenAPI spec shows non-mainnet example values and nullable/default-null `customerReference`.
     - [x] Logs/metrics to check (if applicable): N/A
2. T-002 - Accept null customerReference and run validation
   - Scope: Update the allocate controller to accept `customerReference: null`, add controller coverage, and run repo validation.
   - Output: Swagger docs and actual request decoding stay aligned.
   - Linked requirements: FR-002 / FR-003 / NFR-001 / NFR-002 / NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/inbound/http/controllers`, `SPEC_DIR="specs/2026-03-30-swagger-non-mainnet-defaults" bash scripts/spec-lint.sh`, `bash scripts/precommit-run.sh`
     - [x] Expected result: controller tests pass with a null customer reference payload, spec lint passes, and repo precommit validation stays green.
     - [x] Logs/metrics to check (if applicable): N/A

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001, T-002
- FR-003 -> T-002
- NFR-001 -> T-002
- NFR-002 -> T-002
- NFR-003 -> T-001
- NFR-005 -> T-002
- NFR-006 -> T-001

## Rollout and rollback

- Feature flag: None
- Migration sequencing: update static OpenAPI examples first, then make controller null-tolerant, then validate
- Rollback steps: revert the OpenAPI example changes and the nullable controller decoding change together if Swagger validation or request compatibility regresses

## Validation evidence

- `rg -n "bitcoin-testnet4|ethereum-sepolia|100000000000000|customerReference|nullable|default: null" deployments/swagger/openapi.yaml` passed.
- `go test ./internal/adapters/inbound/http/controllers` passed.
- `SPEC_DIR="specs/2026-03-30-swagger-non-mainnet-defaults" bash scripts/spec-lint.sh` passed.
- `bash scripts/precommit-run.sh` passed.
