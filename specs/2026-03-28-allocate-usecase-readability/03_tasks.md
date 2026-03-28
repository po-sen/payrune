---
doc: 03_tasks
spec_date: 2026-03-28
slug: allocate-usecase-readability
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
  test_plan: null
---

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale: 只做 allocate usecase readability cleanup，不改 integration、data model、或 outward API contract。
- Upstream dependencies (`depends_on`): `2026-03-26-allocate-usecase-decomposition`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 這輪沒有新 integration、沒有 schema 改動、也不改 usecase contract。
  - What would trigger switching to Full mode: 若要同時重寫 transaction model、domain policy、或 outbound port shape。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 每個 task 都附具體驗證指令。

## Milestones

- M1: Remove side-effect error state from allocation transaction flow.
- M2: Verify unchanged allocate behavior and close the spec.

## Tasks (ordered)

1. T-001 - Simplify the allocation transaction failure path
   - Scope: 調整 `issueAllocation(...)` 與其緊鄰 helper，移除 derivation failure 的外部 side-effect error state，讓交易 callback 本身回傳最終 outward error。
   - Output: straighter allocation transaction flow in `allocate_payment_address_use_case.go`.
   - Linked requirements: FR-001 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): inspect `internal/application/usecases/allocate_payment_address_use_case.go` and run `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase'`
     - [x] Expected result: derivation failure no longer depends on outer mutable error state and allocate tests pass.
     - [x] Logs/metrics to check (if applicable): N/A
2. T-002 - Run full validation and close the spec
   - Scope: 跑 full suite、spec lint、precommit，確認 readability cleanup 不影響既有行為。
   - Output: final validated allocate usecase readability refactor and spec updated to `DONE`.
   - Linked requirements: FR-002 / NFR-001 / NFR-002 / NFR-003 / NFR-005
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`, `SPEC_DIR="specs/2026-03-28-allocate-usecase-readability" bash scripts/spec-lint.sh`, `bash scripts/precommit-run.sh`
     - [x] Expected result: full suite and repo validation pass.
     - [x] Logs/metrics to check (if applicable): N/A

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001, T-002
- NFR-001 -> T-002
- NFR-002 -> T-002
- NFR-003 -> T-002
- NFR-005 -> T-002
- NFR-006 -> T-001

## Rollout and rollback

- Feature flag:
- Migration sequencing: simplify transaction failure path first, then run focused/full validation
- Rollback steps:

## Validation evidence

- `internal/application/usecases/allocate_payment_address_use_case.go` no longer uses `derivationFailureErr` side-effect state outside the transaction callback.
- Derivation failure now flows through `handleDerivationFailure(...)`, which persists failure state and returns the same outward error the caller should observe.
- `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase'` passed.
- `go test ./...` passed.
- `SPEC_DIR="specs/2026-03-28-allocate-usecase-readability" bash scripts/spec-lint.sh` passed.
- `bash scripts/precommit-run.sh` passed.
