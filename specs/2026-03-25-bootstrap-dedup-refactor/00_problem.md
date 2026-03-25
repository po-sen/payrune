---
doc: 00_problem
spec_date: 2026-03-25
slug: bootstrap-dedup-refactor
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-03-24-architecture-conformance-refactor
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background: `internal/bootstrap` 經過前一輪 architecture cleanup 後，邊界已經正確，但 `*.go` 與 `*_worker.go` 之間仍保留多段重複的 env parsing 與 request/config builder。這些重複目前主要出現在 API receipt terms、poller dispatch config、receipt webhook dispatcher dispatch config 與 notifier config。
- Users or stakeholders: 維護 `internal/bootstrap` 的開發者、未來要持續收斂 runtime wiring 的人。
- Why now: 使用者明確要求繼續整理 `internal/bootstrap` 的重複程式碼，而且希望這輪一次收完。

## Constraints (optional)

- Technical constraints: 不改變現有 process runtime 與 Cloudflare worker runtime 對外行為；不把 runtime builder 硬抽成 generic framework；維持 `internal/bootstrap` 本地化 ownership。
- Timeline/cost constraints: 一次完成可安全共享的 dedup，不留下半套 helper。
- Compliance/security constraints: 不改動 webhook secret handling、DB env validation、poller / dispatcher 現有驗證語意。

## Problem statement

- Current pain: `internal/bootstrap` 現在的重複主要不是 business logic，而是 lookup source 不同但驗證邏輯相同的 parsing code。這讓 process 與 worker path 很容易在 default、validation message 或欄位映射上慢慢分岔。
- Current pain: `internal/bootstrap` 現在的重複主要不是 business logic，而是 lookup source 不同但驗證邏輯相同的 parsing code。這讓 process 與 worker path 很容易在 default、validation message 或欄位映射上慢慢分岔。這輪 dedup 之後也暴露一個 ownership 收尾問題: `api.go` 目前跨檔依賴 `api_worker.go` 內的 Bitcoin XPub env key constants，語意上不夠乾淨。
- Evidence or examples:
  - `internal/bootstrap/api.go` 與 `internal/bootstrap/api_worker.go` 都在組 `PaymentReceiptTermsScope -> confirmations/expiresAfter` map。
  - `internal/bootstrap/poller.go` 與 `internal/bootstrap/poller_worker.go` 都在 parse `POLL_BATCH_SIZE`、`POLL_RESCHEDULE_INTERVAL`、`POLL_CLAIM_TTL`、`POLL_CHAIN`、`POLL_NETWORK`。
  - `internal/bootstrap/receipt_webhook_dispatcher.go` 與 `internal/bootstrap/receipt_webhook_dispatcher_worker.go` 都在 parse dispatch batch/ttl/retry/max-attempts。
  - receipt webhook notifier config 在 process 與 worker 兩側也都各自 parse `secret`、`timeout`、`insecureSkipVerify`。

## Goals

- G1: 抽出 bootstrap-local shared parsing/builder helper，消除 process / worker 間高訊號重複。
- G2: 保留 runtime-specific builder 與 container ownership，不為 dedup 引入新的 generic framework。
- G3: 用測試鎖住 dedup 後的 validation/default 行為，避免 process / worker 路徑再分岔。

## Non-goals (out of scope)

- NG1: 不合併 process 與 worker 的 runtime/container builder。
- NG2: 不改動 `internal/application`、`internal/domain`、`internal/infrastructure` 邊界。

## Assumptions

- A1: 目前最值得抽的是 parsing / config builder，不是整個 runtime wiring。
- A2: `2026-03-24-architecture-conformance-refactor` 已完成，可以作為本 spec 的前置依賴。

## Open questions

- Q1: 無。
- Q2: 無。

## Success metrics

- Metric: bootstrap duplication reduction
- Target: API、poller、receipt webhook dispatcher 的 process / worker 共享 parsing/builder helper，且不改變既有測試語意。
- Metric: bootstrap runtime locality
- Target: `internal/bootstrap` 不新增新的 `shared/`、`common/`、`framework/` 子目錄；shared helper 仍留在 owning bootstrap file 或同 package 小型 helper。
