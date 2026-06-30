## 1. Update Types (services/monitor-api/types.go)

- [x] 1.1 Add execution result fields to `manualRunResponse` struct (`Outcome`, `DurationMs`, `StatusCode`, `Error`, `ProbeLocationID`, `StartedAt`, `FinishedAt`)
- [x] 1.2 Update `toManualRunResponse()` to include new fields or create `toManualRunResponseWithResult()` helper

## 2. Update Repository Interface (services/monitor-api/repository.go)

- [x] 2.1 Add `RecordExecutionResult(ctx context.Context, monitor monitorconfig.Monitor, runID string, result checkexecution.ExecutionResult) error` to `monitorRepository` interface
- [x] 2.2 Implement `RecordExecutionResult()` in `dynamoMonitorRepository`:
  - Write `RUN#` item (CheckRun record) using `resultstatus.NewCheckRun()`
  - Write `STATUS` item (MonitorStatus) using `resultstatus.NewMonitorStatus()`
  - Check for open incidents and handle creation/resolution
  - Update service status rollup
  - Write all in single transaction

## 3. Update Handler (services/monitor-api/handler.go)

- [x] 3.1 Import `bolt-monitor/shared/checkexecution` package
- [x] 3.2 Modify `runMonitor()` to:
  - Generate runID using `newRunID(now)`
  - Build `checkexecution.ExecutionRequest` from monitor + probe location
  - Create HTTP client with appropriate timeout
  - Call `checkexecution.ExecuteHTTP(ctx, client, request)` inline
  - Call `h.repo.RecordExecutionResult(ctx, monitor, runID, result)` to persist
  - Return `manualRunResponse` with execution result fields
- [x] 3.3 Update response to return HTTP 200 with full result (not HTTP 202 Accepted)

## 4. Verify and Test

- [x] 4.1 Run `make lint-go` to check for linting issues
- [x] 4.2 Run `make test-go-all` to run all Go tests
- [x] 4.3 Build the Lambda: `make build-go`
- [x] 4.4 Manually test the endpoint:
  - Create a service
  - Create a monitor pointing to a known-good URL (e.g., https://example.com)
  - Call `POST /api/v1/services/{id}/monitors/{id}/run`
  - Verify response includes `outcome: "success"`, `statusCode: 200`, `durationMs`, etc.
  - Call `GET /api/v1/services/{id}/monitors/{id}/status` and verify lastCheckedAt updated
  - Call `GET /api/v1/services/{id}/monitors/{id}/runs` and verify run history
- [x] 4.5 Test failure case:
  - Create monitor pointing to invalid URL
  - Call manual run
  - Verify `outcome: "failure"` or `outcome: "error"` with error message
