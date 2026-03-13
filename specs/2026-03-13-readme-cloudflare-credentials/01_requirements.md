---
doc: 01_requirements
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

# Requirements

## Out-of-scope behaviors

- OOS1:
  - Changing Cloudflare deploy scripts or adding new env handling.

## Functional requirements

### FR-001 - README must document Cloudflare credential placement

- Description:
  - The root README must explain where `CLOUDFLARE_ACCOUNT_ID` and `CLOUDFLARE_API_TOKEN` should
    be stored.
- Acceptance criteria:
  - [ ] README says local interactive deploy can use `wrangler login`.
  - [ ] README says `.env.cloudflare` may include `CLOUDFLARE_ACCOUNT_ID` and
        `CLOUDFLARE_API_TOKEN`.
  - [ ] README says CI/non-interactive deploy may also source those values from CI secrets.

### FR-002 - `.env.cloudflare.example` must include optional Cloudflare credentials

- Description:
  - The env template must expose optional Cloudflare account/token entries for deploy flows.
- Acceptance criteria:
  - [ ] `.env.cloudflare.example` contains `CLOUDFLARE_ACCOUNT_ID=`.
  - [ ] `.env.cloudflare.example` contains `CLOUDFLARE_API_TOKEN=`.
  - [ ] The comment explains these are optional for non-interactive deploys.

## Non-functional requirements

- Clarity (NFR-001):
  - The new guidance must be obvious in one quick read.

## Dependencies and integrations

- Internal sources of truth:
  - `README.md`
  - `.env.cloudflare.example`
