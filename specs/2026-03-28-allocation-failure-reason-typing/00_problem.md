---
doc: 00_problem
spec_date: 2026-03-28
slug: allocation-failure-reason-typing
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
  - 2026-03-27-domain-error-contracts
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background: receipt tracking 與 webhook delivery 已經改成 domain-owned typed failure reason，但 allocation derivation failure 仍然沿用自由字串。
- Users or stakeholders: 維護 allocation issuance flow、persistence adapter、以及 status/read model 的開發者與 reviewer。
- Why now: 這是目前唯一還明顯沒跟 receipt 那條線收齊的 failure reason path。

## Constraints (optional)

- Technical constraints: 不做 DB migration；既有 `failure_reason` 欄位維持字串存放。
- Timeline/cost constraints: Quick mode，只收 allocation derivation failure 這條 representation。
- Compliance/security constraints: 不把 lower-level derive error wording 繼續寫進 domain/process state。

## Problem statement

- Current pain: allocation derivation failure 仍由 usecase 直接把 `deriveErr.Error()` 寫入 entity/store，導致 adapter/private wording 進入 core state。
- Evidence or examples:
  - `internal/application/usecases/allocate_payment_address_use_case.go`
  - `internal/domain/entities/payment_address_allocation.go`
  - `internal/adapters/outbound/persistence/postgres/payment_address_allocation_store.go`
  - `internal/adapters/outbound/persistence/cloudflarepostgres/payment_address_allocation_store.go`

## Goals

- G1: 把 allocation derivation failure 收成 domain typed reason code，而不是自由字串。
- G2: 讓 allocate usecase 只做 `lower-level error -> domain reason code` mapping。
- G3: 讓 persistence adapter serialize/parse typed reason，維持 schema 不變。

## Non-goals (out of scope)

- NG1: 不處理 `allocate_payment_address_use_case.go` 的整體可讀性重構。
- NG2: 不調整 outward API contract 或新增 migration。

## Assumptions

- A1: allocation derivation failure reason 屬於 persisted process state，而不是 logging/debug channel。
- A2: 舊資料可能已存在 legacy raw text，需要在 read/write path 安全兼容。

## Open questions

- Q1: 無
- Q2:

## Success metrics

- Metric: allocation derivation failure path 是否改成 domain typed reason
- Target: domain/application/adapter 全線不再依賴 raw derive error text 作為 persisted failure reason
