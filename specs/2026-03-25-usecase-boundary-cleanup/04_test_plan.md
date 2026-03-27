---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - allocation issued-address derivation ownership refactor
  - generate usecase preview capability relocation
  - poller typed scope normalization across bootstrap / scheduler / usecase
- Not covered:
  - HTTP response schema redesign
  - comprehensive DTO formatting cleanup

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, NFR-001, NFR-002, NFR-003, NFR-006
  - Steps: `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase'`
  - Expected: allocate usecase 測試通過，含 Bitcoin path、Ethereum create2 path、derivation failure path。
- TC-002:
  - Linked requirements: FR-002, NFR-002, NFR-006
  - Steps: `go test ./internal/application/usecases -run 'TestGenerateAddressUseCase'`
  - Expected: generate usecase 測試通過，preview unsupported 行為維持不變。
- TC-003:
  - Linked requirements: FR-003, NFR-002, NFR-003, NFR-006
  - Steps: `go test ./internal/application/usecases -run 'TestRunReceiptPollingCycleUseCase'`
  - Expected: poller usecase 測試通過，typed filter 與 polling lifecycle 行為正確。

### Integration

- TC-101:
  - Linked requirements: FR-003, FR-004, NFR-002, NFR-003, NFR-006
  - Steps: `go test ./internal/bootstrap -run 'TestLoadPoller|TestBuildCloudflarePollerRequest|TestHandleCloudflarePollerRequestJSON'`
  - Expected: bootstrap poller path 測試通過，typed scope validation 與 worker request builder 正常。
- TC-102:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-002, NFR-006
  - Steps: `go test ./...`
  - Expected: 全 repo 測試通過，表示 boundary 搬移沒有破壞其他 layer。

### E2E (if applicable)

- Scenario 1: 不適用。
- Scenario 2: 不適用。

## Edge cases and failure modes

- Case: Ethereum create2 salt derivation backend 未配置或回傳錯誤。
  - Expected behavior: allocation 失敗原因仍被正確持久化，usecase 不直接持有 create2 derivation 分支。
- Case: poller 設定了 network 但未設定 chain。
  - Expected behavior: validation 仍拒絕，且錯誤發生在 usecase boundary 之前。
- Case: policy 已存在但不支援 preview。
  - Expected behavior: usecase 仍回傳 `ErrAddressPreviewNotSupported`，但 capability rule 由 domain 表達。

## NFR verification

- Performance: 不新增外部 IO hop；只重組現有 in-process collaboration。
- Reliability: `go test ./internal/application/...`、`go test ./internal/bootstrap/...`、`go test ./...` 必須通過。
- Security: create2 derivation key handling 與 poller scope validation 語意不變。
