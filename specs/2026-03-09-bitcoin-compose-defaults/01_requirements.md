---
doc: 01_requirements
spec_date: 2026-03-09
slug: bitcoin-compose-defaults
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-09-poller-interval-separation
  - 2026-03-09-sticky-paid-unconfirmed-status
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Tick interval:
  - How often the worker wakes up to claim due rows.
- Reschedule interval:
  - How long a processed receipt waits before becoming due again.
- Payment window:
  - The issued receipt lifetime before an unpaid receipt can become `failed_expired`.

## Out-of-scope behaviors

- OOS1:
  - Runtime jitter or fairness algorithms.
- OOS2:
  - Any new expiry rule for `paid_unconfirmed` or `paid_unconfirmed_reverted`.
- OOS3:
  - Runtime changes to `defaultBitcoinReceiptExpiresAfter`.

## Functional requirements

### FR-001 - Lower default poller API pressure

- Description:
  - Compose defaults must favor lower Bitcoin observer API usage by lengthening per-address rescheduling.
- Acceptance criteria:
  - [ ] `poller-mainnet` uses `POLL_RESCHEDULE_INTERVAL=10m` by default.
  - [ ] `poller-testnet4` uses `POLL_RESCHEDULE_INTERVAL=10m` by default.
- Notes:
  - This intentionally trades freshness for lower explorer API consumption.

### FR-002 - Flatten default poller bursts

- Description:
  - Compose defaults must smooth per-tick load by using a small batch and finer worker tick.
- Acceptance criteria:
  - [ ] `poller-mainnet` uses `POLL_TICK_INTERVAL=5s`, `POLL_BATCH_SIZE=2`, and `POLL_CLAIM_TTL=30s` by default.
  - [ ] `poller-testnet4` uses `POLL_TICK_INTERVAL=5s`, `POLL_BATCH_SIZE=2`, and `POLL_CLAIM_TTL=30s` by default.
- Notes:
  - The default profile is smoothing-oriented, not throughput-maximizing.

### FR-003 - Raise Compose default required confirmations

- Description:
  - Bitcoin Compose defaults must require `2` confirmations by default on both networks.
- Acceptance criteria:
  - [ ] `compose.bitcoin.mainnet.yaml` uses `BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS=${BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS:-2}`.
  - [ ] `compose.bitcoin.testnet4.yaml` uses `BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS=${BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS:-2}`.
- Notes:
  - This is the default confirmation threshold for production-like rollout profiles.

### FR-004 - Lower Compose unpaid receipt expiry

- Description:
  - Bitcoin Compose defaults must set the initial unpaid receipt lifetime to `24h`.
- Acceptance criteria:
  - [ ] `compose.bitcoin.mainnet.yaml` uses `BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER=${BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER:-24h}`.
  - [ ] `compose.bitcoin.testnet4.yaml` uses `BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER=${BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER:-24h}`.
- Notes:
  - This only affects the initial unpaid payment window for newly issued receipts.

### FR-005 - Preserve sticky paid semantics in documentation

- Description:
  - The merged spec must not imply a hard two-day lifecycle when runtime behavior does not enforce it.
- Acceptance criteria:
  - [ ] The spec explicitly states that `paid_unconfirmed` and `paid_unconfirmed_reverted` still do not expire via the payment-window rule.
- Notes:
  - This keeps deployment expectations aligned with actual runtime behavior.

### FR-006 - Record sizing results

- Description:
  - The merged spec must include the operational sizing numbers and Validation Cloud cost implications used for rollout planning.
- Acceptance criteria:
  - [ ] The spec records that one address costs about `720-864` HTTP calls over a two-day sizing assumption.
  - [ ] The spec records that the same address costs about `7,200-8,640 CU` over that same assumption.
  - [ ] The spec records that one saturated poller tops out around `2.592M` HTTP calls/month and `25.92M CU/month` for clean addresses.
  - [ ] The spec records that two saturated pollers under one Validation Cloud account consume about `51.84M CU/month` combined.
  - [ ] The spec records that the combined overage above the `50M CU/month` free tier is about `$0.92/month` at `$0.50 / 1M CU`.
- Notes:
  - These numbers are informational rollout data, not runtime-enforced limits.

## Non-functional requirements

- Performance (NFR-001):
  - Default poller cadence must reduce per-address polling frequency from `15s` to `10m`.
- Availability/Reliability (NFR-002):
  - Default `POLL_CLAIM_TTL` must remain greater than the tick interval to avoid rapid duplicate claiming under normal conditions.
- Security/Privacy (NFR-003):
  - No new secrets or endpoints may be introduced.
- Compliance (NFR-004):
  - Not applicable.
- Observability (NFR-005):
  - Existing poller logs and metrics remain unchanged.
- Maintainability (NFR-006):
  - Compose-default tuning and its sizing data must live in one merged spec.

## Dependencies and integrations

- External systems:
  - Public or hosted Esplora-compatible APIs.
  - Validation Cloud CU-based billing assumptions for rollout sizing.
- Internal services:
  - Existing poller env parsing and scheduling flow from `poller-interval-separation`.
  - Sticky paid receipt semantics from `sticky-paid-unconfirmed-status`.
