## 1. Infrastructure (infra/stacks/bootstrap.ts)

- [x] 1.1 Keep a single EventBridge Schedule at `rate(1 minute)`
- [x] 1.2 Do not add a second 30-second-offset EventBridge schedule

## 2. Code: Repository Updates (services/check-runtime/repository.go)

- [x] 2.1 Add `GetLastExecution(ctx, tenantID, serviceID, monitorID)` method to `runtimeRepository` interface
- [x] 2.2 Add `RecordLastExecution(ctx, tenantID, serviceID, monitorID, lastExec)` method to interface
- [x] 2.3 Implement `GetLastExecution` in `dynamoRuntimeRepository` (read from monitor META record)
- [x] 2.4 Implement `RecordLastExecution` in `dynamoRuntimeRepository` (update LastExecutionAt field)

## 3. Code: Scheduler Interval Checking (services/check-runtime/runtime.go)

- [x] 3.1 Add `isMonitorDue(ctx, monitor)` method to check intervalSeconds vs elapsed time
- [x] 3.2 Modify `runScheduler` to call `isMonitorDue` before enqueueing each monitor
- [x] 3.3 After successful SQS send, call `RecordLastExecution` to update timestamp
- [x] 3.4 Handle null LastExecutionAt (first execution) as "always due"
- [x] 3.5 Handle legacy intervalSeconds <= 0 as "always due" at runtime

## 4. Code: Minute-Based Cadence Validation (shared/monitorconfig/model.go)

- [x] 4.1 Add allowed cadence values: 60, 120, 180, 300, 600, 900, 1800, 3600 seconds
- [x] 4.2 Reject unsupported intervalSeconds values in monitor validation
- [x] 4.3 Add validation tests for allowed and unsupported cadences

## 5. Dashboard: Minute-Based Cadence UI (apps/dashboard)

- [x] 5.1 Replace raw intervalSeconds number input with minute-based cadence select
- [x] 5.2 Display monitor cadence as minute/hour labels instead of raw seconds

## 6. Code: Main Entry Point (services/check-runtime/main.go)

- [x] 6.1 No changes needed (same RUNTIME_MODE=scheduler handler works for the one-minute cron)

## 7. Verify and Test

- [x] 7.1 Run `make lint-go` to check for linting issues
- [x] 7.2 Run `make test-go-all` to run all Go tests
- [x] 7.3 Build the Lambda: `make build-go`
- [x] 7.4 Run `make lint-dashboard` to check dashboard linting
- [x] 7.5 Run `make check-dashboard` to check dashboard types
- [x] 7.6 Deploy to staging: `make deploy-infra`
- [x] 7.7 Test scheduler with intervalSeconds=60:
  - Create monitor with intervalSeconds=60
  - Verify first execution happens immediately
  - Verify second execution waits ~60 seconds
- [x] 7.8 Test scheduler with intervalSeconds=120:
  - Create monitor with intervalSeconds=120
  - Verify second execution waits ~2 minutes
- [x] 7.9 Test scheduler rejects unsupported intervalSeconds=90
- [x] 7.10 Verify only one EventBridge scheduler is configured

## 8. Documentation

- [x] 8.1 Update design.md notes about EventBridge rate() minimum and cron second limitations
- [x] 8.2 Add note about minute-based cadence options in user-facing docs (if applicable)
