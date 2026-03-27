---
doc: 00_problem
spec_date: 2026-03-27
slug: application-error-boundaries
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-26-allocate-usecase-decomposition
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: `internal/application/usecases` 目前已把主要 boundary cleanup 做完，但 usecase 本身仍直接 `errors.New(...)` 建立大量 application-facing error。使用者明確指出這類錯誤不應散在 usecase 內部，而應先在 application contract 層集中，讓 adapter 可以穩定共用與判斷。後續又進一步要求盤點 `internal/adapters` 裡那些其實應該升成 `outport.Err...` 的 technical sentinel，並把這套 ownership 規則寫進 `AGENTS.md` 變成 repo 契約。
- Users or stakeholders: 維護 usecase / adapter 邊界的人，以及要做 HTTP / scheduler error mapping 的 adapter 開發者。
- Why now: 這是最後一輪 usecase 邊界收尾；使用者希望明確分出 application error 與 adapter error。

## Constraints (optional)

- Technical constraints: 不改變既有 inbound/outbound flow semantics；只收斂 error 定義 ownership。
- Timeline/cost constraints: Quick mode；不做 API schema redesign。
- Compliance/security constraints: 不放寬既有 error mapping、idempotency、或 polling / dispatch 行為。

## Problem statement

- Current pain: usecase 直接 `errors.New(...)` 會讓 application error 分散在實作檔裡，adapter 難以穩定共用，也模糊了「哪種錯誤是 application contract、哪種是 outbound adapter technical signal」。
- Evidence or examples:
  - `AllocatePaymentAddressUseCase` 直接建立 configuration / consistency error。
  - `RunReceiptPollingCycleUseCase` 與 `RunReceiptWebhookDispatchCycleUseCase` 直接建立 validation / missing dependency error。
  - `CheckHealthUseCase`、`GenerateAddressUseCase`、`ListAddressPoliciesUseCase` 也各自建立 configuration error。
  - 目前只有部分 application-facing errors 集中在 `internal/application/ports/inbound/address_policy_use_cases.go`，其餘仍散落在 usecase 檔案。
  - `postgres` / `cloudflarepostgres` 的相同 persistence port 仍重複回傳同樣字串，例如 receipt tracking claim/save validation、notification outbox delivery-result validation、payment address idempotency validation。

## Goals

- G1: 將 usecase 對外回傳的 application error 集中到 application inbound contract。
- G2: 明確區分 inbound application error 與 outbound adapter error。
- G3: 讓 usecase 不再自己 ad-hoc 建立可共用的錯誤。
- G4: 保持既有 controller / usecase / adapter error mapping 行為不變。
- G5: 將多實作 adapter 共享的 outbound port contract error 提升為 `outport.Err...`。
- G6: 將 error ownership 寫入 `AGENTS.md`，讓後續重構有明確準則。

## Non-goals (out of scope)

- NG1: 不把 adapter 內部所有 technical error 都搬成 shared error catalog。
- NG2: 不改變 domain error ownership。
- NG3: 不改 outward HTTP status mapping 或 response body schema。

## Assumptions

- A1: 真正需要集中的是「application contract error」，不是每一個 adapter/private helper 的 technical error。
- A2: `inport` 適合承接 usecase 對 inbound adapter 暴露的 shared error；`outport` 則繼續承接 application 需要 branch 的 adapter error。

## Open questions

- Q1: 無；方向已足夠明確。
- Q2:

## Success metrics

- Metric: `internal/application/usecases` 不再直接建立 shared application error。
- Target: usecase source 移除 ad-hoc `errors.New(...)`，測試與 spec lint 維持綠燈。
