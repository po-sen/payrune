---
doc: 03_tasks
spec_date: 2026-03-30
slug: compose-eth-sepolia-split
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
  - 2026-03-28-eth-create2-config-update
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
- Rationale: 這次只調整 Compose overlay 分層與 env ownership，不涉及 schema、Go runtime contract、或新 integration。
- Upstream dependencies (`depends_on`): `2026-03-20-create2-eth-payment-receiving`, `2026-03-28-eth-create2-config-update`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 變更集中在兩個 Compose 檔案，沒有新的資料流或架構決策需要額外設計文檔。
  - What would trigger switching to Full mode: 若要新增新的 deployment env contract、改 Compose profiles、或同步修改 Cloudflare / runtime bootstrapping 邏輯。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 每個 task 都附具體 grep / compose config / repo validation 指令。

## Milestones

- M1: Document the Compose split and apply the file changes.
- M2: Validate the merged test config and close the spec.

## Tasks (ordered)

1. T-001 - Move Sepolia API env ownership into the test overlay
   - Scope: 將 `compose.yaml` 的 API `ETHEREUM_SEPOLIA_*` env entries移到 `compose.test.yaml`，並在 test overlay 明確清空 `ETHEREUM_MAINNET_CREATE2_*`。
   - Output: base/test Compose files reflect the intended Ethereum mainnet vs Sepolia split.
   - Linked requirements: FR-001 / FR-002 / NFR-002 / NFR-003 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `rg -n "ETHEREUM_(MAINNET|SEPOLIA)" deployments/compose/compose.yaml deployments/compose/compose.test.yaml`
     - [x] Expected result: `compose.yaml` only keeps Ethereum mainnet API envs, while `compose.test.yaml` owns Sepolia API envs and blank mainnet CREATE2 issuance vars.
     - [x] Logs/metrics to check (if applicable): N/A
2. T-002 - Validate the merged test stack and close the spec
   - Scope: 用 repo 既有 Compose invocation 驗證 merge 結果，再跑 spec lint 與 precommit。
   - Output: validated Compose split with recorded evidence.
   - Linked requirements: FR-003 / NFR-001 / NFR-005 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `docker compose --env-file deployments/compose/compose.test.env -f deployments/compose/compose.yaml -f deployments/compose/compose.test.yaml config`, `SPEC_DIR="specs/2026-03-30-compose-eth-sepolia-split" bash scripts/spec-lint.sh`, `bash scripts/precommit-run.sh`
     - [x] Expected result: merged config resolves successfully with blank `ETHEREUM_MAINNET_CREATE2_*` and present Sepolia API envs; repo validation passes.
     - [x] Logs/metrics to check (if applicable): N/A

## Traceability (optional)

- FR-001 -> T-001
- FR-002 -> T-001
- FR-003 -> T-002
- NFR-001 -> T-002
- NFR-002 -> T-001, T-002
- NFR-003 -> T-001
- NFR-005 -> T-002
- NFR-006 -> T-001, T-002

## Rollout and rollback

- Feature flag: None
- Migration sequencing: update Compose files first, validate merged config second, then reuse the existing `make up` workflow
- Rollback steps: revert the two Compose files and rerun the standard Compose config check

## Validation evidence

- `rg -n "ETHEREUM_(MAINNET|SEPOLIA)" deployments/compose/compose.yaml deployments/compose/compose.test.yaml` confirmed the Sepolia API envs moved into `compose.test.yaml` and `ETHEREUM_MAINNET_CREATE2_*` is blanked there.
- `docker compose --env-file deployments/compose/compose.test.env -f deployments/compose/compose.yaml -f deployments/compose/compose.test.yaml config` passed.
- `SPEC_DIR="specs/2026-03-30-compose-eth-sepolia-split" bash scripts/spec-lint.sh` passed.
- `bash scripts/precommit-run.sh` passed.
