---
doc: 00_problem
spec_date: 2026-03-30
slug: swagger-non-mainnet-defaults
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-03-swagger-ui-container-api-testing
  - 2026-03-20-create2-eth-payment-receiving
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background: `deployments/swagger/openapi.yaml` currently uses mainnet-oriented examples for payment allocation and status flows, and the allocate controller only accepts string-or-omitted `customerReference`.
- Users or stakeholders: Developers using Swagger UI for local API testing and reviewers trying to avoid accidental mainnet-looking examples in docs.
- Why now: The user wants Swagger defaults to steer local testing toward non-mainnet values and wants `customerReference` to default to `null`.

## Constraints (optional)

- Technical constraints: Keep the existing public API shape; only adjust OpenAPI examples/defaults and make request decoding tolerant of `customerReference: null`.
- Timeline/cost constraints: Quick mode, scoped to Swagger contract polish and one small inbound-adapter compatibility change.
- Compliance/security constraints: Swagger examples should not imply mainnet usage by default in local testing docs.

## Problem statement

- Current pain: Swagger examples point users toward `bitcoin-mainnet-*` and `ethereum-mainnet-*` payloads, and `customerReference: null` would currently fail request decoding despite being a sensible local-testing default.
- Evidence or examples:
  - `deployments/swagger/openapi.yaml`
  - `internal/adapters/inbound/http/controllers/allocate_payment_address_controller.go`

## Goals

- G1: Change Swagger payment-allocation defaults/examples to non-mainnet values.
- G2: Set Swagger defaults to Bitcoin `2000` satoshi, Ethereum `0.0001 ETH` (`100000000000000` wei), and `customerReference: null`.
- G3: Make the allocate controller accept `customerReference: null` as an empty optional value.

## Non-goals (out of scope)

- NG1: No business-rule change to amount validation or address issuance.
- NG2: No change to deployment config or non-Swagger UI behavior beyond null-tolerant request decoding for `customerReference`.

## Assumptions

- A1: “Swagger defaults” refers to the examples/default values users see in local Swagger UI for request/response schemas and examples, especially payment-address flows.
- A2: `customerReference: null` should be treated the same as omitting the field.

## Open questions

- Q1: None.
- Q2: None.

## Success metrics

- Metric: Whether Swagger now presents non-mainnet defaults and whether `customerReference: null` works end-to-end.
- Target: OpenAPI examples/defaults for payment allocation/status use testnet4 or sepolia values, BTC default amount is `2000`, ETH default amount is `100000000000000` wei, `customerReference` is documented as nullable/default `null`, and controller tests pass with a null payload.
