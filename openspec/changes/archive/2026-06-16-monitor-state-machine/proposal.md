## Why

The current incident management system uses a binary success/failure model: a single failed check immediately opens an incident, and a single successful check immediately resolves it. This causes flapping (rapid open/close cycles), false positives from transient network blips, and no operator visibility into degradation before an incident is declared. There is also no concept of planned maintenance windows.

## What Changes

- **New 5-state monitor model**: UP, DEGRADED, DOWN, RECOVERING, MAINTENANCE
- **Configurable failure threshold**: Incident opens only after N consecutive failed checks (DEGRADED → DOWN transition)
- **Configurable recovery threshold**: Incident resolves only after M consecutive successful checks (RECOVERING → UP transition)
- **DEGRADED state**: Accumulating failures before threshold is met. No incident open. Monitor status shows "DEGRADED".
- **RECOVERING state**: Incident is open but accumulating consecutive successes before recovery threshold is met.
- **MAINTENANCE state**: Monitor is paused. Open incidents persist but notifications are suppressed.
- **Counter persistence**: Track ConsecutiveFailures and ConsecutiveSuccesses in MonitorStatus record.
- **Manual runs excluded from threshold counting**: Ad-hoc manual checks do not affect threshold counters.
- **Threshold change re-evaluation**: If threshold config changes while in DEGRADED/RECOVERING, immediately re-evaluate against new threshold.

## Capabilities

### New Capabilities

- `monitor-state-machine`: The 5-state model (UP/DEGRADED/DOWN/RECOVERING/MAINTENANCE) with formal state transitions, counter reset rules, and transition guard conditions.
- `failure-threshold`: Configurable consecutive failure threshold per monitor. Default: 1 (current behavior). Must be ≥ 1.
- `recovery-threshold`: Configurable consecutive success threshold per monitor. Default: 1 (current behavior). Must be ≥ 1.
- `maintenance-mode`: Ability to place a monitor in MAINTENANCE state, silencing notifications but preserving open incidents.

### Modified Capabilities

- `check-result-status-model`: MonitorStatus record gains ConsecutiveFailures, ConsecutiveSuccesses, and CurrentStatus fields. CurrentStatus transitions from UP/DOWN to include DEGRADED, RECOVERING, MAINTENANCE.
- `monitor-configuration`: Monitor model gains FailureThreshold and RecoveryThreshold integer fields.
- `incident-management-api`: Incident lifecycle changes — incident opens on DOWN transition (not on first failure), resolves on UP transition (not on first success). Incident stays open through RECOVERING state.
- `check-execution-pipeline`: check-runtime logic changes in `incidentRecordsForResult` to implement threshold accumulation and state transitions.

## Impact

- **Domain models**: `shared/resultstatus/model.go` (MonitorStatus), `shared/monitorconfig/model.go` (Monitor)
- **Check runtime**: `services/check-runtime/repository.go` — `incidentRecordsForResult` state machine logic
- **Monitor API**: `services/monitor-api/handler.go` — new endpoints or fields for threshold config and maintenance mode
- **Dashboard**: `apps/dashboard/` — display new states (DEGRADED, RECOVERING, MAINTENANCE), show accumulation progress
- **DynamoDB**: MonitorStatus record grows new fields; may need schema migration strategy
