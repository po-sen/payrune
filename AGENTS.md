# Project AGENTS

## Mandatory Workflow (Always Apply)

1. Run Spec-Driven Development before coding any feature/change:
   - Create/update `specs/YYYY-MM-DD-slug/` first.
   - Keep spec files as source of truth.
2. Build Go services with:
   - `go-project-layout` rules for directory structure.
   - `clean-architecture-hexagonal-components` rules for layer boundaries and dependency direction.
3. When commit is requested:
   - Use `conventional-commit` rules strictly.
4. Any repository helper automation must live under `scripts/`.

## Project Conventions

- Language/runtime baseline: Go.
- Preferred app shape: single binary under `cmd/<app>/main.go` and private code under `internal/`.
- Architecture baseline: Clean Architecture + Hexagonal (Ports and Adapters).
- Spec lint command:
  - `SPEC_DIR="specs/YYYY-MM-DD-slug" bash scripts/spec-lint.sh`
- Pre-commit validation command:
  - `bash scripts/precommit-run.sh`

## Repo-Specific Agent Guidance

This section is written for the coding agent. Treat it as the repo-specific override layer for the
generic skill snapshots below.

### Precedence

- Use this section to narrow and interpret the generic skills below for this repository.
- Unless a statement here is explicitly wrong, preserve its meaning when rewriting or extending it.
- When a generic skill and this section differ, follow this section for repo-local decisions.

### What this repo optimizes for

- Clear responsibility boundaries over abstract purity.
- Small, explicit Go services over reusable-looking frameworks.
- Concrete names over premature generalization.
- Specs as the source of truth for feature work.
- Code that is easy to review and reason about without long explanation.

### Architecture boundaries

#### `internal/domain`

- Own business rules, invariants, state transitions, validation of domain concepts, and business
  errors.
- Must stay independent from SQL, HTTP, RPC, env parsing, framework types, and transport details.
- If a rule decides what is valid, payable, expired, retryable, or transitionable, it probably
  belongs here.

#### `internal/application`

- Own use-case orchestration only.
- Coordinate transaction boundaries, invoke domain behavior, call outbound ports, and return DTOs.
- May compose multiple steps, but should not become the home for business policy.
- If a use case grows many private helpers or computes state transitions directly, review whether
  domain logic is leaking upward.

#### `internal/adapters`

- Inbound adapters parse and validate transport input, map to commands/DTOs, call application
  ports, and map outputs/errors back to the transport.
- Outbound adapters implement outbound ports and translate core needs into SQL, HTTP, RPC, files,
  or other IO.
- Adapters may do technical validation and protocol mapping, but must not decide business outcomes.

#### `internal/infrastructure/di`, `internal/bootstrap`, `cmd/`

- Own dependency wiring, env parsing, process startup, schedulers, and concrete client
  construction.
- Keep them procedural and thin.
- Do not move policy or business branching into bootstrap code just because it is convenient.

### Port design

- Shape outbound ports around what the application needs, not around vendor SDKs or transport APIs.
- Keep ports small and use-case-focused.
- Separate write-side persistence from read/query concerns when that improves clarity.
- If something is really a delivery pipeline, queue state, or technical process state, the port name
  should reflect that instead of pretending to be a classic aggregate repository.

### Modeling rules

- Use an `Entity` only when the concept has identity, lifecycle, and meaningful business behavior.
- If something mostly mirrors a row, queue item, or delivery state and the rules live elsewhere, do
  not force it into a domain entity.
- Use `Value Object` for small validated concepts defined by value rather than identity.
- Do not create generic abstractions for concepts that only have one real implementation today.

### Persistence naming and transactions

- Use `Repository` when the main job is loading and saving aggregate-like domain objects.
- Use `Store` for technical or process persistence that is not really an aggregate collection.
- Use `Outbox` when the data exists specifically for reliable asynchronous delivery.
- Use `UnitOfWork` to define one transaction boundary across multiple repositories or stores.
- Reusing a shared SQL executor is acceptable when it keeps transaction boundaries clean.
- Repository/store code may enforce persistence invariants, but must not silently own business
  policy.

### Naming and abstraction rules

- Prefer names that describe what the code is today, not what it might support later.
- Prefer explicit chain-specific code over premature multi-chain abstractions.
- Prefer explicit network-specific config over prefix-driven indirection when readability improves.
- Do not introduce routers, registries, or polymorphic wrappers unless there is a real second
  implementation that benefits from them now.
- If a type or function needs a long explanation, assume the model is not clear enough yet.

### Operational preferences

- Prefer domain behavior over large use-case helper chains.
- Prefer bounded due queries over full-table polling scans.
- Prefer Compose-native required env handling over shell-entrypoint validation hacks.
- Prefer committed local test env files for fake non-secret values.

### Review triggers

Stop and reconsider if any of these are true:

- A repository or adapter is deciding business outcomes.
- A use case is calculating state transitions that should live in domain.
- A table-shaped record is being modeled as an entity without domain behavior.
- A generic abstraction exists only for a hypothetical future chain or provider.
- Config becomes harder to read because of prefixes, indirection, or hidden defaults.
- The user says the design feels too abstract or too hard to understand.

## Embedded Skills (Frozen Snapshot)

The following skill definitions are copied into this project so future work does not depend on an external skill registry.

### skill: spec-driven-development

---

name: spec-driven-development
description: >-
Convert rough ideas into Spec-Driven Development artifacts: problem statement, requirements,
design, task plan, and test plan. Use when the user asks for SDD/specs, to clarify requirements,
or to turn a rough idea into an implementable plan before coding.

---

# spec-driven-development

## Purpose

- Turn rough ideas into a clear, verifiable spec package before coding.

## When to Use

- Follow the trigger guidance in the frontmatter description.

## File IO rule

