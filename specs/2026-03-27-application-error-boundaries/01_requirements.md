---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Application error:
  - Error defined by the application layer and intentionally returned by usecases for inbound adapters to inspect or map.
- Adapter error:
  - Error returned by outbound adapters; only centralize in `outport` when the application must branch on it.

## Out-of-scope behaviors

- OOS1: 不抽離 purely local adapter/private helper error。
- OOS2: 不把 domain entity / policy validation error 搬離 domain。

## Functional requirements

### FR-001 - Centralize usecase-returned errors in inbound application contracts

- Description: `internal/application/usecases` 對外回傳的 shared application error 必須集中在 `internal/application/ports/inbound`，usecase 不再自己 ad-hoc `errors.New(...)` 建立這類錯誤。
- Acceptance criteria:
  - [ ] `AllocatePaymentAddressUseCase`、`GenerateAddressUseCase`、`ListAddressPoliciesUseCase`、`GetPaymentAddressStatusUseCase`、`CheckHealthUseCase`、`RunReceiptPollingCycleUseCase`、`RunReceiptWebhookDispatchCycleUseCase` 都不再直接建立 shared application error。
  - [ ] 缺失 dependency、application validation、application consistency 類錯誤改為回傳集中定義的 inbound error。
  - [ ] adapter / tests 可以透過 shared inbound error 做穩定比對。
- Notes: 不是所有 error 都要變 shared；只處理 usecase 對外可見、可共用的 application error。

### FR-002 - Preserve outbound adapter error ownership in outport

- Description: outbound adapter 回給 usecase、且 application 需要 branch 的 error 仍由 `internal/application/ports/outbound` 定義，不搬到 inbound。
- Acceptance criteria:
  - [ ] `ErrPaymentAddressIdempotencyKeyExists`、`ErrAddressIndexExhausted`、`ErrPaymentAddressStatusIncomplete` 仍留在 `outport`。
  - [ ] usecase 仍用 `errors.Is` 對這些 outbound error 做 branch。
  - [ ] 不新增 adapter-only error 到 inbound contract。
- Notes: 這類錯誤是 application 與 outbound adapter 的 port contract，不是 inbound contract。

### FR-003 - Promote shared outbound port contract errors to outport

- Description: 當兩個以上 adapter implementation 為同一個 outbound port 回傳同一種 contract/validation/state error 時，該 error 應集中在 `internal/application/ports/outbound`，避免每個 adapter 重複字串。
- Acceptance criteria:
  - [ ] `payment_address_idempotency_store` 的 shared contract error 升成 `outport.Err...`。
  - [ ] `payment_receipt_tracking_store` 的 shared contract error 升成 `outport.Err...`。
  - [ ] `payment_receipt_status_notification_outbox` 的 shared contract error 升成 `outport.Err...`。
  - [ ] `payment_address_allocation_store` 的 shared contract error 升成 `outport.Err...`。
  - [ ] `payment_address_status_finder` 的 shared contract error 升成 `outport.Err...`。
  - [ ] 只處理多實作共享、且屬於 port contract 的 error；不把 adapter private error 全部抽上去。
- Notes: 這是 outbound port contract cleanup，不是 adapter-internal cleanup。

### FR-004 - Keep controller and usecase error mapping behavior stable

- Description: 集中錯誤定義後，不得改變既有 controller / usecase 的 error mapping 語意。
- Acceptance criteria:
  - [ ] 現有 HTTP controller 對 `inport.Err...` 的 status mapping 維持不變。
  - [ ] usecase 對 domain error 和 outport error 的轉譯維持不變。
  - [ ] usecase validation/configuration 測試改成比對 shared error，而不是散落的字串。
- Notes: 這輪是 ownership cleanup，不是 error semantics redesign。

### FR-005 - Document error ownership in repo architecture rules

- Description: `AGENTS.md` 必須明確定義 domain / inbound application / outbound port / adapter-private error 的 ownership 與使用規則，避免未來重新把 shared error 散回 usecase 或 adapter。
- Acceptance criteria:
  - [ ] `AGENTS.md` 明確說明 usecase-returned shared error 應定義於 `internal/application/ports/inbound`。
  - [ ] `AGENTS.md` 明確說明多實作 outbound port 的 shared contract error 何時應升成 `internal/application/ports/outbound`。
  - [ ] `AGENTS.md` 明確說明 adapter-private technical error 不應無差別提升成 shared error catalog。
- Notes: 這是 repo-local architecture contract，目標是讓之後的 refactor 有一致判斷標準。

## Non-functional requirements

- Performance (NFR-001): 不新增額外 IO 或 transaction。
- Availability/Reliability (NFR-002): `go test ./internal/application/usecases`、`go test ./internal/adapters/inbound/http/controllers`、`go test ./...` 必須通過。
- Security/Privacy (NFR-003): 不改變現有錯誤處理的安全邊界；controller 不得因此暴露新的 internal detail。
- Compliance (NFR-004):
- Observability (NFR-005): 不改動既有 log/response mapping 行為。
- Maintainability (NFR-006): usecase 檔案不再散落 shared error 字串；application vs adapter error ownership 必須可一眼辨識，且多實作 outbound port 不再複製相同 contract error 字串。

## Dependencies and integrations

- External systems: 無新增。
- Internal services:
  - `internal/application/ports/inbound`
  - `internal/application/ports/outbound`
  - `internal/application/usecases`
  - `internal/adapters/inbound/http/controllers`
