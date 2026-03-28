---
doc: 01_requirements
spec_date: 2026-03-28
slug: http-controller-api-locality
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-24-architecture-conformance-refactor
  - 2026-03-27-application-inbound-error-mapping
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- API-local controller file:
- A controller source file that contains one HTTP API's method check, request parsing, usecase call, and error/status mapping.

## Out-of-scope behaviors

- OOS1: No route/path changes
- OOS2: No response payload/status contract changes

## Functional requirements

### FR-001 - Each chain-address API must have its own controller file

- Description: the chain-address HTTP controller should be organized so each API endpoint has its own source file, rather than mixing multiple APIs in one file.
- Acceptance criteria:
  - [x] List address policies handler logic is isolated in its own `*_controller.go` file.
  - [x] Generate address handler logic is isolated in its own `*_controller.go` file.
  - [x] Allocate payment address handler logic is isolated in its own `*_controller.go` file.
  - [x] Get payment address status handler logic is isolated in its own `*_controller.go` file.
- Notes: do not hide the four API controllers behind a shared aggregate controller type.

### FR-002 - API-local files must keep status mapping readable in place

- Description: each endpoint-local file must keep method checks, request parsing, and HTTP error/status mapping together so reviewers can see the contract in one place.
- Acceptance criteria:
  - [x] Endpoint-specific error-to-status mapping remains in the same file as the endpoint handler.
  - [x] Endpoint-specific request parsing helpers stay with the owning endpoint when they are not shared.
  - [x] Common helpers are limited to transport-generic concerns such as JSON writing or shared path parsing.
- Notes: do not replace locality with a central generic mapper.

### FR-003 - Refactor must preserve existing HTTP behavior

- Description: splitting files must not change route behavior, response status codes, or success/error body shapes.
- Acceptance criteria:
  - [x] Existing controller tests continue to pass without changing intended API behavior.
  - [x] Controller tests are aligned to endpoint-local files or clearly shared helpers, so source and test locality match.
  - [x] Shared test scaffolding uses an owner-aligned file name instead of a helper-style outlier.
  - [x] `Idempotency-Replayed` response header behavior remains unchanged.
  - [x] Invalid body/query/path handling remains unchanged.
- Notes: this is a readability refactor, not a contract redesign.

### FR-004 - Bootstrap and router wiring must be explicit per API controller

- Description: bootstrap and router layers must wire list/generate/allocate/get-status controllers explicitly instead of passing one aggregate chain-address controller.
- Acceptance criteria:
  - [x] `internal/bootstrap/api.go` constructs one controller per API and passes them explicitly into the router.
  - [x] `internal/bootstrap/api_worker.go` constructs one controller per API and passes them explicitly into the router.
  - [x] `internal/adapters/inbound/http/router.go` registers routes from explicit per-API controller fields.
- Notes: the purpose is readability of composition, not functional change.

### FR-005 - Controller and router pattern must be regular

- Description: each HTTP controller should follow the same basic shape so readers do not have to infer different conventions per file.
- Acceptance criteria:
  - [x] Each per-API controller exposes a single `ServeHTTP` entrypoint.
  - [x] The router mounts controllers as `http.Handler`, not via special per-controller method names.
  - [x] Tests use the same route-to-handler pattern instead of a helper that reconstructs a removed aggregate model.
- Notes: this is a consistency cleanup inside the inbound HTTP adapter.

## Non-functional requirements

- Performance (NFR-001): No extra network calls, DB calls, or new middleware layers.
- Availability/Reliability (NFR-002): All existing HTTP controller tests and full repo tests pass unchanged in intent.
- Security/Privacy (NFR-003): No new outward error detail is introduced.
- Compliance (NFR-004):
- Observability (NFR-005): Existing tests remain the primary contract verification for status/body behavior.
- Maintainability (NFR-006): A reviewer can inspect one endpoint contract by opening one endpoint-local source file and its corresponding test file, and can inspect composition by opening bootstrap/router without hidden aggregate wiring or mixed controller patterns.

## Dependencies and integrations

- External systems: None new
- Internal services: `internal/adapters/inbound/http/controllers`, existing controller tests, and existing route wiring
