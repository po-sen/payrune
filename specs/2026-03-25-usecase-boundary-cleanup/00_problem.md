---
doc: 00_problem
spec_date: 2026-03-25
slug: usecase-boundary-cleanup
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-24-architecture-conformance-refactor
  - 2026-03-25-bootstrap-dedup-refactor
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: `internal/application/usecases` 大致維持了 orchestration 責任，但仍殘留幾段不屬於業務流程的程式碼，包括 scheme-specific address derivation 細節、poller raw scope parsing，以及少量輸出格式化。
- Users or stakeholders: 維護 `internal/application`、`internal/adapters`、`internal/bootstrap` 邊界的開發者。
- Why now: 使用者明確要求檢查 usecase purity，並直接把不屬於業務流程的程式碼搬離 `internal/application/usecases`。

## Constraints (optional)

- Technical constraints: 不引入 generic framework；保持目前對外 HTTP / scheduler contract 不變；避免把 business rule 倒回 bootstrap 或 adapter。
- Timeline/cost constraints: 這輪要一次完成高訊號 boundary cleanup，不留下半套遷移。
- Compliance/security constraints: 不放寬現有 validation，也不改動 receipt polling / dispatch 的既有安全邏輯。

## Problem statement

- Current pain: `allocate_payment_address_use_case.go` 直接知道 `ethereum/create2` 要如何組 allocation-scoped reference，讓 application layer 持有技術 derivation 細節。
- Current pain: `generate_address_use_case.go` 直接硬編碼 `ethereum/create2` preview 不支援，讓 usecase 持有 policy capability rule。
- Current pain: `run_receipt_polling_cycle_use_case.go` 仍在 parse raw `chain/network` filter，而這些 raw 值其實應該由 bootstrap / inbound adapter 先正規化。
- Evidence or examples:
  - `allocate_payment_address_use_case.go` 直接呼叫 `EthereumCreate2SaltDeriver` 並組 `DeriveEthereumCreate2SaltInput`。
  - `generate_address_use_case.go` 直接檢查 `chain == ethereum && scheme == create2`。
  - `run_receipt_polling_cycle_use_case.go` 內有 `resolveReceiptPollingScope(rawChain, rawNetwork)`。

## Goals

- G1: 將 allocation address derivation 的技術細節移出 usecase，讓 `allocate` usecase 只負責 transaction / idempotency / persistence orchestration。
- G2: 將 address preview capability 判斷收回 domain policy，讓 `generate` usecase 不再硬編碼 chain/scheme 特例。
- G3: 將 poller scope 於 bootstrap / scheduler path 先正規化為 typed value，再交給 usecase。

## Non-goals (out of scope)

- NG1: 不改變 HTTP API 或 worker JSON contract。
- NG2: 不全面重寫 DTO shaping 或 response serialization；像 `PaymentAddressID` 字串化、health timestamp formatting 不在本輪處理。

## Assumptions

- A1: 目前真正高訊號的 boundary 問題是 scheme-specific derivation 與 raw parsing，不是單純 DTO mapping。
- A2: 以 concrete outbound port 封裝 issued-address derivation，比繼續在 usecase 中保存 `create2` 技術分支更符合 repo 風格。

## Open questions

- Q1: 無；目前範圍已足夠明確可直接實作。

## Success metrics

- Metric: `internal/application/usecases` 不再直接 parse raw poller scope，也不再直接知道 `ethereum/create2` derivation 細節。
- Target: 相關 usecase 改以 domain capability / outbound port 協作完成，且 `go test ./...` 持續通過。