- Always create/update files in the repo (not only chat output).
- Must create `specs/YYYY-MM-DD-slug/` and write the spec files there.
- If the user only wants a draft, still write files, but mark them via the document header:
  - Set `status: DRAFT` in YAML frontmatter.
  - Do NOT add any text before the leading `---` (keeps frontmatter valid).

## Inputs

- Rough idea or problem statement.
- Target users or stakeholders and goals.
- In-scope vs out-of-scope.
- Constraints (tech, time, cost, compliance).
- Integrations or dependencies.
- Non-functional requirements or quality targets.

## Outputs

- A spec folder under `specs/YYYY-MM-DD-slug/`.
- Quick mode required output:
  - `00_problem.md`
  - `01_requirements.md`
  - `03_tasks.md`
- Quick mode optional output:
  - `04_test_plan.md`
- Full mode output:
  - `00_problem.md`
  - `01_requirements.md`
  - `02_design.md`
  - `03_tasks.md`
  - `04_test_plan.md`
- Explicit assumptions and open questions.

## Conventions

- Spec folder: `specs/YYYY-MM-DD-slug/` (slug = kebab-case, short and specific).

### Document header rules

- Every spec file MUST start with YAML frontmatter (`---` ... `---`).
- Fill these fields in every produced file:
  - `spec_date`: real date like `2026-01-31` (templates use `null`)
  - `slug`: real slug like `payment-webhook-retry` (templates use `null`)
  - `mode`: `Quick` or `Full` (match selected mode)
  - `status`: `DRAFT`, `READY`, or `DONE`
  - `owners`: `[]` allowed only in DRAFT; add at least one owner before READY/DONE
  - `depends_on`: `[]` or a list of prerequisite spec folder names (`YYYY-MM-DD-slug`) that must be
    `DONE` before this spec can become `READY`
  - Keep `spec_date`, `slug`, `mode`, `status`, and `depends_on` consistent across docs in the same
    spec folder.
- Links MUST NOT point to non-existent files:
  - Keep a consistent key set in all docs: `problem`, `requirements`, `design`, `tasks`,
    `test_plan`.
  - Use `null` when a doc is not produced (e.g., `links.design: null` in Quick mode).
  - If you later produce the doc, update links in the other spec docs immediately.

### Status lifecycle

- `DRAFT`: spec is being prepared; placeholders or open questions may remain.
- `READY`: spec is complete, spec-lint passes, and implementation can start.
- `DONE`: implementation and validation are complete, and the spec reflects final behavior/scope.

### Cross-spec dependencies

- Use `depends_on` to declare prerequisite specs for this spec.
- Format:

  - No prerequisites: `depends_on: []`
  - With prerequisites:

    ```yaml
    depends_on:
      - 2026-01-20-auth-foundation
      - 2026-01-25-shared-api-contract
    ```

  - Non-empty inline list form is not supported: do not use `depends_on: [a, b]`.

- Dependency gate:
  - Before setting this spec to `READY`, every `depends_on` entry must resolve to an existing folder
    under `specs/` and that folder must be folder-wide `status: DONE`.
  - `depends_on` must not include the current spec's own slug.
- Source of truth:
  - `00_problem.md` is canonical for dependency gate checks.
  - Keep `depends_on` aligned across all spec docs in the same folder to avoid drift.
  - Dependency order follows `00_problem.md`; other docs must match the same order.

### Slug rules

- Source: 3-5 keywords from the problem or title.
- Format: lowercase, kebab-case, no punctuation.
- Remove filler words (a/an/the/of/for/and, etc.).
- Max length: 40 characters.
- Example: `add-user-login`.

### IDs and traceability

- Requirement IDs:
  - Functional: `FR-001`, `FR-002`, ...
  - Non-functional: `NFR-001`, ...
- Task IDs: `T-001`, `T-002`, ...
- Test case IDs: `TC-001`, `TC-002`, ...
- Traceability rule:
  - Every `T-XXX` MUST reference one or more `FR/NFR` IDs.
  - Every `TC-XXX` MUST reference one or more `FR/NFR` IDs.

## Quality bar

- Requirements must be verifiable (acceptance criteria / measurable targets, e.g., p95 latency <=
  200 ms).
- Design must cover: flows, data, contracts, failure modes, observability, security.
- Task plan must be ordered, small, independently verifiable, and traceable.

## Modes

### Quick mode (default for small changes)

Use when:

- 1-2 endpoints / a small feature flag / a simple refactor
- No new integrations, no new persistent data model, no risky rollout

Produce:

- Required:
  - `00_problem.md`
  - `01_requirements.md`
  - `03_tasks.md`
- Optional:
  - `04_test_plan.md` (recommended)

Skip:

- `02_design.md` unless any of these are true:
  - New DB schema / migrations
  - New external integration
  - Non-trivial failure modes / async flow
  - Meaningful NFR impact (latency, availability, security)

### Full mode

Use when any of the "Skip" triggers above apply. Produce all 5 files. If unsure, default to Quick
during scaffolding and re-evaluate after clarifying questions.

## Steps

1. Scaffold first (must happen before writing content):
   - Derive `YYYY-MM-DD` and `slug` using "Slug rules".
   - Create `specs/YYYY-MM-DD-slug/` if missing.
   - Create required Quick-mode files by copying templates (not empty files):
     - `00_problem.md` from `assets/00_problem_template.md`
     - `01_requirements.md` from `assets/01_requirements_template.md`
     - `03_tasks.md` from `assets/03_tasks_template.md`
   - Populate document headers (`spec_date`, `slug`, `mode`, `status`, `owners`, `depends_on`)
     immediately.
     - Default `mode: Quick` and `status: DRAFT` during scaffolding (safe defaults).
     - After mode is decided, update `mode` (and `links`) across all produced files to match.
