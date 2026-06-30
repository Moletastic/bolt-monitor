## 1. Domain Model Changes

- [x] 1.1 Add FailureThreshold and RecoveryThreshold fields to Monitor struct in shared/monitorconfig/model.go
- [x] 1.2 Add ConsecutiveFailures, ConsecutiveSuccesses, and CurrentStatus fields to MonitorStatus in shared/resultstatus/model.go
- [x] 1.3 Add new MonitorState enum type with values UP, DEGRADED, DOWN, RECOVERING, MAINTENANCE
- [x] 1.4 Update NewMonitorStatus() to accept threshold config and initialize state

## 2. check-runtime State Machine

- [x] 2.1 Refactor incidentRecordsForResult to accept threshold config and return updated MonitorStatus
- [x] 2.2 Implement state transition logic for UP + failure → DEGRADED
- [x] 2.3 Implement DEGRADED state accumulation and DEGRADED → DOWN transition when threshold met
- [x] 2.4 Implement DEGRADED + success → UP transition
- [x] 2.5 Implement DOWN + success → RECOVERING transition
- [x] 2.6 Implement RECOVERING state accumulation and RECOVERING → UP transition when threshold met
- [x] 2.7 Implement RECOVERING + failure → DOWN transition
- [x] 2.8 Verify manual runs (Trigger=Manual) do not affect ConsecutiveFailures/ConsecutiveSuccesses counters
- [x] 2.9 Add unit tests for state machine transitions covering all 5 states and all transitions

## 3. Monitor API Updates

- [x] 3.1 Add FailureThreshold and RecoveryThreshold to monitor create/update request/response types
- [x] 3.2 Add validation that FailureThreshold >= 1 and RecoveryThreshold >= 1
- [x] 3.3 Add POST /api/v1/services/{serviceId}/monitors/{monitorId}/maintenance/enable endpoint
- [x] 3.4 Add POST /api/v1/services/{serviceId}/monitors/{monitorId}/maintenance/disable endpoint
- [x] 3.5 Wire maintenance mode state into check-runtime scheduler (skip MAINTENANCE monitors)

## 4. Dashboard Updates

- [x] 4.1 Update monitor status display to show DEGRADED state with appropriate visual indicator
- [x] 4.2 Update monitor status display to show RECOVERING state with appropriate visual indicator
- [x] 4.3 Update monitor status display to show MAINTENANCE state with appropriate visual indicator
- [x] 4.4 Add accumulation progress indicator for DEGRADED (e.g., "2/3 failures")
- [x] 4.5 Add accumulation progress indicator for RECOVERING (e.g., "1/2 successes")
- [x] 4.6 Add maintenance mode toggle button to monitor detail view

## 5. Integration and Validation

- [x] 5.6 Run make lint-go and make test-go-all
- [ ] 5.1 Verify existing monitors with default threshold=1 preserve current binary behavior
- [ ] 5.2 Test DEGRADED → DOWN transition with failure threshold > 1
- [ ] 5.3 Test RECOVERING → UP transition with recovery threshold > 1
- [ ] 5.4 Test MAINTENANCE enable/disable flow
- [ ] 5.5 Test threshold change mid-flight triggers immediate re-evaluation
