---
doc: 01_requirements
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

# Requirements

## Out-of-scope behaviors

- OOS1: 不將 allocation flow 拆成多個 public inbound usecase。
- OOS2: 不改變對外 response schema 或 repository/store port 的介面形狀。
- OOS3: 不建立新的 shared readability framework 或 generic utility package。

## Functional requirements

### FR-001 - Keep a single inbound allocate usecase

- Description: `AllocatePaymentAddressUseCase` 仍必須是單一 public usecase，controller / bootstrap 不得開始自行編排多個 allocation-related usecase。
- Acceptance criteria:
  - [ ] `NewAllocatePaymentAddressUseCase(...)` 仍是唯一 public constructor。
  - [ ] controller / bootstrap call site 不新增第二個 allocation inbound usecase。
  - [ ] transaction 與 error semantics 維持既有行為。
- Notes: 這輪整理的是單一 usecase 的可讀性，不是 public usecase API 的切分。

### FR-002 - Keep replay handling readable and local

- Description: idempotency replay lookup 與 consistency 驗證需要清楚，但不必為了責任切分而強制拆成獨立 collaborator file；優先目標是可順讀。
- Acceptance criteria:
  - [ ] replay handling 的程式碼閱讀路徑比目前版本更短。
  - [ ] 不需要跨多個小檔案才看得懂 replay path。
  - [ ] replay path 盡量貼近 `Execute`，避免額外的 file-level 或 top-level 跳轉。
  - [ ] replay 行為與既有測試結果一致。
- Notes: replay handling 仍屬於 application layer，但應貼近主 usecase flow。

### FR-003 - Keep issuance transaction flow linear and understandable

- Description: reservation、issued-address derivation、failure persistence、receipt tracking creation、idempotency completion 的流程需要線性可讀；是否拆 collaborator 不是目的本身。
- Acceptance criteria:
  - [ ] issuance path 的主要步驟可以順著讀，不需要在多個 collaborator file 間跳轉。
  - [ ] issuance path 仍以單一 transaction 協調 reservation / tracking consistency。
  - [ ] derivation failure persistence、idempotency release/complete 語意維持不變。
  - [ ] `issueAllocation` 的 transaction body 應以少量清楚步驟表達，而不是把所有細節塞成一大段。
  - [ ] `Execute` 在 `issueAllocation` 後的 fallback/error path 應由明確 owner 處理，避免把主流程塞滿分支。
- Notes: 流程可以有少量 local helper，但不應再被切成多個難追的 top-level collaborator。

### FR-004 - Move DTO mapping out of the main orchestration flow

- Description: allocation response shaping 不應混在主 usecase 流程末端，應有明確的 local owner。
- Acceptance criteria:
  - [ ] response shaping 不再造成主要流程閱讀中斷。
  - [ ] existing issued allocation replay 的 response shaping 也由同一個 owner 處理。
  - [ ] response payload 與既有測試一致。
- Notes: 這是 application-local mapping，不是 domain logic；避免保留 generic `build...` helper 名稱。

### FR-005 - Keep naming concrete and local

- Description: 這輪重構必須以「可讀性」優先，不得只為了責任對齊而增加檔案數與 helper 數。
- Acceptance criteria:
  - [ ] 命名直接反映目前責任，不引入 `workflow`, `framework`, `shared`, `common` 這類垃圾桶抽象。
  - [ ] 檔案與 helper 的數量不得比目前版本更難追蹤。
  - [ ] 主 allocation flow 的閱讀路徑比目前版本更短、更清楚。
  - [ ] 若 helper 仍保留，必須明顯降低認知負擔，而不是只是把邏輯搬離主流程。
  - [ ] 單次使用且只薄包一層分支的 helper 應優先收回主流程。
- Notes: repo 偏好 concrete names over reusable-looking abstractions。

### FR-006 - Keep generate flow readable in one place

- Description: `GenerateAddressUseCase.Execute` 應直接表達 policy lookup、preview validation、error mapping、address derivation，不保留沒有必要的 single-use helper。
- Acceptance criteria:
  - [ ] `validateGenerateAddressPolicy` 被移除。
  - [ ] `Execute` 仍可直接看懂 preview validation 與 error mapping。
  - [ ] 不引入新的 single-use wrapper helper 取代它。
