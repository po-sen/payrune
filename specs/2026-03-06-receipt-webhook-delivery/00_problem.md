---
doc: 00_problem
spec_date: 2026-03-06
slug: receipt-webhook-delivery
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-receipt-status-change-notification
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background:
  - Receipt status-change events are now persisted as `pending` outbox rows, but nothing delivers them to the wallet/backend receiver yet.
- Users or stakeholders:
  - Platform/backend team operating payment collection and the internal wallet/backend service receiving status updates.
- Why now:
  - The receiving side is fixed and controlled by the platform, so webhook delivery can be implemented with a single configured endpoint instead of per-request callback URLs.

## Problem statement

- Current pain:
  - Status transitions are durable in DB but do not leave the system.
- Evidence or examples:
  - `payment_receipt_status_notifications` only stores `pending` rows and has no delivery lifecycle fields.
  - There is no dispatcher worker, notifier adapter, or runtime config for webhook delivery.

## Goals

- G1:
  - Deliver pending receipt status-change events to one fixed webhook endpoint configured by environment variables.
- G2:
  - Run delivery in a dedicated worker/container separate from blockchain pollers.
- G3:
  - Support retry with lease-based claiming and a terminal failed state to avoid infinite delivery loops.
- G4:
  - Sign webhook payloads so the receiving side can verify authenticity.

## Non-goals (out of scope)

- NG1:
  - Supporting multiple webhook targets or per-merchant callback URLs.
- NG2:
  - Adding user-configurable webhook management APIs.
- NG3:
  - Changing receipt status transition rules or first-phase outbox enqueue behavior.

## Assumptions

- A1:
  - The receiving endpoint is fixed and owned by the same platform.
- A2:
  - Delivery semantics are at-least-once; the receiver will deduplicate by notification ID.
- A3:
  - Fixed endpoint URL and secret belong in runtime config, not in database rows.

## Success metrics

- Metric:
  - Pending webhook events are drained by an independent worker.
- Target:
  - A successful delivery moves the row from `pending` to `sent`.
- Metric:
  - Delivery failures do not retry forever.
- Target:
  - Rows stop retrying after configured `max_attempts` and move to `failed`.
- Metric:
  - Polling and webhook delivery remain operationally independent.
- Target:
  - Webhook delivery runs through its own binary/container and does not call external webhooks from the receipt poller.
