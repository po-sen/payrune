---
doc: 01_requirements
spec_date: 2026-04-12
slug: cloudflare-enabled-sync
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-11-cloudflare-env-location
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
  - Cloudflare operator intent flag
- Definition:
  - A `*_ENABLED` env key used to turn an address policy on or off for the Cloudflare worker runtime.

## Out-of-scope behaviors

- OOS1:
  - Migrating Cloudflare worker env handling away from Wrangler secret sync.
- OOS2:
  - Expanding the set of non-enablement vars beyond those already documented.

## Functional requirements

### FR-001 - Sync Cloudflare `*_ENABLED` flags into worker runtime

- Description:
  - `make cf-up` must sync the Cloudflare worker policy enablement flags when they are set in `deployments/cloudflare/cloudflare.env` or shell env.
- Acceptance criteria:
  - [ ] AC1: `scripts/cf-payrune-worker-deploy.sh` includes all advertised Bitcoin `*_ENABLED` keys in its non-secret deploy-var path.
  - [ ] AC2: `scripts/cf-payrune-worker-deploy.sh` includes all advertised Ethereum `*_ENABLED` keys in its non-secret deploy-var path.
  - [ ] AC3: Empty values still remain unsynced.
- Notes:
  - `*_ENABLED` flags must not be handled as Wrangler secrets.

### FR-002 - Keep Cloudflare docs honest about synced values

- Description:
  - Cloudflare env example and README content must match the actual set of optional worker values supported by `make cf-up`.
- Acceptance criteria:
  - [ ] AC1: `deployments/cloudflare/payrune/README.md` distinguishes non-secret `*_ENABLED` deploy vars from secret-backed optional values.
  - [ ] AC2: `deployments/cloudflare/cloudflare.env.example` wording remains consistent with the actual sync behavior.
- Notes:
  - This requirement is about correctness of the current workflow, not about making the env file a full runtime catalog.

## Non-functional requirements

- Performance (NFR-001):
  - The fix must not add new network round trips beyond syncing the newly supported env keys.
- Availability/Reliability (NFR-002):
  - `make cf-up` must continue to work when these optional values are left blank.
- Security/Privacy (NFR-003):
  - Required secrets handling must remain unchanged, and `*_ENABLED` flags must not be sent through the secret-sync path.
- Compliance (NFR-004):
- Observability (NFR-005):
- Maintainability (NFR-006):
  - The Cloudflare env example should no longer advertise keys that `make cf-up` silently ignores.

## Dependencies and integrations

- External systems:
  - Wrangler secret sync
- Internal services:
  - `scripts/cf-payrune-worker-deploy.sh`
  - Cloudflare worker bootstrap/runtime env consumption
