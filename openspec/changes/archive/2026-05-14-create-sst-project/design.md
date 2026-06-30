## Context

The repository currently contains only OpenSpec configuration and no application or infrastructure code. The change needs to bootstrap a first-class SST project so future backend services, event-driven workflows, and web apps can share one deployment entrypoint and one opinionated infrastructure layout.

SST is a good fit because the project target is an asynchronous, serverless healthcheck platform, and SST provides local development, infrastructure composition, and AWS deployment primitives without forcing raw CDK boilerplate for every addition.

## Goals / Non-Goals

**Goals:**
- Introduce a runnable SST project scaffold in the repository.
- Use TypeScript for SST configuration and stack definitions.
- Establish a minimal baseline stack that can synthesize and deploy successfully.
- Define repository-level scripts and structure that future changes can extend.

**Non-Goals:**
- Implement the healthcheck probers, API, database, or dashboard.
- Model the full production architecture for the platform.
- Migrate any existing infrastructure, because none exists yet.

## Decisions

### Use SST as the primary infrastructure entrypoint
- Decision: Bootstrap the repo around SST instead of raw CDK.
- Rationale: SST provides a faster starting point for a serverless product and simplifies local workflows while still allowing AWS-native infrastructure definition.
- Alternative considered: Start with plain AWS CDK in `infra/`.
- Why not: It adds more boilerplate up front and does not improve the immediate bootstrap outcome.

### Keep the initial stack intentionally minimal
- Decision: Create a starter stack that proves the project shape and deployment path without adding speculative resources.
- Rationale: The first change should establish the foundation only; service-specific resources belong in follow-on specs.
- Alternative considered: Pre-create API Gateway, Lambda, DynamoDB, queues, and alarms.
- Why not: That would encode architecture assumptions before the product requirements exist.

### Preserve room for the monorepo layout described in AGENTS.md
- Decision: Add the SST project in a way that does not block future `apps/`, `services/`, `shared/`, and `infra/` directories.
- Rationale: The repository guidance already points toward a monorepo split, so the bootstrap should align with that expected shape.
- Alternative considered: Place all future code under the SST app directly.
- Why not: That would make later separation of concerns harder.

## Risks / Trade-offs

- [Risk] SST version or template defaults may change over time. -> Mitigation: Pin package versions and commit the generated configuration explicitly.
- [Risk] A bootstrap created without real workloads may need refactoring once backend services are added. -> Mitigation: Keep the first stack minimal and avoid premature resource modeling.
- [Risk] Local development prerequisites for SST and AWS credentials may not be obvious. -> Mitigation: Add clear scripts and bootstrap documentation as part of the scaffold.

## Migration Plan

1. Add the SST project files and Node.js dependencies.
2. Commit a minimal stack that can run SST synth or equivalent validation.
3. Document how developers start the SST app locally and where future resources should be added.

Rollback is straightforward because this is a bootstrap-only change: removing the added SST files returns the repository to its current state.

## Open Questions

- Should the SST app live at the repo root or under `infra/` once implementation begins?
- Which SST major version should be the long-term standard for this repository?
- Should the first stack expose only a placeholder resource, or also include a health endpoint used for smoke testing?
