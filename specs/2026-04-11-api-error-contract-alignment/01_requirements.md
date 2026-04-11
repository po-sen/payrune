---
doc: 01_requirements
spec_date: 2026-04-11
slug: api-error-contract-alignment
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-07-ethereum-contract-readiness
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Address policy not enabled:
  - A valid policy exists, but operator intent/config keeps it disabled.
- Address policy not supported:
  - The supplied `addressPolicyId` is syntactically valid but does not exist for the selected chain.

## Out-of-scope behaviors

- OOS1:
  - Changing Ethereum preview unavailability from `404` to another status.
- OOS2:
  - Adding RFC 7807 or another new response envelope format.

## Functional requirements

### FR-001 - Align address policy error status codes

- Description:
  - The address-generation and payment-allocation endpoints must distinguish malformed policy input, missing policy resources, disabled policies, and unsupported preview behavior with clearer HTTP semantics.
- Acceptance criteria:
  - [ ] AC1: `ErrInvalidAddressPolicyID` continues to map to `400 Bad Request`.
  - [ ] AC2: `ErrAddressPolicyNotFound` maps to `404 Not Found` on both `/v1/chains/{chain}/addresses` and `/v1/chains/{chain}/payment-addresses`.
  - [ ] AC3: `ErrAddressPolicyNotEnabled` maps to `409 Conflict` on both `/v1/chains/{chain}/addresses` and `/v1/chains/{chain}/payment-addresses`.
  - [ ] AC4: `ErrAddressPreviewNotSupported` continues to map to `404 Not Found`.
- Notes:
  - This requirement is limited to inbound controller mapping and its public contract.

### FR-002 - Normalize health error bodies

- Description:
  - `/health` must use the same JSON error envelope as the rest of the API for `405` and `500` responses.
- Acceptance criteria:
  - [ ] AC1: `GET /health` success behavior remains unchanged.
  - [ ] AC2: non-`GET` requests to `/health` return a JSON body with `{"error":"method not allowed"}`.
  - [ ] AC3: health use-case failures return a JSON body with `{"error":"internal server error"}`.
- Notes:
  - This keeps the endpoint contract consistent without changing the success schema.

### FR-003 - Align OpenAPI with runtime behavior

- Description:
  - `deployments/swagger/openapi.yaml` must document the same status codes, descriptions, and JSON error bodies implemented by the controllers.
- Acceptance criteria:
  - [ ] AC1: OpenAPI uses the updated `404` / `409` mappings for address-policy errors.
  - [ ] AC2: `/health` documents JSON error responses for `405` and `500`.
  - [ ] AC3: Error-response schemas are expressed consistently across the spec.
- Notes:
  - The OpenAPI cleanup should stay readable and avoid excessive indirection.

## Non-functional requirements

- Performance (NFR-001):
  - The change must not add new runtime IO or request-path latency.
- Availability/Reliability (NFR-002):
  - Existing successful endpoint behavior must remain unchanged.
- Security/Privacy (NFR-003):
  - The API must preserve the intentional `404` behavior for unsupported Ethereum address preview requests.
- Compliance (NFR-004):
- Observability (NFR-005):
  - Existing controller error logging must continue to log mapped failures after the status changes.
- Maintainability (NFR-006):
  - Controller tests and OpenAPI examples must make the status mapping easy to review without requiring implicit knowledge of old behavior.

## Dependencies and integrations

- External systems:
  - Swagger UI / OpenAPI consumers.
- Internal services:
  - Inbound HTTP controllers.
  - Application inbound error contract in `internal/application/ports/inbound`.
