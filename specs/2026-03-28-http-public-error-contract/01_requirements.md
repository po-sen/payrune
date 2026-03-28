---
doc: 01_requirements
spec_date: 2026-03-28
slug: http-public-error-contract
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-inbound-error-mapping
  - 2026-03-28-http-controller-api-locality
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Requirements

## Glossary (optional)

- Public error text:
- Controller 寫進 `dto.ErrorResponse{Error: ...}` 的對外 HTTP 錯誤文字。

## Out-of-scope behaviors

- OOS1: 不改 application `inport.Err...` 字串。
- OOS2: 不改既有 route shape、request parsing 或 status code mapping。

## Functional requirements

### FR-001 - Controller-owned public error text

- Description: 每個 chain-address HTTP controller 必須自行決定 public error text，不得直接以 `err.Error()` 輸出 `inport.Err...` 作為 client-facing message。
- Acceptance criteria:
  - [ ] `list_address_policies_controller.go`、`generate_address_controller.go`、`allocate_payment_address_controller.go`、`get_payment_address_status_controller.go` 不再用 `err.Error()` 寫入 `dto.ErrorResponse`.
  - [ ] invalid chain path 的 response text 由 HTTP adapter 自己提供，不再直接讀 `inport.ErrChainNotSupported.Error()`.
- Notes: endpoint-local mapping 可以少量重複，不要求抽出 generic mapper。

### FR-002 - Keep existing status mapping stable

- Description: 這輪 cleanup 不得改變每個 API 既有的 status code 分配。
- Acceptance criteria:
  - [ ] list/generate/allocate/get-status 這四個 API 的 success status 與 error status 維持不變。
  - [ ] internal error 仍回 `500 internal server error`。
- Notes: 這輪只調整 public message ownership，不重設 status contract。

### FR-003 - Tests lock public HTTP contract

- Description: controller tests 必須明確驗證 error response 的 public message，而不是只驗 status。
- Acceptance criteria:
  - [ ] 至少每個 chain-address controller 各有一組 error mapping test 檢查 `dto.ErrorResponse.Error`。
  - [ ] unknown chain path test 也檢查 public error text。
- Notes: 這讓 future refactor 不會不小心把 `inport.Err...` 文字重新暴露出去。

## Non-functional requirements

- Performance (NFR-001): 這輪不得新增額外 IO 或顯著增加 handler branching 複雜度。
- Availability/Reliability (NFR-002): 錯誤分類結果與現有 status code 必須保持相容，避免 API regression。
- Security/Privacy (NFR-003): public HTTP response 不得依賴 lower-layer raw error wording。
- Compliance (NFR-004):
- Observability (NFR-005): N/A
- Maintainability (NFR-006): reviewer 打開單一 controller file 時，必須能直接看出 error -> status -> public message 的 mapping。

## Dependencies and integrations

- External systems: None
- Internal services: `internal/application/ports/inbound`, `internal/adapters/inbound/http/dto`
