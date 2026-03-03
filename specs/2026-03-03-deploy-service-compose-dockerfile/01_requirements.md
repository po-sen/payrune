---
doc: 01_requirements
spec_date: 2026-03-03
slug: deploy-service-compose-dockerfile
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

- Compose file: Docker Compose manifest at `deployments/compose/compose.yaml`.
- Runtime image: Final image produced by `build/app/Dockerfile` and used by compose.

## Out-of-scope behaviors

- OOS1: Publish image to remote registry.
- OOS2: Add reverse proxy, TLS termination, or autoscaling.

## Functional requirements

### FR-001 - Service Dockerfile packaging

- Description:
  - The repository MUST provide a Dockerfile to build and run the `payrune` service.
- Acceptance criteria:
  - [ ] `build/app/Dockerfile` exists.
  - [ ] Dockerfile builds `cmd/payrune/main.go` into a Linux binary.
  - [ ] Final image launches the `payrune` binary by default.
  - [ ] Runtime container exposes port `8080`.
- Notes:
  - Multi-stage build is preferred to minimize runtime image size.

### FR-002 - Compose deployment manifest

- Description:
  - The repository MUST provide a compose manifest to run the service container from the local source tree.
- Acceptance criteria:
  - [ ] `deployments/compose/compose.yaml` exists.
  - [ ] Compose defines a `payrune` service using `build/app/Dockerfile` with repo root build context.
  - [ ] Compose maps host `8080` to container `8080`.
  - [ ] Compose sets a restart policy for the service container.
- Notes:
  - Compose file should be directly usable by Docker Compose v2.

### FR-003 - Minimal Makefile lifecycle commands

- Description:
  - The repository MUST provide minimal make targets to bring the service up and down.
- Acceptance criteria:
  - [ ] Root `Makefile` exists.
  - [ ] `make up` runs compose up in detached mode with image build.
  - [ ] `make down` stops and removes resources from the same compose project.
- Notes:
  - Keep the Makefile concise and focused on these lifecycle commands.

## Non-functional requirements

- Performance (NFR-001): After `make up`, service should respond on `/health` within 15 seconds on baseline local machine.
- Availability/Reliability (NFR-002): Compose service must include `restart: unless-stopped` to recover from crashes in local runs.
- Security/Privacy (NFR-003): Runtime container must run as non-root user.
- Compliance (NFR-004): Spec folder must pass `SPEC_DIR="specs/2026-03-03-deploy-service-compose-dockerfile" bash scripts/spec-lint.sh`.
- Observability (NFR-005): Service logs must be visible via `docker compose -f deployments/compose/compose.yaml logs payrune`.
- Maintainability (NFR-006): `Makefile` must stay minimal (only shared compose vars plus `up`/`down` targets).

## Dependencies and integrations

- External systems:
  - Docker Engine
  - Docker Compose v2 plugin
- Internal services:
  - None
