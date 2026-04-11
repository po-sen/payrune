---
doc: 03_tasks
spec_date: 2026-04-07
slug: ethereum-contract-readiness
mode: Full
status: DONE
owners:
  - codex
depends_on:
  - 2026-04-05-ethereum-usdt-payment-receiving
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - This feature adds a new external integration path in the API process, startup-time failure modes, and new API logging.
- Upstream dependencies (`depends_on`): `2026-04-05-ethereum-usdt-payment-receiving`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`

## Milestones

- M1:
  - Define and wire Ethereum issuance readiness into API bootstrap.
- M2:
  - Enforce startup-time readiness before the API serves requests.
- M3:
  - Add minimal API request/error logs so readiness failures are diagnosable in running environments.

## Tasks (ordered)

1. T-001 - Finalize the Ethereum issuance-readiness spec

   - Scope:
     - Complete the Full-mode spec package for pre-issuance Ethereum contract validation.
   - Output:
     - Lintable spec folder `specs/2026-04-07-ethereum-contract-readiness/`.

- Linked requirements: FR-001, FR-002, FR-003, FR-004, NFR-006
- Validation:
  - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-04-07-ethereum-contract-readiness" bash scripts/spec-lint.sh`
  - [x] Expected result: spec lint passes and all docs agree on Full mode, dependencies, and links.
  - [x] Logs/metrics to check (if applicable): none

1. T-002 - Add an explicit Ethereum issuance readiness checker

   - Scope:
     - Add one Ethereum adapter that validates active factory compatibility plus token read compatibility through Ethereum RPC.
   - Output:
     - Ethereum issuance readiness checker with unit tests for factory and token validation behavior.

- Linked requirements: FR-002, FR-003, FR-004, NFR-001, NFR-005, NFR-006
- Validation:
  - [x] How to verify (manual steps or command): `go test ./internal/adapters/outbound/ethereum ./internal/infrastructure/ethereumcreate2assets ./internal/bootstrap`
  - [x] Expected result: readiness-check tests pass for success, missing code, hash mismatch, missing RPC config, and token read failures.
  - [x] Logs/metrics to check (if applicable): internal error context includes policy/network/check phase.

1. T-003 - Enforce readiness during API bootstrap

   - Scope:
     - Add explicit per-policy `*_ENABLED` env intent, validate static config for enabled policies, and wire the readiness checker into API bootstrap so configured enabled Ethereum policies are validated before the HTTP server starts, without per-request readiness checks in generate/allocate.
   - Output:
     - Updated policy construction, bootstrap wiring, simplified use cases, and tests.

- Linked requirements: FR-001, FR-001A, FR-004, NFR-002, NFR-003, NFR-006
- Validation:
  - [x] How to verify (manual steps or command): `go test ./internal/application/usecases ./internal/adapters/inbound/http/controllers ./internal/bootstrap`
  - [x] Expected result: disabled policies are skipped, enabled-but-misconfigured policies fail startup before readiness, configured enabled Ethereum policies fail startup when readiness validation fails, and generate/allocate no longer perform per-request readiness checks.
  - [x] Logs/metrics to check (if applicable): none

1. T-004 - Update docs and verify repo-wide behavior
   - Scope:
     - Update README/OpenAPI/compose env references if needed and run repo-level validation.
   - Output:
     - Finalized docs plus recorded validation evidence.

- Linked requirements: FR-001, FR-004, NFR-001, NFR-005, NFR-006
- Validation:
  - [x] How to verify (manual steps or command): `go test ./...`, `SPEC_DIR="specs/2026-04-07-ethereum-contract-readiness" bash scripts/spec-lint.sh`, `bash scripts/precommit-run.sh`
  - [x] Expected result: repo tests pass and docs reflect the new readiness requirement.
  - [x] Logs/metrics to check (if applicable): none

1. T-005 - Add minimal API logging for request outcomes and mapped failures
   - Scope:
     - Add one small HTTP middleware for request logs and controller-side error logs for mapped failures without introducing a new logging subsystem.
   - Output:
     - API request logs plus controller failure diagnostics for readiness and other error responses.

- Linked requirements: FR-005, NFR-005, NFR-006
- Validation:
  - [x] How to verify (manual steps or command): `go test ./internal/adapters/inbound/http/... ./internal/bootstrap`
  - [x] Expected result: middleware/controller tests pass and manual API runs emit request/status logs plus mapped failure details.
  - [x] Logs/metrics to check (if applicable): startup readiness failures become visible in `api` logs.

## Validation evidence

- `go test ./internal/adapters/outbound/ethereum ./internal/application/usecases ./internal/adapters/inbound/http/controllers ./internal/bootstrap`
- `go test ./internal/adapters/inbound/http/... ./internal/bootstrap`
- `go test ./...`
- `SPEC_DIR="specs/2026-04-07-ethereum-contract-readiness" bash scripts/spec-lint.sh`
- `bash scripts/precommit-run.sh`

## Traceability (optional)

- FR-001 -> T-001, T-003, T-004
- FR-001A -> T-003, T-004
- FR-002 -> T-001, T-002
- FR-003 -> T-001, T-002
- FR-004 -> T-001, T-002, T-003, T-004
- FR-005 -> T-005
- NFR-001 -> T-002, T-004
- NFR-002 -> T-003
- NFR-003 -> T-003
- NFR-005 -> T-002, T-004, T-005
- NFR-006 -> T-001, T-002, T-003, T-004, T-005

## Rollout and rollback

- Feature flag:
  - No dedicated feature flag; rollout follows deployment of the new API binary.
- Migration sequencing:
  - No DB migration required.
- Rollback steps:
  - Revert the API binary to the previous version if readiness checks need to be disabled.
