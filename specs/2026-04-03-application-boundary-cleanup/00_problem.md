---
doc: 00_problem
spec_date: 2026-04-03
slug: application-boundary-cleanup
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-domain-model-boundary-cleanup
  - 2026-04-02-sweep-material-redesign
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background:
  - The repository has already cleaned most domain-layer boundary issues.
  - A fresh review of `internal/application` found two remaining ownership leaks.
- Users or stakeholders:
  - Maintainers extending use cases, outbound ports, and inbound HTTP adapters.
- Why now:
  - The application layer is already close to clean; fixing the last leaks now is lower-risk than
    letting them spread into new features.

## Constraints (optional)

- Technical constraints:
  - Do not introduce new architecture layers, registries, or mini-model frameworks.
  - Keep runtime changes small, explicit, and local.
  - Preserve existing DB schema and external HTTP behavior.
- Timeline/cost constraints:
  - Refactor only; no scope expansion into unrelated cleanup.
- Compliance/security constraints:
  - None beyond preserving current behavior.

## Problem statement

- Current pain:
  - `internal/application/ports/outbound` still exposes `SweepMaterialJSON`, so application code
    knows a persistence/operator JSON representation.
  - `internal/application/dto` still carries HTTP JSON response shape concerns via `json` tags and
    response-only fields.
- Evidence or examples:
  - `DeriveIssuedPaymentAddressOutput.SweepMaterialJSON`
  - `CompletePaymentAddressAllocationInput.SweepMaterialJSON`
  - `dto.HealthResponse`, `dto.ErrorResponse`, and address response DTOs with `json` tags,
    `omitempty`, and `json:"-"`

## Goals

- G1:
  - Remove persistence representation naming from application ports and use cases without
    reintroducing overengineered abstractions.
- G2:
  - Move HTTP JSON response shaping out of `internal/application` and into inbound HTTP adapters.

## Non-goals (out of scope)

- NG1:
  - Redesign DB schema, rename `sweep_material_json`, or change payload content.
- NG2:
  - Introduce new domain/application model hierarchies or extra collaborator layers for sweep
    material.

## Assumptions

- A1:
  - It is acceptable for application to carry an opaque sweep-material document as long as the
    contract no longer names or models it as JSON.
- A2:
  - Existing HTTP response payload field names and omission behavior must remain unchanged.

## Open questions

- None.

## Success metrics

- Metric:
  - `internal/application` no longer contains `SweepMaterialJSON` in runtime code.
- Target:
  - Zero runtime matches.
- Metric:
  - `internal/application/dto` no longer declares HTTP `json` tags.
- Target:
  - Zero matches for `json:"` under `internal/application/dto`.
