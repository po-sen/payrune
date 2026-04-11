---
doc: 03_tasks
spec_date: 2026-04-09
slug: compose-env-example
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-07-ethereum-contract-readiness
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is still a small operational/docs change: keep one checked-in env example readable across both formal/mainnet-style and development paths, plus matching README/Makefile guidance.
- Upstream dependencies (`depends_on`):
  - `2026-04-07-ethereum-contract-readiness`
- Dependency gate before `READY`:
  - The dependency is already folder-wide `DONE`.
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - No new schema, no new integration, and no runtime design change.
  - What would trigger switching to Full mode:
    - If the request expanded into changing the compose env contract itself.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Validation steps are listed under each task below.

## Milestones

- M1:
  - Expand the compose env example file.
- M1A:
  - Trim redundant defaults from the checked-in development env file.
- M2:
  - Validate formal and development profile rendering plus repo checks.
- M3:
  - Simplify Makefile local compose behavior.
- M4:
  - Collapse local development topology into the main compose file.
- M5:
  - Simplify Cloudflare Make targets.

## Tasks (ordered)

1. T-001 - Expand `compose.env.example`
   - Scope:
     - Update `deployments/compose/compose.env.example` so it covers the formal/mainnet-style env contract plus the local development env surface used by the same compose file.
     - Keep Bitcoin testnet4 and Ethereum Sepolia blocks in the example, but disable their policy flags by default.
     - Keep `deployments/compose/compose.dev.env` as the ready-to-run local development env file, but remove entries that simply duplicate `compose.yaml` defaults.
   - Output:
     - A superset `compose.env.example`, plus a short README pointer if useful.
   - Linked requirements: FR-001, FR-002, FR-003, NFR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - Inspect `deployments/compose/compose.env.example` and `deployments/compose/compose.dev.env`.
     - [ ] Expected result:
       - `compose.env.example` covers both formal/mainnet-style and development envs, while `compose.dev.env` remains the ready-to-run development variant with only the necessary development overrides.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.
2. T-002 - Simplify Makefile compose selection
   - Scope:
     - Update `Makefile` so local development and formal/mainnet-style compose paths use explicit targets instead of env-file-presence switching.
     - Keep strict env-file checks on targets that need real operator config.
     - Keep the target recipes direct; avoid extra compose command assembly variables.
     - Add a small built-in `help` target for the supported entrypoints.
     - Keep Cloudflare targets available.
   - Output:
     - A simpler local-compose entrypoint in `Makefile`, plus matching README instructions.
   - Linked requirements: FR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - Inspect `Makefile` and README behavior description.
       - `make help`
       - `make -n up`
       - `cp deployments/compose/compose.env.example deployments/compose/compose.env && make -n up-mainnet && make -n down-mainnet && rm -f deployments/compose/compose.env`
       - `rm -f deployments/compose/compose.env && make -n down-mainnet`
     - [ ] Expected result:
       - The target names make it obvious which path is local development and which path is formal/mainnet-style, the required env file is expressed via plain Make file prerequisites, and `make help` summarizes the supported entrypoints.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.
3. T-003 - Collapse local development topology into `compose.yaml`
   - Scope:
     - Move local development topology into `compose.yaml`.
     - Keep one dedicated `development` profile for dev-only services.
     - Remove the need for a `mainnet` profile.
   - Output:
     - A single compose entrypoint with profile-based topology selection.
   - Linked requirements: FR-006, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command):
       - Inspect `compose.yaml`, `compose.dev.env`, and `Makefile` together.
     - [ ] Expected result:
       - One compose file owns both formal/mainnet-style and development topology, while only dev-only services need the `development` profile.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.
4. T-004 - Validate compose rendering and repo checks
   - Scope:
     - Validate the new env example, Makefile behavior, compose rendering, and repo checks.
   - Output:
     - Evidence that the new example file is syntactically valid and does not break repo formatting/linting.
   - Linked requirements: FR-004, FR-005, FR-006, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command):
       - `docker compose --env-file deployments/compose/compose.env.example -f deployments/compose/compose.yaml config`
       - `docker compose --env-file deployments/compose/compose.env.example --profile development -f deployments/compose/compose.yaml config`
       - `docker compose --env-file deployments/compose/compose.dev.env --profile development -f deployments/compose/compose.yaml config`
       - `make -n up`
       - `make -n up-mainnet`
       - `SPEC_DIR="specs/2026-04-09-compose-env-example" bash scripts/spec-lint.sh`
       - `bash scripts/precommit-run.sh`
     - [ ] Expected result:
       - Compose config rendering and repo checks pass.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.
5. T-005 - Simplify Cloudflare Make targets
   - Scope:
     - Remove the standalone Cloudflare migration Make target.
     - Keep migration inside `cf-up`.
     - Keep worker deploy/delete targets available.
   - Output:
     - A smaller Cloudflare target set in `Makefile`, plus matching README wording.
   - Linked requirements: FR-007
   - Validation:
     - [ ] How to verify (manual steps or command):
       - Inspect `Makefile` and README Cloudflare section.
     - [ ] Expected result:
       - `cf-migrate` is gone and `cf-up` still runs migration before deploying workers.
     - [ ] Logs/metrics to check (if applicable):
       - Not applicable.

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001
- FR-002A -> T-001, T-004
- FR-003 -> T-001
- FR-004 -> T-004
- FR-005 -> T-002, T-004
- FR-006 -> T-003, T-004
- FR-007 -> T-005
- NFR-001 -> T-004
- NFR-002 -> T-004
- NFR-003 -> T-001
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Remove `deployments/compose/compose.env.example` and any accompanying README pointer if the example proves misleading.

## Validation evidence

- `docker compose --env-file deployments/compose/compose.env.example -f deployments/compose/compose.yaml config`
- `docker compose --env-file deployments/compose/compose.env.example --profile development -f deployments/compose/compose.yaml config`
- `docker compose --env-file deployments/compose/compose.dev.env --profile development -f deployments/compose/compose.yaml config`
- `make help`
- `make -n up`
- `make -n up-mainnet`
- `cp deployments/compose/compose.env.example deployments/compose/compose.env && make -n down-mainnet && rm -f deployments/compose/compose.env`
- `rm -f deployments/compose/compose.env && make -n down-mainnet`
- `SPEC_DIR="specs/2026-04-09-compose-env-example" bash scripts/spec-lint.sh`
- `bash scripts/precommit-run.sh`