2. Ask the minimum clarifying questions needed to fill gaps (goal/value, scope, constraints,
   acceptance criteria, integrations, NFRs, and upstream spec dependencies). If answers are missing,
   state assumptions explicitly.
   - If mode is unclear, keep `mode: Quick` from scaffolding and confirm after these questions.
3. Decide mode (Quick or Full) using the triggers under "Modes" and record the decision and
   rationale in `03_tasks.md` under "Mode decision".
   - After deciding mode:
     - Update YAML frontmatter `mode` in every already-produced spec file to match (Quick/Full).
     - If switching to Full:
       - Create `02_design.md` from template and set `links.design` to `02_design.md`.
       - Create `04_test_plan.md` from template and set `links.test_plan` to `04_test_plan.md`.
       - After creating new docs, immediately update their YAML frontmatter fields (`spec_date`,
         `slug`, `mode`, `status`, and `links`) to match the selected mode.
       - Update links in `00_problem.md`, `01_requirements.md`, `03_tasks.md`, and
         `04_test_plan.md`:
         - `links.design: 02_design.md`
         - `links.test_plan: 04_test_plan.md` (in `00_problem.md`, `01_requirements.md`,
           `03_tasks.md`)
     - If staying Quick:
       - Keep `links.design: null`.
       - If you decide to produce `04_test_plan.md`, create it and set `links.test_plan` in
         `00_problem.md`, `01_requirements.md`, and `03_tasks.md`.
       - Keep `links.design: null` in `04_test_plan.md` for Quick mode.
4. Fill `00_problem.md` from `assets/00_problem_template.md` with concrete context, goals,
   non-goals, and success metrics.
5. Fill `01_requirements.md` from `assets/01_requirements_template.md`. Ensure every functional
   requirement has acceptance criteria and NFRs are measurable.
6. If Full mode (or any "Skip" triggers apply):
   - Ensure `02_design.md` exists (create if missing), then fill it from
     `assets/02_design_template.md`.
7. Fill `03_tasks.md` from `assets/03_tasks_template.md`. Order tasks, make each independently
   verifiable, and link tasks back to requirements. Keep task validation steps even when a separate
   test plan exists.
8. If producing `04_test_plan.md`:
   - In Full mode: MUST produce `04_test_plan.md`.
   - In Quick mode: OPTIONAL (recommended) to produce `04_test_plan.md`.
   - Ensure `04_test_plan.md` exists (create if missing), then fill it from
     `assets/04_test_plan_template.md`.
   - Cover unit/integration/e2e as appropriate, plus edge cases and NFR verification.
   - If produced:
     - Set `links.test_plan: 04_test_plan.md` in `00_problem.md`, `01_requirements.md`,
       `03_tasks.md`.
     - If Full mode, set `links.design: 02_design.md` in `04_test_plan.md`.
9. Provide a readiness checklist. Do not change code until the spec package exists. If the user
   requests immediate coding, produce a minimal spec package first and proceed with explicit,
   labeled assumptions (do not invent integrations/constraints silently).
10. If the Ready-to-code checklist is satisfied:
    - Before setting `READY`, run the spec-lint checks below and ensure they pass.
    - If `depends_on` is non-empty, every dependency must already be `DONE`; otherwise keep this
      spec in `DRAFT`.
    - Update `status: READY` in the YAML frontmatter of every produced spec file in the folder (keep
      statuses consistent across docs).
11. After implementation and validation tied to the spec are complete:
    - Update `status: DONE` in the YAML frontmatter of every produced spec file in the folder (keep
      statuses consistent across docs).

## Spec-lint (recommended)

Run these checks against the spec folder before marking `status: READY` or `status: DONE`. The
canonical lint implementation is `scripts/spec-lint.sh`.

```bash
# From repo root:
SPEC_DIR="specs/YYYY-MM-DD-slug" bash skills/spec-driven-development/scripts/spec-lint.sh

# Or from the skill directory:
SPEC_DIR="specs/YYYY-MM-DD-slug" bash scripts/spec-lint.sh
```

## Ready-to-code checklist

### Quick mode checklist

- [ ] `specs/YYYY-MM-DD-slug/` exists and contains `00_problem.md`, `01_requirements.md`,
      `03_tasks.md`
- [ ] Document headers are filled (`spec_date`, `slug`, `mode`, `status`, `owners`, `depends_on`)
      with real values (no `null`/`[]` placeholders except `depends_on: []` when no prerequisites)
- [ ] `owners` includes at least one owner or team
- [ ] Every `depends_on` entry points to an existing `specs/YYYY-MM-DD-slug/` folder and each
      dependency folder is `status: DONE`
- [ ] Frontmatter values are consistent across docs (`spec_date`, `slug`, `mode`, `status`,
      `depends_on`)
- [ ] Every spec doc has the full `links` key set (`problem`, `requirements`, `design`, `tasks`,
      `test_plan`)
- [ ] Mode decision and rationale is recorded in `03_tasks.md`
- [ ] Every `FR-XXX` has acceptance criteria
- [ ] Every `NFR-XXX` is measurable (targets, limits, SLO-like) if applicable
- [ ] Every `T-XXX` links to `FR/NFR` IDs
- [ ] If `04_test_plan.md` is skipped, `03_tasks.md` includes explicit validation steps per task
- [ ] In Quick mode, `links.design` remains `null`
- [ ] All YAML links are valid (either `null` or pointing to existing files)
- [ ] Spec-lint checks pass
- [ ] Set `status: READY` across produced docs (keep statuses consistent)

### Full mode checklist

- [ ] `specs/YYYY-MM-DD-slug/` exists and contains all 5 files
- [ ] Document headers are filled (`spec_date`, `slug`, `mode`, `status`, `owners`, `depends_on`)
      with real values (no `null`/`[]` placeholders except `depends_on: []` when no prerequisites)
