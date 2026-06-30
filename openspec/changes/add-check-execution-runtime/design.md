## Context

Repository already has most domain pieces for monitoring runtime, but they are disconnected. `monitor-api` can accept manual run commands and scheduler config changes, `shared/checkexecution` can build and execute HTTP checks, and `shared/resultstatus` can persist run/status models, yet no service currently consumes accepted run work or recurring scheduler state to execute checks end to end. The current executor also ignores `http.expectedBodyContains`, which leaves one supported HTTP assertion only partially implemented.

## Goals / Non-Goals

**Goals:**
- Turn accepted manual runs and recurring scheduler decisions into one shared execution pipeline.
- Ensure HTTP execution enforces configured status, timeout, and body-content assertions.
- Persist completed execution output into `CheckRun` history and `MonitorStatus` snapshots.
- Define deterministic incident lifecycle behavior driven by execution outcomes.
- Keep existing public API surface stable while making its runtime behavior real.

**Non-Goals:**
- No new public CRUD surface for runs, incidents, or scheduler internals.
- No multi-tenant redesign beyond current single built-in tenant handling.
- No new monitor protocol beyond current HTTP support.
- No broad audit-search redesign or auth/RBAC work.

## Decisions

### Use one execution-work contract for manual and recurring runs
- Decision: both `POST /api/v1/monitors/{id}/run` and recurring scheduler ticks will materialize the same internal execution-work record shape before any check executes.
- Rationale: one downstream worker path prevents drift between manual and recurring semantics and keeps result, status, and incident logic centralized.
- Alternative considered: execute manual runs inline in API and use a different path for recurring runs.
- Why not: would split behavior, increase Lambda latency risk, and create inconsistent persistence and incident outcomes.

### Keep command acceptance separate from check completion
- Decision: manual run API remains an acceptance command that queues execution work and returns a stable `runId`; the worker later produces `CheckRun` and status output.
- Rationale: preserves current command-oriented API shape while allowing asynchronous execution and retry-safe worker behavior.
- Alternative considered: make manual run synchronous and return final outcome.
- Why not: poor fit for Lambda/API response budgets and incompatible with future multi-location or queued execution.

### Treat HTTP expectations as one combined assertion policy
- Decision: HTTP execution will evaluate timeout, response status, and optional body-substring expectation in one normalized result flow.
- Rationale: monitor configuration already exposes these fields, so runtime must honor all declared assertions to make monitor behavior trustworthy.
- Alternative considered: defer body assertion until later and only persist response status outcomes.
- Why not: leaves shipped monitor config misleading and undermines operator confidence.

### Derive incidents from monitor-level result state first
- Decision: v1 incident lifecycle will be monitor-scoped, not monitor-plus-location scoped. First non-success result opens or updates one open incident for that monitor; first later success resolves the open incident.
- Rationale: matches current incident API shape, keeps dashboard/operator semantics simple, and avoids multiplying incidents for multi-location execution before product needs that granularity.
- Alternative considered: incident per failing probe location.
- Why not: adds more cardinality, more complex reads, and harder operator experience before route and UI needs are proven.

### Gate recurring execution at scheduler read time and worker execution time
- Decision: recurring planner must skip scheduling when admin scheduler config disables recurring execution, and worker must still verify monitor enabled state before executing claimed work.
- Rationale: scheduler config is global stop path, while monitor disable remains per-monitor stop path. Both must be enforced where stale queued work could otherwise slip through.
- Alternative considered: only gate work creation and assume queued work stays valid forever.
- Why not: stale queued work could execute after operators believe monitoring is paused or disabled.

## Risks / Trade-offs

- [Risk] Async work records can become stale after monitor edits or disablement. -> Mitigation: worker re-reads current monitor state before execution and skips non-runnable work.
- [Risk] First incident rules may be too coarse for multi-location failures. -> Mitigation: keep monitor-level contract explicit now and revisit with a future spec if location-scoped incidents become necessary.
- [Risk] Shared table records for queued work can complicate repository code. -> Mitigation: keep one explicit execution-work item family with narrow lifecycle fields rather than overloading `CheckRun` or status items.
- [Risk] Body assertion using simple substring matching may be weaker than future needs. -> Mitigation: specify current contract as substring-based only and leave regex/structured body assertions out of scope.

## Migration Plan

1. Add spec deltas for execution pipeline, result/status persistence, manual-run behavior, incident lifecycle, and scheduler runtime gating.
2. Implement execution-work item model and repository methods for enqueue, claim, complete, and skip behavior.
3. Implement execution worker that loads runnable work, executes HTTP checks, persists run/status, and applies incident transitions.
4. Implement recurring planner that reads scheduler config and enqueues recurring work for enabled monitors.
5. Wire new runtime Lambdas and triggers in `infra/`.

Rollback: disable recurring planner trigger and worker trigger, leaving existing monitor CRUD and read APIs intact. Accepted manual runs created during rollback may remain unprocessed internal records but will not affect public API correctness beyond missing downstream execution.

## Open Questions

- Should accepted execution-work records store a full monitor snapshot or only monitor ID plus probe location ID and rely on live re-read for execution?
- Does v1 worker need explicit retry metadata and dead-letter behavior, or is best-effort plus idempotent rerun acceptable for first implementation?
- Should resolved incidents preserve last failing probe location in summary fields, or is monitor-level summary text sufficient?
