---
doc: 01_requirements
spec_date: 2026-03-27
slug: application-inbound-error-mapping
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Glossary (optional)

- Inbound error:
  - `internal/application/ports/inbound` 定義，允許離開 application layer 的 usecase contract error。
- Unexpected dependency failure:
  - outbound port implementation、adapter-private logic、或底層 transaction / IO 在 usecase 既有 business mapping 之外的失敗。
- Unexpected internal failure:
  - usecase/domain/policy 在既有 business mapping 之外出現的 non-user-facing consistency failure。

## Out-of-scope behaviors

- OOS1: 不重寫 HTTP controller 的 public response body。
- OOS2: 不要求 adapter/private technical error 必須改成 shared sentinel。

## Functional requirements

### FR-001 - Usecases only return inbound contract errors

- Description: production `usecase.Execute(...)` 與其 private helper 對外返回時，unexpected 失敗必須先收斂成 `inport.Err...`，不能直接回傳 `outport.Err...`、adapter-private error、或其他 raw error。
- Acceptance criteria:
  - [ ] `internal/application/usecases` production code 不再把 unexpected non-inbound error 直接回傳給 caller。
  - [ ] 既有 known business/config/validation path 仍維持現有 `inport.Err...`。
  - [ ] `GetPaymentAddressStatusUseCase` 不再把 `outport.ErrPaymentAddressStatusIncomplete` 直接回傳給 caller。
- Notes: 只要求離開 application layer 的 contract；usecase 內部仍可用 `errors.Is` 判斷 `outport.Err...`。

### FR-002 - Generic inbound errors distinguish dependency vs internal failure

- Description: application inbound contract 必須新增 generic error，用來承接 unexpected dependency failure 與 unexpected internal consistency failure。
- Acceptance criteria:
  - [ ] `internal/application/ports/inbound/errors.go` 定義 generic inbound error，覆蓋 dependency failure 與 internal failure 兩類。
  - [ ] usecase 針對 outbound call / transaction / adapter failure 轉成 dependency-facing inbound error。
  - [ ] usecase 針對 unexpected domain/policy/internal consistency failure 轉成 internal-facing inbound error。
- Notes: inbound adapter 可以繼續把這兩類都 map 成 500；重點是 application contract 不再 leak raw error。

### FR-003 - Tests lock the boundary

- Description: 關鍵 usecase 測試必須改成驗證 inbound contract，而不是驗證 raw adapter/private error 會穿透。
- Acceptance criteria:
  - [ ] `GenerateAddressUseCase`、`AllocatePaymentAddressUseCase`、`GetPaymentAddressStatusUseCase`、`RunReceiptPollingCycleUseCase`、`RunReceiptWebhookDispatchCycleUseCase` 都有對應 regression test。
  - [ ] 至少一個 inbound adapter 測試覆蓋 generic inbound error 不影響既有 transport 行為。
- Notes: 可以重用現有 fake/outport stub，不需要新增 integration test env。

## Non-functional requirements

- Performance (NFR-001): 不新增額外 network/DB round trip；error mapping 只允許 in-process branching。
- Availability/Reliability (NFR-002): 既有成功路徑與 known business error 路徑行為不變；`go test ./...` 必須通過。
- Security/Privacy (NFR-003): unexpected adapter/private failure detail 不可透過 application contract 對外暴露。
- Compliance (NFR-004): 無新增要求。
- Observability (NFR-005): 不新增 telemetry；維持既有 error propagation 到 outer layer 的 logging 空間。
- Maintainability (NFR-006): generic inbound error 命名必須清楚、少量、可被後續 usecase 直接理解。

## Dependencies and integrations

- External systems: 無新增；沿用既有 outbound adapters。
- Internal services: `internal/application/ports/inbound`, `internal/application/usecases`, `internal/adapters/inbound/http/controllers`, `internal/adapters/inbound/scheduler`。
