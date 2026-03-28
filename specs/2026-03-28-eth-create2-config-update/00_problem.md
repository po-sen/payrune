---
doc: 00_problem
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

# Problem & Goals

## Context

- Background: Payrune 已經用 `ETHEREUM_MAINNET_CREATE2_*` 與 `ETHEREUM_SEPOLIA_CREATE2_*` 這四個環境變數控制 ETH CREATE2 收款地址推導設定；本地 Cloudflare 部署檔與 compose 測試部署檔先前有缺值/舊值，之後又暫時填入了不符合 32-byte contract 的 derivation key。
- Users or stakeholders: 維護 Payrune 部署設定的操作人員，以及依賴固定 collector address / derivation key 產生 ETH 收款地址的服務。
- Why now: 使用者先要求把 ETH collector address 統一成 `0x627e9C4B85a1d486e5A2e4f6D313950A9281a466`，後續又確認這批 derivation key 不會直接上線，因此改成有效的隨機 32-byte fixture 值即可。

## Constraints (optional)

- Technical constraints: 只更新既有設定來源，不改 Go bootstrapping、CREATE2 domain logic、或新增新的 env contract。
- Timeline/cost constraints: Quick mode，直接收斂到 repo 內已追蹤的部署 helper 與測試部署設定。
- Compliance/security constraints: collector address 必須精準保留指定值；derivation key 必須改成有效 32-byte hex，且不新增新的 env 名稱或副本。

## Problem statement

- Current pain: repo 內實際會被部署 helper 與本地 compose 載入的 ETH CREATE2 設定先前不一致，且 derivation key 一度不符合 runtime contract，會造成不同部署路徑推導出不同收款地址，或直接讓 CREATE2 salt deriver 不啟用。
- Evidence or examples:
  - `.env.cloudflare` 原本未包含 `ETHEREUM_MAINNET_CREATE2_*` / `ETHEREUM_SEPOLIA_CREATE2_*`。
  - `deployments/compose/compose.test.env` 原本仍使用舊的 mainnet / sepolia collector 與 derivation key 測試值。
  - `internal/adapters/outbound/ethereum/create2_salt_deriver.go` 只接受 32-byte hex derivation key。

## Goals

- G1: 將 repo 內受影響的 ETH CREATE2 collector 設定統一成 `0x627e9C4B85a1d486e5A2e4f6D313950A9281a466`。
- G2: 讓 Cloudflare 部署 helper 與 compose 測試部署都使用有效的 32-byte Ethereum CREATE2 derivation key。
- G3: 讓 mainnet / sepolia 的最終設定在兩條部署路徑中保持一致。

## Non-goals (out of scope)

- NG1: 不修改 Ethereum CREATE2 contract metadata、factory address、或 receiver init-code hash。
- NG2: 不更動 Go usecase、adapter、domain policy、或新增其他網路/鏈的設定。

## Assumptions

- A1: 使用者提供的 collector address 要同時套用到 Ethereum mainnet 與 Sepolia 的 CREATE2 collector 設定。
- A2: 因為這批 derivation key 不會直接上線，mainnet / sepolia 可以使用不同的隨機 32-byte fixture 值。

## Open questions

- Q1: 無
- Q2:

## Success metrics

- Metric: Repo 內 Cloudflare 與 compose 的 ETH CREATE2 設定是否同時滿足 collector 對齊與 derivation key 有效性
- Target: `.env.cloudflare` 與 `deployments/compose/compose.test.env` 中的 collector address 都等於 `0x627e9C4B85a1d486e5A2e4f6D313950A9281a466`，且 mainnet / sepolia derivation key 都是有效 32-byte hex，並通過 spec lint / repo validation
