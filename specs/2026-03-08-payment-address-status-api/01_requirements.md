---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Payment address status:
  - The latest persisted allocation and receipt-tracking view for one issued `paymentAddressId`.

## Out-of-scope behaviors

- OOS1:
  - Querying by blockchain address, `customerReference`, or `Idempotency-Key`.
- OOS2:
  - Returning webhook delivery attempts or notification history.

## Functional requirements

### FR-001 - Expose payment status read endpoint

- Description:
  - The HTTP API must expose a read endpoint for one issued payment address using chain plus `paymentAddressId`.
- Acceptance criteria:
  - [ ] A `GET /v1/chains/{chain}/payment-addresses/{paymentAddressId}` route exists.
  - [ ] The route accepts only positive integer `paymentAddressId` values.
  - [ ] Unsupported chains return `404`.
  - [ ] Non-`GET` methods return `405`.
- Notes:
  - The route is read-only and separate from the existing create endpoint.

### FR-002 - Return current persisted payment state

- Description:
  - A successful response must return the latest persisted state for the payment address in one payload.
- Acceptance criteria:
  - [ ] `200 OK` includes issued address metadata: `paymentAddressId`, `addressPolicyId`, `chain`, `network`, `scheme`, `minorUnit`, `decimals`, `address`, optional `customerReference`, and `expectedAmountMinor`.
  - [ ] `200 OK` includes current payment state fields sourced from receipt tracking: `paymentStatus`, `observedTotalMinor`, `confirmedTotalMinor`, `unconfirmedTotalMinor`, `conflictTotalMinor`, `requiredConfirmations`, `lastObservedBlockHeight`, `issuedAt`, `firstObservedAt`, `paidAt`, `confirmedAt`, `expiresAt`, and optional `lastError`.
  - [ ] Response timestamps use RFC3339 JSON timestamps.
- Notes:
  - The API returns the latest persisted view; it does not trigger a fresh blockchain poll.

### FR-003 - Return deterministic not-found behavior

- Description:
  - The API must return a deterministic client error when the payment address cannot be read.
- Acceptance criteria:
  - [ ] Unknown `paymentAddressId` returns `404`.
  - [ ] A `paymentAddressId` that is not in `issued` state returns `404`.
  - [ ] If the system cannot build a complete status view because the issued allocation exists but receipt tracking is missing, the API returns `500`.
- Notes:
  - Missing receipt tracking after issuance is treated as an internal invariant failure, not a client error.

## Non-functional requirements

- Performance (NFR-001):
  - The read path should complete with one bounded DB query path and no external blockchain calls.
- Availability/Reliability (NFR-002):
  - The endpoint must remain usable even if webhook delivery is delayed or unavailable because it reads persisted state directly.
- Security/Privacy (NFR-003):
  - The response must not expose xpubs, derivation paths, or internal notification-delivery state.
- Compliance (NFR-004):
  - None beyond existing API handling.
- Observability (NFR-005):
  - Existing HTTP request logging and standard controller error mapping are sufficient; no new metrics are required for this small read endpoint.
- Maintainability (NFR-006):
  - The read path should use a dedicated query-style outbound port or finder rather than expanding write-oriented stores with transport-specific response shaping.

## Dependencies and integrations

- External systems:
  - None.
- Internal services:
  - `address_policy_allocations`
  - `payment_receipt_trackings`
