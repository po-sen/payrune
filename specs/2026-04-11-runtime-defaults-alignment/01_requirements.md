---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Poller reschedule interval:
  - The duration used to schedule the next normal receipt re-poll for a claimed tracking row.
- Ethereum Sepolia required confirmations:
  - The default number of confirmations required before a Sepolia receipt is treated as confirmed.

## Out-of-scope behaviors

- OOS1:
  - Renaming env vars.
- OOS2:
  - Changing unrelated runtime defaults.

## Functional requirements

### FR-001 - Poller checked-in defaults use `5m`

- Description:
  - Checked-in default sources for `POLL_RESCHEDULE_INTERVAL` must use `5m`.
- Acceptance criteria:
  - [ ] Every poller service in `deployments/compose/compose.yaml` uses `${POLL_RESCHEDULE_INTERVAL:-5m}`.
  - [ ] `deployments/compose/compose.env.example` sets `POLL_RESCHEDULE_INTERVAL=5m`.
  - [ ] `deployments/cloudflare/payrune/wrangler.toml` sets `POLL_RESCHEDULE_INTERVAL = "5m"`.
  - [ ] `internal/bootstrap/poller_worker.go` uses a `5m` fallback when the env is absent.
- Notes:
  - Explicit env overrides must still take precedence.

### FR-002 - Sepolia checked-in confirmations defaults use `12`

- Description:
  - Checked-in default sources for `ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS` must use `12`.
- Acceptance criteria:
  - [ ] `deployments/compose/compose.yaml` sets `${ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS:-12}`.
  - [ ] `deployments/compose/compose.env.example` sets `ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS=12`.
  - [ ] `deployments/cloudflare/payrune/wrangler.toml` sets `ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS = "12"`.
  - [ ] `internal/bootstrap/api.go` uses `12` as the Sepolia required-confirmations fallback.
- Notes:
  - Explicit env overrides must still take precedence.

### FR-003 - Validation stays green after both default changes

- Description:
  - The repo must keep focused runtime/config validation passing after both default updates.
- Acceptance criteria:
  - [ ] `go test ./internal/bootstrap` passes.
  - [ ] `docker compose --env-file deployments/compose/compose.env.example -f deployments/compose/compose.yaml config` passes.
  - [ ] `SPEC_DIR="specs/2026-04-11-runtime-defaults-alignment" bash scripts/spec-lint.sh` passes.
  - [ ] `bash scripts/precommit-run.sh` passes.
- Notes:
  - This requirement covers both changes together, not as separate validation tracks.

## Non-functional requirements

- Performance (NFR-001):
  - The change must remain a default-value update only; no new runtime branches or IO.
- Availability/Reliability (NFR-002):
  - Compose config rendering and bootstrap tests must keep passing after the default changes.
- Security/Privacy (NFR-003):
  - No secret handling changes.
- Compliance (NFR-004):
  - No additional compliance requirements.
- Observability (NFR-005):
  - No new observability behavior required.
- Maintainability (NFR-006):
  - All checked-in default sources for both settings must stay aligned to the same value.

## Dependencies and integrations

- External systems:
  - Docker Compose config rendering.
  - Cloudflare Workers config.
- Internal services:
  - [`internal/bootstrap/api.go`](/Users/posen/Desktop/payrune/internal/bootstrap/api.go)
  - [`internal/bootstrap/poller_worker.go`](/Users/posen/Desktop/payrune/internal/bootstrap/poller_worker.go)
  - [`deployments/compose/compose.yaml`](/Users/posen/Desktop/payrune/deployments/compose/compose.yaml)
  - [`deployments/compose/compose.env.example`](/Users/posen/Desktop/payrune/deployments/compose/compose.env.example)
  - [`deployments/cloudflare/payrune/wrangler.toml`](/Users/posen/Desktop/payrune/deployments/cloudflare/payrune/wrangler.toml)
