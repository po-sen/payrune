---
doc: 04_test_plan
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

# Test Plan

## Scope

- Covered:
  - usecase-returned application error centralization
  - controller mapping stability
  - outbound adapter error ownership stability
  - duplicated outbound persistence contract error centralization
- Not covered:
  - adapter-internal technical error redesign
  - domain error redesign

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-003, NFR-002, NFR-006
  - Steps: `go test ./internal/application/usecases`
  - Expected: usecase 測試通過，validation/configuration 測試改為 shared inbound error。
- TC-002:
  - Linked requirements: FR-002, FR-004, NFR-002, NFR-006
  - Steps: `go test ./internal/adapters/inbound/http/controllers`
  - Expected: controller 對 `inport.Err...` 的 HTTP mapping 維持不變。
- TC-003:
  - Linked requirements: FR-002, FR-003, NFR-002, NFR-006
  - Steps: `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
  - Expected: 多實作 persistence adapter 共用 `outport.Err...` contract errors，測試通過。
- TC-004:
  - Linked requirements: FR-005, NFR-006
  - Steps: `rg -n "Error ownership|adapter-private" AGENTS.md`
  - Expected: `AGENTS.md` 可直接讀到 error ownership repo 契約。

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, NFR-002, NFR-003, NFR-005, NFR-006
  - Steps: `go test ./...`
  - Expected: 全 repo 測試通過，無新的 error ownership regression。

### E2E (if applicable)

- Scenario 1:
- Scenario 2:

## Edge cases and failure modes

- Case: usecase configuration dependency 缺失。
  - Expected behavior: 回傳 shared inbound error，不再是散落字串。
- Case: outbound adapter 回傳 `outport.Err...`。
  - Expected behavior: usecase 仍依既有邏輯 branch，不搬成 inbound error。
- Case: 同一個 outbound port 有 2 個 adapter implementation。
  - Expected behavior: 共享的 contract/validation/state error 由 `outport` 集中，不再靠重複字串同步。
- Case: controller 收到既有 application business error。
  - Expected behavior: HTTP status mapping 維持不變。

## NFR verification

- Performance: 不新增額外 IO。
- Reliability: `go test ./internal/application/usecases`、`go test ./internal/adapters/inbound/http/controllers`、`go test ./...` 通過。
- Security: 不新增 internal error detail 暴露到 public response。
