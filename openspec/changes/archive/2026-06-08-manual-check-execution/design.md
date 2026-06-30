## Context

Currently, `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run` accepts a manual run request, creates a `RUN_REQUEST#` item in DynamoDB, and returns immediately with `{runId, status: "accepted"}`. No actual HTTP check is executed. The check-runtime worker (which would process queued work) does not yet exist.

This prevents operators from verifying their monitor configuration works before we invest in building the automated background execution infrastructure (SQS, EventBridge, worker Lambda).

## Goals / Non-Goals

**Goals:**
- Make manual run endpoint execute HTTP checks synchronously and return real results
- Allow operators to validate monitor configuration end-to-end via API
- Record execution results (status, runs, incidents) identically to how background workers will
- Build confidence that the monitoring system works before automating

**Non-Goals:**
- Building automated background execution (SQS, EventBridge, worker Lambda) — this is a separate change
- Supporting multiple probe locations beyond the single hardcoded "iad" location
- Changing the API endpoint path or authentication

## Decisions

### Decision 1: Execute inline in monitor-api Lambda, not via enqueued work

**Choice:** Execute the HTTP check directly within the `runMonitor()` handler, then persist the result.

**Rationale:** Simplest path to verification. We already have all pieces:
- `checkexecution.ExecuteHTTP()` performs the actual HTTP check
- `dynamoMonitorRepository.RecordExecutionResult()` persists results (needs to be added)
- `defaultProbeLocationCatalog()` provides the probe location

**Alternative considered:** Create `RUN_REQUEST#` items and have check-runtime process them synchronously. This would require check-runtime to be deployed and working, which is circular dependency — we want to verify the system works *before* building that.

### Decision 2: Reuse existing result recording logic from check-runtime

**Choice:** Copy the `RecordExecutionResult` pattern from `check-runtime/repository.go` into `monitor-api/repository.go`.

**Rationale:** The worker will eventually need the same logic. Having consistent result recording ensures that manual runs and automated runs produce identical records. The code is well-structured and testable.

**Alternative considered:** Create a shared package for result recording. Defer until background execution is built and we see actual code duplication.

### Decision 3: Response includes full execution result

**Choice:** The `POST /run` response includes outcome, duration, status code, error, timestamps, and probe location.

**Rationale:** Operators need immediate feedback. Returning just `{runId, status: "accepted"}` required additional polling to get results. Now they get results immediately.

**Changed response shape:**
```json
{
  "runId": "RUN_ABC123",
  "outcome": "success",
  "durationMs": 127,
  "statusCode": 200,
  "error": null,
  "probeLocationId": "iad",
  "startedAt": "2026-06-08T14:30:00Z",
  "finishedAt": "2026-06-08T14:30:00Z"
}
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Lambda timeout if target URL is slow or unresponsive | Configure appropriate `timeoutMs` per monitor (user responsibility); API Gateway has 30s timeout |
| HTTP client not properly configured for long-running requests | Use `http.Client` with Timeout matching monitor's `timeoutMs` setting |
| Code duplication between monitor-api and check-runtime | Accept for now; refactor to shared package when background execution is built |
| Incident creation on first failure (may surprise users) | This matches existing behavior — not changing here |

## Implementation Approach

### Changes to `services/monitor-api/handler.go`

In `runMonitor()`:
1. After validating monitor exists and is enabled (existing)
2. Build `checkexecution.ExecutionRequest` from monitor + probe location
3. Call `checkexecution.ExecuteHTTP(ctx, httpClient, request)` inline
4. Call `h.repo.RecordExecutionResult(ctx, monitor, runID, result)` to persist
5. Return `manualRunResponse` with execution result fields

### Changes to `services/monitor-api/repository.go`

1. Add `RecordExecutionResult(ctx, monitor, runID string, result checkexecution.ExecutionResult) error` to `monitorRepository` interface
2. Implement in `dynamoMonitorRepository` — copy logic from `check-runtime/repository.go RecordExecutionResult()`:
   - Write `RUN#` item (CheckRun record)
   - Write `STATUS` item (MonitorStatus)
   - Handle incident creation/resolution
   - Update service rollup

### Changes to `services/monitor-api/types.go`

Modify `manualRunResponse` to include execution result fields:
- Add `Outcome`, `DurationMs`, `StatusCode`, `Error`, `ProbeLocationID`, `StartedAt`, `FinishedAt`

Add `toManualRunResponseWithResult()` helper or extend existing response struct.

### Changes to `services/monitor-api/main.go`

No changes needed — already passes `defaultProbeLocationCatalog()` to handler.

## Migration Plan

1. Deploy modified `monitor-api` with synchronous execution
2. Test manually: `POST /run` returns real results immediately
3. Verify: `/status` shows updated `lastCheckedAt`, `/runs` shows the new run
4. Once verified, proceed to build background execution (separate change)

**Rollback:** Revert `monitor-api` to previous version — manual runs return to async behavior (no execution).

## Open Questions

1. **Should we remove the `CreateManualRun` DynamoDB write?** Currently it writes `RUN_REQUEST#` items even though nothing processes them. Option: keep for audit trail (manual run was requested), or remove since no worker processes them. Decision: Keep — useful for audit that manual run was triggered.

2. **Should we validate the target URL exists or is reachable before queuing?** Not for manual trigger — execute and let the result show success/failure.