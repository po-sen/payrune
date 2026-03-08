---
doc: 03_tasks
spec_date: 2026-03-08
slug: payment-address-idempotency-key
mode: Full
status: DONE
owners:
  - payrune-team
depends_on:
  - 2026-03-04-policy-payment-address-allocation
links:
  problem: 00_problem.md
  requirements: 01_requirements.md
  design: 02_design.md
  tasks: 03_tasks.md
  test_plan: 04_test_plan.md
---

# Payment Address Idempotency Key - Task Plan

## Mode decision

- Selected mode: Full
- Rationale:
  - The change adds a migration, modifies API contract behavior, and changes concurrency semantics.
- Upstream dependencies (`depends_on`):
  - `2026-03-04-policy-payment-address-allocation`
- Dependency gate before `READY`: every dependency is folder-wide `status: DONE`
- If `02_design.md` is skipped (Quick mode):
  - Why it is safe to skip: not applicable
  - What would trigger switching to Full mode: already switched
- If `04_test_plan.md` is skipped:
  - Where validation is specified (must be in each task): not applicable

## Milestones

- M1:
  - Update the idempotency spec to use a dedicated technical store/table instead of allocation-row storage.
- M2:
  - Implement and verify header-based replay protection with an independent idempotency table.

## Tasks (ordered)

1. T-001 - Update the spec package to use `Idempotency-Key`
   - Scope:
     - Rewrite the spec package so retry behavior uses an independent idempotency table rather than allocation-row storage.
   - Output:
     - `specs/2026-03-08-payment-address-idempotency-key/*.md`
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `SPEC_DIR="specs/2026-03-08-payment-address-idempotency-key" bash scripts/spec-lint.sh`
     - [x] Expected result: spec lint passes and docs consistently describe the dedicated idempotency-table model.
     - [x] Logs/metrics to check (if applicable): none
2. T-002 - Add dedicated idempotency persistence and migration
   - Scope:
     - Create `payment_address_idempotency_keys`.
     - Add claim, lookup, complete, and release operations in a dedicated store.
     - Add lookup of issued allocations by `payment_address_id`.
     - Translate duplicate key claims into an application-usable persistence error.
   - Output:
     - `deployments/postgresql/migrations/000008_payment_address_idempotency_key.*.sql`
     - `internal/application/ports/out/payment_address_allocation_store.go`
     - `internal/application/ports/out/payment_address_idempotency_store.go`
     - `internal/adapters/outbound/persistence/postgres/payment_address_allocation_store.go`
     - `internal/adapters/outbound/persistence/postgres/payment_address_idempotency_store.go`
     - related adapter tests
   - Linked requirements: FR-002, FR-003, FR-004, FR-005, FR-006, NFR-002, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/adapters/outbound/persistence/postgres -count=1`
     - [x] Expected result: idempotency-store claim/replay tests and allocation lookup tests pass.
     - [x] Logs/metrics to check (if applicable): none
3. T-003 - Refactor controller and use-case idempotency-key behavior
   - Scope:
     - Read header in the controller.
     - Reuse existing allocation on same-key same-payload replay via the dedicated idempotency store.
     - Emit `Idempotency-Replayed: true` on replayed success responses while keeping `201 Created`.
     - Return conflict on same-key different-payload replay.
     - Release claimed keys on non-issued outcomes.
     - Keep no-header behavior unchanged.
   - Output:
     - `internal/adapters/inbound/http/controllers/chain_address_controller.go`
     - `internal/application/dto/address_policy.go`
     - `internal/application/ports/in/address_policy_use_cases.go`
     - `internal/application/use_cases/allocate_payment_address_use_case.go`
     - related tests
   - Linked requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, NFR-005, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `GOCACHE=/tmp/go-build go test ./internal/application/use_cases ./internal/adapters/inbound/http/controllers ./internal/infrastructure/di -count=1`
     - [x] Expected result: replay path returns prior allocation with `Idempotency-Replayed: true`, conflicting key reuse maps to `409`, failed issuance releases the key, and DI wiring still compiles.
     - [x] Logs/metrics to check (if applicable): none
4. T-004 - Update API docs and run validation
   - Scope:
     - Update OpenAPI docs, keep Swagger browser usage working for the new header, and run repo validations relevant to the change.
   - Output:
     - `deployments/swagger/openapi.yaml`
     - formatted Go files
     - passing validation commands
   - Linked requirements: FR-001, FR-003, FR-004, FR-006, NFR-001, NFR-002, NFR-006
   - Validation:
     - [x] How to verify (manual steps or command): `gofmt -w internal/adapters/inbound/http/middleware/cors.go internal/adapters/inbound/http/middleware/cors_test.go internal/application/dto/address_policy.go internal/application/use_cases/allocate_payment_address_use_case.go internal/application/use_cases/allocate_payment_address_use_case_test.go internal/adapters/inbound/http/controllers/chain_address_controller.go internal/adapters/inbound/http/controllers/chain_address_controller_test.go && GOCACHE=/tmp/go-build go test ./internal/adapters/inbound/http/middleware ./internal/application/use_cases ./internal/adapters/inbound/http/controllers ./internal/adapters/outbound/persistence/postgres ./internal/infrastructure/di -count=1 && GOCACHE=/tmp/go-build go list ./...`
     - [x] Expected result: Swagger-origin preflight allows `Idempotency-Key`, browser clients can read `Idempotency-Replayed`, replayed success responses include `Idempotency-Replayed: true`, and targeted tests/build pass.
     - [x] Logs/metrics to check (if applicable): none

## Traceability (optional)

- FR-001 -> T-001, T-003, T-004
- FR-002 -> T-001, T-002, T-003
- FR-003 -> T-001, T-002, T-003, T-004
- FR-004 -> T-001, T-002, T-003, T-004
- FR-005 -> T-001, T-002, T-003
- FR-006 -> T-001, T-002, T-003, T-004
- NFR-001 -> T-004
- NFR-002 -> T-002, T-004
- NFR-005 -> T-003
- NFR-006 -> T-001, T-002, T-003, T-004

## Rollout and rollback

- Feature flag:
  - None.
- Migration sequencing:
  - Apply `000008` before relying on idempotency-key replay behavior in shared environments.
- Rollback steps:
  - Revert application changes, then drop `payment_address_idempotency_keys` if rollback requires schema reversal.
