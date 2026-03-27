---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - allocation flow readability-oriented restructuring
  - generate preview validation helper inline
  - receipt polling cycle readability cleanup
  - receipt webhook dispatch cycle readability cleanup
  - final usecase boundary ownership cleanup
- Not covered:
  - other usecase refactors
  - API/schema redesign

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-002, FR-005, NFR-002, NFR-003, NFR-006
  - Steps: `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase(ReturnsExistingIssuedAllocation|RejectsIdempotencyKeyConflict|AllowsRetryWhenClaimConflictHasNoCompletedRecord)'`
  - Expected: replay / idempotency 行為不變。
- TC-002:
  - Linked requirements: FR-003, FR-004, FR-005, NFR-001, NFR-002, NFR-003, NFR-006
  - Steps: `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase'`
  - Expected: allocation usecase 全量測試通過，含 derivation failure persistence、tracking creation、idempotency completion。
- TC-003:
  - Linked requirements: FR-006, FR-007, NFR-001, NFR-002, NFR-003, NFR-005, NFR-006
  - Steps: `go test ./internal/application/usecases -run 'TestGenerateAddressUseCase'`
  - Expected: generate usecase 全量測試通過，preview validation error mapping 維持不變。
- TC-004:
  - Linked requirements: FR-008, NFR-001, NFR-002, NFR-003, NFR-005, NFR-006
  - Steps: `go test ./internal/application/usecases -run 'TestRunReceiptPollingCycleUseCase'`
  - Expected: receipt polling usecase 全量測試通過，含 polling error、latest block height cache、status-change enqueue、expiration flow。
- TC-005:
  - Linked requirements: FR-009, NFR-001, NFR-002, NFR-003, NFR-005, NFR-006
  - Steps: `go test ./internal/application/usecases -run 'TestRunReceiptWebhookDispatchCycleUseCase'`
  - Expected: receipt webhook dispatch usecase 全量測試通過，含 sent / retry / terminal failure / validation path。
- TC-006:
  - Linked requirements: FR-010, NFR-002, NFR-006
  - Steps: `go test ./internal/adapters/inbound/http/controllers ./internal/bootstrap/...`
  - Expected: trimming 與 runtime default ownership 仍由 controller / bootstrap 承擔，對應測試通過。

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, FR-009, NFR-002, NFR-006
  - Steps: `go test ./internal/application/usecases`
  - Expected: usecase package 全量測試通過。
- TC-102:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, FR-009, FR-010, NFR-002
  - Steps: `go test ./...`
  - Expected: 全 repo 測試通過。

### E2E (if applicable)

- Scenario 1: 不適用。
- Scenario 2: 不適用。

## Edge cases and failure modes

- Case: idempotency key claim conflict 但沒有 completed record。
  - Expected behavior: 行為維持原本的 retry / error semantics。
- Case: issued-address derivation 失敗或 derivation failure persistence 失敗。
  - Expected behavior: failure path 與既有測試結果一致。
- Case: completed idempotency record 指向不存在的 issued allocation。
  - Expected behavior: consistency error 行為維持不變。
- Case: preview unsupported / disabled policy / chain mismatch。
  - Expected behavior: generate usecase 的 inbound error mapping 維持不變。
- Case: observer error、tip-height error、expired tracking、status change enqueue。
  - Expected behavior: receipt polling cycle 的 processing error / terminal failed / updated counters 與 save/enqueue 行為維持不變。
- Case: webhook notify error retry、terminal failure、successful send。
  - Expected behavior: receipt webhook dispatch cycle 的 sent / retried / failed counters 與 delivery-result persistence 行為維持不變。
- Case: controller / bootstrap 沒有先做 trimming 或 runtime default。
  - Expected behavior: ownership 不再由 usecase 補位；相對應邊界應由 adapter / bootstrap 測試覆蓋。

## NFR verification

- Performance: 不新增額外 transaction 或外部 IO hop。
- Reliability: `go test ./internal/application/usecases` 與 `go test ./...` 持續通過。
- Security: idempotency、failure persistence、preview validation、receipt status/outbox、webhook delivery semantics 維持不變。
