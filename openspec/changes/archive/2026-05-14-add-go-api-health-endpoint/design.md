## Context

Current repo has SST bootstrap in `infra/` and no application service path yet. Next useful step: prove request can enter API Gateway, hit Go Lambda, return stable response. This creates first backend vertical slice without forcing data model, auth, async jobs, or probe execution decisions.

## Goals / Non-Goals

**Goals:**
- Add single public HTTP endpoint `GET /api/health`.
- Implement handler in Go under `services/`.
- Deploy route through SST-managed API Gateway and Lambda.
- Keep response static and deterministic for easy smoke testing.

**Non-Goals:**
- No auth.
- No database, queue, scheduler, or event bus.
- No probe execution.
- No dashboard integration.
- No multi-endpoint API framework unless needed for clean wiring.

## Decisions

### Use `GET /api/health` as first route
- Decision: first route returns service liveness only, under API namespace.
- Rationale: keeps route family coherent for future `/api/v1/...` product endpoints while staying cheap to test locally and after deploy.
- Alternative considered: `/health`, `/ok`, or `/status`.
- Why not: root-level route mixes ops path with future versioned API surface; other names add no gain.

### Put Go handler under `services/`
- Decision: keep runtime code in `services/`, infra glue in `infra/`.
- Rationale: matches repo map in `AGENTS.md`; avoids mixing app logic into SST files.
- Alternative considered: keep Lambda code inside `infra/`.
- Why not: hurts long-term separation once more services exist.

### Keep Lambda single-purpose, static response
- Decision: handler returns `200` and tiny JSON payload like `{ "ok": true }`.
- Rationale: proves packaging and routing path with minimum moving parts.
- Alternative considered: add version/build metadata, AWS checks, or dependency health.
- Why not: adds branching and future maintenance before base path proven.

### Use one API resource in SST for route wiring
- Decision: create one API Gateway resource with one route to one Go Lambda.
- Rationale: enough structure for future routes, still minimal.
- Alternative considered: direct Lambda URL.
- Why not: project explicitly wants API Gateway shape.

## Risks / Trade-offs

- [Risk] Go Lambda build settings in SST may take small trial-and-error. -> Mitigation: keep handler minimal, validate with local SST workflow early.
- [Risk] Route path and response shape may later need expansion. -> Mitigation: choose conventional path and JSON form now.
- [Risk] First API resource may tempt over-generalization. -> Mitigation: keep only one route in scope.

## Migration Plan

1. Add Go module and health handler under `services/`.
2. Extend SST stack with API Gateway and Lambda route.
3. Add docs and validation commands.
4. Verify endpoint in local SST dev flow and deployed environment.

Rollback: remove API resource, Lambda wiring, and Go service files. Bootstrap stack remains intact.

## Open Questions

- Should response include only `{ "ok": true }`, or also service name and timestamp?
