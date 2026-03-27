---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Usecase orchestration:
  - transaction boundary、domain invocation、outbound port coordination、error mapping 這類 application 層本來就應該持有的流程。
- Scheme-specific derivation detail:
  - 像 `ethereum/create2` salt / relative reference 組裝這種與特定技術路徑綁定的細節。

## Out-of-scope behaviors

- OOS1: 不改變對外 API response schema 或 worker request/response envelope。
- OOS2: 不把 `PaymentAddressID`、health timestamp 這類既有輸出格式化全面搬家。

## Functional requirements

### FR-001 - Remove scheme-specific issuance derivation detail from allocate usecase

- Description: `AllocatePaymentAddressUseCase` 不得直接持有 `ethereum/create2` 的 relative reference / salt derivation 細節；它應只協調 reservation、derivation、tracking 與 idempotency flow。
- Acceptance criteria:
  - [ ] `internal/application/usecases/allocate_payment_address_use_case.go` 不再直接依賴 `EthereumCreate2SaltDeriver` 或 `chain == ethereum && scheme == create2` 這類技術分支。
  - [ ] allocation address derivation 經由單一 outbound port 完成，輸入足以支援既有 Bitcoin 與 Ethereum create2 path。
  - [ ] 既有 allocate usecase 測試持續通過，包含 Ethereum create2 與 derivation failure path。
- Notes: 這裡要移走的是技術 derivation 細節，不是 transaction orchestration。

### FR-002 - Move address preview capability rule into domain policy

- Description: address preview 是否支援，應由 address issuance policy 自己表達，而不是由 `GenerateAddressUseCase` 硬編碼 chain/scheme 特例。
- Acceptance criteria:
  - [ ] `internal/application/usecases/generate_address_use_case.go` 不再直接判斷 `ethereum/create2` preview unsupported。
  - [ ] domain entity / policy 能表達 preview support 與其失敗原因。
  - [ ] 既有 preview rejection 行為維持不變，controller 仍收到 `ErrAddressPreviewNotSupported`。
- Notes: usecase 仍可負責將 domain error 映射成 inbound error。

### FR-003 - Normalize poller scope before the usecase boundary

- Description: poller 的 `chain/network` filter 必須在 bootstrap / scheduler path 先正規化成 typed value，再交給 usecase；`RunReceiptPollingCycleUseCase` 不再 parse raw string。
- Acceptance criteria:
  - [ ] `internal/application/usecases/run_receipt_polling_cycle_use_case.go` 不再包含 raw `ParseChainID` / `ParseNetworkID` parsing helper。
  - [ ] `RunReceiptPollingCycleInput` 與 scheduler request path 使用 typed chain/network filter。
  - [ ] 現有 `poll chain is required when poll network is set` 驗證語意維持不變。
- Notes: usecase 可保留 typed filter consistency validation，但不應再負責 raw parsing。

### FR-004 - Keep boundary cleanup concrete and local

- Description: 這輪 refactor 必須用 concrete naming 與 local ownership 落地，不得引入過度抽象的新 framework。
- Acceptance criteria:
  - [ ] 不新增 `shared`, `common`, `framework`, `registry` 類型目錄或命名。
  - [ ] 新增的 outbound port / adapter 名稱直接描述目前責任，例如 issued address derivation，而不是假設未來會有一整個 generic platform。
  - [ ] bootstrap 與 adapter 的 call site 更新後仍維持可直接追蹤的 wiring。
- Notes: repo 偏好 concrete names over reusable-looking abstractions。

## Non-functional requirements

- Performance (NFR-001): refactor 不得新增外部 IO hop；地址 derivation 仍在本地 adapter/service 邏輯內完成。
- Availability/Reliability (NFR-002): `go test ./internal/application/...`、`go test ./internal/bootstrap/...`、`go test ./...` 必須持續通過。
- Security/Privacy (NFR-003): create2 derivation key handling 與 poller scope validation 語意不得放寬。
- Compliance (NFR-004): 不適用。
- Observability (NFR-005): 不改動現有 scheduler / controller 的 response 與 log 行為。
- Maintainability (NFR-006): `internal/application/usecases` 只保留 orchestration；scheme-specific technical detail 與 raw parsing 必須落到更適合的 owner。

## Dependencies and integrations

- External systems: 無新增。
- Internal services:
  - `internal/domain/entities`
  - `internal/application/ports/outbound`
  - `internal/adapters/outbound/blockchain`
  - `internal/bootstrap`
