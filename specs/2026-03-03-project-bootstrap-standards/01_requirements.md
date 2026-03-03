---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Spec package: The five markdown documents under `specs/YYYY-MM-DD-slug/`.
- Core layer: `internal/domain` + `internal/application` packages.
- Adapter layer: `internal/adapters` packages handling IO/transport.

## Out-of-scope behaviors

- OOS1: Business-specific payment flows.
- OOS2: Runtime deployment manifests and cloud infrastructure.

## Functional requirements

### FR-001 - Project AGENTS policy file

- Description:
  - The repository MUST contain a root-level `AGENTS.md` that defines mandatory development workflow and preserves guidance from required skills.
- Acceptance criteria:
  - [ ] `AGENTS.md` exists at repo root.
  - [ ] File explicitly mandates spec-first workflow in `specs/`.
  - [ ] File explicitly mandates Go project layout and Clean Architecture + Hexagonal boundaries.
  - [ ] File explicitly mandates Conventional Commit process when commit is requested.
  - [ ] File instructs scripts placement under `scripts/`.
- Notes:
  - Preserve enough detail so future sessions do not depend on external skill registry.

### FR-002 - Spec-first workflow baseline

- Description:
  - The repository MUST include one bootstrap spec package created before implementation artifacts.
- Acceptance criteria:
  - [ ] Folder `specs/2026-03-03-project-bootstrap-standards/` exists.
  - [ ] Files `00_problem.md`, `01_requirements.md`, `02_design.md`, `03_tasks.md`, and `04_test_plan.md` exist.
  - [ ] Frontmatter values are consistent across all five docs.
  - [ ] `03_tasks.md` contains explicit mode decision and rationale.
- Notes:
  - Selected mode is `Full` for architecture and quality-gate definition.

### FR-003 - Go service bootstrap layout

- Description:
  - The repository MUST provide a compilable Go baseline aligned with project layout and clean hexagonal boundaries.
- Acceptance criteria:
  - [ ] `go.mod` exists at root.
  - [ ] Entry point exists under `cmd/<app>/main.go`.
  - [ ] Domain/application/adapters/infrastructure/bootstrap packages exist under `internal/`.
  - [ ] Domain/application packages do not import adapters/infrastructure/bootstrap.
  - [ ] `go list ./...` succeeds.
- Notes:
  - Include a minimal health check flow to validate wiring.

### FR-004 - Pre-commit baseline and verification

- Description:
  - The repository MUST include the provided pre-commit hooks, plus any required local support files/scripts, and a documented verification path.
- Acceptance criteria:
  - [ ] `.pre-commit-config.yaml` exists and matches required hook set.
  - [ ] `.markdownlint.json` exists and is referenced by hook configuration.
  - [ ] Validation scripts used for setup/verification are located under `scripts/`.
  - [ ] Running default-stage hooks succeeds with `pre-commit run --all-files`.
- Notes:
  - `govulncheck` remains manual stage per provided configuration.

### FR-005 - Commit policy for future implementation work

- Description:
  - Project guidance MUST define commit message generation and execution rules based on Conventional Commits.
- Acceptance criteria:
  - [ ] `AGENTS.md` includes message format and type/scope inference rules.
  - [ ] `AGENTS.md` states default behavior: stage all and commit unless user asks draft-only.
  - [ ] `AGENTS.md` includes breaking-change notation requirements.
- Notes:
  - This requirement defines policy; no commit is required in this task.

## Non-functional requirements

- Performance (NFR-001): `go test ./... -short -count=1` should complete within 60 seconds on local dev machine baseline.
- Availability/Reliability (NFR-002): Bootstrap HTTP health endpoint returns deterministic JSON with HTTP 200 for GET requests and HTTP 405 for unsupported methods.
- Security/Privacy (NFR-003): Pre-commit default stages must include secret leakage detection via `detect-private-key`.
- Compliance (NFR-004): Spec documents must pass repository-local `scripts/spec-lint.sh` checks before status can be `READY`.
- Observability (NFR-005): Bootstrap server should emit fatal startup errors to stderr through standard logger.
- Maintainability (NFR-006): Architectural boundaries must remain enforceable by package structure and import direction.

## Dependencies and integrations

- External systems:
  - GitHub-hosted hook repos referenced by `.pre-commit-config.yaml`.
- Internal services:
  - None for bootstrap baseline.
