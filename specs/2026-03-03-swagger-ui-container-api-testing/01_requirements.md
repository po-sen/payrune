---
doc: 01_requirements
spec_date: 2026-03-03
slug: swagger-ui-container-api-testing
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-deploy-service-compose-dockerfile
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Glossary (optional)

- Direct Swagger call: browser request from `http://localhost:8081` to `http://localhost:8080`.
- CORS preflight: `OPTIONS` request used by browser to validate cross-origin permissions.

## Out-of-scope behaviors

- OOS1: Dynamic policy engine for CORS.
- OOS2: Credential-based CORS sessions.

## Functional requirements

### FR-001 - OpenAPI direct server target

- Description:
  - OpenAPI spec MUST direct Swagger try-it-out calls to API host `http://localhost:8080`.
- Acceptance criteria:
  - [ ] `deployments/swagger/openapi.yaml` has `servers` URL set to `http://localhost:8080`.
  - [ ] `/health` operation remains documented and callable.
- Notes:
  - Generated curl command from Swagger should show 8080 host.

### FR-002 - CORS support in payrune HTTP server

- Description:
  - API server MUST return CORS headers allowing browser requests from Swagger origin.
- Acceptance criteria:
  - [ ] Request with `Origin: http://localhost:8081` includes `Access-Control-Allow-Origin: http://localhost:8081`.
  - [ ] Response includes `Access-Control-Allow-Methods` and `Access-Control-Allow-Headers`.
  - [ ] `OPTIONS` preflight requests return success without calling business handlers.
- Notes:
  - CORS behavior must be implemented at inbound HTTP layer.

### FR-003 - Compose workflow continuity

- Description:
  - Compose workflow MUST keep both `payrune` and `swagger` services operational with existing make targets.
- Acceptance criteria:
  - [ ] `make up` starts both services.
  - [ ] `http://localhost:8081` remains available for Swagger UI.
  - [ ] `make down` stops and removes both services.
- Notes:
  - Remove obsolete proxy-only configuration if not required for direct mode.

## Non-functional requirements

- Performance (NFR-001): `GET /health` direct call from Swagger should succeed within 15 seconds after startup.
- Availability/Reliability (NFR-002): Existing compose restart policies remain `unless-stopped`.
- Security/Privacy (NFR-003): CORS allow-origin is explicit `http://localhost:8081` (no wildcard).
- Compliance (NFR-004): `SPEC_DIR="specs/2026-03-03-swagger-ui-container-api-testing" bash scripts/spec-lint.sh` passes.
- Observability (NFR-005): CORS behavior is diagnosable via response headers and compose logs.
- Maintainability (NFR-006): CORS logic is isolated in reusable HTTP middleware.

## Dependencies and integrations

- External systems:
  - Browser CORS enforcement
  - Docker image `swaggerapi/swagger-ui`
- Internal services:
  - payrune HTTP service on port 8080
