## 1. Scheduler: Skip Draft Services (services/check-runtime/runtime.go)

- [x] 1.1 Add `GetService(ctx, tenantID, serviceID)` method to `runtimeRepository` interface
- [x] 1.2 Implement `GetService` in `dynamoRuntimeRepository`
- [x] 1.3 Modify `runScheduler` to check service lifecycleState before enqueueing
- [x] 1.4 Skip monitors where service lifecycleState is "draft"

## 2. Repository: Service Lifecycle Support (services/check-runtime/repository.go)

- [x] 2.1 Add `GetService` method to interface (if not already present from monitor-api)
- [x] 2.2 Implement `GetService` to read service record from DynamoDB

## 3. Monitor-API: Auto-Transition Service Lifecycle (services/monitor-api)

- [x] 3.1 Modify `enableMonitor` handler to check service lifecycleState
- [x] 3.2 If service is "draft", transition to "active" before enabling monitor
- [x] 3.3 Use transaction to ensure atomicity (both service transition and monitor enable)
- [x] 3.4 Verify Active → Archived remains manual only (no auto-transition)

## 4. Dashboard: Fix Status Display (apps/dashboard)

- [x] 4.1 Investigate monitor status display component
- [x] 4.2 Verify `currentStatus` field is being read correctly from API response
- [x] 4.3 Fix any incorrect field path or conditional logic
- [x] 4.4 Verify status shows "up", "down", or "unknown" based on actual outcome

## 5. Dashboard: Fix Duration Display (apps/dashboard)

- [x] 5.1 Investigate duration display component
- [x] 5.2 Verify `lastDurationMs` field is being read correctly from API response
- [x] 5.3 Fix display to show duration in milliseconds (e.g., "45ms")
- [x] 5.4 Handle null/undefined case (show "-" or "N/A" only when truly unavailable)

## 6. Dashboard: Remove Manual Lifecycle Dropdown (apps/dashboard)

- [x] 6.1 Remove manual lifecycle state dropdown from service edit form
- [x] 6.2 Verify lifecycle transitions happen automatically per spec
- [x] 6.3 Add UI indicator showing current lifecycle state (read-only)

## 7. Verify and Test

- [x] 7.1 Run `make lint-go` to check for linting issues
- [x] 7.2 Run `make test-go-all` to run all Go tests
- [x] 7.3 Build the Lambda: `make build-go`
- [x] 7.4 Deploy to staging: `make deploy-infra`
- [x] 7.5 Test scheduler skips draft services:
  - Note: Cannot have draft service with enabled monitor (auto-transition happens on enable)
  - Verified via code review: scheduler correctly skips draft services
- [x] 7.6 Test status display:
  - Check monitor status shows correct value (up/down/unknown)
  - Check duration shows actual value
- [x] 7.7 Test lifecycle auto-transition:
  - Create draft service with disabled monitor
  - Enable monitor
  - Verify service transitions to active

## Notes

- Task 7.8 (archive UI testing) is deferred to `service-archive-ui` change since no archive UI exists yet