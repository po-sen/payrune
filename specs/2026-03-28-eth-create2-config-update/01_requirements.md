---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Term: ETH CREATE2 deployment settings
- Definition: 由 repo 內部署 helper 或 compose env 檔提供的 `ETHEREUM_MAINNET_CREATE2_*` 與 `ETHEREUM_SEPOLIA_CREATE2_*` 變數。

## Out-of-scope behaviors

- OOS1: 不新增新的部署流程、secret 同步腳本、或環境變數名稱。
- OOS2: 不調整 README 內對 derivation key 格式的通用說明。

## Functional requirements

### FR-001 - Update Cloudflare deploy inputs

- Description: `.env.cloudflare` 必須包含 Ethereum mainnet 與 Sepolia 的 CREATE2 collector / derivation key 設定；collector 使用指定值，derivation key 使用有效的隨機 32-byte fixture 值。
- Acceptance criteria:
  - [ ] `ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS=0x627e9C4B85a1d486e5A2e4f6D313950A9281a466`
  - [ ] `ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS=0x627e9C4B85a1d486e5A2e4f6D313950A9281a466`
  - [ ] `ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY=0x10b7a8d60db72f48e9e41e17d3e6f0b89abe1a802644c9cb7cf1f8064569ba57`
  - [ ] `ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY=0x132a4fb6dc5766792c7fe25f9d42d3c3079331715ffdf6e9ea1a8bfcfe378d75`
- Notes: `.env.cloudflare.example` 保持 generic 範例用途，不嵌入實際 secret 值。

### FR-002 - Update compose deploy inputs

- Description: `deployments/compose/compose.test.env` 必須把 Ethereum mainnet 與 Sepolia 的 CREATE2 collector / derivation key 測試部署值更新成最終 fixture 設定。
- Acceptance criteria:
  - [ ] `deployments/compose/compose.test.env` 的四個 `ETHEREUM_*_CREATE2_*` 變數都存在。
  - [ ] 兩個 collector address 都等於 `0x627e9C4B85a1d486e5A2e4f6D313950A9281a466`。
  - [ ] `ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY=0x10b7a8d60db72f48e9e41e17d3e6f0b89abe1a802644c9cb7cf1f8064569ba57`
  - [ ] `ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY=0x132a4fb6dc5766792c7fe25f9d42d3c3079331715ffdf6e9ea1a8bfcfe378d75`
- Notes: 這個檔案是 `make up` 預設會載入的 compose env。

## Non-functional requirements

- Performance (NFR-001): N/A，本次變更不得新增任何 runtime 查詢、輪詢、或部署前額外步驟。
- Availability/Reliability (NFR-002): 受影響的兩條部署路徑都必須保留原有 env 變數名稱，避免因 rename 導致讀值失敗。
- Security/Privacy (NFR-003): derivation key 只能寫入既有 secret/env 欄位，不新增額外副本或替代名稱，且必須是有效的 32-byte hex。
- Compliance (NFR-004): N/A，無新增合規需求。
- Observability (NFR-005): N/A，無新增 log / metric / trace 行為。
- Maintainability (NFR-006): 每個 network 的 collector / derivation key 必須在受影響檔案間保持一致，避免同 network 設定 drift。

## Dependencies and integrations

- External systems: Cloudflare Worker deploy helper (`make cf-up` / `scripts/cf-payrune-worker-deploy.sh`)、Docker Compose deploy (`make up`)。
- Internal services: `internal/bootstrap/api.go` 既有 env contract，無需改動。
