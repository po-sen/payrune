---
doc: 03_tasks
spec_date: 2026-03-28
slug: http-controller-api-locality
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-24-architecture-conformance-refactor
  - 2026-03-27-application-inbound-error-mapping
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
- Rationale: 只做 HTTP inbound controller 的 locality refactor，不改 route、不改 usecase contract、也不改 outward API 行為。
- Upstream dependencies (`depends_on`): `2026-03-24-architecture-conformance-refactor`, `2026-03-27-application-inbound-error-mapping`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 這輪沒有新 integration 或資料模型，只調整檔案切分與責任可讀性。
  - What would trigger switching to Full mode: 若要重設 HTTP error contract、route structure、或 transport abstraction。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 每個 task 都附具體測試/驗證指令。

## Milestones

- M1: Split chain-address controller into endpoint-local files.
- M2: Verify unchanged HTTP behavior and close the spec.

## Tasks (ordered)

1. T-001 - Split chain-address handling into explicit per-API controllers
   - Scope: 把 list/generate/allocate/get-status 的 handler 邏輯拆成各自 controller type 與各自檔案，移除 aggregate chain-address controller，並統一 endpoint 檔名為 `*_controller.go`。
   - Output: one endpoint-local `*_controller.go` source file and controller type per chain-address API.
   - Linked requirements: FR-001 / FR-002 / FR-004 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): inspect `internal/adapters/inbound/http/controllers/` tree and run `go test ./internal/adapters/inbound/http/controllers`
     - [x] Expected result: each endpoint has its own file and controller tests pass.
     - [x] Logs/metrics to check (if applicable): N/A
2. T-002 - Make bootstrap, router, and tests explicit per API
   - Scope: 調整 bootstrap、router、controller tests，讓 composition 和 test locality 都直接對到每個 API controller。
   - Output: explicit per-API router/bootstrap wiring and matching endpoint-local test files.
   - Linked requirements: FR-003 / FR-004 / FR-005 / NFR-002 / NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/inbound/http/controllers ./internal/adapters/inbound/http`
     - [x] Expected result: existing HTTP tests pass without intended contract drift.
     - [x] Logs/metrics to check (if applicable): N/A
3. T-003 - Run full validation and close the spec
   - Scope: 跑 full suite 與 spec lint，更新 spec 到 `DONE`。
   - Output: verified controller locality refactor with final evidence.
   - Linked requirements: FR-001 / FR-002 / FR-003 / NFR-001 / NFR-002 / NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`, `SPEC_DIR="specs/2026-03-28-http-controller-api-locality" bash scripts/spec-lint.sh`, `bash scripts/precommit-run.sh`
     - [x] Expected result: full suite passes and spec reflects final behavior.
     - [x] Logs/metrics to check (if applicable): N/A

## Traceability (optional)

- FR-001 -> T-001, T-003
- FR-002 -> T-001, T-003
- FR-003 -> T-002, T-003
- FR-004 -> T-001, T-002, T-003
- NFR-001 -> T-003
- NFR-002 -> T-002, T-003
- NFR-003 -> T-002, T-003
- NFR-005 -> T-002, T-003
- NFR-006 -> T-001, T-003

## Rollout and rollback

- Feature flag: None
- Migration sequencing: split files first, then run controller/full validation
- Rollback steps: revert the file split if any HTTP contract regression appears

## Validation evidence

- `ls internal/adapters/inbound/http/controllers` shows endpoint-local files for list/generate/allocate/get-status.
- Each per-API controller now exposes `ServeHTTP`, so request parsing, method checks, usecase calls, and status mapping are visible from one file.
- `internal/bootstrap/api.go` and `internal/bootstrap/api_worker.go` now construct one controller per API and pass them explicitly into `httpadapter.RouterControllers`.
- `internal/adapters/inbound/http/router.go` now registers explicit per-API `http.Handler` fields instead of a shared aggregate controller or per-controller special method names.
- Controller tests now mount routes directly to the owning controller instead of rebuilding an aggregate-style helper.
- `go test ./internal/adapters/inbound/http/controllers ./internal/adapters/inbound/http` passed.
- `go test ./...` passed.
- `SPEC_DIR="specs/2026-03-28-http-controller-api-locality" bash scripts/spec-lint.sh` passed.
- `bash scripts/precommit-run.sh` passed.
