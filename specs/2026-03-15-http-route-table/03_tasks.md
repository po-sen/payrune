---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a localized inbound HTTP routing refactor with no new integration or contract.
- Upstream dependencies (`depends_on`):
  - None.
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The route-table change stays inside existing inbound HTTP adapters.
  - What would trigger switching to Full mode:
    - Any change to public API contract or the addition of a new transport framework.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Centralize public HTTP route registrations in the routing file.
- M2:
  - Remove nested controller routing helpers and update tests.
- M3:
  - Rename routing composition files and exports to use `router` naming.

## Tasks (ordered)

1. T-001 - Centralize public route registrations
   - Scope:
     - Update `router.go` to register concrete public API paths directly.
   - Output:
     - A visible route table for public HTTP endpoints.
   - Linked requirements: FR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): run HTTP handler and controller tests.
     - [ ] Expected result: the registered routes are visible in one place and route-based tests
           still pass.
     - [ ] Logs/metrics to check (if applicable): none
2. T-002 - Remove nested controller routing
   - Scope:
     - Replace shared-prefix chain routing with endpoint-specific controller handlers and remove
       obsolete parsing helpers.
   - Output:
     - Slimmer controllers without sub-router logic.
   - Linked requirements: FR-002, FR-003, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted HTTP tests, `go list ./...`, and
           full `go test ./...`.
     - [ ] Expected result: public HTTP behavior stays green after the route-table refactor.
     - [ ] Logs/metrics to check (if applicable): none
3. T-003 - Align routing file and export naming
   - Scope:
     - Rename the routing composition file and exported constructors/types from `handler` wording
       to `router` wording.
   - Output:
     - Routing-specific file and API names that match the actual responsibility.
   - Linked requirements: FR-004, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): run targeted HTTP tests, `go list ./...`, and
           full `go test ./...`.
     - [ ] Expected result: all call sites compile against the router-named API and behavior stays
           unchanged.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-002
- FR-003 -> T-002, T-003
- FR-004 -> T-003
- NFR-001 -> T-002, T-003
- NFR-002 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Centralize route registrations first, then remove nested controller routing, then align router
    naming.
- Rollback steps:
  - Restore the prior handler-named routing file and controller-managed route registration if the
    refactor causes regressions.

## Validation evidence

- `SPEC_DIR="specs/2026-03-15-http-route-table" bash scripts/spec-lint.sh`
- `gofmt -w internal/adapters/inbound/http/router.go internal/adapters/inbound/http/router_test.go internal/adapters/inbound/http/controllers/health_controller.go internal/adapters/inbound/http/controllers/health_controller_test.go internal/adapters/inbound/http/controllers/chain_address_controller.go internal/adapters/inbound/http/controllers/chain_address_controller_test.go internal/infrastructure/di/container.go internal/infrastructure/di/cloudflare_api_worker.go`
- `go test ./internal/adapters/inbound/http/...`
- `go list ./...`
- `go test ./...`
- `rg -n "RegisterRoutes\\(|parseChainRoute\\(|NewHandler\\(|NewPublicHandler\\(|type Dependencies struct" internal cmd -g '*.go'`
