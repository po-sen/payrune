---
doc: 04_test_plan
spec_date: 2026-03-25
slug: bootstrap-dedup-refactor
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-24-architecture-conformance-refactor
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Test Plan

## Scope

- Covered:
  - API receipt terms parsing dedup。
  - poller dispatch config / worker request parsing dedup。
  - receipt webhook dispatcher dispatch config 與 notifier shared parsing dedup。
  - bootstrap-locality constraint 驗證。
- Not covered:
  - runtime builder / container 大型重寫。
  - `internal/bootstrap` 以外的結構重構。

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, NFR-002, NFR-006
  - Steps: `go test ./internal/bootstrap -run 'TestLoadReceipt|TestExecuteAPIWorkerRequest'`
  - Expected: API process / worker 的 receipt terms 相關測試通過，表示 shared helper 沒破壞既有語意。
- TC-002:
  - Linked requirements: FR-002, NFR-002, NFR-006
  - Steps: `go test ./internal/bootstrap -run 'TestLoadPoller|TestBuildCloudflarePollerRequest|TestHandleCloudflarePollerRequestJSON'`
  - Expected: poller process / worker 的 parsing 與 validation 測試通過。
- TC-003:
  - Linked requirements: FR-003, NFR-002, NFR-003, NFR-006
  - Steps: `go test ./internal/bootstrap -run 'TestLoadReceiptWebhookDispatcher|TestLoadPaymentReceiptWebhookNotifier|TestBuildCloudflareReceiptWebhookDispatcherRequest|TestHandleCloudflareReceiptWebhookDispatcherRequestJSON|TestLoadCloudflareReceiptWebhookNotifierConfig'`
  - Expected: dispatcher process / worker 的 parsing 與 notifier config 測試通過。

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-002, NFR-006
  - Steps: `go test ./internal/bootstrap/...`
  - Expected: bootstrap package 全量測試通過。
- TC-102:
  - Linked requirements: FR-004, NFR-006
  - Steps: `go test ./...`
  - Expected: repo 全量測試通過，表示 dedup 沒破壞其他 call site。

### E2E (if applicable)

- Scenario 1: 不適用。
- Scenario 2: 不適用。

## Edge cases and failure modes

- Case: process 與 worker 共享 helper 後 accidentally 吃到同一組 defaults。
- Expected behavior: 測試能驗證 process / worker 仍保留各自應有的 fallback 值。
- Case: chain/network validation 被 dedup 時誤改錯誤訊息或 required 條件。
- Expected behavior: poller / dispatcher worker 測試持續驗證 invalid JSON、missing chain、invalid bool/int/duration path。

## NFR verification

- Performance: 不新增外部 IO；只抽本地 helper。
- Reliability: `go test ./internal/bootstrap/...` 與 `go test ./...` 通過。
- Security: secret/bool/int/duration validation path 維持既有測試覆蓋。
