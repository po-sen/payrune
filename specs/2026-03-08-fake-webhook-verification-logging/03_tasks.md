---
doc: 03_tasks
spec_date: 2026-03-08
slug: fake-webhook-verification-logging
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-webhook-delivery
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
  - This is a small, test-environment-only enhancement with no schema change and no production contract change.
- Upstream dependencies (`depends_on`):
  - `2026-03-06-receipt-webhook-delivery`
- Dependency gate before `READY`:
  - The dependency spec is already `DONE`.
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The change is limited to a local test receiver, handler logic, and compose wiring.
  - What would trigger switching to Full mode:
    - A change to the production webhook contract or to persistent delivery state.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Add receiver-side verification and logs.
- M2:
  - Add tests and test-compose env wiring.

## Tasks (ordered)

1. T-001 - Implement receiver-side signature verification and logs

   - Scope:
     - Add helper logic to the fake receiver that reads the raw body, computes the expected HMAC-SHA256 signature, and keeps only one readable request log block.
     - Reject requests with invalid signatures.
   - Output:
     - The fake receiver shows the full incoming request in one readable log block.
   - Linked requirements: FR-001, FR-002, NFR-001, NFR-002, NFR-003
   - Validation:
     - [x] `go test ./cmd/fake_webhook_receiver -count=1`
     - [x] Valid signed requests return `204`; invalid signatures return non-2xx.
     - [x] Receiver logs include full headers and raw body.
     - [x] Receiver logs are formatted in readable multi-line sections.
     - [x] Receiver no longer emits separate `verification` or `payload` logs for successful requests.

1. T-002 - Wire test env secret into the fake receiver
   - Scope:
     - Pass the webhook secret into the fake receiver service in `compose.test.yaml`.
     - Keep the existing dispatcher env contract unchanged.
   - Output:
     - Local compose test runs can exercise receiver-side signature verification end-to-end.
   - Linked requirements: FR-003, NFR-001, NFR-002
   - Validation:
     - [x] `docker compose --env-file deployments/compose/compose.test.env -f deployments/compose/compose.yaml -f deployments/compose/compose.test.yaml config`
     - [x] `bash scripts/precommit-run.sh`

## Traceability

- FR-001 -> T-001
- FR-002 -> T-001
- FR-003 -> T-002
- NFR-001 -> T-001, T-002
- NFR-002 -> T-001, T-002
- NFR-003 -> T-001

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Revert the fake receiver verification handler and remove the extra compose env wiring.
