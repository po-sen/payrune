---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - The HTTP API currently mixes transport semantics across endpoints. In particular, `address policy is not enabled` is mapped to `501 Not Implemented`, and `/health` uses plain-text error bodies while the rest of the API mostly returns JSON.
- Users or stakeholders:
  - Local operators using the API directly or through Swagger UI.
  - Developers integrating against the local API contract.
- Why now:
  - Recent `*_ENABLED` policy-intent work makes disabled policies a normal operator state rather than an implementation gap, so the current `5xx`-style semantics are misleading.

## Constraints (optional)

- Technical constraints:
  - Keep the change within the existing controller / inbound-port error model.
  - Do not introduce a new error framework or large OpenAPI abstraction layer.
- Timeline/cost constraints:
  - Small repo-local cleanup; keep this in Quick mode.
- Compliance/security constraints:
  - Preserve the current intentional `404` behavior for Ethereum preview unavailability on `/v1/chains/{chain}/addresses`.

## Problem statement

- Current pain:
  - Disabled address policies currently surface as `501`, which misclassifies an intentional operator configuration state as a server-side implementation problem.
  - `address policy is not supported` currently surfaces as `400` even though it is closer to a missing subresource than malformed input.
  - `/health` emits plain-text `405` / `500`, while the documented API shape otherwise expects JSON error bodies.
  - `deployments/swagger/openapi.yaml` still documents the old status mapping and repeats error schemas inline.
- Evidence or examples:
  - `POST /v1/chains/ethereum/payment-addresses` with a disabled policy returns `501 address policy is not enabled`.
  - `internal/adapters/inbound/http/controllers/health_controller.go` uses `http.Error(...)` instead of `writeErrorJSON(...)`.

## Goals

- G1:
  - Align controller status codes with clearer API semantics for disabled and unsupported address policies.
- G2:
  - Make error body format consistent across the public API, including `/health`.
- G3:
  - Update OpenAPI so documented responses, examples, and error schemas match the implemented contract.

## Non-goals (out of scope)

- NG1:
  - Changing domain/use-case error taxonomy beyond what is needed for controller mapping.
- NG2:
  - Redesigning success response payloads or adding versioned API paths.

## Assumptions

- A1:
  - `address policy not enabled` should be treated as a client-visible state conflict, not as a server fault.
- A2:
  - `address policy not supported` should be treated as a missing referenced resource when the identifier is syntactically valid.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Disabled policy requests no longer return any `5xx`-class status.
- Target:

  - `ErrAddressPolicyNotEnabled` maps to a non-`5xx` status in all affected endpoints.

- Metric:
  - Swagger examples and status codes match controller behavior for address-policy and health errors.
- Target:
  - OpenAPI validation and controller tests pass with the new mappings.