- [ ] `owners` includes at least one owner or team
- [ ] Every `depends_on` entry points to an existing `specs/YYYY-MM-DD-slug/` folder and each
      dependency folder is `status: DONE`
- [ ] Frontmatter values are consistent across docs (`spec_date`, `slug`, `mode`, `status`,
      `depends_on`)
- [ ] Every spec doc has the full `links` key set (`problem`, `requirements`, `design`, `tasks`,
      `test_plan`)
- [ ] Every `FR-XXX` has acceptance criteria
- [ ] Every `NFR-XXX` is measurable (targets, limits, SLO-like)
- [ ] Design covers flows, data, contracts, failure modes, observability, security
- [ ] Every `T-XXX` links to `FR/NFR` IDs
- [ ] Every `TC-XXX` links to `FR/NFR` IDs
- [ ] All YAML links are valid (either `null` or pointing to existing files)
- [ ] Spec-lint checks pass
- [ ] Set `status: READY` across all docs (keep statuses consistent)

## Done checklist

- [ ] Implementation tasks in `03_tasks.md` are complete
- [ ] Validation evidence is recorded (tests/manual checks/metrics as applicable)
- [ ] Spec docs reflect final behavior and scope (including any approved changes)
- [ ] Set `status: DONE` across produced docs (keep statuses consistent)

## Notes

- Treat the spec as the source of truth; update the spec before changing code.
- Keep templates minimal. Adapt to existing repo conventions (ADR/RFC/docs) but preserve the section
  structure.
- Example spec package (Full mode): `examples/specs/2026-02-06-lint-pass-example/`.
- Avoid inventing integrations or requirements. Ask or mark as assumptions.
- Prefer concise, testable statements over narrative prose.
- Use `DONE` only after implementation and validation are complete; otherwise keep `READY`.

### skill: go-project-layout

---

name: go-project-layout
description: |
Enforce Go project directory structure using golang-standards/project-layout with pragmatic
defaults and official Go module/package guidance. Use when creating or refactoring Go repos so
code placement, visibility, and boundaries stay consistent.

---

# Go Project Layout

## Purpose

- Place Go code in predictable directories with clear visibility boundaries.
- Keep layouts minimal by default and scale only when project complexity requires it.

## When to Use

- Follow the trigger guidance in the frontmatter description.

## Inputs

- Change request or feature scope.
- Existing tree (`go.mod`, folders, binaries/libraries, deployment targets).
- Whether the repo is single-binary, multi-binary, service, or shared library.
- Existing constraints (monorepo rules, CI paths, proto/openapi generation, container packaging).

## Outputs

- New or updated directory tree that matches the project type.
- Files moved/created in the correct directories with imports fixed.
- Updated Go entry points and package boundaries (`internal` vs `pkg`) with no cycles.
- Validation evidence from `go list ./...` and the chosen test workflow (`go test -short ./...`,
  plus full/e2e runs when required), or a clear reason if skipped.

## Decision Cheatsheet

- Single-binary service: `cmd/<app>/main.go` plus `internal/<domain>/...`.
- Multi-binary repo: `cmd/<app>/main.go` per app plus `internal/<app>/<domain>/...`.
- Reusable library: organize packages under the module root (for example `foo/`, `bar/`), and use
  optional `pkg/<name>` only for intentional, stable public APIs.
- Mixed apps plus library: keep app logic in `internal/`; expose only intentional public APIs in
  `pkg/`.

## Steps

1. Detect current shape before editing:
   - Run `go env GOMOD` and `go env GOWORK`, then scan top-level folders.
   - If `go env GOMOD` is empty or `/dev/null`, the repo likely has no module yet. Initialize with
     `go mod init <module>` unless the task explicitly disallows it.
   - If module initialization is intentionally skipped, record the reason and skip `go list ./...`
     until a module exists.
   - If `go env GOWORK` is non-empty, the repo may be in workspace mode. Keep layout decisions
     scoped to the current module unless the task explicitly targets workspace-level structure.
   - If a module exists, run `go list ./...` and ensure no import cycles.
2. Classify the repo:
   - Single deployable app
   - Multiple deployable binaries
   - Reusable library
   - Mixed (apps + library)
3. Start from the smallest viable layout:
   - Keep `go.mod` and `go.sum` at repo root.
   - Do not add `src/`; place Go code at module root or subfolders.
4. Place executable entry points in `cmd/<app>/main.go`:
   - `cmd` only contains wiring/bootstrap code.
   - Move business logic out of `cmd` into packages.
5. Place non-exported app code in `internal/`:
   - Use `internal/<domain>` or `internal/<app>/<domain>` for services, handlers, repositories, and
     use cases.
   - Treat `internal` as default for code not intended for external import.
   - Decision rule:
     - Single-binary service: prefer `internal/<domain>` to keep structure shallow.
     - Multi-binary or mixed apps: prefer `internal/<app>/<domain>` to avoid collisions.
6. Use `pkg/` only for genuinely reusable public packages:
   - If a package is not meant for outside consumers, keep it in `internal/`.
   - Avoid mirroring everything under both `internal/` and `pkg/`.
   - For any `pkg/` package, require clear ownership, semantic version tags, and compatibility
     commitments.
   - If publishing v2+, ensure the module path uses the `/vN` suffix (for example `.../v2`).
7. Add optional folders only when they have concrete artifacts:
   - `api/` for OpenAPI, protobuf, or API contracts and generator settings only.
   - `configs/` for example/default config files (not secrets).
   - `scripts/` for dev/CI scripts.
   - `build/` and `deployments/` for packaging/deploy manifests.
   - `test/` for black-box/integration test assets; keep unit tests next to code.
   - Do not place handlers, controllers, or server implementations in `api/`.
   - If `test/` contains slow tests, isolate them with build tags (for example
     `//go:build integration` or `//go:build e2e`) or `-short` workflows.
