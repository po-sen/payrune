---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale: This combined work includes tooling configuration, pre-commit integration,
  cross-package application port relocation, adapter/use-case refactoring, and strict import-policy
  enforcement.
- Upstream dependencies (`depends_on`): []
- Dependency gate before `READY`: no upstream dependencies.
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: Not applicable.
  - What would trigger switching to Full mode: Already Full.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): Not applicable.

## Milestones

- M1: Install/configure keepline and pre-commit policy hooks.
- M2: Move adapter-facing contracts under application ports.
- M3: Refactor adapters and use cases to use port records.
- M4: Enforce strict layer and adapter-family import policy.
- M5: Consolidate incremental specs into this single source of truth.

## Tasks (ordered)

1. T-001 - Add keepline configuration
   - Scope: Add root `keepline.toml` with commit-message, file, and import policy sections.
   - Output: Repository keepline policy file.
   - Linked requirements: FR-001 / NFR-003 / NFR-004 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `keepline config-check`
     - [x] Expected result: `OK: keepline.toml is valid`.
     - [x] Logs/metrics to check (if applicable): keepline output.
2. T-002 - Update pre-commit integration
   - Scope: Pin keepline provider to `v0.17.0` and add local import-check enforcement.
   - Output: Updated `.pre-commit-config.yaml`.
   - Linked requirements: FR-002 / NFR-002 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `bash scripts/precommit-run.sh`
     - [x] Expected result: pre-commit passes and includes `keepline import check`.
     - [x] Logs/metrics to check (if applicable): pre-commit hook output.
3. T-003 - Inventory and remove forbidden adapter imports
   - Scope: Find adapter imports of domain, application internals, bootstrap, cmd, and concrete
     adapter families.
   - Output: Adapter imports no longer cross forbidden boundaries.
   - Linked requirements: FR-005 / FR-006 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `rg -n "internal/(domain|application|bootstrap)" internal/adapters -g "**/*.go"`
     - [x] Expected result: no forbidden core/composition adapter imports remain.
     - [x] Logs/metrics to check (if applicable): search output.
4. T-004 - Move contracts under application ports
   - Scope: Move inbound and outbound contract interfaces/records under
     `internal/application/ports/inbound` and `internal/application/ports/outbound`.
   - Output: Application-owned port packages.
   - Linked requirements: FR-004 / FR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `rg -n "payrune/internal/ports" internal cmd`
     - [x] Expected result: no old `internal/ports` import path remains.
     - [x] Logs/metrics to check (if applicable): search output.
5. T-005 - Refactor adapters, use cases, and bootstrap
   - Scope: Update inbound/outbound adapters, use cases, tests, and bootstrap wiring to exchange
     port records instead of domain/application implementation types at adapter boundaries.
   - Output: Behavior-preserving package boundary refactor.
   - Linked requirements: FR-004 / FR-005 / FR-008 / NFR-001 / NFR-002
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`
     - [x] Expected result: all packages pass.
     - [x] Logs/metrics to check (if applicable): Go test output.
6. T-006 - Strengthen import policy
   - Scope: Encode domain, application, port, adapter, adapter-family, and infrastructure import
     rules in `keepline.toml`.
   - Output: Strict keepline import policy.
   - Linked requirements: FR-003 / FR-004 / FR-005 / FR-006 / FR-007 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `keepline import-check` and
           `keepline check --scope all`
     - [x] Expected result: all checks pass.
     - [x] Logs/metrics to check (if applicable): keepline output.
7. T-007 - Consolidate spec package
   - Scope: Merge the incremental keepline and architecture specs into this Full spec and remove
     the superseded spec folders.
   - Output: One authoritative spec package for the keepline architecture policy work.
   - Linked requirements: FR-001 / FR-002 / FR-003 / FR-004 / FR-005 / FR-006 / FR-007 / FR-008 /
     NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-04-29-keepline-architecture-policy" bash scripts/spec-lint.sh`
     - [x] Expected result: spec lint passes.
     - [x] Logs/metrics to check (if applicable): spec-lint output.

## Validation evidence

- `keepline config-check`: passed.
- `keepline import-check`: passed.
- `keepline check --scope all`: passed.
- `go test ./...`: passed.
- `bash scripts/precommit-run.sh`: passed when run with permissions for Go build cache writes and
  local `httptest` listeners.
- `SPEC_DIR="specs/2026-04-29-keepline-architecture-policy" bash scripts/spec-lint.sh`: passed.

## Traceability (optional)

- FR-001 -> T-001, T-007
- FR-002 -> T-002, T-007
- FR-003 -> T-006, T-007
- FR-004 -> T-004, T-005, T-006, T-007
- FR-005 -> T-003, T-005, T-006, T-007
- FR-006 -> T-003, T-006, T-007
- FR-007 -> T-006, T-007
- FR-008 -> T-005, T-007
- NFR-001 -> T-005
- NFR-002 -> T-002, T-005
- NFR-003 -> T-001, T-006
- NFR-004 -> T-001
- NFR-005 -> T-007
- NFR-006 -> T-001, T-002, T-003, T-004, T-006, T-007

## Rollout and rollback

- Feature flag: none.
- Migration sequencing: configure keepline, refactor ports/imports, strengthen policy, validate,
  and consolidate specs.
- Rollback steps: revert the port/package refactor, keepline policy changes, and pre-commit import
  hook together if the strict architecture policy is intentionally relaxed by a later spec.
