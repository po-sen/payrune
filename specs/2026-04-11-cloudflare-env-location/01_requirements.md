---
doc: 01_requirements
spec_date: 2026-04-11
slug: cloudflare-env-location
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

- Cloudflare env file:
  - The local operator env file used by `make cf-up` and related Cloudflare helper scripts.

## Out-of-scope behaviors

- OOS1:
  - Changing Cloudflare secret names or deploy order.
- OOS2:
  - Introducing a new env-loading library or central shell framework.

## Functional requirements

### FR-001 - Move the checked-in Cloudflare env example under deployments

- Description:
  - The checked-in Cloudflare env example must move out of repo root into the Cloudflare deployment area.
- Acceptance criteria:
  - [ ] AC1: The example file lives at `deployments/cloudflare/cloudflare.env.example`.
  - [ ] AC2: The old root `.env.cloudflare.example` file is removed.
- Notes:
  - Naming should stay consistent with the existing `compose.env.example` pattern.

### FR-002 - Prefer the new deployment-local env path in scripts and docs

- Description:
  - Cloudflare deploy/migrate scripts and user-facing docs must point operators at the new deployment-local env path.
- Acceptance criteria:
  - [ ] AC1: README instructions use `deployments/cloudflare/cloudflare.env.example` and `deployments/cloudflare/cloudflare.env`.
  - [ ] AC2: Cloudflare deployment docs under `deployments/cloudflare/` reference the new path.
  - [ ] AC3: Cloudflare helper scripts look for the new deployment-local env file first.
- Notes:
  - Root-level path references should be removed from normal documentation.

### FR-003 - Preserve backward compatibility for existing local root env files

- Description:
  - Existing local `.env.cloudflare` users should not be broken immediately by the path move.
- Acceptance criteria:
  - [ ] AC1: If `deployments/cloudflare/cloudflare.env` exists, scripts load it.
  - [ ] AC2: If the new file does not exist but root `.env.cloudflare` exists, scripts still load the root file.
  - [ ] AC3: Script/operator messages make it clear which file location is being used.
- Notes:
  - This is a transition compatibility requirement, not a permanent guarantee.
  - The legacy root file does not need to stay gitignored once the deployment-local path exists.

## Non-functional requirements

- Performance (NFR-001):
  - The path move must not add meaningful overhead beyond one extra file-existence check in shell scripts.
- Availability/Reliability (NFR-002):
  - Existing Cloudflare deploy flows must continue working after the change when either the new or legacy local env file exists.
- Security/Privacy (NFR-003):
- Compliance (NFR-004):
- Observability (NFR-005):
  - Script output should clearly state the env-loading behavior.
- Maintainability (NFR-006):
  - Cloudflare deployment-local config should now live under `deployments/cloudflare/`, matching repo organization.

## Dependencies and integrations

- External systems:
  - Cloudflare Wrangler CLI.
- Internal services:
  - `Makefile`
  - `scripts/cf-*.sh`
  - `README.md`
  - `deployments/cloudflare/payrune/README.md`
