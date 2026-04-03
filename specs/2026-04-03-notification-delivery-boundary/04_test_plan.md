---
doc: 04_test_plan
spec_date: 2026-04-03
slug: notification-delivery-boundary
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-domain-model-boundary-cleanup
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
  - notification delivery status / failure reason ownership move
  - webhook dispatch use case call sites
  - postgres / cloudflarepostgres outbox persistence call sites
- Not covered:
  - external webhook receiver behavior
  - DB migration, because this spec does not change schema

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001 / NFR-006
  - Steps:
    - `go test ./internal/application/outbox`
  - Expected:
    - moved delivery workflow types 與 delivery result tests 通過。
- TC-002:
  - Linked requirements: FR-001 / FR-002 / NFR-002
  - Steps:
    - `go test ./internal/application/usecases`
  - Expected:
    - webhook dispatch use case 仍正確處理 sent / pending / failed 分支。

### Integration

- TC-101:
  - Linked requirements: FR-002 / NFR-002 / NFR-005
  - Steps:
    - `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
  - Expected:
    - outbox stores 仍正確 parse / persist delivery status 與 failure reason。
    - persisted string values 維持不變。

### E2E (if applicable)

- Scenario 1:
  - 不適用。
- Scenario 2:
  - 不適用。

## Edge cases and failure modes

- Case:
  - persisted delivery status / failure reason 為空或無效
- Expected behavior:
  - 仍沿既有 adapter normalization / contract error 路徑處理，不因 type 搬移而改變語意。

## NFR verification

- Performance:
  - `go test ./...` 通過，且沒有新增 round trip。
- Reliability:
  - webhook dispatch 與 outbox persistence 相關測試維持通過。
- Security:
  - 無新增 payload surface 或 secrets。
