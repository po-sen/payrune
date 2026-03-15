---
doc: 00_problem
spec_date: 2026-03-15
slug: infrastructure-driver-extraction
mode: Quick
status: DONE
owners:
  - payrune-team
depends_on: []
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: null
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Problem & Goals

## Context

- Background:
  - `internal/infrastructure` currently only contains `di`, while some low-level runtime and
    connection setup concerns are still mixed into DI or adapter packages.
- Users or stakeholders:
  - payrune maintainers who want to see whether a small `drivers` extraction improves clarity
    without creating a vague infrastructure dumping ground.
- Why now:
  - The repo is evaluating whether `drivers` should exist at all, and two concrete low-level
    concerns now stand out:
    - standalone PostgreSQL open-and-ping setup
    - Cloudflare Postgres raw JS bridge calls
    - Cloudflare webhook raw JS bridge calls

## Constraints (optional)

- Technical constraints:
  - The extraction must not move adapter mapping or use-case wiring out of DI.
  - Cloudflare adapter semantics must stay in adapter packages; only the raw bridge layer may move.

## Problem statement

- Current pain:
  - `container.go`, `poller_container.go`, and `receipt_webhook_dispatcher_container.go` each parse
    `DATABASE_URL`, open a PostgreSQL connection, and run the same ping check inline.
- Evidence or examples:
  - The repeated `sql.Open` and `db.PingContext` sequence lives in multiple DI files even though it
    is pure low-level connection setup rather than runtime composition policy.
  - `internal/adapters/outbound/persistence/cloudflarepostgres/js_bridge_*.go` owns the raw
    JS/WASM bridge implementation even though the rest of the package is adapter-side persistence
    mapping.
  - `internal/adapters/outbound/webhook/cloudflare_payment_receipt_status_notifier_bridge_*.go`
    owns raw JS/WASM webhook posting even though the notifier adapter itself should remain the
    place where webhook semantics live.

## Goals

- G1:
  - Introduce concrete infrastructure drivers for low-level PostgreSQL and Cloudflare Postgres
    runtime mechanics.
- G2:
  - Reduce repeated standalone database connection setup in DI containers without changing runtime
    behavior.
- G3:
  - Move the raw Cloudflare Postgres JS bridge implementation out of adapter code while keeping the
    adapter's persistence mapping intact.
- G4:
  - Move the raw Cloudflare webhook JS bridge implementation out of adapter code while keeping the
    notifier adapter intact.

## Non-goals (out of scope)

- NG1:
  - Moving repository/store adapters out of `internal/adapters/outbound/persistence`.
- NG2:
  - Creating a generic `drivers` abstraction for every external dependency.
- NG3:
  - Moving Bitcoin Cloudflare Esplora bridges in this iteration.

## Assumptions

- A1:
  - A PostgreSQL connection opener is a legitimate low-level driver concern in this repo.
- A2:
  - Keeping env parsing for non-driver business/runtime config in DI remains the right boundary.
- A3:
  - The Cloudflare Postgres raw JS bridge is a low-level driver concern, but the persistence
    executor/unit-of-work/stores remain adapter concerns.
- A4:
  - The Cloudflare webhook raw JS bridge is a low-level driver concern, but the notifier remains an
    outbound adapter concern.

## Open questions

- Q1:
  - None.

## Success metrics

- Metric:
  - Low-level driver concerns move out of DI/adapter code without creating a vague infrastructure
    dumping ground.
- Target:
  - A standalone PostgreSQL driver package is used by DI containers, the Cloudflare Postgres raw
    JS bridge lives under `internal/infrastructure/drivers`, the Cloudflare webhook raw bridge
    lives under `internal/infrastructure/drivers`, and all relevant Go tests remain green.
