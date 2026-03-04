---
doc: 03_tasks
spec_date: 2026-03-04
slug: compose-bitcoin-test-env
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-03-deploy-service-compose-dockerfile
  - 2026-03-04-bitcoin-address-vectors
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - Change only adds one compose override and one Makefile rule without new data model/integration complexity.
- Upstream dependencies (`depends_on`):
  - `2026-03-03-deploy-service-compose-dockerfile`
  - `2026-03-04-bitcoin-address-vectors`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - No architecture boundary or protocol design changes.
  - What would trigger switching to Full mode:
    - Introducing runtime configuration service, secrets backend, or complex rollout controls.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1: Spec package created and linted.
- M2: Compose test env override file added.
- M3: Makefile startup rule added and validated.

## Tasks (ordered)

1. T-001 - Author spec for compose test env profile

   - Scope:
     - Capture requirements and validation for hardcoded xpub override + make startup rule.
   - Output:
     - `specs/2026-03-04-compose-bitcoin-test-env/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, NFR-004
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-04-compose-bitcoin-test-env" bash scripts/spec-lint.sh`
     - [x] Expected result: lint exits with code 0.
     - [x] Logs/metrics to check (if applicable): no spec-lint errors.

2. T-002 - Add compose override with hardcoded bitcoin xpub fixtures

   - Scope:
     - Add a new compose override file that sets app bitcoin xpub env values from current unit-test fixtures.
   - Output:
     - `deployments/compose/compose.test-env.yaml`
   - Linked requirements: FR-001, NFR-002, NFR-003, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `docker compose -f deployments/compose/compose.yaml -f deployments/compose/compose.test-env.yaml config`
     - [x] Expected result: merged config contains all bitcoin xpub env values under `services.app.environment`.
     - [x] Logs/metrics to check (if applicable): no unresolved variable placeholders for those keys.

3. T-003 - Add make startup rule for test env override
   - Scope:
     - Add an additive Makefile rule to launch compose using the new override with direct path (no extra variable).
   - Output:
     - `Makefile`
   - Linked requirements: FR-002, FR-003, NFR-001, NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `make up-test-env`
     - [x] Expected result: stack starts via existing `up` flow with override file applied.
     - [x] Logs/metrics to check (if applicable): app service has expected bitcoin xpub environment values in compose config.

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-003
- FR-003 -> T-001, T-003
- NFR-001 -> T-003
- NFR-002 -> T-002
- NFR-003 -> T-002
- NFR-004 -> T-001
- NFR-005 -> T-003
- NFR-006 -> T-002

## Rollout and rollback

- Feature flag:
  - Not applicable.
- Migration sequencing:
  - Not applicable.
- Rollback steps:
  - Remove the new compose override file and the new Makefile rule.
