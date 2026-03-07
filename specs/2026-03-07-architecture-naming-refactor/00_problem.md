---
doc: 00_problem
spec_date: 2026-03-07
slug: architecture-naming-refactor
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
  - The codebase now has a few architecture-level names that do not match their actual responsibility.
- Users or stakeholders:
  - Future coding agents, the project owner, and maintainers reviewing architecture consistency.
- Why now:
  - Recent review cycles repeatedly surfaced the same naming problems: outbox state called a repository, a chain selector called a router, and `Chain` vs `ChainID` remaining harder to distinguish than necessary.

## Problem statement

- Current pain:
  - A few key names still blur the intended architecture, which makes the code harder to review and easier to model incorrectly in future changes.
- Evidence or examples:
  - The receipt notification persistence port behaves like an outbox but is named as a repository.
  - The multi-chain observer adapter is named like a router even though it is a selector/dispatcher over chain-specific observers.
  - `Chain` and `ChainID` represent different concepts, but the difference is not obvious enough from the current naming.
  - Tracking and allocation persistence still decide parts of receipt and reservation lifecycle inside SQL-facing adapters.
  - Some use cases still have orchestration issues: inconsistent per-item timestamps, inconsistent dependency fail-fast behavior, and output counters whose names do not match what they count.
  - `AllocatePaymentAddressUseCase` still exposed both a default constructor and a `WithConfig` constructor even though only one constructor surface was actually needed.
  - `AllocatePaymentAddressUseCase` and `RunReceiptPollingCycleUseCase` still keep some remaining business policy in application code, especially issuance defaults, reservation priority, expiry reasons, and lifecycle extensions.
  - `AllocatePaymentAddressUseCase` now has the right responsibilities, but its transaction body is still denser than it needs to be for routine maintenance.
  - `TxStores` is still named too narrowly for a transaction-scoped bundle that already contains mixed persistence collaborators such as stores and outboxes.
  - `AllocatePaymentAddressUseCase` still uses a local `func() time.Time` clock hook instead of the shared outbound clock port already used by the other time-aware use cases.
  - `GenerateAddressUseCase` still depends directly on `BitcoinAddressDeriver`, so its name and route shape look chain-generic while its outbound dependency stays bitcoin-specific.
  - Address generation still wires a bitcoin-only deriver directly into use cases, so adding a second chain would require changing existing runtime composition instead of only adding a new chain-specific adapter and registration.
  - The postgres allocation store, receipt tracking store, and unit-of-work paths still have very thin direct test coverage even though they carry cursor, lease, and transaction semantics.
  - `json_response.go` still relies on indirect coverage only, and the claimed receipt-notification outbox row still lives under `application/dto` even though it is not an external boundary DTO.
  - The claimed receipt-notification outbox row no longer lives under `application/dto`, but the generic `application/messages` package is still broader than the actual ownership of an outbox workflow payload.
  - `BitcoinAddressDeriver` still lives in `application/ports/out` even though it is now only an internal collaborator of the bitcoin outbound adapter, not a core application port.
  - `AddressPolicy` still mixes public policy metadata with derivation/issuance configuration, so listing policies and deriving addresses still depend on one over-broad type.

## Goals

- G1:
  - Rename the receipt notification persistence contract and implementations to outbox-oriented names.
- G2:
  - Rename the multi-chain observer adapter to a name that matches its responsibility.
- G3:
  - Rename the supported-chain value object so it is clearly distinct from `ChainID`.
- G4:
  - Keep the refactor behavior-preserving and limited to names that are materially misleading.
- G5:
  - Move business rules currently embedded in persistence adapters into domain objects and domain policies.
- G6:
  - Rename remaining workflow-oriented persistence contracts away from repository terminology when they are not true aggregate repositories.
- G7:
  - Clean up use case orchestration so timestamps, dependency checks, and output counters match actual runtime behavior.
- G8:
  - Collapse redundant `WithConfig` constructor naming into a single explicit constructor surface.
- G9:
  - Move the remaining allocation-issuance and receipt-lifecycle policy out of use cases into domain policies.
- G10:
  - Make the allocate-payment-address orchestration readable without changing its transaction boundary or behavior.
- G11:
  - Rename transaction-scoped dependency bundles away from `TxStores` to a name that reflects transaction scope rather than one specific collaborator type.
- G12:
  - Align allocate-payment-address clock handling with the shared `outport.Clock` contract already used elsewhere in application code.
- G13:
  - Make `GenerateAddressUseCase` depend on a chain-generic outbound port so the use case boundary matches its chain-scoped API surface.
- G14:
  - Make address-derivation composition extensible so adding a new chain only requires adding a new chain-specific adapter and registering it, without refactoring existing chain adapters or use cases.
- G15:
  - Raise direct test coverage on high-risk postgres persistence and transaction code so SQL state transitions are protected by focused tests.
- G16:
  - Tighten small application-layer naming and test gaps so helper files and message types are placed and tested according to their real responsibility.
- G17:
  - Make application-layer data-shape ownership explicit by keeping use case boundary data in `application/dto` and outbox workflow payloads in a dedicated `application/outbox` package.
- G18:
  - Move bitcoin-only derivation internals out of application ports so core ports are shaped only by application use cases.
- G19:
  - Split public address-policy metadata from derivation/issuance configuration so list and issuance flows stop sharing one over-broad domain type.

## Non-goals (out of scope)

- NG1:
  - Changing runtime behavior, schema behavior, or delivery semantics.
- NG2:
  - Renaming every domain type just for stylistic consistency.

## Assumptions

- A1:
  - A focused rename is better than a broad codebase-wide naming churn.
- A2:
  - The current `assets/` templates and repo validation flow remain the source of truth for this refactor.

## Success metrics

- Metric:
  - The renamed types should better match the architecture described in `AGENTS.md`.
- Target:
  - Outbox, multi-chain adapter, and supported-chain names are unambiguous in code review.
- Metric:
  - Workflow persistence names should reflect store/outbox semantics instead of repository semantics.
- Target:
  - Allocation and receipt tracking persistence no longer use repository naming in production code.
- Metric:
  - The refactor should not change behavior.
- Target:
  - `go test ./... -short -count=1` and `bash scripts/precommit-run.sh` pass after the rename.
