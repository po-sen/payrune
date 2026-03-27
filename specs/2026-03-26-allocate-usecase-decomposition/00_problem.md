---
doc: 00_problem
spec_date: 2026-03-26
slug: allocate-usecase-decomposition
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-25-usecase-boundary-cleanup
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: 這輪在 `internal/application/usecases` 持續做 readability cleanup。`AllocatePaymentAddressUseCase` 先前已把 `create2` 技術細節與 raw parsing 移出 application layer，但使用者明確指出目前的 collaborator 拆分只是把複雜度分散到多個檔案，沒有真正降低理解成本；同樣標準也延伸到 `GenerateAddressUseCase`、`RunReceiptPollingCycleUseCase`、以及最後一輪整體 usecase 架構審查。
- Users or stakeholders: 維護 `internal/application/usecases` 的開發者，以及之後要繼續整理 allocation flow 的人。
- Why now: 使用者明確要求「優化程式碼到看得懂」，且不希望為了責任邊界而持續增加 spec 與 helper 碎片。

## Constraints (optional)

- Technical constraints: 保持單一 inbound usecase；不把 transaction boundary 拆散；不引入 generic workflow framework。
- Timeline/cost constraints: 本輪只整理 `internal/application/usecases` 內的可讀性問題，不改外部行為。
- Compliance/security constraints: 不放寬 idempotency、receipt tracking、derivation failure persistence 的既有語意。

## Problem statement

- Current pain: 目前版本雖然邊界正確，但就算回到單檔，若 top-level helper 過多或重複分支過密，閱讀時仍會在多個小段落間來回切換。
- Current pain: 使用者的核心需求是「程式碼順著讀得懂」，不是「責任被拆成更多檔案」或「每個小檢查各一個 helper」。
- Evidence or examples:
  - `Execute` 同時看得到 replay lookup、issuance plan、transaction execution、error mapping。
  - 若 replay、issuance、response 都拆成太多 top-level helper，即使同檔也仍然難順讀。
  - 目前 `Execute` 仍需跳到獨立的 replay helper 才能看懂 duplicate-idempotency path。
  - `buildAllocatePaymentAddressResponse` 這類 generic helper name 沒有幫助理解，反而讓 response mapping 的責任變模糊。
  - `issueAllocation` 仍然把 idempotency、reservation、derivation failure、tracking create 全部擠在同一段 transaction body 裡，閱讀密度過高。
  - `Execute` 裡的 `issueAllocation` error branch 同時做 fallback replay 與 error mapping，主流程會被一大段分支打斷。
  - 若只是單次使用的薄 error helper，可能只是把分支搬走，未必真的提升可讀性。
  - `GenerateAddressUseCase` 的 preview validation 先前有 single-use helper，實際上直接 inline 回主流程更清楚。
  - `RunReceiptPollingCycleUseCase` 在主 loop 內重複 4 次 polling-error save path，且重複 2 次 save+enqueue path，讓 `Execute` 過於擁擠。
  - `RunReceiptWebhookDispatchCycleUseCase` 的單筆 dispatch 流程同時混了 notifier call、delivery-result resolve/mark sent、save result 與 counter branching，且 `SaveDeliveryResult` transaction 重複兩次。
  - 最後一輪審查仍發現 `AllocatePaymentAddressUseCase` 在 usecase 內做 `TrimSpace`，這屬於 inbound transport normalization。
  - `RunReceiptPollingCycleUseCase` 與 `RunReceiptWebhookDispatchCycleUseCase` 仍在 usecase 內注入 runtime default，這屬於 bootstrap / scheduler ownership，不是 application orchestration。

## Goals

- G1: 保留各自單一 public usecase，並讓主要流程能在少量檔案中線性讀懂。
- G2: 減少不必要的 top-level helper 數量，避免理解 flow 時來回跳轉。
- G3: 讓 allocation replay/fallback path 更貼近主流程，降低視線跳轉。
- G4: 讓 generate 與 receipt polling 也符合同樣的 readability 標準。
- G5: 讓 receipt webhook dispatch 也符合同樣的 readability 標準。
- G6: 不改變對外 contract、transaction semantics、或現有測試行為。
- G7: 移除 usecase 內殘留的 transport normalization 與 runtime default ownership。

## Non-goals (out of scope)

- NG1: 不新增第二個 public inbound usecase。
- NG2: 不改動 idempotency / allocation / receipt tracking 的持久化模型與 adapter contract。
- NG3: 不建立新的 shared readability framework 或 utility package。

## Assumptions

- A1: 目前複雜度的主要來源不是 business rule 太多，而是 flow 被切成過多小塊，或重複錯誤/save 分支讓主流程擁擠。
- A2: 使用者偏好少檔案、少 helper、線性流程，這優先於「每個責任各一個小檔」的整齊感。

## Open questions

- Q1: 無；範圍已足夠明確。

## Success metrics

- Metric: 主要 usecase flow 能在少量檔案內順著讀完，且重複錯誤/save 分支被收斂。
- Target: 清掉不必要的 helper/spec 碎片並保持 `go test ./...` 綠燈。
