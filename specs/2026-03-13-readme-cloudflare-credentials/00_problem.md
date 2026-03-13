---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - The root README and `.env.cloudflare.example` explain Cloudflare deployment, but they do not
    clearly state that `CLOUDFLARE_ACCOUNT_ID` and `CLOUDFLARE_API_TOKEN` may live in
    `.env.cloudflare`.
- Users or stakeholders:
  - maintainers deploying Payrune locally or from CI
  - AI agents reading the repo to infer the deployment contract
- Why now:
  - Deployment guidance should clearly state that Cloudflare credentials can be managed in the same
    Cloudflare deployment env file.

## Constraints (optional)

- Technical constraints:
  - Keep the change documentation-only.
  - Keep the guidance concise.

## Problem statement

- Current pain:
  - The current docs mention `wrangler login`, but they do not explicitly say that Cloudflare
    account id and API token can be stored in `.env.cloudflare`.

## Goals

- G1:
  - Clarify that `CLOUDFLARE_ACCOUNT_ID` and `CLOUDFLARE_API_TOKEN` can live in `.env.cloudflare`.
- G2:
  - Keep support for `wrangler login` and CI secrets without requiring them.

## Non-goals (out of scope)

- NG1:
  - Changing deployment scripts or runtime behavior.

## Assumptions

- A1:
  - The project will continue to support both interactive `wrangler login` and CI/non-interactive
    deployment.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - A reader can tell where Cloudflare credentials should be stored.
- Target:
  - README and `.env.cloudflare.example` both state that account/token may be stored in
    `.env.cloudflare`.
