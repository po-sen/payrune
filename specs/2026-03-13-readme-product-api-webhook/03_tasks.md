---
doc: 03_tasks
spec_date: 2026-03-13
slug: readme-product-api-webhook
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-10-cloudflare-workers-postgres
  - 2026-03-11-cloudflare-poller-workers
  - 2026-03-13-cloudflare-webhook-dispatcher-worker
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
  - This is a documentation-only change with no runtime, schema, or contract change.
- Upstream dependencies (`depends_on`):
  - `2026-03-10-cloudflare-workers-postgres`
  - `2026-03-11-cloudflare-poller-workers`
  - `2026-03-13-cloudflare-webhook-dispatcher-worker`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - The task is limited to a new root README.
  - What would trigger switching to Full mode:
    - Any runtime or contract change.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - Not skipped.

## Milestones

- M1:
  - Draft concise README content from current product behavior.
- M2:
  - Validate README against current API, webhook, and deployment entrypoints.

## Tasks (ordered)

1. T-001 - Add concise top-level product README

   - Scope:
     - Create `README.md` with product summary, public API list, main parameters, and local /
       Cloudflare deployment commands.
   - Output:
     - New repo-root `README.md`
   - Linked requirements: FR-001, FR-004, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): inspect README content against current compose,
           worker docs, and env template.
     - [ ] Expected result: README is concise and matches current deploy entrypoints.
     - [ ] Logs/metrics to check (if applicable): none

1. T-002 - Add API and webhook integration quick reference
   - Scope:
     - Document payment address API usage, status lookup, webhook headers, payload, and signature
       verification rule.
   - Output:
     - Updated README integration sections
   - Linked requirements: FR-002, FR-003, NFR-001, NFR-002
   - Validation:
     - [ ] How to verify (manual steps or command): compare examples and field names against
           OpenAPI and webhook notifier code.
     - [ ] Expected result: a human or AI can wire API requests and verify webhook signatures from
           the README alone.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-002
- FR-003 -> T-002
- FR-004 -> T-001
- NFR-001 -> T-001, T-002
- NFR-002 -> T-001, T-002

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore previous README state or remove the new document.
