---
doc: 03_tasks
spec_date: 2026-03-13
slug: readme-cloudflare-credentials
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-13-readme-product-api-webhook
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale:
  - This is a small documentation-only clarification.
- Upstream dependencies (`depends_on`):
  - `2026-03-13-readme-product-api-webhook`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip:
    - No code or runtime behavior changes are involved.
  - What would trigger switching to Full mode:
    - Any deployment script or secret-sync behavior change.
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task):
    - In the task validation steps below.

## Milestones

- M1:
  - Clarify credential placement in README and env template.

## Tasks (ordered)

1. T-001 - Clarify Cloudflare credential storage
   - Scope:
     - Update root README and `.env.cloudflare.example` to state that
       `CLOUDFLARE_ACCOUNT_ID` and `CLOUDFLARE_API_TOKEN` may live in `.env.cloudflare`, while
       still allowing `wrangler login` and CI secrets.
   - Output:
     - Updated `README.md`
     - Updated `.env.cloudflare.example`
   - Linked requirements: FR-001, FR-002, NFR-001
   - Validation:
     - [ ] How to verify (manual steps or command): inspect README and `.env.cloudflare.example`.
     - [ ] Expected result: both files show account/token as optional `.env.cloudflare` entries and
           still mention `wrangler login` / CI secrets as valid alternatives.
     - [ ] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001
- NFR-001 -> T-001

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - None.
- Rollback steps:
  - Restore previous README and `.env.cloudflare.example` text.
