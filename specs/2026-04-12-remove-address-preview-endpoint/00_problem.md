---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - The public API still exposes `GET /v1/chains/{chain}/addresses?addressPolicyId=...&index=...`.
  - In practice this route only serves deterministic Bitcoin preview. Ethereum CREATE2 already rejects preview on purpose.
- Users or stakeholders:
  - API operators and maintainers.
- Why now:
  - The route is not part of payment address issuance, creates asymmetric API behavior across chains, and keeps unnecessary preview-only code in controller, use case, DTO, and domain policy layers.

## Constraints (optional)

- Technical constraints:
  - Removal must cover router, bootstrap wiring, application contracts, controller helpers, OpenAPI, and preview-only domain methods.
- Timeline/cost constraints:
  - Keep this as a Quick spec and a single focused code cleanup.

## Problem statement

- Current pain:
  - The repo carries a public preview endpoint that is not needed for payment receiving.
  - The preview route forces extra abstractions:
    - `GenerateAddressUseCase`
    - preview-only DTOs
    - preview-only controller and JSON mapping
    - `SupportsAddressPreview` / `ValidateForAddressPreview`
    - preview-only error contracts
- Evidence or examples:
  - `deployments/swagger/openapi.yaml` still documents `/v1/chains/{chain}/addresses`.
  - `internal/bootstrap/api.go` and `internal/bootstrap/api_worker.go` still wire `GenerateAddressUseCase`.

## Goals

- G1:
  - Remove the public address preview endpoint and all runtime code that exists only for that endpoint.
- G2:
  - Keep the remaining address-policy listing and payment-address issuance/status APIs unchanged.

## Non-goals (out of scope)

- NG1:
  - Change payment address allocation behavior.
- NG2:
  - Introduce a replacement preview endpoint.

## Assumptions

- A1:
  - Public deterministic preview is no longer required by current product behavior.
- A2:
  - Historical specs may mention the removed route, but the new spec becomes the source of truth for this removal.

## Open questions

- None.

## Success metrics

- Metric:
  - Runtime surface
- Target:
  - No public route, Swagger path, controller, use case, DTO, or preview-only domain method remains for address preview.
