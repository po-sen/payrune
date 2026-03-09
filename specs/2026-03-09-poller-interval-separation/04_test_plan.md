---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - Poller env parsing for the new interval vars and explicit rejection of the legacy env name.
  - Poller bootstrap/runtime wiring of separate interval fields.
  - Receipt polling reschedule behavior using the renamed interval input.
  - Compose-related poller config compilation/tests and YAML parse validation after poller env reordering.
- Not covered:
  - Manual production rollout validation.
  - Status-aware polling cadence changes.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, NFR-002
  - Steps:
    - Set `POLL_TICK_INTERVAL`, `RECEIPT_POLL_INTERVAL`, `POLL_CLAIM_TTL`, `POLL_BATCH_SIZE`, `POLL_CHAIN`, and `POLL_NETWORK`.
    - Call `loadPollerConfigFromEnv`.
  - Expected:
    - The config loads successfully with distinct tick and receipt poll interval values.
- TC-002:
  - Linked requirements: FR-001, NFR-002
  - Steps:
    - Set only legacy `POLL_INTERVAL` plus the other required envs.
    - Call `loadPollerConfigFromEnv`.
  - Expected:
    - The config ignores `POLL_INTERVAL`; only defaults or the explicit new env names affect the loaded intervals.
- TC-003:
  - Linked requirements: FR-002, NFR-006
  - Steps:
    - Execute `RunReceiptPollingCycleUseCase` with a distinct receipt poll interval.
    - Observe persisted `next_poll_at` in the fake store.
  - Expected:
    - `next_poll_at` is derived from the receipt poll interval only.

### Integration

- TC-101:
  - Linked requirements: FR-002, NFR-005
  - Steps:
    - Run targeted Go tests for `cmd/poller`, `internal/bootstrap`, and `internal/application/use_cases`.
  - Expected:
    - Poller config parsing and runtime wiring remain green after the refactor.
- TC-102:
  - Linked requirements: FR-003, NFR-002
  - Steps:
    - Run targeted Go tests for compose-related DI/config packages, `go list ./...`, and YAML parse validation for `deployments/compose/*.yaml`.
  - Expected:
    - Updated compose/config wiring compiles cleanly with the new env names and the reordered Compose files remain valid YAML.

### E2E (if applicable)

- Scenario 1:
  - Not applicable for this refactor.

## Edge cases and failure modes

- Case:
  - Only the removed legacy env var is set.
- Expected behavior:

  - The poller ignores the legacy env and falls back to code defaults for both interval fields.

- Case:
  - No interval env vars are set.
- Expected behavior:
  - The poller uses existing default durations and starts successfully.

## NFR verification

- Performance:
  - Confirm the change is limited to config parsing and in-memory scheduling; no new IO is introduced.
- Reliability:
  - Confirm removing the legacy env path does not change runtime behavior when the new env names are configured or omitted.
- Security:
  - Confirm no new credentials or external endpoints are introduced.
