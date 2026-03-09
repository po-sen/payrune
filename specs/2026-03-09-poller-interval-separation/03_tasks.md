---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a focused configuration and naming refactor with no schema change, no new integration, and no business flow redesign.
- Upstream dependencies (`depends_on`):
  - None.
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The change is limited to env parsing, poller wiring, and use case input naming.
  - What would trigger switching to Full mode:
    - A scheduler redesign, status-aware polling strategy, or persistence change.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable; `04_test_plan.md` is included.

## Milestones

- M1:
  - Finalize the interval separation spec and removal of the legacy env.
- M2:
  - Update poller config, runtime wiring, compose defaults, and tests.

## Tasks (ordered)

1. T-001 - Finalize interval separation spec
   - Scope:
     - Capture the split between worker tick cadence and receipt reschedule cadence, including removal of the legacy env name.
   - Output:
     - `specs/2026-03-09-poller-interval-separation/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-09-poller-interval-separation" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec lint passes and all docs describe the same env split and removal rules.
     - [ ] Logs/metrics to check (if applicable): none
2. T-002 - Separate poller config and runtime interval wiring
   - Scope:
     - Update poller config, env parsing, bootstrap wiring, and receipt polling input naming so worker tick and receipt reschedule use different fields, with no legacy env fallback.
   - Output:
     - Poller runtime uses separate interval fields without changing receipt lifecycle behavior.
   - Linked requirements: FR-001, FR-002, NFR-001, NFR-002, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./cmd/poller ./internal/bootstrap ./internal/application/use_cases -count=1`
     - [ ] Expected result: poller config parsing, bootstrap wiring, and polling use case tests pass with separated interval semantics.
     - [ ] Logs/metrics to check (if applicable): none
3. T-003 - Expose explicit interval envs in compose defaults
   - Scope:
     - Update poller compose files to use the new env names only and keep the touched env blocks grouped in a stable concern-based order.
   - Output:
     - Compose defaults show `POLL_TICK_INTERVAL` and `RECEIPT_POLL_INTERVAL` explicitly, with the touched poller env blocks ordered by concern.
   - Linked requirements: FR-003, NFR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./cmd/poller ./internal/infrastructure/di -count=1`, `GOCACHE=/tmp/go-build go list ./...`, and `ruby --disable-gems -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }' deployments/compose/*.yaml`
     - [ ] Expected result: poller config tests pass, packages compile cleanly, and all Compose YAML files still parse after env reordering.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002
- FR-003 -> T-001, T-003
- NFR-001 -> T-002
- NFR-002 -> T-002, T-003
- NFR-005 -> T-002
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Update environments to `POLL_TICK_INTERVAL` and `RECEIPT_POLL_INTERVAL` before deploying this change where explicit configuration is required.
- Rollback steps:
  - Revert to the previous single-interval wiring and restore `POLL_INTERVAL` support if the env rename causes regressions.
