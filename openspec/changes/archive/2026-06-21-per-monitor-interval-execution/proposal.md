## Why

The current scheduler ignores `intervalSeconds` on monitors - every enabled monitor gets enqueued on every scheduler cycle (every 1 minute). This wastes resources on monitors with long intervals and doesn't respect user-configured cadences. We need per-monitor interval checking so monitors only execute when their configured minute-based cadence has elapsed.

## What Changes

- **Modified**: Scheduler to check `intervalSeconds` vs time since last execution before enqueueing
- **New**: DynamoDB tracking of `LastExecutionAt` per monitor
- **New**: Repository methods for `GetLastExecution` and `RecordLastExecution`
- **Modified**: Monitor validation to allow only minute-based cadence presets

## Capabilities

### New Capabilities
- `scheduler-interval-enforcement`: Scheduler respects per-monitor `intervalSeconds` and only enqueues when sufficient time has elapsed since last execution
- `monitor-minute-cadence-validation`: Monitor cadence is selected from supported minute-based presets

### Modified Capabilities
- `check-runtime-scheduler-mode`: Add requirement that scheduler SHALL respect `intervalSeconds` and skip monitors not yet due
- `scheduler-eventbridge-trigger`: Keep a single one-minute EventBridge trigger; interval enforcement happens in scheduler code

## Impact

- **Code**: `services/check-runtime/runtime.go` - add interval checking logic
- **Code**: `services/check-runtime/repository.go` - add last execution tracking
- **Code**: `shared/monitorconfig/model.go` - constrain intervalSeconds to minute-based presets
- **DynamoDB**: Add `LastExecutionAt` field to monitor records
