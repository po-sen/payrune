---
doc: 01_requirements
spec_date: 2026-03-28
slug: outbound-review-clarity
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-outbound-port-error-conformance
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Adapter-private collaborator:
- An interface or helper used only inside an adapter package or its bootstrap wiring, not an application `outport`.

## Out-of-scope behaviors

- OOS1: No new `outport.Err...` work in this spec.
- OOS2: No directory reshuffle.

## Functional requirements

### FR-001 - Adapter-private interfaces should look adapter-private

- Description: Interfaces in `internal/adapters/outbound` that are not application `outport` contracts should use unexported naming when they are only package-private collaborators.
- Acceptance criteria:
  - [ ] `CloudflareEsploraBridge`, `AddressEncoder`, and persistence `Executor` / `Rows` / `Row` / `Result` style collaborators are renamed to unexported names where they are not part of an application port.
  - [ ] Corresponding constructors, tests, and bootstrap call sites still compile and behave the same.
- Notes: This is a readability/locality cleanup, not a behavior change.

### FR-002 - Review heuristics must be explicit in AGENTS

- Description: `AGENTS.md` must explain how to review error ownership and how to distinguish `outport` boundaries from adapter-private collaborators.
- Acceptance criteria:
  - [ ] `AGENTS.md` states that only methods implementing `internal/application/ports/outbound` must restrict outward errors to `outport.Err...`.
  - [ ] `AGENTS.md` states that constructor/configuration paths and adapter-private collaborator interfaces may keep package-local errors.
- Notes: The goal is to reduce reviewer ambiguity.

## Non-functional requirements

- Performance (NFR-001): No additional IO or runtime branching.
- Availability/Reliability (NFR-002): No runtime behavior changes beyond naming / review guidance.
- Security/Privacy (NFR-003): No new outward error detail exposure.
- Compliance (NFR-004): N/A.
- Observability (NFR-005): Existing test coverage for outbound adapters remains green after renames.
- Maintainability (NFR-006): Reviewers should be able to identify the correct error-ownership rule by naming and file ownership without reading deep call chains.

## Dependencies and integrations

- External systems: None new.
- Internal services: `AGENTS.md`, bootstrap wiring, outbound adapter tests.
