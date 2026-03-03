---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - CORS middleware behavior for allowed origin and preflight.
  - OpenAPI direct server target on 8080.
  - Compose startup and direct request validation from swagger origin.
- Not covered:
  - Credentialed CORS scenarios and multi-origin policy expansion.

## Tests

### Unit

- TC-001: Allowed-origin CORS headers

  - Linked requirements: FR-002, NFR-003, NFR-006
  - Steps:
    - Execute middleware with `Origin: http://localhost:8081` and standard GET request.
  - Expected:
    - Response contains expected CORS headers and downstream handler executes.

- TC-002: Preflight short-circuit
  - Linked requirements: FR-002
  - Steps:
    - Execute `OPTIONS` preflight request with allowed origin.
  - Expected:
    - Status is `204` and business handler is not invoked.

### Integration

- TC-101: Compose config and service availability

  - Linked requirements: FR-003, NFR-002
  - Steps:
    - Run `docker compose -f deployments/compose/compose.yaml config`.
    - Run `make up` and inspect `docker compose ... ps`.
  - Expected:
    - swagger/payrune services are both up and reachable.

- TC-102: Direct API call with CORS header check
  - Linked requirements: FR-001, FR-002, FR-003, NFR-001, NFR-005
  - Steps:
    - Run `curl -i -H "Origin: http://localhost:8081" http://localhost:8080/health`.
  - Expected:
    - 200 response includes explicit allow-origin header and health payload.

### E2E (if applicable)

- Scenario 1:
  - Open `http://localhost:8081`, execute `GET /health` from Swagger UI, and verify successful response.

## Edge cases and failure modes

- Case: Request origin is not in allow list.
- Expected behavior:

  - CORS allow-origin header is omitted.

- Case: API is up but swagger is down.
- Expected behavior:
  - Direct 8080 curl remains successful; UI path unavailable until swagger recovers.

## NFR verification

- Performance:
  - Confirm health endpoint success within 15 seconds after startup.
- Reliability:
  - Confirm both services restart policy remains unchanged.
- Security:
  - Confirm explicit origin allow list is enforced.
