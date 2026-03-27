---
doc: 00_problem
spec_date: 2026-03-27
slug: outbound-port-error-conformance
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-27-application-error-boundaries
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: null
---

# Problem & Goals

## Context

- Background: `internal/application` 的 inbound/outbound error ownership 已經先收斂，但使用者發現 `internal/adapters/outbound` 的公開 port method 仍直接回很多 adapter-local `errors.New(...)` / `fmt.Errorf(...)`。這代表同一個 outbound port boundary 上，同時混用 shared contract error 與 ad-hoc adapter error。
- Users or stakeholders: 維護 usecase 與 outbound adapter 邊界的人，以及後續要判讀 port contract 的開發者。
- Why now: 使用者明確要求除了 `NewXXX(...)` 這類 constructor/configuration path 之外，其餘 outbound port method 都不應再洩漏 adapter 自定義 error，而要完全對齊 port contract。

## Constraints (optional)

- Technical constraints: 不改變 usecase 對外行為；只收斂 outbound port error contract。constructor/bootstrap error 可維持 package-local。
- Timeline/cost constraints: Quick mode，一次做完整 sweep，不拆多輪。
- Compliance/security constraints: 不放寬既有 input validation 或 dependency failure handling。

## Problem statement

- Current pain: 多個 outbound adapter 的公開 port method 直接回 raw adapter error，讓 application 無法只依賴 `outport.Err...` 來理解 outbound contract，也讓 error ownership 與 repo 規則不一致。
- Evidence or examples:
  - `internal/adapters/outbound/blockchain/multi_chain_address_deriver.go` 的 `DeriveAddress(...)` 直接回 `"chain is invalid"`、`"network is invalid"`、`"not configured for chain"`。
  - `internal/adapters/outbound/blockchain/multi_chain_receipt_observer.go` 的 `ObserveAddress(...)` / `FetchLatestBlockHeight(...)` 會回傳 `resolveObserver(...)` 的 raw error。
  - `internal/adapters/outbound/bitcoin/esplora_receipt_observer.go` 的 `ObserveAddress(...)` 直接回 input validation error。
  - `internal/adapters/outbound/ethereum/rpc_receipt_observer.go` 的 `ObserveAddress(...)` 同樣直接回 input validation error。

## Goals

- G1: 讓 outbound port 的公開 method 只回傳 port 定義的 shared error contract，或可被 `errors.Is(..., outport.Err...)` 穩定辨識的包裝。
- G2: 保留 constructor/bootstrap/configuration path 的 package-local error，不把所有 adapter error 一律升成 shared catalog。
- G3: 讓 chain/address/receipt observer 這幾組 port 的 contract 明確、可一致測試。

## Non-goals (out of scope)

- NG1: 不把 adapter 內部 helper、parser、scan、constructor 的所有 error 都升成 `outport.Err...`。
- NG2: 不重寫 usecase orchestration 或 HTTP/worker outward error mapping。

## Assumptions

- A1: 公開 port method 指的是實作 `internal/application/ports/outbound` interface 的 method，不包含 constructor。
- A2: 若 application 對某些 outbound error 仍只做 generic dependency mapping，也仍應先把該 error 收斂成 port contract，而不是留 adapter-local 字串。

## Open questions

- Q1: 無；邊界已足夠明確。
- Q2:

## Success metrics

- Metric: outbound adapter 的公開 port method 是否仍直接建立 adapter-local raw error。
- Target: `internal/adapters/outbound` 中所有實作 outbound port 的公開 method 都改為只回 port-defined error contract；測試與 spec lint 維持綠燈。
