---
doc: 02_design
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

# Technical Design

## High-level approach

- Summary:
  - Add one explicit Ethereum issuance-readiness checker to the API process.
  - Add explicit per-policy `*_ENABLED` env flags so operator intent is not inferred from partial config.
  - Reuse the existing Ethereum RPC env configuration already used by the poller.
  - Run readiness validation during API bootstrap after policy construction and env validation, before the HTTP server starts.
  - Fail closed for API startup when any enabled Ethereum policy's metadata factory or token contract does not match runtime expectations.
- Key decisions:
  - Keep this feature explicit to Ethereum issuance instead of inventing a generic multi-chain contract health subsystem.
  - Treat `enabled` as explicit operator intent from env, not as a derived property of partial config.
  - Validate static config only for explicitly enabled policies, and fail startup when required config is missing or malformed.
  - Validate the active issuance factory by runtime-code hash against the checked-in `Create2ReceiverFactory` artifact.
  - Validate token-backed Ethereum policies with code existence plus `balanceOf(address)` and `decimals()` read compatibility.
  - Native ETH policies only require factory validation.
  - Reuse the poller's Ethereum RPC config loaders instead of duplicating env-name conventions.

## System context

- Components:
  - API bootstrap
  - Address policy reader
  - generate-address use case
  - allocate-payment-address use case
  - new Ethereum issuance readiness checker adapter
  - checked-in CREATE2 factory artifact metadata
- Interfaces:
  - `GET /v1/chains/{chain}/generate-address`
  - `POST /v1/chains/{chain}/payment-addresses`
  - Ethereum JSON-RPC `eth_getCode` and `eth_call`

## Key flows

- Flow 1:
  - API bootstrap builds the policy catalog from explicit `*_ENABLED` env flags plus per-policy static config values.
  - Disabled policies remain visible in the catalog as disabled and are skipped by later validation phases.
  - API bootstrap loads Ethereum RPC config with the existing poller env keys and constructs an Ethereum issuance readiness checker.
  - Compose wiring should keep the base API service on mainnet-ready defaults, while the test overlay adds Sepolia RPC/env for test issuance readiness.
- Flow 2:
  - After policy construction, API bootstrap validates required static config only for enabled policies.
  - API bootstrap then runs readiness validation only for configured enabled Ethereum policies before wiring the HTTP handler.
- Flow 3:
  - `generate-address` and `allocate-payment-address` keep their existing business validation and derivation paths without per-request readiness RPC calls.
- Flow 4:
  - The readiness checker loads the active issuance metadata for the policy network, fetches factory code, compares runtime-code hash with the checked-in factory artifact, and for token policies also validates token read compatibility.
- Flow 5:
  - On readiness failure, API bootstrap returns an error and the process fails before serving requests.
  - The API middleware logs request outcomes, and startup failure logs include the readiness failure reason.

## Diagrams (optional)

- Mermaid sequence / flow:

  ```mermaid
  sequenceDiagram
    participant Bootstrap
    participant Checker
    participant RPC
    Bootstrap->>Checker: Check(enabled Ethereum policy)
    Checker->>RPC: eth_getCode(factory)
    Checker->>RPC: eth_call/balanceOf + decimals (token policy)
    Checker-->>Bootstrap: ready / not ready
    Bootstrap-->>Bootstrap: start server or fail startup
  ```

## Data model

- Entities:
  - No new persisted domain entities.
- Schema changes or migrations:
  - None.
- Consistency and idempotency:
  - Readiness runs before the API starts, so failed checks do not consume address indexes or create partial DB state.

## API or contracts

- Endpoints or events:
  - No endpoint shape changes.
  - No new generate/allocate response codes are introduced by startup-time readiness validation.

## Backward compatibility (optional)

- API compatibility:
  - Request/response shapes stay the same.
- Data migration compatibility:
  - No DB migration required.

## Failure modes and resiliency

- Retries/timeouts:
  - Use existing RPC timeout config when calling readiness checks.
- Backpressure/limits:
  - Startup validates only enabled Ethereum policies; there is no background scan after the server starts.
- Degradation strategy:
  - Disabled policies are skipped entirely by static-config validation and startup readiness.
  - Missing required static config for an enabled policy fails API startup immediately.
  - Missing or mismatched factory code fails API startup when the affected Ethereum policy is enabled and configured.
  - Missing RPC config for an enabled configured Ethereum network fails API startup.
  - Token-contract read failures fail API startup only when the affected token-backed policy is enabled and configured; unrelated disabled policies remain ignored.

## Observability

- Logs:
  - Add one small API request-log middleware that records method, path, status, and duration.
  - Controllers log mapped internal errors with request method/path plus the public status code.
  - Readiness-check startup errors already include policy id, chain, network, and the failed contract-check phase.
- Metrics:
  - None added in this slice.
- Traces:
  - None added.
- Alerts:
  - Existing API startup/error monitoring can treat repeated readiness bootstrap failures as an operational signal.

## Security

- Authentication/authorization:
  - No auth changes.
- Secrets:
  - Reuse existing Ethereum RPC credentials.
- Abuse cases:
  - Readiness checks must not leak recovery payload internals.
  - Invalid Ethereum asset references must still fail closed before any RPC calls that assume a normalized address.

## Alternatives considered

- Option A:
  - Only validate at API startup and trust that state forever.
- Option B:
  - Run readiness validation on every generate/allocate request.
- Why chosen:
  - Option A matches the desired operational model, keeps request paths simpler, and avoids adding repeated Ethereum RPC cost to every issuance request.
  - Option B adds per-request latency and operational coupling that the user explicitly does not want.

## Risks

- Risk:
  - API startup gains a new operational dependency on per-network RPC configuration when enabled Ethereum policies are present.
- Mitigation:
  - Keep checks narrow, use existing timeout config, and skip disabled Ethereum policies.
- Risk:
  - Factory runtime-code hash checks can fail after intentional redeploys if checked-in artifacts or metadata drift.
- Mitigation:
  - Treat this as fail-closed by design and document that artifact/metadata updates are required before issuance resumes.
