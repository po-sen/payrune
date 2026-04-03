---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Opaque sweep material:
  - A chain-specific recovery/sweep document that application passes between outbound collaborators
    without interpreting its serialized representation.
- Transport DTO:
  - An HTTP-specific request/response type owned by the inbound adapter layer.

## Out-of-scope behaviors

- OOS1:
  - Changing DB schema or persisted `sweep_material_json` payload content.
- OOS2:
  - Redesigning notification outbox workflow or other unrelated application-layer packages.

## Functional requirements

### FR-001 - Remove persistence representation naming from application ports

- Description:
  - Application outbound contracts and use cases must stop referring to sweep material as JSON or
    `_json` representation.
- Acceptance criteria:
  - [ ] Runtime code under `internal/application` contains no `SweepMaterialJSON` field or
        parameter names.
  - [ ] `allocate_payment_address_use_case` passes an opaque sweep-material value between the
        issued-address deriver and allocation store without interpreting serialization format.
  - [ ] Existing adapter behavior and persisted `sweep_material_json` column content remain
        unchanged.
- Notes:
  - Keeping the value as an opaque string-like document is acceptable if the application contract no
    longer names or models it as JSON.

### FR-002 - Keep HTTP response shaping inside inbound HTTP adapters

- Description:
  - Application DTOs returned by use cases must be transport-agnostic. HTTP JSON field naming,
    omission, and response-only formatting belong in inbound HTTP adapters.
- Acceptance criteria:
  - [ ] Runtime code under `internal/application/dto` contains no `json` struct tags.
  - [ ] `internal/application/dto` no longer contains HTTP-only error response types.
  - [ ] Existing HTTP response payload field names and omission behavior remain unchanged in current
        controller tests.
- Notes:
  - Controllers may map application DTOs into HTTP response structs with `json` tags.

### FR-003 - Preserve current application orchestration behavior

- Description:
  - The refactor must not move business policy into application use cases or change current use-case
    behavior beyond transport-neutral typing.
- Acceptance criteria:
  - [ ] `internal/application/usecases` continues to orchestrate ports and domain behavior without
        adding new persistence or transport branching.
  - [ ] Existing use-case tests remain green after the refactor.
- Notes:
  - This requirement keeps the cleanup boring and prevents new layers from being invented during the
    refactor.

## Non-functional requirements

- Performance (NFR-001):
  - No additional DB, network, or JSON round-trip may be introduced by this refactor.
- Availability/Reliability (NFR-002):
  - Address allocation, health check, and status retrieval behavior must remain unchanged.
- Security/Privacy (NFR-003):
  - No change.
- Compliance (NFR-004):
  - No change.
- Observability (NFR-005):
  - No change.
- Maintainability (NFR-006):
  - Boundary ownership must be clearer after the refactor: application DTOs are transport-agnostic,
    and outbound contracts do not expose persistence representation details.

## Dependencies and integrations

- External systems:
  - None added.
- Internal services:
  - Inbound HTTP controllers
  - Address allocation use case
  - Issued payment address deriver outbound port
  - Payment address allocation store outbound port