8. Keep package boundaries clean:
   - Prevent cyclic imports (`go list ./...` must pass).
   - Keep package names short, lowercase, and purpose-driven.
   - Prefer composition interfaces near consumers to reduce cross-package coupling.
9. For multi-binary repos, share code through `internal/` first:
   - Use `internal/platform` or `internal/shared` only when sharing is real and stable.
   - Promote to `pkg/` only when external reuse is intentional and versioned.
10. For multi-module needs, split modules deliberately:
    - Default to one module per repo.
    - Introduce extra modules only with explicit release or ownership boundaries.
    - If multiple modules are required, wire them with `go work` and document ownership.
11. Verify after edits:
    - Use the repo/CI formatter when present (for example `goimports` or `golangci-lint` format
      rules); otherwise run `go fmt ./...`.
    - `go mod tidy`
    - `go list ./...`
    - Prefer `go test -short ./...` for default CI/local verification.
    - Run full tests separately when needed (for example on nightly/release: `go test ./...`).
    - Run e2e tests separately with explicit build tags/jobs (for example
      `go test -tags=e2e ./...`).
    - `go vet ./...` (when enabled in repo/CI)
12. Document non-obvious layout decisions:
    - Add or update top-level comments or README sections when creating new major folders.
    - Explain why `pkg/` exists (or why everything stays in `internal/`).

## Notes

- `golang-standards/project-layout` is a widely used reference, not an official Go standard; treat
  it as a toolbox, not a mandatory checklist.
- Common baseline for applications:

```text
.
├── go.mod
├── go.sum
├── cmd/
│   └── <app>/main.go
├── internal/
│   ├── <domain>/...        # or internal/<app>/<domain>/...
│   └── platform/           # optional, only for real shared platform code
├── api/            # optional
├── configs/        # optional
├── scripts/        # optional
├── build/          # optional
├── deployments/    # optional
└── test/           # optional integration/e2e assets
```

- Example tree: single-binary service.

```text
.
├── go.mod
├── cmd/
│   └── billing-api/main.go
└── internal/
    ├── billing/
    │   ├── service.go
    │   └── repository.go
    └── platform/
        └── httpserver/server.go
```

- Example tree: multi-binary repo.

```text
.
├── go.mod
├── cmd/
│   ├── api/main.go
│   └── worker/main.go
└── internal/
    ├── api/
    │   └── orders/handler.go
    └── worker/
        └── orders/consumer.go
```

- Example tree: reusable library.
- Use `pkg/` only when you intentionally ship a stable public API/SDK; otherwise prefer packages
  under the module root.

```text
.
├── go.mod
├── parser/
│   └── parser.go
└── validate/
    └── rules.go
```

- Directory intent from project-layout:
  - `cmd/`: executable programs.
  - `internal/`: private code enforced by the Go compiler.
  - `pkg/`: public/reusable packages (optional, convention only; not compiler-enforced).
  - `api/`: API contracts only (OpenAPI/proto), not runtime server implementations.
  - `configs/`, `scripts/`, `build/`, `deployments/`, `test/`: operational/supporting artifacts.
- Anti-patterns:
  - Creating deep folder trees before they are needed.
  - Putting domain logic in `cmd/`.
  - Putting runtime handler/impl code in `api/`.
  - Using `pkg/` as a generic dumping ground.
  - Adding `src/` just because other languages do it.
- Priority order when rules conflict:
  1. Keep build/test green and avoid import cycles.
  2. Preserve minimal, clear boundaries (`cmd` vs `internal` vs optional `pkg`).
  3. Add optional folders only for real, present needs.

### skill: clean-architecture-hexagonal-components

---

name: clean-architecture-hexagonal-components
description: |
Apply strict Clean Architecture + Hexagonal (Ports & Adapters) with optional component (bounded
context) packaging. Use when creating or modifying features that must enforce
domain/application/adapter boundaries, inward dependencies, and (when enabled) cross-component
isolation.

---

# Clean Architecture Hexagonal (Components Optional)

## Purpose

Enforce a Clean Architecture + Hexagonal (Ports & Adapters) structure with strict dependency and
layering rules, using component (bounded context) packaging when appropriate. Components may also be
named `modules/` or similar; treat them equivalently as bounded contexts.

## When to Use

Use when building or refactoring features that must follow strict Clean Architecture + Hexagonal
boundaries. Components (bounded contexts) are optional for small projects.

## Inputs

- Feature request or change description.
- Existing architecture cues and folder structure (if any).
- Target component (bounded context) name if components are used; otherwise the module name.
- Inbound interface(s) (HTTP, CLI, MQ, etc.).
- IO needs (persistence, external services, messaging).
- Existing code conventions and DI/composition patterns.

## Outputs

- Coherent patch that creates/updates files, moves files if needed, and fixes imports.
- Component-structured directories when components are enabled; otherwise a single-module layout.
- Thin inbound controllers/handlers and pure domain logic.
- Tests aligned to domain, use case, and adapter layers.
- No forbidden imports across layers or components.

## Steps

1. Scan the repo for existing architecture cues (e.g., a `components/` or `modules/` directory,
   `bounded_contexts/`, `domain/`); summarize what you found.
2. Default to the existing structure. If none exists, default to a single-module layout unless there
   are clear signs of multiple bounded contexts (e.g., distinct feature folders or multiple
   domains).
3. If components are used, identify the target component (bounded context). If unspecified, infer it
   from domain language; ask a question only if strictly necessary to avoid incorrect placement.
4. Ensure the directory structure exists for `shared_kernel/` and `bootstrap/`, plus
   `components/<component>/` (or equivalent such as `modules/<module>/`) when bounded contexts are
   enabled. Place these at the project source root (repo root or optional `src/`).
