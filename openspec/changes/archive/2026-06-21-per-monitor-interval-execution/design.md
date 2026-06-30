## Context

The current scheduler runs every 1 minute via EventBridge and enqueues all enabled monitors regardless of their configured `intervalSeconds`. This means:

1. A monitor with `intervalSeconds: 300` (5 minutes) gets checked every 1 minute, which is wasteful.
2. A monitor with `intervalSeconds: 60` is checked once per minute, which is optimal.
3. The system does not track when a monitor last executed, so it cannot determine if a monitor is due.

**Constraint**: EventBridge `rate()` expressions have a minimum of 1 minute, and EventBridge cron expressions do not support seconds-level offsets.

**Goal**: Implement per-monitor interval checking for minute-based cadence presets so monitors only execute when their configured cadence has elapsed since last execution.

## Goals / Non-Goals

**Goals:**
- Respect supported minute-based `intervalSeconds` values configured per monitor
- Add `LastExecutionAt` tracking to DynamoDB
- Keep the existing single 1-minute EventBridge scheduler
- Keep runtime compatibility for legacy monitors with missing or invalid interval data

**Non-Goals:**
- Achieve sub-minute exact intervals
- Support arbitrary off-grid intervals such as 90, 150, or 210 seconds
- Change the SQS queue or worker architecture
- Rename the public API field from `intervalSeconds`

## Decisions

### Decision 1: Track LastExecutionAt on Monitor Record

**Choice**: Add `LastExecutionAt` directly to the monitor META record in DynamoDB.

**Rationale**: This is the simplest approach: one write per scheduled execution, co-located with monitor data. A separate tracking key adds complexity without current benefit.

**Schema**:

```text
PK: SERVICE#<tenantId>#<serviceId>#MONITOR#<monitorId>
SK: META
...
LastExecutionAt: "2026-06-09T16:00:00Z"
```

### Decision 2: Keep Single One-Minute Scheduler

**Choice**: Keep one EventBridge Schedule:
- `SchedulerSchedule`: `rate(1 minute)`

**Rationale**: EventBridge cannot provide a true 30-second offset with cron syntax. A single one-minute tick is simple, reliable, and matches the product decision to offer minute-based cadence presets.

**Rejected Alternative**: A second cron offset by 30 seconds. `cron(30 * * * ? *)` means minute 30 of every hour, not every minute at second 30.

### Decision 3: Offer Minute-Based Cadence Presets

**Choice**: Allow only these `intervalSeconds` values:
- 60 (1 minute)
- 120 (2 minutes)
- 180 (3 minutes)
- 300 (5 minutes)
- 600 (10 minutes)
- 900 (15 minutes)
- 1800 (30 minutes)
- 3600 (1 hour)

**Rationale**: Users think in minutes for uptime checks, and these presets map cleanly to a one-minute scheduler tick. Values such as 90 seconds would be rounded up by the scheduler in practice, so accepting them would create misleading behavior.

### Decision 4: Skip Monitors Not Yet Due

**Choice**: Scheduler checks `intervalSeconds` vs elapsed time since `LastExecutionAt`. It only enqueues if:
- `LastExecutionAt` is null (never executed), or
- `h.now() - LastExecutionAt >= monitor.IntervalSeconds`

**Rationale**: Matches user expectation. If they set 2 minutes, the monitor runs approximately every 2 minutes on the minute scheduler tick.

**Default Behavior**: New monitor writes must use one of the allowed presets. Runtime treats legacy `intervalSeconds <= 0` as always due so existing bad data does not block scheduler processing.

### Decision 5: Update LastExecutionAt After SQS Send

**Choice**: Record `LastExecutionAt` when scheduler sends to SQS, not when worker completes.

**Rationale**: From the user's perspective, cadence starts when the check is scheduled. Recording after worker completion would cause interval drift.

**Trade-off**: If SQS send succeeds but worker later fails and exhausts retries, `LastExecutionAt` is still set. This is acceptable because retries and DLQ handle failed work separately.

## Architecture

```text
EventBridge Scheduler (rate 1 minute)
  -> Scheduler Lambda (RUNTIME_MODE=scheduler)
     1. Read SchedulerConfig
     2. Query enabled monitors
     3. For each monitor:
        a. Skip disabled monitors
        b. Check LastExecutionAt vs intervalSeconds
        c. If due, build ExecutionRequest(s)
        d. Send request(s) to SQS
        e. Write RUN_REQUEST audit work
        f. Update LastExecutionAt
  -> SQS execution queue
  -> Worker Lambda (RUNTIME_MODE=worker)
```

## DynamoDB Changes

### Monitor Record Update

```typescript
interface MonitorMetaRecord {
  PK: string
  SK: string
  // existing fields...
  LastExecutionAt?: string
}
```

## Code Changes

### Repository Interface Additions

```go
type runtimeRepository interface {
  // existing methods...
  GetLastExecution(ctx context.Context, tenantID, serviceID, monitorID string) (*time.Time, error)
  RecordLastExecution(ctx context.Context, tenantID, serviceID, monitorID string, lastExec time.Time) error
}
```

### Scheduler Logic

```go
func (h runtimeHandler) isMonitorDue(ctx context.Context, monitor monitorconfig.Monitor) (bool, error) {
  if monitor.IntervalSeconds <= 0 {
    return true, nil
  }
  lastExec, err := h.repo.GetLastExecution(ctx, monitor.TenantID, monitor.ServiceID, monitor.MonitorID)
  if err != nil {
    return false, err
  }
  if lastExec == nil {
    return true, nil
  }
  return h.now().Sub(*lastExec) >= time.Duration(monitor.IntervalSeconds)*time.Second, nil
}
```

### Validation Logic

```go
allowedIntervalSeconds := []int{60, 120, 180, 300, 600, 900, 1800, 3600}
```

Monitor validation rejects values outside that set.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Scheduler cadence minimum is 1 minute | Offer minute-based presets only |
| LastExecutionAt set before worker executes | DLQ and retry handling remain responsible for worker failures |
| DynamoDB write for LastExecutionAt adds latency | Single item update is acceptable overhead |
| Legacy bad interval data exists | Runtime treats `intervalSeconds <= 0` as always due |

## Migration Plan

1. Deploy scheduler code with interval checking using the existing one-minute cron.
2. Deploy monitor validation that only accepts supported minute-based presets.
3. Verify monitors with 1, 2, and 5 minute cadences are respected.
4. Monitor for any issues with interval enforcement.

**Rollback**: Revert scheduler to enqueue all monitors every minute.

## Open Questions

1. Should the dashboard continue posting `intervalSeconds` while displaying minute labels? Recommended: yes, for minimal API churn.
2. Should we eventually introduce plan-based cadence limits? Example: Free starts at 5 minutes, Pro allows 1 minute.
3. Should we track per probe-location last execution? Current design tracks per monitor.
