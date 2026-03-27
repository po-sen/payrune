---
doc: 03_tasks
spec_date: 2026-03-27
slug: outbound-port-error-conformance
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
  test_plan: null
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale: 這輪是既有 outbound ports 與 adapter error contract 的收斂，不新增 integration、schema、或 async 設計。
- Upstream dependencies (`depends_on`): `2026-03-27-application-error-boundaries`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 只需要在既有 port 與 adapter 邊界上收斂 shared error contract，不涉及新資料流或新元件。
  - What would trigger switching to Full mode: 若需要改寫 port shape、新增 error code payload、或變更 outward API 行為才需要 Full mode。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 每個 task 都包含對應 `go test` / grep 驗證。

## Milestones

- M1: 盤點並補齊 outbound port shared error contract。
- M2: 收斂 adapter 實作與測試，確保 port-facing methods 不再回 raw adapter error。

## Tasks (ordered)

1. T-001 - Audit current outbound port-facing errors and define shared contracts
   - Scope: 盤點 `internal/adapters/outbound` 中所有公開 port method 仍直接回 raw error 的位置，並在對應 `internal/application/ports/outbound/*.go` 補齊 `outport.Err...`。
   - Output: 更新過的 outbound port error definitions，覆蓋 address deriver、issued address deriver、receipt observer、notifier 等目前 concrete ports。
   - Linked requirements: FR-001 / FR-003 / NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `rg -n "var Err|errors.New\\(|fmt.Errorf\\(" internal/application/ports/outbound --glob '*.go'`
     - [ ] Expected result: relevant outbound port files define the shared error sentinels needed by current adapters.
     - [ ] Logs/metrics to check (if applicable): N/A
2. T-002 - Refactor port implementations to return only outport-defined errors
   - Scope: 更新 `internal/adapters/outbound` 的公開 port methods，將 raw adapter error 改成 `outport.Err...` 或可用 `errors.Is(...)` 辨識的包裝。
   - Output: cleaned adapter implementations with constructor-local errors preserved.
   - Linked requirements: FR-001 / FR-002 / NFR-002 / NFR-003 / NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `rg -n "errors.New\\(|fmt.Errorf\\(" internal/adapters/outbound --glob '*.go'`
     - [ ] Expected result: remaining inline errors are limited to constructors/helpers or wrapped into `outport.Err...`; public port methods no longer directly emit adapter-local raw errors.
     - [ ] Logs/metrics to check (if applicable): N/A
3. T-003 - Lock the contract with tests and final validation
   - Scope: 更新 adapter tests 與必要的 usecase-facing tests，確保 port-facing error paths 可用 `errors.Is(..., outport.Err...)` 驗證。
   - Output: passing focused tests, full test suite, and DONE spec.
   - Linked requirements: FR-003 / NFR-001 / NFR-002 / NFR-005 / NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/adapters/outbound/... ./internal/application/usecases ./...`
     - [ ] Expected result: tests pass and error contract assertions use `errors.Is(...)` against `outport.Err...`.
     - [ ] Logs/metrics to check (if applicable): N/A

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-002
- FR-003 -> T-001, T-003
- NFR-001 -> T-003
- NFR-002 -> T-002, T-003
- NFR-003 -> T-002
- NFR-005 -> T-003
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag: None.
- Migration sequencing: Refactor port definitions first, then adapters, then tests.
- Rollback steps: Revert the port error contract cleanup commit if application/usecase contract expectations regress.
