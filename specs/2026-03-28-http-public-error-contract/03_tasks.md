---
doc: 03_tasks
spec_date: 2026-03-28
slug: http-public-error-contract
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-inbound-error-mapping
  - 2026-03-28-http-controller-api-locality
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
- Rationale: 只調整 HTTP inbound adapter 的 error message ownership 與對應測試，不涉及新 integration、schema、或 async design。
- Upstream dependencies (`depends_on`): `2026-03-27-application-inbound-error-mapping`, `2026-03-28-http-controller-api-locality`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 這輪不改 route、不改 usecase、不改 persistence，只收 transport-level public contract。
  - What would trigger switching to Full mode: 若要加入 HTTP error code schema、版本化 API error contract、或跨 transport 共用 abstraction。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 每個 task 都附具體測試指令與 expected result。

## Milestones

- M1: Replace `err.Error()`-based HTTP public messages with controller-owned mappings.
- M2: Lock the contract in tests and close the spec.

## Tasks (ordered)

1. T-001 - Move public error text ownership into HTTP controllers
   - Scope: 調整 chain-path parsing 與 list/generate/allocate/get-status controllers，讓 public error text 由 transport layer 明確決定，不再直接輸出 `inport.Err...` 字串。
   - Output: controller-local error mapping functions or constants with unchanged status mapping.
   - Linked requirements: FR-001 / FR-002 / NFR-002 / NFR-003 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `rg -n "err\\.Error\\(\\)" internal/adapters/inbound/http/controllers --glob '*.go'`
     - [x] Expected result: chain-address controllers no longer use `err.Error()` in `dto.ErrorResponse`.
     - [x] Logs/metrics to check (if applicable): N/A
2. T-002 - Lock HTTP public messages in controller tests
   - Scope: 更新 per-controller tests 與 shared chain path test，讓 error mapping 同時驗 status 與 public message。
   - Output: controller tests that assert stable `dto.ErrorResponse.Error` values.
   - Linked requirements: FR-003 / NFR-002 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/inbound/http/controllers ./internal/adapters/inbound/http`
     - [x] Expected result: controller tests pass and explicitly cover status/message contract.
     - [x] Logs/metrics to check (if applicable): N/A
3. T-003 - Run full validation and close the spec
   - Scope: 跑 full suite、spec lint、更新 spec 狀態到 `DONE`。
   - Output: final validated HTTP public error contract cleanup.
   - Linked requirements: FR-001 / FR-002 / FR-003 / NFR-001 / NFR-002 / NFR-003 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`, `SPEC_DIR=\"specs/2026-03-28-http-public-error-contract\" bash scripts/spec-lint.sh`, `bash scripts/precommit-run.sh`
     - [x] Expected result: full suite and spec lint pass with no HTTP contract regression.
     - [x] Logs/metrics to check (if applicable): N/A

## Traceability (optional)

- FR-001 -> T-001, T-003
- FR-002 -> T-001, T-003
- FR-003 -> T-002, T-003
- NFR-001 -> T-003
- NFR-002 -> T-001, T-002, T-003
- NFR-003 -> T-001, T-003
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag: None
- Migration sequencing: update controller mapping first, then tests, then full validation
- Rollback steps: revert controller-local public message mapping if any client-visible regression is detected

## Validation evidence

- `internal/adapters/inbound/http/controllers/` no longer uses `err.Error()` to produce `dto.ErrorResponse`.
- `chain_path.go` now owns the unknown-chain public message instead of forwarding `inport.ErrChainNotSupported.Error()`.
- Each chain-address controller now exposes a single `map...Error(err) -> status, message` block, so `ServeHTTP` keeps request flow while error/status/message contract stays readable in one place.
- Controller tests now assert both status code and `dto.ErrorResponse.Error`.
- `go test ./internal/adapters/inbound/http/controllers ./internal/adapters/inbound/http` passed.
- `go test ./...` passed.
- `SPEC_DIR="specs/2026-03-28-http-public-error-contract" bash scripts/spec-lint.sh` passed.
