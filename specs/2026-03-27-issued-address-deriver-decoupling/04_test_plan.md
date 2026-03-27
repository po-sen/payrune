---
doc: 04_test_plan
spec_date: 2026-03-27
slug: issued-address-deriver-decoupling
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null # set to 02_design.md in Full mode
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Test Plan

## Scope

- Covered:
  - `IssuedPaymentAddressDeriver` package ownership shape
  - bootstrap wiring compatibility
  - bitcoin / ethereum create2 issuance behavior regression checks
- Not covered:
  - `Create2SaltDeriver` algorithm redesign
  - application port redesign

## Tests

### Unit

- TC-001:
  - Linked requirements: FR-001, FR-002, FR-003, NFR-002, NFR-006
  - Steps: `go test ./internal/adapters/outbound/blockchain`
  - Expected: `blockchain` 只剩 multi-chain dispatch 行為測試，chain-specific issued derivation 不再留在這個 package。
- TC-002:
  - Linked requirements: FR-002, FR-003, NFR-002, NFR-005, NFR-006
  - Steps: `go test ./internal/adapters/outbound/bitcoin ./internal/adapters/outbound/ethereum ./internal/bootstrap/...`
  - Expected: bitcoin / ethereum issued derivation 與 bootstrap 組裝維持通過。

### Integration

- TC-101:
  - Linked requirements: FR-001, FR-002, FR-003, NFR-002, NFR-003, NFR-005, NFR-006
  - Steps: `go test ./...`
  - Expected: 全 repo 測試通過，無 adapter dependency shape regression。

### E2E (if applicable)

- Scenario 1:
- Scenario 2:

## Edge cases and failure modes

- Case: ethereum create2 deriver 未配置 salt deriver。
- Expected behavior: 仍回傳與現行行為一致的 not-configured error。
- Case: 非 ethereum/create2 policy。
- Expected behavior: 不需要 salt collaborator 也能照常導出地址。

## NFR verification

- Performance: 不新增額外 IO 或多餘 indirection。
- Reliability: `go test ./internal/adapters/outbound/blockchain ./internal/bootstrap/...`、`go test ./...` 通過。
- Security: 不改變 derivation key / secret material 使用方式。
