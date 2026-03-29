---
doc: 00_problem
spec_date: 2026-03-30
slug: compose-eth-sepolia-split
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
  - 2026-03-28-eth-create2-config-update
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background: `deployments/compose/compose.yaml` currently exposes Ethereum mainnet and Sepolia API envs together, while `deployments/compose/compose.test.yaml` already owns the Sepolia poller and the Bitcoin test-only split.
- Users or stakeholders: Operators using the local Compose test stack and reviewers maintaining deployment config readability.
- Why now: The user wants Sepolia-specific API wiring to live only in the test overlay, and wants Ethereum mainnet issuance in test mode to be explicitly disabled the same way Bitcoin mainnet issuance is disabled there.

## Constraints (optional)

- Technical constraints: Only change Compose layering; do not rename env vars or change Go bootstrap contracts.
- Timeline/cost constraints: Quick-mode config refactor only.
- Compliance/security constraints: Test overlay must not accidentally inherit Ethereum mainnet CREATE2 issuance credentials.

## Problem statement

- Current pain: Base Compose mixes Sepolia API envs into the default stack, and the test overlay does not explicitly clear Ethereum mainnet CREATE2 issuance vars the way it clears Bitcoin mainnet XPUBs.
- Evidence or examples:
  - `deployments/compose/compose.yaml`
  - `deployments/compose/compose.test.yaml`
  - `Makefile` loads both files together with `deployments/compose/compose.test.env`

## Goals

- G1: Keep Sepolia API env wiring in `deployments/compose/compose.test.yaml` instead of the base `compose.yaml`.
- G2: Make the test overlay explicitly blank Ethereum mainnet CREATE2 issuance vars so test stacks default to Sepolia-only ETH issuance behavior.

## Non-goals (out of scope)

- NG1: No changes to `.env.cloudflare` or `deployments/compose/compose.test.env` values.
- NG2: No changes to Go bootstrap logic, CREATE2 metadata, or poller runtime behavior.

## Assumptions

- A1: “Set MAINNET empty like Bitcoin” refers to the Ethereum mainnet issuance-enabling API vars, analogous to Bitcoin mainnet XPUBs, not to receipt-term defaults.
- A2: Sepolia RPC poller wiring remains in `compose.test.yaml` and does not need structural changes beyond the API env split.

## Open questions

- Q1: None.
- Q2: None.

## Success metrics

- Metric: Whether base/test Compose responsibilities are clearer and merged test config resolves to Sepolia-only ETH issuance by default.
- Target: `compose.yaml` no longer contains `ETHEREUM_SEPOLIA_*` API env entries, `compose.test.yaml` defines those Sepolia API envs, and `compose.test.yaml` explicitly blanks `ETHEREUM_MAINNET_CREATE2_*`.
