---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - The repo has deployment docs and worker-specific READMEs, but no single top-level product
    README.
- Users or stakeholders:
  - humans deploying or integrating Payrune
  - AI agents that need enough contract detail to wire API and webhook clients correctly
- Why now:
  - The project needs one concise source that explains what the product is, how to deploy it, which
    parameters matter, and how to integrate with its API and webhook contract.

## Constraints (optional)

- Technical constraints:
  - Keep the document concise.
  - Use existing code and specs as the source of truth.
  - Do not change runtime behavior while writing docs.

## Problem statement

- Current pain:
  - The entrypoint for understanding Payrune is fragmented across compose, Cloudflare worker
    READMEs, OpenAPI, and code.
- Evidence or examples:
  - There is no `README.md` at repo root.

## Goals

- G1:
  - Add one top-level README that explains the product and deployment model in one pass.
- G2:
  - Include the minimum API and webhook contract details needed for direct integration.
- G3:
  - Include the main runtime parameters and the shortest local and Cloudflare deployment flows.

## Non-goals (out of scope)

- NG1:
  - Redesigning runtime config or deployment flows.
- NG2:
  - Replacing OpenAPI or worker-specific README documents.

## Assumptions

- A1:
  - A concise README with concrete examples is more useful than a long architectural guide.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - A reader can identify the product, its public API surface, webhook contract, and deployment
    flow from one file.
- Target:
  - Root `README.md` contains product summary, main parameters, deploy commands, and integration
    examples.
