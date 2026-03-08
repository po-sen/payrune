---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - The repo already has a fixed-endpoint webhook dispatcher and a fake HTTPS receiver used by the local test environment.
- Users or stakeholders:
  - The project owner and maintainers who want to inspect webhook verification behavior from the receiver side.
- Why now:
  - The current fake receiver only logs that a request arrived; it does not show how the receiver should verify the webhook signature and payload.

## Constraints (optional)

- Technical constraints:
  - Keep the production webhook contract unchanged.
  - Limit changes to the fake receiver, its tests, and the test compose wiring.
- Compliance/security constraints:
  - Use the same HMAC-SHA256 verification approach as the production webhook sender.

## Problem statement

- Current pain:
  - The fake receiver does not demonstrate the correct signature verification flow, so local webhook debugging lacks a trustworthy receiver-side reference.
- Evidence or examples:
  - `cmd/fake_webhook_receiver/main.go` currently logs only method and path.
  - The local test receiver does not validate `X-Payrune-Signature-256` against the shared secret.
  - Even after adding verification, the fake receiver logs still do not show the full received request content, which makes local debugging less direct than it should be.

## Goals

- G1:
  - Make the fake receiver verify the incoming webhook signature using the same method as the sender.
- G2:
  - Keep receiver-side logs focused on the full incoming request so local debugging is direct and uncluttered.
- G3:
  - Wire the fake receiver in `compose.test.yaml` so it receives the webhook secret in local test runs.
- G4:
  - Include the full received headers and raw body in the fake receiver logs so local debugging can inspect the exact webhook input.
- G5:
  - Format the single request log into readable sections so developers can scan the incoming webhook directly from `docker logs`.

## Non-goals (out of scope)

- NG1:
  - Do not change the production webhook payload schema or header contract.
- NG2:
  - Do not add replay protection or new webhook protocol features in this change.

## Assumptions

- A1:
  - The fake receiver is only used in local or test environments.
- A2:
  - `PAYMENT_RECEIPT_WEBHOOK_SECRET` already exists in `compose.test.env` and can be reused for the receiver.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Local fake receiver logs show one readable request block with full received headers and raw body for each webhook request.
- Target:
  - A developer can inspect the exact incoming webhook from the receiver logs during a local compose test run.
