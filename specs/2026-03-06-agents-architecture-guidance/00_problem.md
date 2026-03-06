---
doc: 00_problem
spec_date: 2026-03-06
slug: agents-architecture-guidance
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
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
  - `AGENTS.md` currently mixes workflow notes, copied skill snapshots, and a few high-level architecture statements.
- Users or stakeholders:
  - Future coding agents working in this repository and the project owner reviewing their output.
- Why now:
  - Recent implementation work showed repeated ambiguity around repository vs store naming, domain vs application logic, and how strictly to model outbox-style data.

## Problem statement

- Current pain:
  - The repository instructions do not clearly encode the owner's architecture bar, so the agent can satisfy dependency direction while still missing the intended modeling quality.
- Evidence or examples:
  - The agent needed repeated correction on repository/store naming, entity vs record modeling, and what belongs in domain vs application.
  - The existing file also contains embedded skill snapshots and other important instructions that must not be discarded.

## Goals

- G1:
  - Add a clear repo-specific operating section for the coding agent without removing existing important instructions.
- G2:
  - Make architecture boundaries explicit, especially domain/application/adapters/infrastructure responsibilities.
- G3:
  - Encode naming and modeling rules for entity, value object, repository, store, outbox, and unit of work.
- G4:
  - Add clear anti-patterns and decision rules so future changes require fewer corrective review cycles.
- G5:
  - Restate the repo-specific architecture guidance in stronger best-practice language without changing the meaning of correct existing instructions.

## Non-goals (out of scope)

- NG1:
  - Refactoring production code or renaming packages to match the new guidance.
- NG2:
  - Replacing the repo's external developer/system instructions.
- NG3:
  - Writing a generic Clean Architecture tutorial unrelated to this repository.

## Assumptions

- A1:
  - The primary reader is the coding agent, not a general engineering audience.
- A2:
  - The existing embedded skill content remains intentionally valuable and should be preserved.

## Success metrics

- Metric:
  - Future agent behavior should be constrained by explicit project-specific architecture rules.
- Target:
  - `AGENTS.md` contains concrete rules for layer responsibility, naming, modeling, and validation expectations.
- Metric:
  - The guidance should stay usable during implementation.
- Target:
  - The document becomes clearer for the agent while preserving the pre-existing important content.
