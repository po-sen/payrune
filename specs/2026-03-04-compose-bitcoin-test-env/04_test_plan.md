---
doc: 04_test_plan
spec_date: 2026-03-04
slug: compose-bitcoin-test-env
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-deploy-service-compose-dockerfile
  - 2026-03-04-bitcoin-address-vectors
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Test Plan

## Scope

- Covered:
  - Compose override injection of bitcoin xpub fixture values.
  - Makefile startup rule for test env profile.
- Not covered:
  - Runtime business logic of address derivation APIs.

## Tests

### Unit

- TC-001: Spec consistency
  - Linked requirements: FR-001, FR-002, FR-003, NFR-004
  - Steps:
    - Run `SPEC_DIR="specs/2026-03-04-compose-bitcoin-test-env" bash scripts/spec-lint.sh`.
  - Expected:
    - Spec lint passes.

### Integration

- TC-101: Compose config merge

  - Linked requirements: FR-001, NFR-002, NFR-003, NFR-006
  - Steps:
    - Run `docker compose -f deployments/compose/compose.yaml -f deployments/compose/compose.test-env.yaml config`.
  - Expected:
    - Output includes all eight bitcoin xpub keys for `services.app.environment` with hardcoded fixture values.

- TC-102: Make startup rule
  - Linked requirements: FR-002, FR-003, NFR-001, NFR-005
  - Steps:
    - Run `make up-test-env`.
    - Optionally inspect config via `docker compose -f deployments/compose/compose.yaml -f deployments/compose/compose.test-env.yaml config`.
  - Expected:
    - Services start and make rule applies the new override file while preserving existing `up/down` behavior.

### E2E (if applicable)

- Scenario 1:
  - Not applicable.
- Scenario 2:
  - Not applicable.

## Edge cases and failure modes

- Case:
  - Fixture drift between unit tests and compose override.
- Expected behavior:
  - Validation review catches mismatched xpub values before merge.

## NFR verification

- Performance:
  - Startup command runtime remains comparable to `make up`.
- Reliability:
  - Re-running startup with same files yields same env values.
- Security:
  - Confirm no private keys exist in override file.
