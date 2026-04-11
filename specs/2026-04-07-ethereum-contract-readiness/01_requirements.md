---
doc: 01_requirements
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

# Requirements

## Glossary (optional)

- Enabled policy:
  - A policy whose explicit operator `ENABLED` flag is set to true.
- Configured policy:
  - An enabled policy whose required static configuration is present and well-formed.
- Ethereum issuance readiness:
  - The pre-issuance check that confirms a configured Ethereum policy has the on-chain contracts needed for safe address generation and allocation.
- Factory compatibility:
  - The requirement that the on-chain factory at the active metadata address has code and that its runtime-code hash matches the checked-in `Create2ReceiverFactory` artifact.
- Token read compatibility:
  - The requirement that an enabled token-backed Ethereum policy points to a deployed contract that responds to the read calls this runtime depends on.

## Out-of-scope behaviors

- OOS1:
  - Bitcoin or non-Ethereum readiness checks.
- OOS2:
  - Background polling/caching of readiness state after API startup.
- OOS3:
  - Automatic on-chain repair when readiness fails.
- OOS4:
  - Broad ERC-20 behavioral certification beyond code presence and required read calls.

## Functional requirements

### FR-001 - Validate Ethereum issuance readiness before generating or allocating addresses

- Description:
  - Enabled Ethereum policies must pass an on-chain readiness check during API startup before the API serves address preview or allocation requests.
- Acceptance criteria:
  - [ ] API bootstrap validates every configured Ethereum issuance policy before the HTTP server starts serving requests.
  - [ ] If any configured Ethereum issuance policy fails readiness validation, API startup fails closed instead of serving requests with a degraded readiness state.
  - [ ] Bitcoin flows remain unaffected and do not require Ethereum RPC configuration.
  - [ ] Disabled Ethereum policies do not require readiness validation and do not block API startup.
- Notes:
  - The readiness check is required only for enabled Ethereum issuance policies.

### FR-001A - Keep enabled semantics limited to static policy configuration

- Description:
  - Operator intent, static config completeness, and startup readiness must be represented as separate decisions.
- Acceptance criteria:
  - [ ] Each address policy has one explicit `*_ENABLED` env that decides whether the operator intends that policy to be active.
  - [ ] Disabled policies are skipped by static-config validation and startup readiness checks.
  - [ ] If an enabled Bitcoin policy is missing or has an invalid xpub, API startup fails closed.
  - [ ] If an enabled native Ethereum CREATE2 policy is missing required CREATE2 static config, API startup fails closed.
  - [ ] If an enabled Ethereum USDT CREATE2 policy is missing or has an invalid `assetReference`, API startup fails closed.
  - [ ] Startup readiness never changes whether a policy is enabled; it only blocks startup for configured enabled Ethereum policies that are not ready.
- Notes:
  - This requirement separates intent (`ENABLED`), configured state, and startup readiness.

### FR-002 - Confirm the configured CREATE2 factory matches the checked-in runtime contract

- Description:
  - Ethereum CREATE2 issuance must only proceed when the active metadata factory for that policy network is deployed and matches the checked-in factory artifact.
- Acceptance criteria:
  - [ ] For configured Ethereum policies, readiness validation checks that the metadata factory address has on-chain code.
  - [ ] For configured Ethereum policies, readiness validation compares the on-chain factory runtime-code hash with the checked-in `Create2ReceiverFactory` runtime-code hash.
  - [ ] If the metadata factory is missing or mismatched, the policy is treated as not ready for issuance.
- Notes:
  - This feature validates the active issuance factory, not every historical row-owned recovery factory.

### FR-003 - Confirm token-backed Ethereum policies have a compatible token contract

- Description:
  - Token-backed Ethereum policies must prove the configured asset-reference contract is deployed and satisfies the read calls required by the runtime.
- Acceptance criteria:
  - [ ] If a configured Ethereum policy has non-empty `assetReference`, readiness validation confirms that address has on-chain code.
  - [ ] Readiness validation confirms `balanceOf(address)` succeeds for the configured asset-reference contract.
  - [ ] Readiness validation confirms `decimals()` succeeds and matches the policy decimals.
  - [ ] Native ETH policies skip token-contract validation when `assetReference` is empty.
- Notes:
  - This rollout is explicit to the current Ethereum token model where `assetReference` is the ERC-20 contract address.

### FR-004 - Reuse explicit Ethereum RPC config and fail startup when readiness cannot be established

- Description:
  - The API process must load Ethereum RPC config explicitly and fail startup when enabled Ethereum readiness cannot be established.
- Acceptance criteria:
  - [ ] API bootstrap reuses the existing network-scoped Ethereum RPC env contract instead of inventing new API-only env names.
  - [ ] If a configured Ethereum policy needs readiness validation but no RPC client is configured for its network, API startup fails closed.
  - [ ] Generate/allocate use cases do not perform per-request readiness RPC checks after startup succeeds.
  - [ ] HTTP success/error mapping for generate/allocate remains unchanged by this feature.
- Notes:
  - Startup may still succeed with partial Ethereum RPC coverage as long as no enabled Ethereum policy depends on the missing network.

### FR-005 - Emit API diagnostics for request outcomes and readiness failures

- Description:
  - The API process must emit enough logs to diagnose issuance failures, including startup-time readiness failures.
- Acceptance criteria:
  - [ ] The public API router logs each request with method, path, status code, and duration.
  - [ ] When generate/allocate map a use-case error to an HTTP error, API logs include the internal error value plus the mapped status code.
  - [ ] Startup readiness failures include enough context to distinguish missing RPC config, factory-code problems, and token-contract problems.
- Notes:
  - This rollout prefers concise text logs with the existing standard logger over a new logging framework.

## Non-functional requirements

- Performance (NFR-001):
  - The added readiness validation should run only during API startup and should not introduce mandatory Ethereum RPC work for individual generate or allocate requests after startup.
- Availability/Reliability (NFR-002):
  - Readiness failures must fail closed for the affected Ethereum policy and must not corrupt idempotency or partial-allocation state.
- Security/Privacy (NFR-003):
  - Public errors and logs must not expose CREATE2 salts, private keys, or raw recovery payloads.
- Compliance (NFR-004):
  - No change.
- Observability (NFR-005):
  - Readiness failures should identify policy id, chain, network, and which contract check failed in internal diagnostics, and API logs should make startup readiness failures visible without attaching a debugger.
- Maintainability (NFR-006):
  - Keep the implementation explicit to Ethereum issuance readiness; do not introduce a generic health-registry abstraction for hypothetical future chains.

## Dependencies and integrations

- External systems:
  - Ethereum JSON-RPC endpoints already configured via `ETHEREUM_MAINNET_RPC_*` / `ETHEREUM_SEPOLIA_RPC_*`
- Internal services:
  - API bootstrap
  - address generation and allocation use cases
  - `internal/infrastructure/ethereumcreate2assets`
  - Ethereum adapter package
