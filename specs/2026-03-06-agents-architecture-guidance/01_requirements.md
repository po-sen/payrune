---
doc: 01_requirements
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

# Requirements

## Functional requirements

### FR-001 - State the agent-facing workflow rules first

- Description:
  - `AGENTS.md` must clearly expose the required working process the agent must follow in this repository without removing existing important instructions.
- Acceptance criteria:
  - [x] The document explicitly requires spec-first changes, pre-commit validation, and conventional commits when requested.
  - [x] The document states that it is written for the coding agent.

### FR-002 - Make layer responsibilities explicit

- Description:
  - The document must describe what belongs in domain, application, adapters, infrastructure, bootstrap, and `cmd`.
- Acceptance criteria:
  - [x] Domain rules vs application orchestration are distinguished clearly.
  - [x] The document defines what outbound ports may and may not contain.
  - [x] The document explains where IO is allowed and where it is forbidden.

### FR-003 - Define project-specific naming and modeling rules

- Description:
  - The document must clarify how this repo should use terms like entity, value object, repository, store, outbox, and unit of work.
- Acceptance criteria:
  - [x] The document gives naming rules or decision criteria for repository vs store vs outbox.
  - [x] The document clarifies when a row-like structure should not be modeled as a domain entity.
  - [x] The document sets expectations for unit-of-work and transaction boundaries.

### FR-004 - Call out anti-patterns and review triggers

- Description:
  - The document must tell the agent what to avoid and what requires extra caution or explicit justification.
- Acceptance criteria:
  - [x] The document lists architecture anti-patterns relevant to this repo.
  - [x] The document includes explicit review triggers for suspicious designs, such as business rules drifting into repositories or adapters.

### FR-005 - Preserve important existing guidance

- Description:
  - The rewrite must preserve the important workflow and embedded-skill content already present in `AGENTS.md`, then layer repo-specific overrides on top.
- Acceptance criteria:
  - [x] Existing mandatory workflow and embedded skill sections remain in the file.
  - [x] New repo-specific guidance is added as an override section rather than replacing the existing content wholesale.

### FR-006 - Preserve the meaning of correct existing instructions

- Description:
  - The rewritten repo-specific guidance should improve clarity and best-practice framing without changing the meaning of correct existing instructions.
- Acceptance criteria:
  - [x] The update strengthens the structure and wording of the repo-specific architecture section.
  - [x] The update does not contradict or silently narrow correct existing rules unless the user has explicitly asked for that change.

## Non-functional requirements

- Maintainability (NFR-001):
  - The rewritten document must be concise enough to scan quickly during implementation and still remain more explicit than the current version.
- Clarity (NFR-002):
  - The guidance must be phrased as actionable directives instead of abstract principles.
- Consistency (NFR-003):
  - The new instructions must align with the repo's current folder structure under `cmd/`, `internal/`, `deployments/`, `scripts/`, and `specs/`.

## Dependencies and integrations

- Internal dependencies:
  - Existing project layout and current architectural boundaries already present in the repo.
