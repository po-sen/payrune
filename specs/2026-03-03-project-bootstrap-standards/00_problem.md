---
doc: 00_problem
spec_date: 2026-03-03
slug: project-bootstrap-standards
mode: Full
status: READY
owners:
  - payrune-team
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

- Background: This repository is a new project and currently has no implementation baseline.
- Users or stakeholders: Engineering team, reviewers, and CI maintainers.
- Why now: The team wants a stable implementation contract before development starts.

## Constraints (optional)

- Technical constraints:
  - Must create specs before coding.
  - Service implementation language is Go.
  - Architecture must follow Clean Architecture + Hexagonal boundaries.
  - Project tree must follow pragmatic Go project layout.
  - Any required automation helpers must live under `scripts/`.
- Timeline/cost constraints:
  - Baseline must be light enough to bootstrap quickly.
- Compliance/security constraints:
  - Pre-commit checks must include private key detection and vulnerability scan workflow.

## Problem statement

- Current pain:
  - No project-local AGENTS policy file exists.
  - No codified development workflow exists for specs, architecture, and commits.
  - No validated pre-commit baseline is configured.
- Evidence or examples:
  - Repository root is effectively empty besides `.git`.

## Goals

- G1: Create a project-local `AGENTS.md` that embeds required workflow and skill guidance.
- G2: Establish a Go service bootstrap layout that is compatible with Go layout and Clean Hexagonal rules.
- G3: Add and validate pre-commit configuration from the provided baseline.

## Non-goals (out of scope)

- NG1: Implement production business features beyond bootstrap health path.
- NG2: Introduce external runtime integrations (databases, MQ, third-party APIs).

## Assumptions

- A1: Team ownership label `payrune-team` is acceptable for spec ownership.
- A2: Module path `payrune` is acceptable for initial local bootstrap.
- A3: Running manual-stage hook `govulncheck` is optional during normal commit flow.

## Open questions

- Q1: Should module path be changed to a fully qualified VCS path after remote is created?
- Q2: Should we enforce additional CI rules (e.g., required integration/e2e jobs) now or later?

## Success metrics

- Metric: Spec package completeness.
- Target: All five spec documents exist with consistent frontmatter and lint passes.
- Metric: Baseline project quality gate.
- Target: `pre-commit run --all-files` passes for default stages.
- Metric: Architecture baseline health.
- Target: `go list ./...` and `go test ./... -short -count=1` pass without boundary violations.
