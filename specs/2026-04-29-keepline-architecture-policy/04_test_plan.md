---
doc: 04_test_plan
spec_date: 2026-04-29
slug: keepline-architecture-policy
mode: Full
status: DONE
owners:
  - repo-maintainers
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Test Plan

## Scope

- Covered: Spec lint, keepline configuration, keepline import policy, Go tests, adapter import
  searches, and pre-commit default-stage validation.
- Not covered: Live external blockchain, PostgreSQL, Cloudflare, or deployed API validation.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-004, FR-005, FR-008, NFR-001, NFR-002
  - Steps: Run `go test ./...`.
  - Expected: Existing tests pass.
- TC-002:
  - Linked requirements: FR-004, FR-005, NFR-006
  - Steps: Run `rg -n "payrune/internal/ports" internal cmd`.
  - Expected: No old top-level port imports remain.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-003, FR-004, FR-005, FR-006, FR-007, NFR-003, NFR-006
  - Steps: Run `keepline config-check`, `keepline import-check`, and `keepline check --scope all`.
  - Expected: Strict config, file, and import policies pass.
- TC-102:
  - Linked requirements: FR-005, FR-006, NFR-006
  - Steps: Run `rg -n '"payrune/internal/adapters/' internal/adapters -g '*.go'`.
  - Expected: Only allowed same-family inbound HTTP references appear.
- TC-103:
  - Linked requirements: FR-002, FR-008, NFR-002, NFR-003, NFR-004, NFR-006
  - Steps: Run `bash scripts/precommit-run.sh`.
  - Expected: Default-stage hooks pass and include `keepline import check`.
- TC-104:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, NFR-005
  - Steps: Run `SPEC_DIR="specs/2026-04-29-keepline-architecture-policy" bash scripts/spec-lint.sh`.
  - Expected: Consolidated spec lint passes.

### E2E (if applicable)

- Scenario 1: Not applicable; external APIs and deployment behavior are unchanged.
- Scenario 2: Not applicable.

## Edge cases and failure modes

- Case: A future adapter imports `internal/domain/**`.
  - Expected behavior: `keepline import-check` fails.
- Case: A future adapter imports `internal/application/dto/**` or `internal/application/outbox/**`.
  - Expected behavior: `keepline import-check` fails.
- Case: A future inbound adapter imports an outbound adapter implementation.
  - Expected behavior: `keepline import-check` fails.
- Case: A future outbound adapter family imports another outbound adapter family.
  - Expected behavior: `keepline import-check` fails.
- Case: A port package imports domain or implementation packages.
  - Expected behavior: `keepline import-check` fails.

## NFR verification

- Performance: Review confirms the refactor adds only in-process mapping and no new external IO.
- Reliability: `go test ./...` and pre-commit pass.
- Security: keepline file policy and private-key detection pass in pre-commit.
- Maintainability: keepline import policy covers layer and adapter-family boundaries.

## Validation evidence

- `keepline config-check`: passed.
- `keepline import-check`: passed.
- `keepline check --scope all`: passed.
- `go test ./...`: passed.
- `bash scripts/precommit-run.sh`: passed.
- `SPEC_DIR="specs/2026-04-29-keepline-architecture-policy" bash scripts/spec-lint.sh`: passed.
