---
doc: 01_requirements
spec_date: 2026-03-09
slug: shared-tip-height-polling
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

# Requirements

## Glossary (optional)

- Shared tip height:
  - The latest block height fetched once and reused for multiple address observations in the same poll cycle.

## Out-of-scope behaviors

- OOS1:
  - Changing the receipt polling persistence schema.
- OOS2:
  - Changing public payment status responses.

## Functional requirements

### FR-001 - Fetch tip height once per claimed network

- Description:
  - The receipt polling flow must fetch latest block height at most once for each claimed chain/network pair in one poll cycle.
- Acceptance criteria:
  - [ ] The polling use case resolves one latest block height for each claimed chain/network pair before reusing it across matching address observations.
  - [ ] Address observations on the same claimed chain/network pair receive the same latest block height value during that poll cycle.
  - [ ] A failure fetching latest block height for one chain/network pair is handled as an observation failure for trackings on that pair.
- Notes:
  - The optimization targets repeated tip-height calls, not address history calls.

### FR-002 - Preserve receipt observation behavior

- Description:
  - The observer must continue to compute confirmed and unconfirmed totals using the supplied latest block height without changing receipt status behavior.
- Acceptance criteria:
  - [ ] Confirmation counting still uses latest block height plus transaction block height.
  - [ ] Existing status transitions and polling error persistence remain unchanged.
  - [ ] Multi-chain observer routing still validates chain/network and forwards the shared tip height to the chain-specific observer.
- Notes:
  - This refactor changes orchestration, not business policy.

## Non-functional requirements

- Performance (NFR-001):
  - The successful fresh poll path must avoid redundant latest-block-height calls for repeated addresses on the same chain/network pair within a single poll cycle.
- Availability/Reliability (NFR-002):
  - Polling must still deterministically mark observer failures as polling errors and continue to honor existing retry scheduling.
- Security/Privacy (NFR-003):
  - No new external data is requested beyond the latest block height already used today.
- Compliance (NFR-004):
  - None.
- Observability (NFR-005):
  - Stored `LastObservedBlockHeight` and status API output remain populated as before.
- Maintainability (NFR-006):
  - Poll-cycle orchestration should clearly own shared tip-height fetching rather than hiding the optimization in the Bitcoin adapter.

## Dependencies and integrations

- External systems:
  - Bitcoin Esplora-compatible HTTP API.
- Internal services:
  - `internal/application/use_cases/run_receipt_polling_cycle_use_case.go`
  - `internal/application/ports/out/blockchain_receipt_observer.go`
  - `internal/adapters/outbound/blockchain/multi_chain_receipt_observer.go`
  - `internal/adapters/outbound/bitcoin/esplora_receipt_observer.go`
