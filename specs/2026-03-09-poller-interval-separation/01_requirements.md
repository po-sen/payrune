---
doc: 01_requirements
spec_date: 2026-03-09
slug: poller-interval-separation
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Glossary (optional)

- Poll tick interval:
  - The worker wake-up cadence used by the poller process ticker.
- Receipt poll interval:
  - The reschedule duration written into `next_poll_at` after one receipt tracking is processed.

## Out-of-scope behaviors

- OOS1:
  - Dynamic per-status polling intervals.
- OOS2:
  - Automatic migration of deployed environments that still use the removed legacy env name.

## Functional requirements

### FR-001 - Separate poller interval inputs

- Description:
  - Poller configuration must accept distinct values for worker tick cadence and receipt reschedule cadence.
- Acceptance criteria:
  - [ ] `loadPollerConfigFromEnv` reads `POLL_TICK_INTERVAL` for worker ticker cadence.
  - [ ] `loadPollerConfigFromEnv` reads `RECEIPT_POLL_INTERVAL` for receipt `next_poll_at` cadence.
  - [ ] `loadPollerConfigFromEnv` does not read or depend on `POLL_INTERVAL`.
- Notes:
  - This requirement is limited to poller configuration and does not change webhook dispatcher config.

### FR-002 - Apply separated intervals at runtime

- Description:
  - Poller execution must use the separated interval values in their correct runtime responsibilities.
- Acceptance criteria:
  - [ ] The worker ticker uses only the configured poll tick interval.
  - [ ] `RunReceiptPollingCycleUseCase` receives only the receipt poll interval for `next_poll_at` scheduling.
  - [ ] Receipt lifecycle and status transitions remain unchanged apart from using the renamed interval input.
- Notes:
  - This change is naming and wiring cleanup; business policy stays the same.

### FR-003 - Align compose defaults with explicit names

- Description:
  - Compose defaults must expose the new env names so operators can tune the two intervals explicitly, and the touched poller env blocks must keep related settings grouped together.
- Acceptance criteria:
  - [ ] Bitcoin poller compose files use `POLL_TICK_INTERVAL` and `RECEIPT_POLL_INTERVAL` instead of `POLL_INTERVAL`.
  - [ ] Poller config tests cover the new env names and confirm the poller no longer accepts the removed legacy env name.
  - [ ] In the touched poller Compose env blocks, DB connection, poll scope, poll cadence, provider config, and receipt behavior settings appear in a stable concern-based order.
- Notes:
  - Default duration values may stay unchanged in this refactor.

## Non-functional requirements

- Performance (NFR-001):
  - The refactor must not add extra DB queries or external API calls to one poll cycle.
- Availability/Reliability (NFR-002):
  - Poller startup must remain deterministic: it either boots with the new env names or falls back to code defaults when they are unset.
- Security/Privacy (NFR-003):
  - No new secrets or network integrations are introduced.
- Compliance (NFR-004):
  - Not applicable.
- Observability (NFR-005):
  - Existing poller logs and output counts must remain unchanged.
- Maintainability (NFR-006):
  - Naming in bootstrap/use case input and ordering in touched poller Compose env blocks must make the poller settings easy to scan without extra comments.

## Dependencies and integrations

- External systems:
  - Docker Compose environment configuration for poller services.
- Internal services:
  - `cmd/poller`
  - `internal/bootstrap/poller.go`
  - `internal/application/use_cases/run_receipt_polling_cycle_use_case.go`
