---
doc: 00_problem
spec_date: 2026-03-08
slug: payment-address-status-api
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-04-policy-payment-address-allocation
  - 2026-03-05-blockchain-receipt-polling-service
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Payment Address Status API - Problem & Goals

## Context

- Background:
  - The create API already returns a stable `paymentAddressId`.
  - Receipt polling already persists the latest on-chain payment state in `payment_receipt_trackings`.
  - Merchant integrations currently rely on webhook delivery to learn payment progress after address issuance.
- Users or stakeholders:
  - Merchant backend teams that need to query the current payment state on demand.
  - Payrune backend maintainers who need a first-party read path that matches persisted receipt state.
- Why now:
  - The user explicitly needs a direct API to inspect current payment progress instead of depending only on webhook delivery.

## Constraints (optional)

- Technical constraints:
  - Reuse existing allocation and receipt-tracking data; do not add a new persistence model for this read path.
  - Keep the change within the current Go clean-architecture layout.
  - Preserve the existing `POST /v1/chains/{chain}/payment-addresses` contract.
- Timeline/cost constraints:
  - Prefer one explicit read endpoint over a broader query/list API.
- Compliance/security constraints:
  - No auth or merchant-identity model is introduced in this change.

## Problem statement

- Current pain:
  - After creating a payment address, clients have no API to fetch the latest payment state directly.
  - Webhook delivery is push-oriented and insufficient when the client wants pull-based status checks, manual investigation, or webhook recovery workflows.
- Evidence or examples:
  - Existing specs already asked whether a follow-up retrieval endpoint by `paymentAddressId` should be added.
  - The system already persists current receipt totals and status, so the missing piece is an API read path rather than a new tracking mechanism.

## Goals

- G1:
  - Provide a direct API to fetch current payment status by `paymentAddressId`.
- G2:
  - Return both issued address metadata and the latest persisted receipt-tracking state in one response.
- G3:
  - Make the API usable alongside webhook delivery rather than replacing it.
- G4:
  - Keep not-found and invalid-input behavior deterministic for clients.

## Non-goals (out of scope)

- NG1:
  - Listing or searching payments by arbitrary filters.
- NG2:
  - Replacing webhook delivery or changing webhook payload semantics.
- NG3:
  - Introducing merchant authentication or access-control changes.
- NG4:
  - Adding a new write path to refresh or mutate payment status from the API.

## Assumptions

- A1:
  - `paymentAddressId` is the primary lookup key for this read API because it is already returned by the create endpoint and called out in earlier specs as the likely retrieval key.
- A2:
  - Every issued payment address should have a corresponding receipt-tracking row once the allocation transaction commits.

## Open questions

- Q1:
  - None for this scope.

## Success metrics

- Metric:
  - Merchant clients can retrieve current payment status without waiting for a webhook.
- Target:
  - `GET` by a valid issued `paymentAddressId` returns the latest persisted status and totals in one response.
