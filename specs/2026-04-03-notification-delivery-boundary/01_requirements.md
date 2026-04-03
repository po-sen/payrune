---
doc: 01_requirements
spec_date: 2026-04-03
slug: notification-delivery-boundary
mode: Quick
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-02-domain-model-boundary-cleanup
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Glossary (optional)

- Notification delivery workflow:
  - receipt status webhook 的 outbox delivery lifecycle，例如 `pending`、`sent`、`failed`。
- Payment domain state:
  - payment receipt 本身的 business state，例如 watching / partially paid / paid / expired。

## Out-of-scope behaviors

- OOS1:
  - 不更改 delivery status / failure reason 的 string 值。
- OOS2:
  - 不新增新的 delivery status 或 retry 規則。

## Functional requirements

### FR-001 - Move notification delivery workflow types out of domain

- Description:
  - `PaymentReceiptNotificationDeliveryStatus` 與 `PaymentReceiptNotificationDeliveryFailureReason` 不得再留在 `internal/domain/valueobjects`。
- Acceptance criteria:
  - [ ] 這兩個型別與其 parse/helper API 由 `internal/domain` 移出。
  - [ ] 新位置屬於 `internal/application/outbox` 或等價的 application workflow boundary。
  - [ ] `internal/domain` 不再匯出 notification delivery workflow type。
- Notes:
  - 這兩個型別是否進 DB，不影響它們的 ownership 判定。

### FR-002 - Keep webhook delivery behavior and storage unchanged

- Description:
  - ownership 重分類後，webhook delivery 與 outbox persistence 行為必須保持不變。
- Acceptance criteria:
  - [ ] persisted status / failure reason values 維持原樣。
  - [ ] webhook dispatch use case 與 outbox stores 的 sent / pending / failed 行為不變。
  - [ ] parse / canonicalization 行為在新位置保持相容。
- Notes:
  - 本輪是 boundary cleanup，不是流程改版。

### FR-003 - Keep the refactor local and explicit

- Description:
  - 本輪只做 type ownership 更正，不得為此新增新的 generic bucket 或抽象層。
- Acceptance criteria:
  - [ ] 不新增新的 top-level architecture folder。
  - [ ] 不新增與這兩個型別無關的 helper framework。
  - [ ] call site 更新只限於 application/outbox、use case、outbox persistence、與測試。
- Notes:
  - 目標是讓 reviewer 一眼看出這是 workflow type，不是 payment domain type。

## Non-functional requirements

- Performance (NFR-001):
  - 本輪不得新增任何 DB 或 network round trip。
- Availability/Reliability (NFR-002):
  - receipt webhook dispatch / outbox persistence 既有測試應維持通過。
- Security/Privacy (NFR-003):
  - 不新增新的 secret 或 payload surface。
- Compliance (NFR-004):
  - 不適用。
- Observability (NFR-005):
  - 既有 error mapping 與測試訊號不得退化。
- Maintainability (NFR-006):
  - reviewer 可從 package 位置直接判斷 delivery status/reason 屬於 workflow，而非 domain。

## Dependencies and integrations

- External systems:
  - PostgreSQL / Cloudflare Postgres outbox tables
- Internal services:
  - `internal/application/outbox`
  - `internal/application/usecases`
  - `internal/adapters/outbound/persistence/postgres`
  - `internal/adapters/outbound/persistence/cloudflarepostgres`