5. Define or extend a single inbound port per use case in `application/ports/in`, using
   command-style input and explicit output DTOs.
6. Implement the use case in `application/use_cases`, orchestrating domain behavior and interacting
   with external systems only via outbound ports (no direct drivers/framework calls).
7. Define outbound ports in `application/ports/out` for any IO needs; shape them by core needs, not
   external APIs.
8. For query-heavy read use cases, define a read-side outbound port (`*ReadModel` / `*QueryService`
   / `*Finder`) returning DTOs/views, separate from aggregate repositories.
9. Implement outbound adapters in `adapters/outbound/*`, mapping through ACLs for external systems
   and using infrastructure drivers as needed.
10. Implement inbound adapters in `adapters/inbound/*`: validate input, map to command/DTO, call the
    inbound port or command bus, map errors.
11. Wire dependencies only in composition roots (`components/<name>/infrastructure/di` when
    components are used, otherwise the module-level DI area) and in the bootstrap entry point (e.g.,
    `bootstrap/main.*`).
12. Add tests according to the Testing taxonomy below: unit (domain + use case with mocked outbound
    ports), integration + contract (adapters), and functional tests for critical user flows.
13. Verify dependency boundaries by checking imports; fix any violations before finalizing.
14. Run the SOLID review gate in Notes. Treat any "No" answer as a design defect; if you must accept
    a "No", document the exception + trade-offs before finalizing.

## Testing taxonomy

Use these definitions when planning and implementing Step 12.

- Unit tests
  - Definition: Verify a single domain rule or use-case orchestration path in isolation.
  - Typical scope: `domain/**` entities, value objects, policies, domain services; application use
    cases with mocked/fake outbound ports.
  - Real DB: Not allowed.
  - Placement: Near the layer under test (for example `domain/**` and `application/**` test files).
  - Allowed dependencies: Same-layer code, `shared_kernel/` primitives, and test doubles only. No
    adapter, infrastructure, or bootstrap dependencies.
- Integration tests
  - Definition: Verify collaboration across architectural boundaries (for example application port
    to adapter to driver/real dependency).
  - Typical scope: `adapters/outbound/**` implementations against real dependencies and
    adapter-level contract tests at inbound/outbound boundaries.
  - Real DB: Not required for contract tests; allowed and preferred for persistence-adapter
    integration (use ephemeral/local test DB or containerized DB).
  - Placement: Adapter/infrastructure test locations (for example `adapters/**` or
    `infrastructure/**` test files).
  - Allowed dependencies: Application port contracts/DTOs, adapter code, infrastructure drivers, and
    test fixtures. Do not move business rules into these tests.
  - Contract vs integration: Contract tests may use in-memory harnesses or stubs without real
    dependencies; integration tests should exercise real dependencies (DB/SDK sandbox) when
    feasible.
  - Boundary with functional: Integration may use in-memory transport harnesses (for example
    `httptest`) without full service bootstrap.
- Functional tests
  - Definition: Black-box verification of user-visible behavior through real inbound interfaces
    (HTTP/CLI/MQ).
  - Typical scope: End-to-end feature flows via public endpoints/commands/messages.
  - Real DB: Allowed when needed for realistic behavior; reset state between tests.
  - Placement: Top-level `tests/functional` (or equivalent repo-standard e2e location).
  - Allowed dependencies: Public interface clients/test harness and fixtures. Avoid asserting on
    domain internals or private adapter details.
  - Boundary with integration: Prefer fully wired application bootstrap/composition root and
    external-client-style assertions (or an equivalent black-box setup).
- Testing code note: Testing code may reference cross-layer public interfaces/components when
  required for verification; this does not permit breaking production dependency direction or import
  boundaries.

## Notes

Default structure (components optional for small projects; adjust layout to fit your language's
conventions):

```text
shared_kernel/
  domain/
    events/
    value_objects/
    specifications/
  application/
    events/
    messaging/
components/ (optional)
  <component_name>/
    domain/
      entities/
      value_objects/
      services/
      policies/
      events/
    application/
      ports/
        in/
        out/
      use_cases/
      dto/
      mappers/
    adapters/
      inbound/
        http/
          controllers/
          middleware/
        cli/
        mq/
      outbound/
        persistence/
        external/
        messaging/
    infrastructure/
      drivers/
      di/
bootstrap/
  main.*
```

Single-module structure (when components are not used; adjust layout to fit your language's
conventions):

```text
domain/
  entities/
  value_objects/
  services/
  policies/
  events/
application/
  ports/
    in/
    out/
  use_cases/
  dto/
  mappers/
adapters/
  inbound/
    http/
      controllers/
      middleware/
    cli/
    mq/
  outbound/
    persistence/
    external/
    messaging/
infrastructure/
  drivers/
  di/
bootstrap/
  main.*
```

Note: The above directory structure is illustrative. Adjust the top-level placement to fit your
language's standard project layout. For example, some projects keep source code in `src/` or
`src/main` (common in Java/.NET), while others place the folders at the repository root (common in
Go, Python, etc.). Follow the standard conventions of your language, as long as the separation of
architectural layers (domain, application, adapters, etc.) remains intact.

### SOLID review gate (required before finalize)

- Single Responsibility Principle (SRP)
  - One use case represents one primary business intent (one "reason to change").
  - Inbound ports define core-facing contracts; inbound adapters are transport-facing wrappers.
  - Inbound adapters only parse/validate/map transport data and delegate to use cases.
  - Repositories persist aggregates only; read-model/query ports return DTOs/views.
  - Domain services contain only domain logic, with no IO or framework dependencies.
