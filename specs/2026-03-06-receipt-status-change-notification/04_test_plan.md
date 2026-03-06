---
doc: 04_test_plan
spec_date: 2026-03-06
slug: receipt-status-change-notification
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-polling-expiration-guard
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Test Plan

## Unit

- TC-001:

  - Linked requirements: FR-001, FR-004
  - Steps:
    - observation 後狀態不變。
  - Expected:
    - 不 enqueue notification。

- TC-002:

  - Linked requirements: FR-001, FR-002, FR-003
  - Steps:
    - observation 造成 `watching -> paid_confirmed`。
  - Expected:
    - enqueue 1 筆 pending 事件，from/to 與金額欄位正確。

- TC-003:

  - Linked requirements: FR-004
  - Steps:
    - expiry path 造成 `* -> failed_expired`。
  - Expected:
    - enqueue 1 筆 pending 事件。

- TC-004:
  - Linked requirements: FR-002
  - Steps:
    - 模擬 enqueue repository 回錯。
  - Expected:
    - use case 回錯，該輪不計入 updated success。

## Integration

- TC-101:

  - Linked requirements: FR-003, NFR-003, NFR-004
  - Steps:
    - migration up 後檢查 table/constraints/index 存在。
  - Expected:
    - schema 與索引符合設計。

- TC-102:

  - Linked requirements: FR-003, NFR-002
  - Steps:
    - 直接呼叫 postgres notification repository enqueue。
  - Expected:
    - DB 成功新增 pending 事件，customer_reference 自動帶入。

- TC-103:
  - Linked requirements: FR-002, NFR-001
  - Steps:
    - 在同一 tx 內故意讓 enqueue 失敗。
  - Expected:
    - tracking 更新與事件寫入皆不提交。

## Functional

- TC-201:
  - Linked requirements: FR-001, FR-002, FR-003, FR-004
  - Steps:
    - 實際執行 polling cycle 觸發一次狀態轉換。
  - Expected:
    - tracking 狀態更新，且 DB 出現對應 pending notification 事件。
