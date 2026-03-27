---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Outbound port method:
- A public method on an outbound adapter that satisfies an interface in `internal/application/ports/outbound`.

## Out-of-scope behaviors

- OOS1: Constructor/bootstrap/configuration errors in `NewXXX(...)` remain package-local.
- OOS2: Adapter-private helper errors that never cross the outbound port boundary do not need promotion.

## Functional requirements

### FR-001 - Outbound port methods must expose only port-defined errors

- Description: Any public outbound adapter method that implements an `outport` interface must return only errors defined by the corresponding outbound port contract, or wrapped errors that are stably matchable via `errors.Is(..., outport.Err...)`.
- Acceptance criteria:
  - [ ] Public methods implementing `ChainAddressDeriver`, `IssuedPaymentAddressDeriver`, `BlockchainReceiptObserver`, `ChainReceiptObserver`, or `PaymentReceiptStatusNotifier` no longer return adapter-local raw `errors.New(...)` / `fmt.Errorf(...)` values directly.
  - [ ] Invalid input, unsupported chain/network, missing runtime configuration, and dependency failures that can cross the port boundary are represented by `outport.Err...`.
- Notes: Internal helper and constructor errors are out of scope unless they are returned directly by a port method.

### FR-002 - Constructor and adapter-private errors remain local

- Description: Do not over-promote constructor or helper errors into shared contract errors when they never cross an outbound port boundary.
- Acceptance criteria:
  - [ ] `NewXXX(...)` and internal helper functions may still use package-local errors where appropriate.
  - [ ] The refactor does not create a global adapter error catalog disconnected from port ownership.
- Notes: This preserves the repo preference for ownership and locality.

### FR-003 - Port contracts must become testable and explicit

- Description: The outbound port packages must define the shared sentinel errors needed by public adapter methods, and tests must verify those contracts.
- Acceptance criteria:
  - [ ] Relevant outbound port definition files declare the shared `outport.Err...` values needed by current implementations.
  - [ ] Adapter tests assert `errors.Is(...)` against `outport.Err...` for port-facing error paths.
- Notes: Focus on current concrete ports; do not generalize for hypothetical providers.

## Non-functional requirements

- Performance (NFR-001): The refactor must not add extra IO or change polling/derivation/notifier runtime complexity.
- Availability/Reliability (NFR-002): Existing success/failure behavior of the outbound adapters must remain unchanged aside from error identity cleanup.
- Security/Privacy (NFR-003): No new raw vendor/internal detail should be promoted into a shared outward contract beyond port-scoped sentinel meanings.
- Compliance (NFR-004): N/A.
- Observability (NFR-005): Existing wrapped dependency errors may still preserve internal cause chains for logs/tests while exposing stable `errors.Is(...)` matches.
- Maintainability (NFR-006): Port error ownership must be local to the relevant `outport` file so future adapters can implement the same contract without copying strings.

## Dependencies and integrations

- External systems: Bitcoin Esplora, Ethereum RPC, webhook delivery targets.
- Internal services: `internal/application/ports/outbound`, usecase tests, bootstrap wiring that constructs outbound adapters.
