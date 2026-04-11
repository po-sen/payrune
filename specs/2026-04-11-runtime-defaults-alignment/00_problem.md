---
doc: 00_problem
spec_date: 2026-04-11
slug: runtime-defaults-alignment
mode: Quick
status: DONE
owners:
  - codex
depends_on: []
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
  - Several checked-in runtime defaults had drifted from the currently desired operating profile.
  - Two concrete defaults needed adjustment on the same day:
    - `POLL_RESCHEDULE_INTERVAL` from `10m` to `5m`
    - `ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS` from `1` to `12`
- Users or stakeholders:
  - Operators running local Compose and Cloudflare deployments.
  - Developers relying on checked-in defaults and examples.
- Why now:
  - The repo should keep runtime defaults, checked-in env examples, and worker/bootstrap fallbacks aligned instead of carrying multiple tiny spec folders for closely related default tuning.

## Constraints (optional)

- Technical constraints:
  - Keep this as a Quick-mode default-value change.
  - Do not rename env vars or change override semantics.
- Timeline/cost constraints:
  - Prefer a small, explicit update across the existing checked-in default sources.
- Compliance/security constraints:
  - No secrets or provider credentials are involved.

## Problem statement

- Current pain:
  - Checked-in defaults for poller cadence and Sepolia confirmations can drift between Compose, env examples, Cloudflare config, and bootstrap/runtime fallbacks.
  - Keeping one tiny spec per default tweak creates unnecessary spec fragmentation for closely related operational tuning.
- Evidence or examples:
  - [`deployments/compose/compose.yaml`](/Users/posen/Desktop/payrune/deployments/compose/compose.yaml)
  - [`deployments/compose/compose.env.example`](/Users/posen/Desktop/payrune/deployments/compose/compose.env.example)
  - [`deployments/cloudflare/payrune/wrangler.toml`](/Users/posen/Desktop/payrune/deployments/cloudflare/payrune/wrangler.toml)
  - [`internal/bootstrap/api.go`](/Users/posen/Desktop/payrune/internal/bootstrap/api.go)
  - [`internal/bootstrap/poller_worker.go`](/Users/posen/Desktop/payrune/internal/bootstrap/poller_worker.go)

## Goals

- G1:
  - Change the checked-in default `POLL_RESCHEDULE_INTERVAL` from `10m` to `5m`.
- G2:
  - Change the checked-in default `ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS` from `1` to `12`.
- G3:
  - Keep local Compose, env example, Cloudflare worker vars, and bootstrap/runtime defaults aligned for both changes.
- G4:
  - Preserve existing explicit env override behavior.

## Non-goals (out of scope)

- NG1:
  - Changing unrelated tick, claim-TTL, batch-size, or expiry defaults.
- NG2:
  - Changing Ethereum mainnet confirmation defaults or env names.

## Assumptions

- A1:
  - These changes apply to checked-in defaults only; operator-owned local env files are not rewritten automatically.
- A2:
  - The requested changes are still default-value updates, not new behavioral flow changes.

## Open questions

- None for this scoped change.

## Success metrics

- Metric:
  - Every checked-in default source now reflects the desired values for both settings.
- Target:
  - `deployments/compose/compose.yaml` defaults poller rescheduling to `5m`.
  - `deployments/compose/compose.env.example` exposes `POLL_RESCHEDULE_INTERVAL=5m`.
  - `deployments/cloudflare/payrune/wrangler.toml` sets `POLL_RESCHEDULE_INTERVAL = "5m"`.
  - `deployments/compose/compose.yaml` defaults `ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS` to `12`.
  - `deployments/compose/compose.env.example` exposes `ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS=12`.
  - `deployments/cloudflare/payrune/wrangler.toml` sets `ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS = "12"`.
  - Related tests and repo validation pass.
