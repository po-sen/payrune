---
doc: 04_test_plan
spec_date: 2026-03-15
slug: http-route-table
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
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
  - Public HTTP route registration centralization, controller routing cleanup, and router naming
    alignment.
- Not covered:
  - Public API contract changes.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002, NFR-002
  - Steps:
    - Run router and controller unit tests after centralizing the route table.
  - Expected:
    - Exact public routes still resolve, and controller endpoint handlers still execute correctly.
- TC-002:
  - Linked requirements: FR-003, NFR-001
  - Steps:
    - Run the existing chain-address and health controller tests.
  - Expected:
    - Existing route and response behavior stays intact.

### Integration

- TC-101:
  - Linked requirements: FR-003, FR-004, NFR-001, NFR-002
  - Steps:
    - Run:
      - `go test ./internal/adapters/inbound/http/...`
      - `go list ./...`
      - `go test ./...`
  - Expected:
    - The HTTP layer compiles cleanly and the repo remains green.

## Edge cases and failure modes

- Case:
  - Unknown or partially matched `/v1/chains/...` paths.
- Expected behavior:
  - The router returns `404` without requiring controller-internal path parsing.

## NFR verification

- Reliability:
  - Targeted and full Go verification commands pass.
- Maintainability:
  - Public route definitions are visible from one routing file with router-specific naming.