- Notes: 這輪優先考慮順讀性，而不是責任對齊式拆分。

### FR-007 - Preserve generate preview validation semantics

- Description: `GenerateAddressUseCase` inline 後必須維持既有 preview validation 和 inbound error mapping。
- Acceptance criteria:
  - [ ] `ErrAddressPolicyChainMismatch` 仍映射成 `ErrAddressPolicyNotFound`。
  - [ ] `ErrAddressPolicyNotEnabled` 仍映射成 `ErrAddressPolicyNotEnabled`。
  - [ ] `ErrAddressPolicyPreviewNotSupported` 仍映射成 `ErrAddressPreviewNotSupported`。
  - [ ] 其他錯誤仍原樣回傳。
- Notes: 行為穩定性比 helper 存在與否更重要。

### FR-008 - Keep receipt polling cycle readable without changing orchestration semantics

- Description: `RunReceiptPollingCycleUseCase.Execute` 仍保有 claim -> observe -> lifecycle -> save/outbox 的 orchestration，但要收斂重複的 polling-error save path 與 save+enqueue path。
- Acceptance criteria:
  - [ ] 重複的 `MarkPollingError -> Save -> ProcessingErrorCount++` 路徑被收斂成明確 owner。
  - [ ] 重複的 `Save -> maybe enqueue status-changed event` 路徑被收斂成明確 owner。
  - [ ] main loop 仍能直接看懂單筆 tracking 的高階流程。
  - [ ] latest-block-height cache 行為與 status/outbox 行為不變。
- Notes: 不拆成多個 public usecase；只整理 internal readability。

### FR-009 - Keep receipt webhook dispatch cycle readable without changing delivery semantics

- Description: `RunReceiptWebhookDispatchCycleUseCase.Execute` 仍保有 claim -> notify -> resolve delivery result -> save result 的 orchestration，但要收斂單筆 dispatch 流程與重複的 result-save transaction。
- Acceptance criteria:
  - [ ] 單筆 notification dispatch 流程由明確 owner 處理，`Execute` main loop 只保留高階步驟。
  - [ ] 重複的 `SaveDeliveryResult` transaction 路徑被收斂成明確 owner。
  - [ ] sent / retried / failed counter 行為與 clock 使用語意維持不變。
  - [ ] notifier input 與 delivery result resolution 行為不變。
- Notes: 不需要大拆，只整理重複與分支密度。

### FR-010 - Keep transport normalization and runtime defaults out of usecases

- Description: `internal/application/usecases` 不應再持有 inbound transport normalization 或 scheduler/bootstrap runtime default ownership；這些責任應留在 inbound adapters 與 bootstrap。
- Acceptance criteria:
  - [ ] `AllocatePaymentAddressUseCase` 不再對 `AddressPolicyID`、`CustomerReference`、`IdempotencyKey` 做 `TrimSpace`。
  - [ ] `RunReceiptPollingCycleUseCase` 不再在 usecase 內注入 `RescheduleInterval` / `ClaimTTL` default。
  - [ ] `RunReceiptWebhookDispatchCycleUseCase` 不再在 usecase 內注入 `DispatchTTL` default。
  - [ ] 現有 controller / bootstrap / scheduler call path 仍能提供必要 normalization 與 default 值。
- Notes: 這是 ownership cleanup，不是 outward contract redesign。

## Non-functional requirements

- Performance (NFR-001): 不新增額外 DB round trip 或外部 IO hop。
- Availability/Reliability (NFR-002): `go test ./internal/application/usecases` 與 `go test ./...` 必須持續通過。
- Security/Privacy (NFR-003): idempotency claim/release/complete、failure persistence、tracking creation、preview validation、receipt status/outbox、webhook delivery semantics 不得放寬。
- Compliance (NFR-004): 不適用。
- Observability (NFR-005): 不改動現有 outward response / log behavior。
- Maintainability (NFR-006): 主 allocation flow 應能在少量檔案內看出完整流程；減少不必要的 helper / file 跳轉，且 usecase 不再持有 transport/runtime ownership。

## Dependencies and integrations

- External systems: 無新增。
- Internal services:
  - `internal/application/ports/outbound`
  - `internal/domain/entities`
  - `internal/domain/policies`
