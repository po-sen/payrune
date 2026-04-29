---
doc: 00_problem
spec_date: 2026-04-29
slug: keepline-architecture-policy
mode: Full
status: DONE
owners:
  - repo-maintainers
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: The maintainer requested keepline installation, a repository `keepline.toml`, and
  validation against strict commit-message, file, and import policies. During review, the requested
  architecture standard evolved into a stricter Clean Architecture and Hexagonal boundary:
  adapters may depend on application ports, but not on domain models, use cases, generic DTO
  packages, outbox internals, bootstrap wiring, cmd entrypoints, or other concrete adapter
  families.
- Users or stakeholders: repository maintainers, reviewers, and future contributors.
- Why now: The repository is adopting keepline as the executable architecture contract, so the final
  policy and refactor should be documented as one source of truth instead of several incremental
  specs.

## Constraints (optional)

- Technical constraints: Keep the Go layout under `cmd/` and `internal/`; keep helper automation
  under `scripts/`; keep ports under `internal/application/ports/**`.
- Timeline/cost constraints: Preserve runtime behavior and avoid broad domain redesign beyond the
  package boundary work needed for strict import enforcement.
- Compliance/security constraints: Preserve the requested secret/file-deny policy and Conventional
  Commit rules.

## Problem statement

- Current pain: Architecture guidance in `AGENTS.md` was previously enforced by convention, and
  adapters could import core implementation details or concrete sibling adapters unless keepline
  encoded the boundary.
- Evidence or examples: The original adapter boundary allowed imports of application DTOs, outbox
  types, and domain types. The final code now exposes only explicit application port contracts to
  adapters.

## Goals

- G1: Install and configure keepline at the repository root.
- G2: Configure pre-commit so keepline validates commit messages, file policy, and import policy.
- G3: Refactor adapter-facing contracts so they live under `internal/application/ports/**`.
- G4: Ensure adapters import only application port contracts and infrastructure helpers.
- G5: Encode strict domain, application, infrastructure, adapter, and adapter-family import rules
  in `keepline.toml`.
- G6: Preserve runtime behavior and validate with keepline, Go tests, spec lint, and pre-commit.

## Non-goals (out of scope)

- NG1: Change HTTP API routes, response schemas, scheduler behavior, database schemas, migrations,
  or deployment contracts.
- NG2: Add a new CI workflow.
- NG3: Introduce alternate top-level architecture such as `pkg/`, `shared_kernel/`, `components/`,
  or feature-oriented domain package trees.
- NG4: Reintroduce a generic adapter-visible `internal/application/dto/**` contract package.

## Assumptions

- A1: `internal/application/ports/inbound` and `internal/application/ports/outbound` are stable
  application boundary contracts.
- A2: Application port packages may define interfaces, commands, results, records, and shared
  contract errors, but must stay free of domain, use-case, adapter, infrastructure, bootstrap, and
  cmd imports.
- A3: A generic `internal/application/dto/**` package is not adapter-visible because it does not
  encode whether a shape belongs to an inbound port, outbound port, or use-case-private workflow.
- A4: Bootstrap and `cmd` remain composition roots and may wire concrete adapters, use cases,
  domain policies, and infrastructure.

## Open questions

- None.

## Success metrics

- Metric: `keepline config-check`
- Target: passes.
- Metric: `keepline import-check` and `keepline check --scope all`
- Target: pass with strict import policy enabled.
- Metric: `go test ./...`
- Target: all packages pass.
- Metric: `bash scripts/precommit-run.sh`
- Target: default-stage hooks pass and include keepline import checking.
