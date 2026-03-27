---
doc: 03_tasks
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

# Task Plan

## Mode decision

- Selected mode: Quick
- Rationale: 這輪是單一 outbound adapter slice 的 ownership cleanup，不新增 integration / schema / runtime flow；需要 test plan 鎖住 dispatch / chain-specific behavior 不變即可。
- Upstream dependencies (`depends_on`): `2026-03-27-application-error-boundaries`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 不改 architecture layer shape，只收斂 `blockchain` / `bitcoin` / `ethereum` 三者的責任分界。
  - What would trigger switching to Full mode: 若需要改 application port、domain model、或新增新的 provider/runtime path。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 不適用，本 spec 產出 test plan。

## Milestones

- M1: 將 ownership 目標更新到 spec。
- M2: 拆出 chain-specific issued derivers 並讓 blockchain 只做 dispatch。
- M3: 驗證與 spec closeout。

## Tasks (ordered)

1. T-001 - Define neutral dispatch target
   - Scope: 更新 spec，明確 `blockchain` 只負責 multi-chain dispatch，chain-specific issued derivation 回到 `bitcoin` / `ethereum`。
   - Output: 清楚的 package ownership 目標。
   - Linked requirements: FR-001, FR-002, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-27-issued-address-deriver-decoupling" bash scripts/spec-lint.sh`
     - [ ] Expected result: spec-lint 通過，ownership decision 在 spec 中明確。
     - [ ] Logs/metrics to check (if applicable):
2. T-002 - Split issued address derivation by chain
   - Scope: 將 bitcoin / ethereum issued derivation 搬回各自 package，並把 `blockchain` 改成 multi-chain issued deriver dispatch。
   - Output: `blockchain` 中立、chain-specific owner 清楚、bootstrap wiring 仍可直讀。
   - Linked requirements: FR-001, FR-002, FR-003, NFR-001, NFR-002, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./internal/adapters/outbound/blockchain ./internal/bootstrap/...`
     - [ ] Expected result: blockchain adapter 與 bootstrap 測試通過，create2 issuance 行為不變。
     - [ ] Logs/metrics to check (if applicable):
3. T-003 - Final verification and spec closeout
   - Scope: 跑全 repo 驗證並更新 spec 狀態。
   - Output: 綠燈驗證與 DONE frontmatter。
   - Linked requirements: FR-001, FR-002, FR-003, NFR-002, NFR-003, NFR-005, NFR-006
   - Validation:
     - [ ] How to verify (manual steps or command): `go test ./...`、`SPEC_DIR="specs/2026-03-27-issued-address-deriver-decoupling" bash scripts/spec-lint.sh`
     - [ ] Expected result: 所有驗證通過，spec 狀態一致。
     - [ ] Logs/metrics to check (if applicable): 無。

## Traceability (optional)

- FR-001 -> T-001, T-002, T-003
- FR-002 -> T-001, T-002, T-003
- FR-003 -> T-002, T-003
- NFR-001 -> T-002
- NFR-002 -> T-002, T-003
- NFR-003 -> T-003
- NFR-005 -> T-002, T-003
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag: 無。
- Migration sequencing: 先定義 collaborator contract，再改 constructor / tests / bootstrap wiring，最後跑全量驗證。
- Rollback steps: revert 本輪 `issued_payment_address_deriver`、bootstrap wiring、tests、與 spec 變更；不涉及 schema 或 config migration。
