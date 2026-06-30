## Context

This phase implements the backend half of the product decision made by `spec-remove-probe-location-concept`. The current code has several layers of location-specific behavior:

```text
CreateMonitorRequest.probeLocations
  -> Monitor.ProbeLocations
  -> DynamoDB MonitorRecord.ProbeLocations
  -> BuildExecutionRequests loops over locations
  -> ExecutionWork.ProbeLocationID
  -> ExecutionResult.ProbeLocationID
  -> CheckRun.ProbeLocationID / MonitorStatus.LastProbeLocationID
  -> API response fields
```

The target shape is flatter:

```text
Monitor
  -> one due execution request
  -> one execution result
  -> check run + latest status
```

## Goals / Non-Goals

**Goals:**

- Remove location fields from backend public contracts.
- Remove location fan-out and catalog validation from execution planning.
- Remove hard-coded `iad` and `US East` from services/shared tests and runtime code.
- Keep response envelope behavior unchanged.

**Non-Goals:**

- Dashboard UI removal; that is handled by `dashboard-and-docs-remove-locations`.
- Migration tooling for production DynamoDB records.
- Adding a new explicit runtime-environment identifier.

## Decisions

### Decision: Remove fields instead of retaining empty/default values

**Choice:** Delete location fields from request/response/domain types where they only served probe-location semantics.

**Rationale:** Returning `probeLocationId: "default"` or `lastProbeLocationId: "default"` would preserve accidental complexity and imply an operator-visible dimension that no longer exists.

### Decision: One work item per monitor run

**Choice:** The scheduler and manual-run path produce one check attempt per monitor run.

**Rationale:** This is the simplest faithful model for single-environment health checking and removes deduplication/claiming keyed by `runID + probeLocationID`.

## Risks / Trade-offs

- **[Risk] OpenAPI/dashboard clients break until phase 3 lands.** Mitigation: land backend and dashboard changes close together, or keep branch-local until both are complete.
- **[Risk] DynamoDB record constructors and repository tests may need broad updates.** Mitigation: prefer deleting location-specific fields over adding adapter helpers.
- **[Trade-off] No compatibility for old monitor records with `ProbeLocations`.** Accepted because there is no production environment.
