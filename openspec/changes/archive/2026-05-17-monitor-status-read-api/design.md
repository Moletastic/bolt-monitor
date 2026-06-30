## Context

Basic monitor CRUD can manage configuration, but operators and future frontend also need current status and recent execution history. Read API should focus on operational visibility, not mutation, and should reuse persisted `MonitorStatus` and `CheckRun` models.

## Goals / Non-Goals

**Goals:**
- Expose current monitor status through HTTP API.
- Expose recent run history for individual monitor.
- Expose dashboard-oriented monitor listing with embedded status summary.

**Non-Goals:**
- Full incident views or alerting logic.
- Rich filtering, sorting, and pagination beyond basic read needs.
- Authentication, RBAC, or billing views.

## Decisions

### Read from persisted status, not recompute on every request
- Decision: status endpoints should read from stored `MonitorStatus` and `CheckRun` records.
- Rationale: faster, simpler, consistent with single-table design.
- Alternative considered: derive status from latest runs dynamically.
- Why not: unnecessary read cost and complexity.

### Separate config detail from operational detail
- Decision: monitor detail can include config plus status summary, while run history stays separate endpoint.
- Rationale: keeps response payloads focused and efficient.
- Alternative considered: one huge monitor detail response with full history embedded.
- Why not: wasteful and hard to paginate.

### Start with operationally useful reads only
- Decision: first read API should target status summary and recent runs, not every analytics use case.
- Rationale: enough for dashboard bootstrap and operator inspection.
- Alternative considered: broad report/query API first.
- Why not: too large before frontend and incidents exist.

## Risks / Trade-offs

- [Risk] Read API may need pagination soon for run history. -> Mitigation: keep endpoint shape ready for basic cursor or limit fields.
- [Risk] Dashboard list may need sorting/filtering later. -> Mitigation: keep initial response simple, based on status summary.
- [Risk] If result/status model changes, read API must adapt. -> Mitigation: sequence this work after result/status contract stabilizes.

## Migration Plan

1. Define status and run-history endpoints.
2. Implement read repository paths using persisted result/status records.
3. Use these endpoints as backend base for frontend dashboard work.

Rollback is easy because this adds read surface only; no new mutation paths.

## Open Questions

- Should monitor list return status summary inline by default?
- What recent-run window should detail endpoint expose by default?
- Do we need separate dashboard endpoint or can monitor list cover v1 needs?
