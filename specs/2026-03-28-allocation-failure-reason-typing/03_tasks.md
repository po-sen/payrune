---
doc: 03_tasks
spec_date: 2026-03-28
slug: allocation-failure-reason-typing
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
  - 2026-03-27-domain-error-contracts
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
- Rationale: 只做 allocation derivation failure reason 的型別化，不新增 migration、integration、或 outward contract。
- Upstream dependencies (`depends_on`): `2026-03-27-application-error-boundaries`, `2026-03-27-domain-error-contracts`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 這是一條既有 failure path 的 representation refactor，不涉及新流程或 schema。
  - What would trigger switching to Full mode: 若需要改 schema、重做 allocation aggregate 邊界、或改 outward API。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 每個 task 都附具體 `go test` / spec-lint 驗證。

## Milestones

- M1: Introduce a domain typed allocation derivation failure reason.
- M2: Adapt usecase, persistence, and tests; validate and close the spec.

## Tasks (ordered)

1. T-001 - Add a typed allocation derivation-failure reason to the domain model
   - Scope: 在 `internal/domain/valueobjects` 定義 allocation derivation failure reason，並調整 `PaymentAddressAllocation` 改吃 typed reason。
   - Output: allocation entity uses typed failure reason instead of free-form string.
   - Linked requirements: FR-001 / NFR-002 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/domain/... -run 'TestPaymentAddressAllocation|TestPaymentAddressAllocationDerivationFailureReason'`
     - [x] Expected result: domain tests pass with typed derivation failure reason.
     - [x] Logs/metrics to check (if applicable): N/A
2. T-002 - Map derive failures to typed reasons in the allocation usecase
   - Scope: 調整 `AllocatePaymentAddressUseCase` derivation failure path，不再寫入 raw derive error text。
   - Output: usecase marks allocation derivation failures with domain reason codes.
   - Linked requirements: FR-002 / NFR-002 / NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase'`
     - [x] Expected result: allocation usecase tests pass and assert typed failure reasons.
     - [x] Logs/metrics to check (if applicable): N/A
3. T-003 - Serialize and parse typed derivation-failure reasons in allocation stores
   - Scope: 調整 postgres/cloudflarepostgres allocation stores 與相關測試，兼容 legacy raw text。
   - Output: allocation persistence writes typed codes and reads typed reasons safely.
   - Linked requirements: FR-003 / NFR-001 / NFR-002 / NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres`
     - [x] Expected result: allocation store tests pass with typed reason serialization/parsing.
     - [x] Logs/metrics to check (if applicable): N/A
4. T-004 - Run full validation and close the spec
   - Scope: 跑 full test/spec lint，並把 spec 收回 `DONE`。
   - Output: verified allocation failure reason cleanup with final evidence.
   - Linked requirements: FR-001 / FR-002 / FR-003 / NFR-001 / NFR-002 / NFR-003 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./...`, `SPEC_DIR="specs/2026-03-28-allocation-failure-reason-typing" bash scripts/spec-lint.sh`, `bash scripts/precommit-run.sh`
     - [x] Expected result: full suite passes and spec reflects final typed-reason model.
     - [x] Logs/metrics to check (if applicable): N/A

## Traceability (optional)

- FR-001 -> T-001, T-004
- FR-002 -> T-002, T-004
- FR-003 -> T-003, T-004
- NFR-001 -> T-003, T-004
- NFR-002 -> T-001, T-002, T-003, T-004
- NFR-003 -> T-002, T-003, T-004
- NFR-005 -> T-001, T-002, T-003, T-004
- NFR-006 -> T-001, T-002, T-003, T-004

## Rollout and rollback

- Feature flag: None
- Migration sequencing: define domain reason first, adapt usecase next, then persistence/tests, then full validation
- Rollback steps: revert the typed-reason refactor if any persistence compatibility issue appears

## Validation evidence

- `go test ./internal/domain/... -run 'TestPaymentAddressAllocation|TestPaymentAddressAllocationDerivationFailureReason'` passed.
- `go test ./internal/application/usecases -run 'TestAllocatePaymentAddressUseCase'` passed.
- `go test ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres` passed.
- `go test ./internal/domain/... ./internal/application/usecases ./internal/adapters/outbound/persistence/postgres ./internal/adapters/outbound/persistence/cloudflarepostgres` passed.
- `go test ./...` passed.
- `SPEC_DIR="specs/2026-03-28-allocation-failure-reason-typing" bash scripts/spec-lint.sh` passed.
- `bash scripts/precommit-run.sh` passed.
