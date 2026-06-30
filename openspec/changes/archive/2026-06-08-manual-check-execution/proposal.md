## Why

Currently, `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run` accepts a manual run request and returns immediately without executing the check. This makes it impossible to verify that the monitoring system works before building automated background workloads. We need manual triggers to execute synchronously and return real results, so operators can validate their monitor configuration end-to-end.

## What Changes

- **Modified**: `POST /run` endpoint to execute HTTP check synchronously within the API Lambda
- **Modified**: Response now includes execution result (outcome, duration, status code, error)
- **Unchanged**: The endpoint path, authentication, and error handling for missing/disabled monitors

## Capabilities

### New Capabilities
- `manual-check-sync-execution`: Execute monitor checks synchronously on demand and return real-time results. This allows operators to verify monitor configuration works before enabling automated background execution.

### Modified Capabilities
- `manual-run-api`: Change requirement from async accept-and-return-id to synchronous execute-and-return-result. The manual run endpoint shall wait for HTTP check completion and return the execution outcome inline in the response.

## Impact

- **Code**: `services/monitor-api/handler.go` (`runMonitor` function) — execute inline instead of enqueueing
- **Code**: `services/monitor-api/repository.go` — add `RecordExecutionResult` method
- **Code**: `services/monitor-api/main.go` — pass probe location catalog to handler
- **API Response**: `POST /run` response shape changes from minimal `{runId, status}` to full result `{runId, outcome, durationMs, statusCode, error, startedAt, finishedAt, probeLocationId}`