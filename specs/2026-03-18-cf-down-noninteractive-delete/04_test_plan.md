---
doc: 04_test_plan
spec_date: 2026-03-18
slug: cf-down-noninteractive-delete
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
  - Static verification of the delete scripts used by `make cf-down`.
  - Static verification that `make cf-down` still expands to the same delete steps in the same
    order.
- Not covered:
  - Live Cloudflare deletion against a real account.

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, NFR-001
  - Steps:
    - Inspect each delete script under `scripts/` that is called by `make cf-down`.
  - Expected:
    - Each script invokes `wrangler delete --force`.
- TC-002:
  - Linked requirements: FR-002, NFR-002
  - Steps:
    - Run `make -n cf-down`.
  - Expected:
    - The dry-run output lists the same five delete scripts in the same order and preserves the
      poller target arguments.

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, NFR-002
  - Steps:
    - Run `SPEC_DIR="specs/2026-03-18-cf-down-noninteractive-delete" bash scripts/spec-lint.sh`.
  - Expected:
    - The spec passes repository lint checks before and after implementation completion.

### E2E (if applicable)

- Scenario 1:
  - Optional operator check: run `make cf-down` in a configured Cloudflare environment and confirm
    no extra prompt appears.

## Edge cases and failure modes

- Case:
  - A delete script gains `--force` but accidentally drops an existing argument such as
    `--env "$TARGET_NETWORK"`.
- Expected behavior:
  - Static verification catches the missing argument before operators run teardown.

## NFR verification

- Reliability:
  - Confirm the implementation uses Wrangler's supported flag rather than shell-specific stdin
    automation.
- Maintainability:
  - Confirm the diff is limited to delete scripts and the directly related spec package.
