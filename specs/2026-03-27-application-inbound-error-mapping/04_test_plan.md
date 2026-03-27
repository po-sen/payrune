---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered: generic inbound error definition、usecase error mapping、至少一條 inbound adapter regression。
- Not covered: adapter-private error catalog cleanup、HTTP response schema redesign、logging/telemetry enhancement。

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002
  - Steps: 執行 `go test ./internal/application/usecases -run 'TestGenerateAddressUseCase|TestAllocatePaymentAddressUseCase|TestGetPaymentAddressStatusUseCase'`
  - Expected: unexpected reader/deriver/finder error 會被映射成 generic inbound error；既有 known business mapping 維持不變。
- TC-002:
  - Linked requirements: FR-001, FR-002
  - Steps: 執行 `go test ./internal/application/usecases -run 'TestRunReceiptPollingCycleUseCase|TestRunReceiptWebhookDispatchCycleUseCase'`
  - Expected: unexpected transaction/store/notifier failure 會被映射成 generic inbound error；output counter 與成功路徑不變。

### Integration

- TC-101:
  - Linked requirements: FR-003, NFR-002
  - Steps: 執行 `go test ./internal/adapters/inbound/...`
  - Expected: HTTP / scheduler inbound adapter 測試維持綠燈，generic inbound error 不改既有 transport contract。

### E2E (if applicable)

- Scenario 1: 不適用。
- Scenario 2: 不適用。

## Edge cases and failure modes

- Case: `GetPaymentAddressStatusUseCase` 遇到 `outport.ErrPaymentAddressStatusIncomplete`
  - Expected behavior: 不再直接回傳 `outport` error，改成 generic inbound error。
- Case: `AllocatePaymentAddressUseCase` 的 derivation failure / persistence failure
  - Expected behavior: 交易內部既有補償流程維持，但對外回傳 generic inbound error。
- Case: `RunReceiptPollingCycleUseCase` 在 save/enqueue 時失敗
  - Expected behavior: 對 caller 回 generic inbound error，不把 raw persistence error 往外丟。

## NFR verification

- Performance: 只新增 in-process error mapping；不應增加 IO。
- Reliability: `go test ./...` 必須通過。
- Security: unexpected internal/adpater failure detail 不應作為 application contract 對外傳播。