- Open/Closed Principle (OCP)
  - Add new behavior by introducing new adapters, policies, or strategy implementations.
  - Avoid modifying stable domain/use-case code when only transport/vendor concerns change.
  - Shape ports by core needs so implementations can vary without changing core contracts.
- Liskov Substitution Principle (LSP)
  - Every adapter implementation must preserve the semantics of its port contract.
  - Nullability, error contracts, ordering, and idempotency guarantees must stay compatible across
    implementations.
  - Require outbound adapter contract tests (see Testing taxonomy: integration + contract) so
    alternate implementations are safely swappable.
- Interface Segregation Principle (ISP)
  - Keep ports small and use-case-focused; avoid "god interfaces."
  - Separate read-side query ports from write-side repository ports.
  - Do not force adapters to implement methods they do not need.
- Dependency Inversion Principle (DIP)
  - Domain/application depend only on abstractions (ports), never concrete drivers/SDKs/frameworks.
  - Bind abstractions to concrete adapters only in composition roots and bootstrap wiring.
  - Keep vendor DTOs and SDK models in outbound adapters/ACLs, mapped to core DTOs/domain types.

SOLID review checklist (all should be "Yes"; if "No", document exception + trade-offs):

- [ ] SRP: If one feature changes, do edits stay localized to one primary module per layer?
- [ ] OCP: Can a new transport or vendor be added by creating an adapter instead of editing core
      rules?
- [ ] LSP: Can implementation A and B for the same port pass the same contract test suite unchanged?
- [ ] ISP: Does each consumer depend only on methods it actually uses?
- [ ] DIP: Are all framework/driver objects created outside domain/application in composition roots?

Non-negotiable rules (treat violations as errors):

The following are strict architecture boundaries. Treat violations as errors, regardless of language
or tooling.

- Domain must not import `application/`.
- Domain must not import `adapters/`, `infrastructure/`, or `bootstrap/`.
- Domain and application may import `shared_kernel/` (but `shared_kernel/` must be dependency-free
  with respect to feature modules; no imports from `components/` or `modules/`).
- Domain events live in `domain/events` and describe in-model state changes. Cross-component
  communication uses integration events in `shared_kernel/domain/events` (or equivalent path if
  already present, or explicit ports), never direct imports. Do not place feature-specific DTOs in
  shared_kernel.
- Application may import `domain/`, `shared_kernel/`, and other `application/**` modules.
  Application must not import `adapters/`, `infrastructure/`, or `bootstrap/`.
- Inbound adapters must not execute domain business logic or mutate aggregates. They may reference
  domain types (value objects, error codes) for parsing and error mapping, but must call the use
  case (inbound port or bus) to perform any business action. Perform transport validation/parsing in
  inbound adapters before calling the use case. Business validation/invariants are enforced in
  domain/application.
- Adapters may import `application/ports/**`, application DTOs, and domain types as needed for
  mapping, but must not move business logic into adapters.
- Transport-specific schemas/validators live in `adapters/inbound/<transport>/middleware` (or
  equivalent), not in `application/` or `domain/`.
- Vendor/SDK DTOs must not appear in application/domain. Map them in outbound adapters/ACL to
  application DTOs or domain types.
- Outbound adapters implement application outbound ports; may use infrastructure drivers.
- Components/modules (bounded contexts) must not import each other's domain/application directly.
  Use shared_kernel events, outbound ACLs, or explicit query ports.
- Ports live inside application core; adapters live outside.
- One use case equals one inbound port/handler.
- Outbound ports represent required capabilities (repositories, gateways, publishers) and are shaped
  by core needs.
- Use cases orchestrate: load entities, invoke domain behavior, persist, publish application events
  if needed.
- Prefer value objects/policies/specifications for pure rules; use domain services only when the
  logic does not fit naturally on an entity/value object.
- Domain services contain domain logic that does not naturally belong to a single entity/value
  object. They have no IO and no repository dependencies.
- Pure domain rules (policies/strategies) live in `domain/policies`.
- Repository ports are only for aggregate persistence (get/save by aggregate identity). For queries,
  use `*ReadModel` / `*QueryService` / `*Finder` returning DTOs/views.
- Repository ports must accept/return aggregates (or aggregate IDs). They must not return view
  models/DTOs.
- Read-side ports must return DTOs/views and must not return aggregates.
- Transport payloads (HTTP/MQ/CLI) must be mapped in inbound adapters to application DTOs/commands.
  Do not leak transport DTOs into application/domain.
- Outbound adapters must not call inbound ports/use cases. All orchestration happens in application
  use cases.
- Errors are structured (type + code + message + optional metadata). Inbound adapters map these
  errors to transport-specific responses.
- Only composition roots may bind ports to adapter implementations. Do not instantiate drivers/SDK
  clients inside domain/application.
- Command bus interface (if used) lives in `application/ports/in`. In-memory bus may live in
  application; framework-driven bus wiring stays in infrastructure.

Output requirements:

- Keep controllers/handlers thin; no business logic in adapters.
- Keep domain pure; no frameworks, IO, or persistence models.
- Do not add external integrations unless requested; use ports and adapters with clear boundaries.
- Do not introduce new architectural concepts (e.g., CQRS, event sourcing, command bus) unless
  requested or already present in the repo.

Naming guidance:

- Inbound: `<Verb><Noun>UseCase` / `<Verb><Noun>Handler` / `<Verb><Noun>Port`.
- Outbound: `<Noun>Repository` / `<Capability>Gateway` / `<Capability>Publisher`.
- Query read side: `<Noun>ReadModel` / `<Noun>QueryService` / `<Noun>Finder`.
- External vendors: `<Vendor><Capability>Client` with ACL mappers in outbound adapters.

### skill: conventional-commit

---

