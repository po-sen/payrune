---
doc: 01_requirements
spec_date: 2026-03-06
slug: receipt-polling-expiration-guard
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-06-write-through-receipt-tracking
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Requirements

## Functional requirements

### FR-001 - Persist expiration deadline

- Description:
  - Every receipt tracking row must persist `expires_at`.
- Acceptance criteria:
  - [x] Schema includes non-null `expires_at`.
  - [x] Issue-time registration writes initial expiry.
  - [x] Existing rows are backfilled.

### FR-002 - Terminal expired status

- Description:
  - Expired active rows transition to `failed_expired` and stop observer processing.
- Acceptance criteria:
  - [x] Domain status includes `failed_expired`.
  - [x] Poller checks expiry before observer call.
  - [x] Expired rows are persisted with terminal status and not re-polled as active.

### FR-003 - Dynamic expiry extension by status

- Description:
  - Poller extends expiry only when status transitions into `paid_unconfirmed`.
- Acceptance criteria:
  - [x] Transition into `paid_unconfirmed` extends expiry forward.
  - [x] Repeated polls with unchanged `paid_unconfirmed` status do not re-extend expiry.
  - [x] `partially_paid` and `paid_confirmed` do not trigger expiry extension.

### FR-004 - Configurable expiry windows via env

- Description:
  - Initial and status-based expiry windows must be configurable through environment variables.
- Acceptance criteria:
  - [x] App container reads an env duration for initial issue-time expiry window.
  - [x] Poller container reads an env duration for `paid_unconfirmed` extension window.
  - [x] Invalid/non-positive env values fail fast with explicit errors.
  - [x] Compose overlays expose `PAYMENT_RECEIPT_PAID_UNCONFIRMED_EXPIRY_EXTENSION`.

### FR-005 - Separate claim lease from poll schedule

- Description:
  - Poll claim lock must use `lease_until` and not reuse `next_poll_at`.
- Acceptance criteria:
  - [x] Schema includes nullable `lease_until`.
  - [x] `ClaimDue` claims only rows with expired/empty lease and sets `lease_until = claim_until`.
  - [x] Save paths clear `lease_until` while writing final `next_poll_at`.
  - [x] `next_poll_at` remains the scheduling field only.

## Non-functional requirements

- Reliability (NFR-001):
  - Expiration logic must prevent infinite observer loops on stale rows.
- Maintainability (NFR-002):
  - Expiry lifecycle logic remains in domain/use-case boundaries; SQL only handles persistence/claim filters.
- Operability (NFR-003):
  - Expiry window tuning requires only env updates and container restart, with no code changes.
- Correctness (NFR-004):
  - Claim locking and poll scheduling semantics must be independent and unambiguous in SQL paths.

## Dependencies and integrations

- External systems:
  - PostgreSQL migrations and polling persistence queries.
- Internal services:
  - Allocation issue use case and receipt polling cycle use case.
