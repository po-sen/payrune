---
doc: 01_requirements
spec_date: 2026-04-12
slug: remove-address-preview-endpoint
mode: Quick
status: DONE
owners:
  - codex
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Out-of-scope behaviors

- OOS1:
  - No replacement API for deterministic preview.
- OOS2:
  - No changes to address derivation internals used by issuance.

## Functional requirements

### FR-001 - Remove public address preview route

- Description:
  - The service must stop exposing `GET /v1/chains/{chain}/addresses`.
- Acceptance criteria:
  - [ ] `internal/adapters/inbound/http/router.go` no longer registers `/v1/chains/{chain}/addresses`.
  - [ ] `internal/bootstrap/api.go` and `internal/bootstrap/api_worker.go` no longer construct or inject preview-specific handlers/use cases.
  - [ ] `deployments/swagger/openapi.yaml` no longer documents `/v1/chains/{chain}/addresses`.
- Notes:
  - A request to that path should now fall through as an unregistered route.

### FR-002 - Remove preview-only application and domain code

- Description:
  - The repo must remove preview-only runtime artifacts that become unused after the route is deleted.
- Acceptance criteria:
  - [ ] `GenerateAddressUseCase`, `GenerateAddressInput`, and `GenerateAddressResponse` are removed.
  - [ ] Preview-only inbound error `ErrAddressPreviewNotSupported` is removed.
  - [ ] Preview-only policy methods and errors are removed when no longer referenced.
- Notes:
  - Shared derivation code that is still needed by issuance must remain.

### FR-003 - Keep payment receiving APIs intact

- Description:
  - Removing preview must not change the remaining public payment receiving APIs.
- Acceptance criteria:
  - [ ] `GET /v1/chains/{chain}/address-policies` remains available.
  - [ ] `POST /v1/chains/{chain}/payment-addresses` remains available.
  - [ ] `GET /v1/chains/{chain}/payment-addresses/{paymentAddressId}` remains available.

## Non-functional requirements

- Maintainability (NFR-001):
  - Preview-only runtime code paths must be removed rather than left unused.
- Reliability (NFR-002):
  - `bash scripts/precommit-run.sh` must pass after the removal.
- Maintainability (NFR-003):
  - `SPEC_DIR="specs/2026-04-12-remove-address-preview-endpoint" bash scripts/spec-lint.sh` must pass.

## Dependencies and integrations

- External systems:
  - OpenAPI/Swagger consumers.
- Internal services:
  - Public API bootstrap and Cloudflare API worker bootstrap.