name: conventional-commit
description: |
Generate a Conventional Commits message from the current working tree and commit by default.
Stage all changes, infer type/scope from the diff, produce a compliant header with optional
body/footer, and run git commit. Use draft-only mode only when the user explicitly asks for a
message without committing.

---

# Conventional Commit

## Purpose

- Draft a Conventional Commits message from the current working tree.
- Default to staging all changes and committing with the drafted message.
- Keep message generation rules separate from git execution steps for clarity.

## When to Use

- Follow the trigger guidance in the frontmatter description; do not add new criteria here.

## Inputs

- Optional change summary or constraints provided by the user.
- Repo state and diffs (when available).
- Optional: preferred type, scope, breaking-change details, references, or test notes.

## Outputs

- Conventional commit message with a single-line header.
- Optional body and footer when they add value.
- Draft-only mode: return only the commit message.
- Default mode: stage all changes, show the final message, run `git commit`, and report the new
  commit hash on success.

## Steps

### 1) Mode Selection

- Commit mode (default): stage all changes and run `git commit`.
- Draft-only mode: only when the user explicitly asks for message-only output (e.g., "draft only",
  "message only", "no commit", "message only please").

### 2) Commit Message Spec

#### Header

- Format: `type(scope): summary` or `type: summary`
- Imperative mood; no trailing period
- Header length (including type/scope): <= 72 characters

#### Body (optional, recommended when needed)

Body is optional. Include it only when rationale is needed, risk is higher, behavior changes, or the
diff is non-obvious.

Non-trivial examples: public API change, business logic change, data schema/migration, auth/payment,
error handling, concurrency.

- Default body: prefer a short Why (1-3 lines).
- Add `Tests:` only when tests were run, risk is higher, or behavior changed.
- For `Tests:` trigger rules and format, follow Inference Rules -> Tests (body note).
- Include What only when the diff is non-obvious or spans multiple areas (for example: `- api: ...`,
  `- ci: ...`).
- Include Impact only when there is behavior change, compatibility/migration concerns, or likely
  pitfalls (1-3 lines).

Line wrapping: wrap body lines at ~72 characters when practical.

#### Footer

- Breaking changes:
  - Add `!` in header (`type(scope)!: summary` or `type!: summary`)
  - Add `BREAKING CHANGE: <what changed> <how to migrate>`
- References: keep in footer unless user requests otherwise.

### 3) Inference Rules

#### Type (priority heuristics)

1. docs-only -> `docs`
2. test-only -> `test`
3. ci/workflows -> `ci`
4. build/deps/tooling -> `build`
5. perf-only or clear performance improvement -> `perf`
6. formatting-only or lint-only -> `style`
7. behavior change: new -> `feat`, bug fix -> `fix`
8. no behavior change -> `refactor`
9. explicit revert request -> `revert`
10. otherwise -> `chore`

If reverting, use header `revert: <original summary>` and include `This reverts commit <hash>.` in
the body when the hash is available.

#### Scope

- Use the smallest meaningful scope (lowercase).
- Keep scope length <= 16 characters; if longer, omit scope.
- If spanning many areas, omit scope or use `repo` and list key areas in the body.
- If mixed changes clearly include multiple areas, still commit by default but summarize the
  sections in the body (for example: `- api: ...`, `- ci: ...`, `- docs: ...`).

#### Breaking change

- Likely breaking if public APIs, routes, config contracts, or schemas change.
- If unclear, ask only: "Is this a breaking change?"

#### Untracked files

- If untracked files exist:
  - If <= 10: list them in Body "What"
  - If > 10: summarize count and list the first 10

#### Tests (body note)

- Add a `Tests:` section when either is true:
  - Tests were run.
  - Risk is higher or externally observable behavior changed.
- Treat risk as higher for changes like API/route contracts, schema/migrations, auth/payment,
  concurrency, or error-handling behavior.
- Keep only tests that can run in this repo; do not include commands/tooling unrelated to this repo.
- If test coverage items are many, list only high-signal tests (for example: highest-risk paths,
  changed behavior paths, and one happy path). Do not enumerate every test case.
- When omitting long test lists, add one summary line for the remainder (for example:
  `- plus <n> additional checks`).
- Format:
  - `Tests:`
  - `- <command>`
  - `- manual - <scenario>`
  - `- not run (reason)`

#### Body inclusion heuristics

- This section clarifies the Body rules above; if conflicts arise, follow Commit Message Spec.
- `feat` / `fix` / `perf`: usually add Why; add `Tests:` when run or risk is higher.
- `refactor`: add Why/Impact only when diff is non-obvious or risk is higher.
- `docs` / `style` / `ci` / `build` / `test`: usually no body; if needed, add only a 1-line Why and
  a `Tests:` section with `- not run (reason)` when a Tests note is needed.

### 4) Execution Steps

1. `git rev-parse --is-inside-work-tree`
2. If not a repo: draft from user summary; do not commit.
3. `git status --porcelain`
4. If no changes: report "nothing to commit" and stop.
5. If unmerged paths exist (`git diff --name-only --diff-filter=U`): stop and ask to resolve
   conflicts before committing.
6. Commit mode:
   - `git add -A`
   - Inspect: `git --no-pager diff --cached --stat` and `git --no-pager diff --cached`
7. Draft-only mode:
   - Inspect: `git --no-pager diff --stat` and `git --no-pager diff`
8. Generate final message and show it to the user.
9. Commit (commit mode):
   - Use `mktemp`, write message, run `git commit -F <tmpfile>`
   - Clean up temp file (best effort)
10. On success: report short hash and final message.
11. On failure: return message and error summary; do not retry blindly.

## Notes

- Do not invent details; ask for missing essentials only when inference is unclear.
- Prefer consistency in type/scope naming across the repo.
- Default to staging all changes and committing without extra confirmation.
- Reference `references/conventional-commits.md` for the v1.0.0 spec.
