---
doc: 03_tasks
spec_date: 2026-03-28
slug: eth-create2-config-update
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-20-create2-eth-payment-receiving
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
- Rationale: 這次只更新既有 ETH CREATE2 env 值，先對齊 collector，再把無效 derivation key fixture 換成有效 32-byte 值；不新增 integration、schema、或 runtime flow。
- Upstream dependencies (`depends_on`): `2026-03-20-create2-eth-payment-receiving`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: 不改 bootstrapping contract、contract metadata、資料模型、或 secret sync 流程，只更新既有設定值。
  - What would trigger switching to Full mode: 若需新增新的 env contract、改 Cloudflare deploy mechanism、或重設 CREATE2 derivation 規則。
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): 每個 task 都附具體檢查指令與預期結果。

## Milestones

- M1: Prepare a ready-to-code spec for the ETH CREATE2 config change.
- M2: Update deployment env files with the final collector and derivation key values, validate, and close the spec.

## Tasks (ordered)

1. T-001 - Finalize and lint the spec package
   - Scope: 補齊 problem / requirements / tasks，記錄 mode decision、假設、目標值、與驗證方式，並在 dependency gate 成立後把 spec 切到 `READY`。
   - Output: lintable Quick spec folder for `2026-03-28-eth-create2-config-update`.
   - Linked requirements: FR-001 / FR-002 / NFR-002 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-28-eth-create2-config-update" bash scripts/spec-lint.sh`
     - [x] Expected result: spec lint passes and the spec can be marked `READY`.
     - [x] Logs/metrics to check (if applicable): N/A
2. T-002 - Apply the ETH CREATE2 configuration updates
   - Scope: 更新 `.env.cloudflare` 與 `deployments/compose/compose.test.env` 的 Ethereum CREATE2 env 值，使 collector address 對齊，並把 derivation key 修正為有效的 32-byte fixture 值。
   - Output: tracked deployment env files with aligned ETH CREATE2 values.
   - Linked requirements: FR-001 / FR-002 / NFR-002 / NFR-003 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `rg -n "ETHEREUM_(MAINNET|SEPOLIA)_CREATE2_(COLLECTOR_ADDRESS|DERIVATION_KEY)" .env.cloudflare deployments/compose/compose.test.env`
     - [x] Expected result: both files contain the final collector address and the valid mainnet / sepolia 32-byte derivation keys recorded in this spec.
     - [x] Logs/metrics to check (if applicable): N/A
3. T-003 - Run repo validation and close the spec
   - Scope: 執行 repo validation，確認這次設定更新沒有破壞既有檢查流程，完成後把 spec 切到 `DONE`。
   - Output: validated config change and closed spec.
   - Linked requirements: NFR-001 / NFR-002 / NFR-003 / NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `bash scripts/precommit-run.sh`
     - [x] Expected result: default-stage hooks pass.
     - [x] Logs/metrics to check (if applicable): N/A

## Traceability (optional)

- FR-001 -> T-001, T-002
- FR-002 -> T-001, T-002
- NFR-001 -> T-003
- NFR-002 -> T-001, T-002, T-003
- NFR-003 -> T-002, T-003
- NFR-006 -> T-001, T-002, T-003

## Rollout and rollback

- Feature flag:
- Migration sequencing: update tracked env files first, then redeploy with existing helper commands.
- Rollback steps: revert `.env.cloudflare` 與 `deployments/compose/compose.test.env` 的四個 Ethereum CREATE2 env 值並重新部署對應環境。

## Validation evidence

- `SPEC_DIR="specs/2026-03-28-eth-create2-config-update" bash scripts/spec-lint.sh` passed.
- `rg -n "ETHEREUM_(MAINNET|SEPOLIA)_CREATE2_(COLLECTOR_ADDRESS|DERIVATION_KEY)" .env.cloudflare deployments/compose/compose.test.env` confirmed both files use collector `0x627e9C4B85a1d486e5A2e4f6D313950A9281a466`, mainnet key `0x10b7a8d60db72f48e9e41e17d3e6f0b89abe1a802644c9cb7cf1f8064569ba57`, and sepolia key `0x132a4fb6dc5766792c7fe25f9d42d3c3079331715ffdf6e9ea1a8bfcfe378d75`.
- `bash scripts/precommit-run.sh` passed, including secret detection, markdown/json/yaml validation, `golangci-lint`, `go test (short)`, `spec lint`, and `go mod tidy (and verify clean diff)`.
- `.env.cloudflare` remains local-only because it is ignored by `.gitignore`; `deployments/compose/compose.test.env` remains the tracked deployment config changed by this spec.
