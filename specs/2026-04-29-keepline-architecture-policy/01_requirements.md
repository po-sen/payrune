---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- keepline: The policy runner used for commit-message, file-policy, and import-boundary checks.
- Application port: An inbound or outbound application boundary contract under
  `internal/application/ports/**`.
- Adapter family: A cohesive concrete adapter package group under `internal/adapters/inbound` or
  `internal/adapters/outbound`, such as HTTP, scheduler, bitcoin, persistence, or webhook.
- Generic DTO package: A broad data-transfer package such as `internal/application/dto/**` that is
  not tied to a specific application port boundary.

## Out-of-scope behaviors

- OOS1: Live external blockchain, PostgreSQL, or Cloudflare deployment validation.
- OOS2: Runtime API, database, scheduler, migration, or OpenAPI behavior changes.
- OOS3: Allowing adapters to import domain models, use cases, generic DTOs, outbox internals, or
  concrete sibling adapter families.

## Functional requirements

### FR-001 - Provide keepline configuration

- Description: Add a repository-root `keepline.toml` containing the requested commit-message,
  denied-file, and import-boundary policies.
- Acceptance criteria:
  - [x] `keepline.toml` exists at the repository root.
  - [x] Commit-message policy has max header length 72, requested allowed types, vague subject
        checks, special commit allowances, and uppercase ticket-prefix allowance.
  - [x] File policy keeps `allowed = []` and includes the requested denied patterns.
  - [x] Import policy is enabled.
- Notes: The final import policy is stricter than the original sample because the codebase was
  refactored to satisfy it.

### FR-002 - Pre-commit runs keepline policy

- Description: The pre-commit configuration must run keepline checks before commit.
- Acceptance criteria:
  - [x] The keepline provider is pinned to `v0.17.0`.
  - [x] `keepline-commit-msg` is configured for commit-message validation.
  - [x] `keepline-file-check` is configured for file policy validation.
  - [x] A local `keepline import-check` hook runs as part of default pre-commit validation.
- Notes: `keepline import-check` is added locally because the provider hook set covers commit
  message and file policy hooks.

### FR-003 - Domain remains pure

- Description: Domain packages must not depend on application, adapters, infrastructure, bootstrap,
  or cmd.
- Acceptance criteria:
  - [x] `internal/domain/**` denies `internal/application/**`, `internal/adapters/**`,
        `internal/infrastructure/**`, `internal/bootstrap/**`, and `cmd/**`.
  - [x] `internal/domain/entities/**` denies `internal/domain/policies/**`.
  - [x] `internal/domain/valueobjects/**` denies entities, events, and policies.
- Notes: Policies may depend on value objects and entities when necessary; entities and value
  objects do not depend back on policies.

### FR-004 - Application owns use cases and ports

- Description: Application code owns orchestration and port contracts without depending on
  implementation layers.
- Acceptance criteria:
  - [x] `internal/application/**` denies adapters, infrastructure, bootstrap, and cmd.
  - [x] `internal/application/ports/**` denies domain, use cases, adapters, infrastructure,
        bootstrap, and cmd.
  - [x] No production code imports the old `payrune/internal/ports/**` path.
- Notes: Port contract records are intentionally under application, not under a separate top-level
  `internal/ports` package.

### FR-005 - Adapters see ports, not core internals

- Description: Inbound and outbound adapters may import application port contracts and
  infrastructure helpers, but not core implementation details.
- Acceptance criteria:
  - [x] `internal/adapters/**` denies `internal/domain/**`.
  - [x] `internal/adapters/**` denies `internal/application/usecases/**`,
        `internal/application/dto/**`, and `internal/application/outbox/**`.
  - [x] `internal/adapters/**` denies `internal/bootstrap/**` and `cmd/**`.
  - [x] Adapters may import `internal/application/ports/inbound/**` or
        `internal/application/ports/outbound/**` where appropriate.
- Notes: Generic DTOs are not adapter-visible contracts; records used by adapters live beside the
  port interface they belong to.

### FR-006 - Concrete adapter families stay isolated

- Description: Concrete adapter families must not import each other directly.
- Acceptance criteria:
  - [x] `internal/adapters/inbound/**` denies `internal/adapters/outbound/**`.
  - [x] `internal/adapters/outbound/**` denies `internal/adapters/inbound/**`.
  - [x] HTTP inbound adapters deny scheduler inbound adapters, and scheduler denies HTTP.
  - [x] Outbound adapter families deny the other outbound families.
  - [x] Postgres and Cloudflare Postgres persistence implementations deny each other.
- Notes: Shared behavior belongs in application ports, infrastructure helpers, package-local helpers,
  or bootstrap composition.

### FR-007 - Infrastructure stays technical

- Description: Infrastructure packages must not depend on core, application, adapters, bootstrap, or
  cmd.
- Acceptance criteria:
  - [x] `internal/infrastructure/**` denies domain, application, adapters, bootstrap, and cmd.
- Notes: Adapters may consume infrastructure helpers; infrastructure does not call back into
  adapters or application logic.

### FR-008 - Runtime behavior is preserved

- Description: The refactor must preserve existing API, scheduler, persistence, blockchain,
  webhook, and bootstrap behavior.
- Acceptance criteria:
  - [x] Existing Go tests pass.
  - [x] Existing pre-commit hooks pass.
  - [x] No database migration files are changed.
  - [x] No OpenAPI route or schema file is changed.
- Notes: Field moves between domain/application DTOs and port records are internal only.

## Non-functional requirements

- Performance (NFR-001): Port mapping must not add new external IO or asymptotic query/runtime cost.
- Availability/Reliability (NFR-002): Existing tests and pre-commit validation must pass.
- Security/Privacy (NFR-003): Denied-file and private-key checks remain enabled.
- Compliance (NFR-004): Conventional Commit policy remains enforced.
- Observability (NFR-005): Validation evidence must record the command names and outcomes.
- Maintainability (NFR-006): Future boundary violations must fail fast through keepline and
  pre-commit.

## Dependencies and integrations

- External systems: `github.com/po-sen/keepline` CLI and pre-commit.
- Internal services: existing Go packages under `cmd/`, `internal/`, `deployments/`, `scripts/`,
  `assets/`, and `specs/`.
