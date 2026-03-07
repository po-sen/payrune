---
doc: 04_test_plan
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

# Test Plan

## Functional

- TC-201:

  - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, NFR-001, NFR-002, NFR-003
  - Steps:
    - Read the rewritten `AGENTS.md` from top to bottom as if starting a new feature.
  - Expected:
    - The document preserves the existing important sections and also makes workflow, layering, naming, and review expectations explicit for the agent without changing the meaning of correct existing instructions.

- TC-202:
  - Linked requirements: FR-007
  - Steps:
    - Verify that the repository contains the five spec template files under `assets/`.
  - Expected:
    - The local assets exist and match the repo's documented scaffolding flow.

## Validation commands

- TC-901:

  - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008
  - Steps:
    - Run `SPEC_DIR="specs/2026-03-06-agents-architecture-guidance" bash scripts/spec-lint.sh`.
  - Expected:
    - Spec docs pass lint.

- TC-902:
  - Linked requirements: NFR-001, NFR-002, NFR-003
  - Steps:
    - Run `bash scripts/precommit-run.sh`.
  - Expected:
    - Repo validations pass after the `AGENTS.md` rewrite.
