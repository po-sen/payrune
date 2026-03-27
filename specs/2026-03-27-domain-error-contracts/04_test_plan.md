---
doc: 04_test_plan
spec_date: 2026-03-27
slug: domain-error-contracts
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-inbound-error-mapping
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
  - domain sentinel error extraction across `entities`, `valueobjects`, `events`, `policies`
  - domain tests using `errors.Is(...)`
  - application usecase regressions affected by new domain error contract
- Not covered:
  - application/outbound/bootstrap error redesign
  - HTTP contract changes

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002
  - Steps: 執行 `go test ./internal/domain/...`。
  - Expected: domain tests 通過，並對新的 sentinel errors 做穩定 assertions。
- TC-002:
  - Linked requirements: FR-002, FR-003
  - Steps: 執行 `go test ./internal/application/usecases`。
  - Expected: usecase tests 維持綠燈，既有 domain-to-inport mapping 不回歸。

### Integration

- TC-101:
  - Linked requirements: FR-003, NFR-002
  - Steps: 執行 `go test ./...`。
  - Expected: 全 repo 測試通過。

### E2E (if applicable)

- Scenario 1: 不適用。
- Scenario 2: 不適用。

## Edge cases and failure modes

- Case: 呼叫端仍以字串比對 domain error。
- Expected behavior: 測試會暴露不穩定 assertion，並改成 `errors.Is(...)`。

- Case: 抽取 error 後 accidentally 改變 validation path。
- Expected behavior: domain/application tests fail 並阻擋回歸。

## NFR verification

- Performance: 不應有可觀察的 runtime 成本變化。
- Reliability: 全 repo tests 維持通過。
- Security: 不適用。
