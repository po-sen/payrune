---
doc: 01_requirements
spec_date: 2026-04-12
slug: compose-entrypoint-wording
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-09-compose-env-example
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Term:
  - Base stack
- Definition:
  - The unprofiled Compose services that start without enabling the `development` profile.
- Term:
  - Development profile
- Definition:
  - The additional Compose services enabled by `--profile development`.

## Out-of-scope behaviors

- OOS1:
  - Reprofiling mainnet pollers or moving services between base and development topology.
- OOS2:
  - Changing target names or env-file conventions.

## Functional requirements

### FR-001 - Describe `make up` as base stack plus development services

- Description:
  - `Makefile` help and `README.md` must describe `make up` / `make down` / `make config` as using `compose.dev.env` with the `development` profile added on top of the base stack.
- Acceptance criteria:
  - [ ] AC1: `Makefile` help no longer implies that `make up` excludes unprofiled base services.
  - [ ] AC2: `README.md` explains that `make up` adds development-profile services to the base stack.
- Notes:
  - This is a wording-only change; runtime behavior must stay the same.

### FR-002 - Describe `make up-mainnet` as base stack only

- Description:
  - `Makefile` help and `README.md` must describe `make up-mainnet` / `make down-mainnet` / `make config-mainnet` as using the base stack with no `development` profile.
- Acceptance criteria:
  - [ ] AC1: `Makefile` help no longer calls the path “formal/mainnet-style” without clarifying that it means base stack only.
  - [ ] AC2: `README.md` states that `make up-mainnet` starts the base stack without the extra development-profile services.
- Notes:
  - The target name remains unchanged in this cleanup.

## Non-functional requirements

- Performance (NFR-001):
  - No runtime or Compose behavior changes are introduced.
- Availability/Reliability (NFR-002):
  - Existing `make up`, `make down`, `make up-mainnet`, and `make down-mainnet` behavior must remain unchanged.
- Security/Privacy (NFR-003):
  - No changes to secrets or env-file handling.
- Compliance (NFR-004):
- Observability (NFR-005):
- Maintainability (NFR-006):
  - Compose entrypoint wording should match the actual topology closely enough that operators do not need to inspect `docker compose config --services` to understand it.

## Dependencies and integrations

- External systems:
  - Docker Compose
- Internal services:
  - `Makefile`
  - `README.md`
