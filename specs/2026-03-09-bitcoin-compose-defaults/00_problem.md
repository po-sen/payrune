---
doc: 00_problem
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

# Problem & Goals

## Context

- Background:
  - Bitcoin Compose defaults currently determine both newly issued receipt lifetime and poller API pressure in production-like setups.
- Users or stakeholders:
  - Operators running the Bitcoin `app`, `poller-mainnet`, and `poller-testnet4` services from `deployments/compose/`.
- Why now:
  - Production rollout is being prepared, and the desired profile is a lower-cost, smoother poller plus a shorter unpaid payment window.

## Constraints (optional)

- Technical constraints:
  - Only Compose defaults are changed in this spec.
  - Sticky paid semantics remain unchanged: once a receipt reaches `paid_unconfirmed`, it no longer expires via payment-window expiry.
- Timeline/cost constraints:
  - Prefer lower third-party explorer API usage and smoother per-tick bursts over fastest possible freshness.

## Problem statement

- Current pain:
  - `15s` rescheduling and `50`-row batches are too aggressive for public or hosted Esplora usage.
  - `7d` unpaid receipt expiry keeps newly issued addresses active too long.
  - The uncommitted rollout decision was split across two separate specs even though it is one Compose-default tuning pass.
- Evidence or examples:
  - Current Compose defaults poll a single active address every `15s`.
  - Current Compose defaults keep unpaid receipts alive for `168h`.
  - Current Compose defaults require only `1` confirmation on both Bitcoin networks.
  - Under the target `10m` cadence and a two-day sizing assumption, one address consumes about `720-864` HTTP calls (`7,200-8,640 CU`) over its lifecycle.
  - With `POLL_TICK_INTERVAL=5s` and `POLL_BATCH_SIZE=2`, one poller caps at about `2.592M` HTTP calls/month (`25.92M CU/month`) for clean addresses, which stays below Validation Cloud's `50M CU` free tier for a single saturated poller.
  - If both mainnet and testnet4 pollers saturate under the same Validation Cloud account, combined usage is about `51.84M CU/month`, or roughly `$0.92/month` of overage at `$0.50 / 1M CU` beyond the free tier.

## Goals

- G1:
  - Reduce default steady-state observer API usage by changing poller cadence to `10m` rescheduling with small batches.
- G2:
  - Make per-tick observer bursts flatter by capping the default batch size at `2`.
- G3:
  - Lower the default unpaid receipt window to `24h`.
- G4:
  - Keep mainnet and testnet4 Compose defaults aligned.
- G5:
  - Raise the default required confirmations to `2` on both Bitcoin networks.
- G6:
  - Record the operational sizing numbers directly in the merged spec.
- G7:
  - Record the Validation Cloud free-tier and overage implication for the default profile.

## Non-goals (out of scope)

- NG1:
  - Adding jitter or fairness algorithms to runtime scheduling.
- NG2:
  - Reintroducing confirmation grace-period expiry for `paid_unconfirmed` or `paid_unconfirmed_reverted`.
- NG3:
  - Changing runtime code defaults outside Compose.

## Assumptions

- A1:
  - Compose defaults are optimized for low-cost public/hosted Esplora usage, not fastest possible status freshness.
- A2:
  - A two-day address lifecycle is a sizing assumption used for API/CU estimation, not a hard runtime guarantee under current sticky paid rules.
- A3:
  - Addresses are clean enough that `/txs/chain` normally fits within one page.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Default per-address observer cadence.
- Target:
  - No more than one observation every `10m`.
- Metric:
  - Default per-tick burst size.
- Target:
  - No more than `2` addresses claimed per tick.
- Metric:
  - Default unpaid receipt window.
- Target:
  - `24h` on both Bitcoin networks.
- Metric:
  - Default required confirmations.
- Target:
  - `2` on both Bitcoin networks.
