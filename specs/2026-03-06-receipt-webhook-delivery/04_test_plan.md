---
doc: 04_test_plan
spec_date: 2026-03-06
slug: receipt-webhook-delivery
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-status-change-notification
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Test Plan

## Unit

- TC-001:

  - Linked requirements: FR-004
  - Steps:
    - Build webhook payload and send through test server.
  - Expected:
    - Request body and signature headers match expected event contract.

- TC-002:

  - Linked requirements: FR-005
  - Steps:
    - Dispatch use case receives a transport/non-`2xx` failure under max attempts.
  - Expected:
    - Row is rescheduled as `pending` with incremented attempts and `last_error`.

- TC-003:

  - Linked requirements: FR-005
  - Steps:
    - Dispatch use case receives a failure at `max_attempts`.
  - Expected:
    - Row is marked `failed` and no further retry is scheduled.

- TC-004:
  - Linked requirements: FR-006
  - Steps:
    - Load dispatcher env with invalid URL, secret, or durations.
  - Expected:
    - Startup/config parsing fails fast.

## Integration

- TC-101:

  - Linked requirements: FR-002, NFR-001
  - Steps:
    - Claim pending rows with expired and active leases.
  - Expected:
    - Only due/unleased rows are claimed and leased.

- TC-102:

  - Linked requirements: FR-005, NFR-004
  - Steps:
    - Persist success, retry, and terminal failed outcomes through repository methods.
  - Expected:
    - Delivery status fields update correctly and clear leases.

- TC-103:
  - Linked requirements: FR-003, FR-004, NFR-002
  - Steps:
    - Use TLS test server and notifier adapter.
  - Expected:
    - Adapter accepts HTTPS URL, signs request, and treats only `2xx` as success.

## Functional

- TC-201:

  - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007
  - Steps:
    - Seed a pending notification row and run one dispatch cycle against a controlled webhook server.
  - Expected:
    - Worker claims the row, POSTs the event once, and marks it `sent`.

- TC-202:
  - Linked requirements: FR-007, NFR-005
  - Steps:
    - Render `compose.yaml + compose.test.yaml` with `--env-file deployments/compose/compose.test.env`.
  - Expected:
    - Dispatcher service is present in base compose, fake webhook URL/secret are loaded from the committed env file, and test env adds the fake receiver service.
