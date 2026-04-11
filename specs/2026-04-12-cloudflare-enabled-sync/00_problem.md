---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - Cloudflare worker runtime now reads policy enablement from env keys such as `BITCOIN_MAINNET_LEGACY_ENABLED` and `ETHEREUM_SEPOLIA_CREATE2_ENABLED`.
  - The checked-in Cloudflare env example already exposes these keys as local operator inputs.
- Users or stakeholders:
  - Operators using `deployments/cloudflare/cloudflare.env` and `make cf-up`.
  - Developers relying on the Cloudflare worker to honor the same policy intent flags as the local Docker path.
- Why now:
  - The env example suggests these flags are configurable for Cloudflare, but the deploy script does not currently sync them into Wrangler.

## Constraints (optional)

- Technical constraints:
  - Keep the current Cloudflare deploy flow based on `make cf-up` and Wrangler secret syncing.
  - Do not redesign the worker runtime env model in this cleanup.
- Timeline/cost constraints:
  - Small Quick-mode bug fix.
- Compliance/security constraints:
  - Preserve shell-env-overrides-file behavior in the scripts.

## Problem statement

- Current pain:
  - Operators can set `*_ENABLED` in `deployments/cloudflare/cloudflare.env`, but those values are not actually synced into the worker runtime by `make cf-up`.
  - This creates a misleading configuration surface where policy intent looks configurable but may stay at the checked-in default inside the worker.
- Evidence or examples:
  - `deployments/cloudflare/cloudflare.env.example` lists `BITCOIN_*_ENABLED` and `ETHEREUM_*_ENABLED`.
  - `scripts/cf-payrune-worker-deploy.sh` currently syncs xpubs, CREATE2 config, and provider auth, but not the enablement flags.

## Goals

- G1:
  - Make Cloudflare `*_ENABLED` values actually reach the worker runtime when set in `deployments/cloudflare/cloudflare.env`.
- G2:
  - Align Cloudflare docs with the real set of env values synced by `make cf-up`.
- G3:
  - Stop treating `*_ENABLED` flags as secrets in the Cloudflare deploy path.

## Non-goals (out of scope)

- NG1:
  - Reworking Cloudflare worker runtime defaults in `wrangler.toml`.
- NG2:
  - Changing local Docker Compose env behavior.

## Assumptions

- A1:
  - Wrangler `deploy --var KEY:VALUE` is a suitable transport for non-secret operator intent flags.
- A2:
  - The current worker runtime already reads these env keys if they are present.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Cloudflare operator intent flags are no longer misleading.
- Target:

  - Every `*_ENABLED` key advertised in `deployments/cloudflare/cloudflare.env.example` is included in the `make cf-up` worker deploy path as a non-secret runtime var.

- Metric:
  - Cloudflare docs match actual deploy behavior.
- Target:
  - README text for Cloudflare optional worker values and env examples matches what `scripts/cf-payrune-worker-deploy.sh` really syncs.
