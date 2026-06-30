## Context

The current check-runtime incident logic in `services/check-runtime/repository.go:414` (`incidentRecordsForResult`) uses a pure binary model:

- Any failure outcome + no open incident → immediately open incident
- Any success outcome + open incident → immediately resolve incident

There is no concept of warning states, accumulation periods, or graduated recovery. This causes:
- **Flapping**: Transient failures (network blips) immediately create then resolve incidents
- **False positives**: Single anomalous check results open incidents for self-healing failures
- **No operator anticipation**: Operators only see the incident after threshold is already crossed
- **No maintenance windows**: No ability to silence monitoring during planned downtime

The system already supports multi-location probing and a 5-state domain model in the dashboard. The missing piece is the threshold accumulation logic and state machine implementation.

## Goals / Non-Goals

**Goals:**
- Implement a 5-state monitor model: UP, DEGRADED, DOWN, RECOVERING, MAINTENANCE
- Add configurable failure threshold (consecutive failures before incident opens)
- Add configurable recovery threshold (consecutive successes before incident resolves)
- Track consecutive outcome counters persistently in MonitorStatus
- Exclude manual runs from threshold accumulation
- Support monitor maintenance windows

**Non-Goals:**
- Probe-location-aware threshold aggregation (not in this change — will be separate)
- Automatic maintenance window scheduling (manual only in v1)
- Threshold breach notifications (notify only when incident opens/resolve, not during accumulation)
- Historical state reconstruction (counters reset on service restart — acceptable tradeoff)

## Decisions

### Decision 1: State Storage — MonitorStatus Record

**Choice**: Store ConsecutiveFailures, ConsecutiveSuccesses, and CurrentStatus directly on the existing `MonitorStatus` DynamoDB record.

**Rationale**: MonitorStatus is already updated on every check run. Adding counters there avoids a new entity type and extra DynamoDB operations. The record already contains LastOutcome, LastCheckedAt, and current state.

**Alternatives considered**:
- Separate `CONSECUTIVE#` counter record: More resilient to service restarts but adds DynamoDB writes and an extra entity type
- Query last N CheckRun records to derive counters: No extra storage but adds read latency on every check execution

**Trade-off**: If check-runtime pod restarts mid-flight, counters reset to 0. Acceptable for v1.

### Decision 2: Threshold Configuration — On Monitor Model

**Choice**: Add `failureThreshold int` and `recoveryThreshold int` fields directly to the `Monitor` model in `shared/monitorconfig/model.go`.

**Rationale**: Monitors are already the unit of configuration. Adding threshold fields there keeps configuration co-located and consistent with existing patterns.

**Alternatives considered**:
- Separate `ThresholdPolicy` referenced by name: More flexible (reuse across monitors) but adds indirection
- Global defaults in scheduler config: Simpler but less flexible per-monitor

**Constraint**: Both thresholds must be ≥ 1. Default to 1 (preserves current binary behavior).

### Decision 3: State Machine Location — `incidentRecordsForResult`

**Choice**: Implement state transition logic in `services/check-runtime/repository.go` within `incidentRecordsForResult`.

**Rationale**: This function is already the decision point for incident lifecycle. Extending it to implement threshold accumulation and state transitions keeps all incident logic co-located.

**Contract change**: `incidentRecordsForResult` currently returns `(records []any, notificationEvent string, incidentID string, error)`. It will need to also return the updated `MonitorStatus` with new state fields.

### Decision 4: MAINTENANCE State Entry/xit

**Choice**: MAINTENANCE is entered via explicit API action (`POST .../maintenance/enable`). It is exited via `POST .../maintenance/disable` which re-evaluates current check state.

**Rationale**: Simple manual trigger for v1. When exiting MAINTENANCE, the monitor re-evaluates its current outcome state — if still failing, it may immediately go to DEGRADED or DOWN.

**Open**: Should exiting MAINTENANCE reset counters? Yes — fresh start on resume.

### Decision 5: Counter Reset Rules

**Choice**:
- On opposite outcome: Counter resets (failures reset to 0 on success, successes reset to 0 on failure)
- On config change: If threshold changes while in DEGRADED/RECOVERING, immediately re-evaluate against new threshold
- On manual run: No counter change (manual runs are excluded)
- On MAINTENANCE entry: Counters reset to 0
- On MAINTENANCE exit: Counters reset to 0, fresh evaluation

### Decision 6: State Machine Formalization

```
State transitions (per check result):

UP + failure → DEGRADED (ConsecutiveFailures = 1)
UP + success → UP (ConsecutiveSuccesses += 1, reset)

DEGRADED + failure → DEGRADED (ConsecutiveFailures += 1)
  └─ if ConsecutiveFailures >= FailureThreshold → DOWN (open incident)
DEGRADED + success → UP (reset ConsecutiveFailures)

DOWN + failure → DOWN (ConsecutiveFailures += 1, incident already open)
DOWN + success → RECOVERING (ConsecutiveSuccesses = 1)

RECOVERING + failure → DOWN (ConsecutiveSuccesses reset, incident still open)
RECOVERING + success → RECOVERING (ConsecutiveSuccesses += 1)
  └─ if ConsecutiveSuccesses >= RecoveryThreshold → UP (resolve incident)

Any state + MAINTENANCE enable → MAINTENANCE (reset counters)
MAINTENANCE + MAINTENANCE disable → UP (re-evaluate from current check result)
```

## Risks / Trade-offs

[Risk] Counter loss on service restart
→ Mitigation: Acceptable for v1. Counters persist in DynamoDB but reset if check-runtime pod restarts mid-flight. Operators see DEGRADED→DOWN transition after first failure post-restart.

[Risk] DynamoDB record size growth
→ Mitigation: MonitorStatus record grows by ~30 bytes per monitor. For most deployments this is negligible. Monitor with ~50 probes × 8 bytes each = ~400 bytes growth, still well under DynamoDB 400KB limit.

[Risk] Threshold change re-evaluation requires check-runtime to observe config changes
→ Mitigation: check-runtime reads Monitor config from DynamoDB on every execution cycle. New thresholds are picked up on next scheduled run. Immediate re-evaluation is best-effort.

[Risk] No probe-location-aware aggregation
→ Mitigation: This design treats all probe location results identically. A separate change can introduce location-aware threshold logic if needed.

## Migration Plan

1. **Phase 1 — Schema and model**: Add threshold fields to Monitor model, add state/counter fields to MonitorStatus. No behavior change yet.
2. **Phase 2 — check-runtime state machine**: Implement `incidentRecordsForResult` state machine. FailureThreshold/RecoveryThreshold default to 1, preserving current binary behavior.
3. **Phase 3 — API exposure**: Expose threshold fields in monitor CRUD API. Add maintenance mode endpoints.
4. **Phase 4 — Dashboard**: Update monitor status display to show new states (DEGRADED, RECOVERING, MAINTENANCE). Show accumulation progress.
5. **Phase 5 — Validation**: Verify no flapping on existing monitors when thresholds are set to 1 (default).

**Rollback**: If issues arise, set FailureThreshold=1 and RecoveryThreshold=1 on all monitors to restore binary behavior.

## Open Questions

1. Should the API expose the raw counter values (ConsecutiveFailures, ConsecutiveSuccesses) for dashboard progress indicators?
2. Does the dashboard need a visual "accumulation progress" indicator (e.g., "2/3 failures") in DEGRADED/RECOVERING states?
3. What is the initial state for a newly created monitor? UP? Or does it need to pass its first check first?
4. Should there be a maximum threshold value (e.g., cap at 10) to prevent unbounded accumulation?
