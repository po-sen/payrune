---
doc: 01_requirements
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

# Requirements

## Out-of-scope behaviors

- OOS1:
  - Changing any existing public API path shape.
- OOS2:
  - Changing business logic or use-case orchestration.

## Functional requirements

### FR-001 - Public HTTP route table is centralized in router.go

- Description:
  - Public HTTP path registrations must be declared in `internal/adapters/inbound/http/router.go`.
- Acceptance criteria:
  - [ ] `router.go` explicitly registers `/health`.
  - [ ] `router.go` explicitly registers each `/v1/chains/...` route used by the public API.
  - [ ] `router.go` no longer delegates route discovery to controller `RegisterRoutes(...)`
        methods.
- Notes:
  - `ServeMux` path patterns may be used.

### FR-002 - Controllers stop acting as nested routers

- Description:
  - Controllers must handle endpoint-specific requests without manually dispatching across nested
    `/v1/chains/...` resources.
- Acceptance criteria:
  - [ ] `parseChainRoute(...)` is removed from the chain-address controller.
  - [ ] The chain-address controller no longer switches on resource names parsed from a shared path
        prefix.
  - [ ] Endpoint-specific controller methods are directly callable from the route table.
- Notes:
  - Controllers may still validate methods, parse path values, and map errors.

### FR-003 - Public HTTP behavior remains unchanged

- Description:
  - Existing public HTTP routes must keep the same externally visible behavior after the refactor.
- Acceptance criteria:
  - [ ] `/health` still responds successfully for `GET`.
  - [ ] Existing `/v1/chains/...` route tests continue to pass.
  - [ ] `go test ./...` passes after the refactor.
- Notes:
  - This is a readability and boundary refactor, not an API redesign.

### FR-004 - Routing naming is aligned with router semantics

- Description:
  - File and exported function/type names for inbound HTTP routing must use `router` terminology
    where they own route composition.
- Acceptance criteria:
  - [ ] `handler.go` is renamed to `router.go`.
  - [ ] Route composition constructors/types use `router` naming instead of generic `handler`
        naming.
  - [ ] All internal call sites compile against the renamed routing API.
- Notes:
  - Controller names do not need to change because they still represent endpoint handlers.

## Non-functional requirements

- Reliability (NFR-001):
  - Targeted and full Go verification commands must pass after the refactor.
- Maintainability (NFR-002):
  - A reader should be able to identify the public HTTP surface from one routing file with
    routing-specific naming.

## Dependencies and integrations

- External systems:
  - None.
- Internal services:
  - `internal/adapters/inbound/http`
  - `internal/application/ports/inbound`
