---
doc: 03_tasks
spec_date: 2026-03-28
slug: outbound-review-clarity
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-outbound-port-error-conformance
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
- Rationale: 只做 naming/locality 與 repo guidance 補強，不涉及新 data flow 或 integration。
- Upstream dependencies (`depends_on`): `2026-03-27-outbound-port-error-conformance`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 沒有新系統設計，只有 reviewability cleanup。
  - What would trigger switching to Full mode: 若要移動 package、改 bootstrap architecture、或新增新的 adapter hierarchy。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 每個 task 都有 grep / go test / spec-lint 驗證。

## Milestones

- M1: 收斂 adapter-private interface naming。
- M2: 補上 `AGENTS.md` review heuristics 並驗證全綠。

## Tasks (ordered)

1. T-001 - Rename adapter-private interfaces for review clarity
   - Scope: 將 `internal/adapters/outbound` 裡僅供 adapter 內部使用的 exported interface 改成 unexported naming，並修正呼叫點與測試。
   - Output: renamed adapter-private collaborator interfaces with no behavior changes.
   - Linked requirements: FR-001 / NFR-002 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `rg -n "^type [A-Z][A-Za-z0-9_]* interface" internal/adapters/outbound --glob '*.go'`
     - [x] Expected result: only intentionally outward-facing or package-required exported interfaces remain; adapter-private collaborators are unexported.
     - [x] Logs/metrics to check (if applicable): N/A
2. T-002 - Add explicit review heuristics to AGENTS
   - Scope: 補強 `AGENTS.md` 的 error ownership / review trigger 規則，讓 reviewer 能快速判斷哪些錯誤該收斂到 `outport`。
   - Output: updated AGENTS guidance for outbound port vs adapter-private collaborator review.
   - Linked requirements: FR-002 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `rg -n "outport.Err|constructor|adapter-private collaborator|implements internal/application/ports/outbound" AGENTS.md`
     - [x] Expected result: AGENTS documents the review rule in concrete terms.
     - [x] Logs/metrics to check (if applicable): N/A
3. T-003 - Run validation and close the spec
   - Scope: 跑受影響測試、full suite、spec lint，並將 spec 收回 `DONE`。
   - Output: passing validation and completed spec.
   - Linked requirements: FR-001 / FR-002 / NFR-001 / NFR-002 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `go test ./internal/adapters/outbound/...`, `go test ./...`, `SPEC_DIR="specs/2026-03-28-outbound-review-clarity" bash scripts/spec-lint.sh`, `bash scripts/precommit-run.sh`
     - [x] Expected result: all tests pass after renames and doc updates.
     - [x] Logs/metrics to check (if applicable): N/A

## Validation evidence

- `rg -n "^type [A-Z][A-Za-z0-9_]* interface" internal/adapters/outbound --glob '*.go'` returned no matches.
- `rg -n "outport.Err|constructors, bootstrap/configuration paths|adapter-private collaborators|application boundaries" AGENTS.md` matched the new review guidance.
- `go test ./internal/adapters/outbound/...` passed.
- `go test ./...` passed.
- `SPEC_DIR="specs/2026-03-28-outbound-review-clarity" bash scripts/spec-lint.sh` passed.
- `bash scripts/precommit-run.sh` passed.

## Traceability (optional)

- FR-001 -> T-001, T-003
- FR-002 -> T-002, T-003
- NFR-001 -> T-003
- NFR-002 -> T-001, T-003
- NFR-005 -> T-001, T-003
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag: None.
- Migration sequencing: Rename internal collaborator interfaces first, then AGENTS, then validate.
- Rollback steps: Revert the reviewability cleanup if any exported call site or tests regress.
